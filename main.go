package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	_ "net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/ipfs-force-community/sophon-messager/publisher"
	"github.com/ipfs-force-community/sophon-messager/publisher/pubsub"

	"github.com/filecoin-project/venus/fixtures/networks"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	gatewayAPI "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/ipfs-force-community/sophon-auth/jwtclient"
	"github.com/ipfs-force-community/sophon-messager/metrics"
	"github.com/ipfs-force-community/sophon-messager/utils"
	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	"github.com/urfave/cli/v2"
	"go.uber.org/fx"

	"github.com/ipfs-force-community/sophon-messager/api"
	ccli "github.com/ipfs-force-community/sophon-messager/cli"
	"github.com/ipfs-force-community/sophon-messager/config"
	"github.com/ipfs-force-community/sophon-messager/filestore"
	"github.com/ipfs-force-community/sophon-messager/gateway"
	"github.com/ipfs-force-community/sophon-messager/models"
	"github.com/ipfs-force-community/sophon-messager/service"
	"github.com/ipfs-force-community/sophon-messager/version"
)

var log = logging.Logger("main")

func main() {
	app := &cli.App{
		Name:  "sophon message",
		Usage: "used for manage message",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "repo",
				Value:   ccli.DefRepoPath,
				EnvVars: []string{"SOPHON_MESSAGER_PATH"},
			},
		},
		Commands: []*cli.Command{
			ccli.MsgCmds,
			ccli.AddrCmds,
			ccli.SharedParamsCmds,
			ccli.ActorCfgCmds,
			ccli.NodeCmds,
			ccli.LogCmds,
			ccli.SendCmd,
			ccli.SwarmCmds,
			runCmd,
		},
	}

	app.Version = version.Version
	app.Setup()
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		return
	}
}

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "run messager",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "listen",
			Usage: "specify endpoint for listen",
			Value: "/ip4/127.0.0.1/tcp/39812",
		},
		&cli.StringFlag{
			Name:  "auth-url",
			Usage: "url for auth server",
		},

		// node
		&cli.BoolFlag{
			Name:  "disable-push",
			Usage: "disable push messager function, Warn only one instance can used to push message",
		},
		&cli.StringFlag{
			Name:  "node-url",
			Usage: "url for connection lotus/venus",
		},
		&cli.StringFlag{
			Name:  "node-token",
			Usage: "token auth for lotus/venus",
		},

		// database
		&cli.StringFlag{
			Name:  "db-type",
			Usage: "which db to use. sqlite/mysql",
		},
		&cli.StringFlag{
			Name:  "mysql-dsn",
			Usage: "mysql connection string",
		},
		&cli.StringSliceFlag{
			Name:  "gateway-url",
			Usage: "gateway url",
		},
		&cli.StringFlag{
			Name:  "gateway-token",
			Usage: "gateway token",
		},
		&cli.StringFlag{
			Name:  "auth-token",
			Usage: "auth token",
		},
		&cli.StringFlag{
			Name: "rate-limit-redis",
		},
	},
	Action: runAction,
}

