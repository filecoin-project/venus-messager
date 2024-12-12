package integration

import (
	"context"
	"fmt"
	"net"
	"strings"

	"go.uber.org/fx"

	"github.com/mitchellh/go-homedir"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	"github.com/filecoin-project/go-jsonrpc"

	"github.com/ipfs-force-community/sophon-auth/jwtclient"

	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	gatewayAPI "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	"github.com/filecoin-project/venus/venus-shared/api/messager"

	"github.com/ipfs-force-community/sophon-messager/api"
	ccli "github.com/ipfs-force-community/sophon-messager/cli"
	"github.com/ipfs-force-community/sophon-messager/config"
	"github.com/ipfs-force-community/sophon-messager/filestore"
	"github.com/ipfs-force-community/sophon-messager/gateway"
	"github.com/ipfs-force-community/sophon-messager/metrics"
	"github.com/ipfs-force-community/sophon-messager/models"
	"github.com/ipfs-force-community/sophon-messager/publisher"
	"github.com/ipfs-force-community/sophon-messager/publisher/pubsub"
	"github.com/ipfs-force-community/sophon-messager/service"
	"github.com/ipfs-force-community/sophon-messager/testhelper"
	"github.com/ipfs-force-community/sophon-messager/utils"
)

type messagerServer struct {
	walletCli *gateway.MockWalletProxy
	fullNode  *testhelper.MockFullNode

	token string
	port  string

	app         *fx.App
	appStartErr chan error
}

func mockMessagerServer(ctx context.Context, repoPath string, cfg *config.Config, authClient jwtclient.IAuthClient) (*messagerServer, error) {
	repoPath, err := homedir.Expand(repoPath)
	if err != nil {
		return nil, err
	}

	remoteAuthClient := &jwtclient.AuthClient{}

	localAuthCli, token, err := jwtclient.NewLocalAuthClient()
	if err != nil {
		return nil, fmt.Errorf("failed to generate local auth client %v", err)
	}

	fsRepo, err := filestore.InitFSRepo(repoPath, cfg)
	if err != nil {
		return nil, err
	}
	utils.SetupLogLevels()

	fmt.Printf("node info url: %s, token: %s\n", cfg.Node.Url, cfg.Node.Token)
	fmt.Printf("auth info url: %s\n", cfg.JWT.AuthURL)
	fmt.Printf("gateway info url: %s, token: %s\n", cfg.Gateway.Url, cfg.Node.Token)
	fmt.Printf("rate limit info: redis: %s \n", cfg.RateLimit.Redis)

	fullNode, err := testhelper.NewMockFullNode(ctx, cfg.MessageService.WaitingChainHeadStableDuration*2)
	if err != nil {
		return nil, err
	}
	if err := ccli.LoadBuiltinActors(ctx, fullNode); err != nil {
		return nil, err
	}

	networkParams, err := fullNode.StateGetNetworkParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("get network params failed %v", err)
	}

	mAddr, err := ma.NewMultiaddr(cfg.API.Address)
	if err != nil {
		return nil, err
	}

	walletCli := gateway.NewMockWalletProxy()

	// Listen on the configured address in order to bind the port number in case it has
	// been configured as zero (i.e. OS-provided)
	apiListener, err := manet.Listen(mAddr)
	if err != nil {
		return nil, err
	}
	lst := manet.NetListener(apiListener)

	provider := fx.Options(
		// prover
		fx.Supply(cfg, &cfg.DB, &cfg.API, &cfg.JWT, &cfg.Node, &cfg.MessageService, cfg.Libp2pNet,
			&cfg.Gateway, &cfg.RateLimit, cfg.Trace, cfg.Metrics, cfg.Publisher),
		fx.Supply(fullNode),
		fx.Supply(networkParams.NetworkName),
		fx.Supply(remoteAuthClient),
		fx.Supply(localAuthCli),
		fx.Supply(networkParams),
		fx.Provide(func() gatewayAPI.IWalletClient {
			return walletCli
		}),
		fx.Provide(func() jwtclient.IAuthClient {
			return authClient
		}),
		fx.Provide(func() v1.FullNode {
			return fullNode
		}),
		fx.Provide(func() filestore.FSRepo {
			return fsRepo
		}),

		// service
		service.MessagerService(),
		// api
		fx.Provide(api.NewMessageImp),

		fx.Provide(func() context.Context {
			return ctx
		}),

		// middleware

		fx.Provide(func() net.Listener {
			return lst
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

	app := fx.New(provider,
		models.Options(),
		publisher.Options(),
		pubsub.Options(),
		invoker,
		apiOption,
	)

	return &messagerServer{
		walletCli:   walletCli,
		fullNode:    fullNode,
		token:       string(token),
		port:        strings.Split(lst.Addr().String(), ":")[1],
		app:         app,
		appStartErr: make(chan error),
	}, nil
}

func (ms *messagerServer) start(ctx context.Context) {
	ms.appStartErr <- ms.app.Start(ctx)
}

func (ms *messagerServer) stop(ctx context.Context) error {
	return ms.app.Stop(ctx)
}

func newMessagerClient(ctx context.Context, port, token string) (messager.IMessager, jsonrpc.ClientCloser, error) {
	return messager.DialIMessagerRPC(ctx, fmt.Sprintf("/ip4/127.0.0.1/tcp/%s", port), token, nil)
}
