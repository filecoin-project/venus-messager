package cli

import (
	"context"
	"path/filepath"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/venus-messager/filestore"
	"github.com/filecoin-project/venus-messager/service"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	builtinactors "github.com/filecoin-project/venus/venus-shared/builtin-actors"
	"github.com/filecoin-project/venus/venus-shared/types"
	"github.com/filecoin-project/venus/venus-shared/utils"
	"github.com/ipfs-force-community/venus-common-utils/apiinfo"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/venus-messager/config"

	"github.com/filecoin-project/venus/venus-shared/api/messager"
)

func getAPI(ctx *cli.Context) (messager.IMessager, jsonrpc.ClientCloser, error) {
	cfg, err := getConfig(ctx)
	if err != nil {
		return nil, func() {}, err
	}

	apiInfo := apiinfo.NewAPIInfo(cfg.API.Address, cfg.JWT.Local.Token)
	addr, err := apiInfo.DialArgs("v0")
	if err != nil {
		return nil, nil, err
	}

	client, closer, err := messager.NewIMessagerRPC(ctx.Context, addr, apiInfo.AuthHeader())

	return client, closer, err
}

func getNodeAPI(ctx *cli.Context) (v1.FullNode, jsonrpc.ClientCloser, error) {
	cfg, err := getConfig(ctx)
	if err != nil {
		return nil, func() {}, err
	}
	return service.NewNodeClient(ctx.Context, &cfg.Node)
}

func getConfig(ctx *cli.Context) (*config.Config, error) {
	repoPath, err := homedir.Expand(ctx.String("repo"))
	if err != nil {
		return nil, err
	}

	return config.ReadConfig(filepath.Join(repoPath, filestore.ConfigFile))
}

func LoadBuiltinActors(ctx context.Context, cfg *config.Config) error {
	full, closer, err := service.NewNodeClient(ctx, &cfg.Node)
	if err != nil {
		return err
	}
	defer closer()
	networkName, err := full.StateNetworkName(ctx)
	if err != nil {
		return err
	}
	if err := builtinactors.SetNetworkBundle(networkNameToNetworkType(networkName)); err != nil {
		return err
	}
	utils.ReloadMethodsMap()

	return nil
}

func networkNameToNetworkType(networkName types.NetworkName) types.NetworkType {
	switch networkName {
	case "mainnet":
		return types.NetworkMainnet
	case "calibrationnet":
		return types.NetworkCalibnet
	case "butterflynet":
		return types.NetworkButterfly
	case "interopnet":
		return types.NetworkInterop
	default:
		return types.Network2k
	}
}