func runAction(cctx *cli.Context) error {
	var fsRepo filestore.FSRepo
	cfg := config.DefaultConfig()

	ctx, cancel := context.WithCancel(cctx.Context)
	defer cancel()

	// Set the log level. The default log level is info
	utils.SetupLogLevels()

	repoPath, err := ccli.GetRepoPath(cctx)
	if err != nil {
		return err
	}
	hasFSRepo, err := hasFSRepo(repoPath)
	if err != nil {
		return err
	}
	if hasFSRepo {
		fsRepo, err = filestore.NewFSRepo(repoPath)
		if err != nil {
			return err
		}
		cfg = fsRepo.Config()
	}

	if err = updateFlag(cfg, cctx); err != nil {
		return err
	}

	if !hasFSRepo {
		fsRepo, err = filestore.InitFSRepo(repoPath, cfg)
		if err != nil {
			return err
		}
	}

	log.Infof("node info url: %s, token: %s\n", cfg.Node.Url, cfg.Node.Token)
	log.Infof("auth info url: %s\n", cfg.JWT.AuthURL)
	log.Infof("gateway info url: %s, token: %s\n", cfg.Gateway.Url, cfg.Gateway.Token)
	log.Infof("rate limit info: redis: %s \n", cfg.RateLimit.Redis)
	log.Infof("default timeout: %v, sign message timeout: %v, estimate message timeout: %v", cfg.MessageService.DefaultTimeout,
		cfg.MessageService.SignMessageTimeout, cfg.MessageService.EstimateMessageTimeout)

	remoteAuthCli, err := jwtclient.NewAuthClient(cfg.JWT.AuthURL, cfg.JWT.Token)
	if err != nil {
		return err
	}

	localAuthCli, token, err := jwtclient.NewLocalAuthClient()
	if err != nil {
		return fmt.Errorf("failed to generate local auth client %v", err)
	}

	client, closer, err := v1.DialFullNodeRPC(ctx, cfg.Node.Url, cfg.Node.Token, nil)
	if err != nil {
		return fmt.Errorf("connect to node failed %v", err)
	}
	defer closer()

	// TODO: delete this when relative issue is fixed in lotus https://github.com/filecoin-project/venus/issues/5247
	log.Info("wait for height of chain bigger than zero ...")
	ticker := time.NewTicker(10 * time.Second)
	for {
		head, err := client.ChainHead(ctx)
		if err != nil {
			return err
		}
		if head.Height() > 0 {
			break
		}
		select {
		case <-ctx.Done():
			fmt.Println("\nExit by user")
			return nil
		case <-ticker.C:
		}
	}
	ticker.Stop()

	networkParams, err := client.StateGetNetworkParams(ctx)
	if err != nil {
		return fmt.Errorf("get network params failed %v", err)
	}

	// The 2k network block delay is 4s, which will be less than WaitingChainHeadStableDuration (8s)
	// and will not push messages
	if networkParams.BlockDelaySecs <= uint64(cfg.MessageService.WaitingChainHeadStableDuration/time.Second) {
		cfg.MessageService.WaitingChainHeadStableDuration = time.Duration(networkParams.BlockDelaySecs/2) * time.Second
		if err := fsRepo.ReplaceConfig(cfg); err != nil {
			return err
		}
	}

	if err := ccli.LoadBuiltinActors(ctx, client); err != nil {
		return err
	}
	if err := setEip155ChainID(string(networkParams.NetworkName)); err != nil {
		return fmt.Errorf("failed to set Eip155ChainID: %v", err)
	}

	mAddr, err := ma.NewMultiaddr(cfg.API.Address)
	if err != nil {
		return err
	}

	walletCli, walletCliCloser, err := gateway.NewWalletClient(ctx, &cfg.Gateway)
	if err != nil {
		return err
	}
	defer walletCliCloser()

	// Listen on the configured address in order to bind the port number in case it has
	// been configured as zero (i.e. OS-provided)
	apiListener, err := manet.Listen(mAddr)
	if err != nil {
		return err
	}
	lst := manet.NetListener(apiListener)

	err = fsRepo.SaveToken(token)
	if err != nil {
		return fmt.Errorf("failed to save token %v", err)
	}

	provider := fx.Options(
		fx.Logger(fxLogger{}),
		// prover
		fx.Supply(cfg, &cfg.DB, &cfg.API, &cfg.JWT, &cfg.Node, &cfg.MessageService, cfg.Libp2pNet,
			&cfg.Gateway, &cfg.RateLimit, cfg.Trace, cfg.Metrics, cfg.Publisher),
		fx.Supply(networkParams.NetworkName),
		fx.Supply(networkParams),
		fx.Supply(remoteAuthCli),
		fx.Supply(localAuthCli),
		fx.Provide(func() gatewayAPI.IWalletClient {
			return walletCli
		}),
		fx.Provide(func() jwtclient.IAuthClient {
			return remoteAuthCli
		}),
		fx.Provide(func() v1.FullNode {
			return client
		}),
		fx.Provide(func() filestore.FSRepo {
			return fsRepo
		}),

		// service
		service.MessagerService(),
		// api
		fx.Provide(api.NewMessageImp),

		// middleware

		fx.Provide(func() net.Listener {
			return lst
		}),

		fx.Provide(func() context.Context {
			return ctx
		}),
	)

	invoker := fx.Options(
		// invoke
		fx.Invoke(service.StartNodeEvents),
		fx.Invoke(metrics.SetupJaeger),
		fx.Invoke(metrics.SetupMetrics),
	)

	apiOption := fx.Options(
		fx.Provide(api.BindRateLimit),
		fx.Invoke(api.RunAPI),
	)

	app := fx.New(
		models.Options(),
		publisher.Options(),
		pubsub.Options(),
		provider,
		invoker,
		apiOption,
	)
	if err := app.Start(ctx); err != nil {
		// comment fx.NopLogger few lines above for easier debugging
		return fmt.Errorf("starting app: %w", err)
	}

	shutdownChan := make(chan struct{})
	// wait for exit to complete
	finishCh := make(chan struct{})
	go func() {
		<-shutdownChan

		log.Warn("received shutdown")

		log.Warn("Shutting down...")
		if err := app.Stop(ctx); err != nil {
			log.Errorf("graceful shutting down failed: %s", err)
		}
		log.Info("Graceful shutdown successful")

		close(finishCh)
	}()

	<-app.Done()

	shutdownChan <- struct{}{}

	<-finishCh

	return nil
}

func updateFlag(cfg *config.Config, ctx *cli.Context) error {
	if ctx.IsSet("listen") {
		cfg.API.Address = ctx.String("listen")
	}

	if ctx.IsSet("disable-push") {
		cfg.MessageService.SkipPushMessage = true
		cfg.MessageService.SkipProcessHead = true
	}

	if ctx.IsSet("auth-url") {
		cfg.JWT.AuthURL = ctx.String("auth-url")
	}

	if ctx.IsSet("node-url") {
		cfg.Node.Url = ctx.String("node-url")
	}

	if ctx.IsSet("gateway-url") {
		cfg.Gateway.Url = ctx.StringSlice("gateway-url")
	}

	if ctx.IsSet("auth-token") {
		token := ctx.String("auth-token")
		cfg.Node.Token = token
		cfg.Gateway.Token = token
		cfg.JWT.Token = token
	}

	if ctx.IsSet("node-token") {
		cfg.Node.Token = ctx.String("node-token")
	}

	if ctx.IsSet("gateway-token") {
		cfg.Gateway.Token = ctx.String("gateway-token")
	}

	if ctx.IsSet("db-type") {
		cfg.DB.Type = ctx.String("db-type")
		switch cfg.DB.Type {
		case "sqlite":
		case "mysql":
			if ctx.IsSet("mysql-dsn") {
				cfg.DB.MySql.ConnectionString = ctx.String("mysql-dsn")
			}
		default:
			return fmt.Errorf("unexpected db type %s", cfg.DB.Type)
		}
	}
	if ctx.IsSet("rate-limit-redis") {
		cfg.RateLimit.Redis = ctx.String("rate-limit-redis")
	}
	return nil
}

type fxLogger struct{}

func (l fxLogger) Printf(str string, args ...interface{}) {
	log.Infof(str, args...)
}

func hasFSRepo(repoPath string) (bool, error) {
	fi, err := os.Stat(repoPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	if !fi.IsDir() {
		return false, fmt.Errorf("%s is not a folder", repoPath)
	}

	return true, nil
}

// setEip155ChainID set the Eip155ChainID by different network.
// Use Eip155ChainID in https://github.com/filecoin-project/venus/blob/master/venus-shared/actors/types/message.go#L230
func setEip155ChainID(networkName string) error {
	var eip155ChainID int

	// 2k
	if strings.Contains(networkName, "localnet-") {
		eip155ChainID = networks.Net2k().Network.Eip155ChainID
	} else {
		netCfg, err := networks.GetNetworkConfigFromName(networkName)
		if err != nil {
			return err
		}
		eip155ChainID = netCfg.Network.Eip155ChainID
	}
	log.Infof("Eip155ChainID: %d", eip155ChainID)
	types.SetEip155ChainID(eip155ChainID)

	return nil
}
