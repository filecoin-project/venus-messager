package cli

import (
	"context"
	"path/filepath"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/venus-messager/filestore"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	"github.com/filecoin-project/venus/venus-shared/utils"
	"github.com/ipfs-force-community/venus-common-utils/apiinfo"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/venus-messager/config"
	utils2 "github.com/filecoin-project/venus-messager/utils"

	"github.com/filecoin-project/venus/venus-shared/api/messager"
)

func getAPI(ctx *cli.Context) (messager.IMessager, jsonrpc.ClientCloser, error) {
	cfg, err := getConfig(ctx)
	if err != nil {
		return nil, func() {}, err
	}

	return NewMessagerAPI(ctx.Context, cfg.API.Address, cfg.JWT.Local.Token)
}

func NewMessagerAPI(ctx context.Context, addr, token string) (messager.IMessager, jsonrpc.ClientCloser, error) {
	apiInfo := apiinfo.NewAPIInfo(addr, token)
	addr, err := apiInfo.DialArgs("v0")
	if err != nil {
		return nil, nil, err
	}

	client, closer, err := messager.NewIMessagerRPC(ctx, addr, apiInfo.AuthHeader())

	return client, closer, err
}

func getNodeAPI(ctx *cli.Context) (v1.FullNode, jsonrpc.ClientCloser, error) {
	cfg, err := getConfig(ctx)
	if err != nil {
		return nil, func() {}, err
	}
	return v1.DialFullNodeRPC(ctx.Context, cfg.Node.Url, cfg.Node.Token, nil)
}

func NewNodeAPI(ctx context.Context, addr, token string) (v1.FullNode, jsonrpc.ClientCloser, error) {
	return v1.DialFullNodeRPC(ctx, addr, token, nil)
}

func getConfig(ctx *cli.Context) (*config.Config, error) {
	repoPath, err := homedir.Expand(ctx.String("repo"))
	if err != nil {
		return nil, err
	}
	cfg := new(config.Config)

	err = utils2.ReadConfig(filepath.Join(repoPath, filestore.ConfigFile), cfg)

	return cfg, err
}

func LoadBuiltinActors(ctx context.Context, nodeAPI v1.FullNode) error {
	if err := utils.LoadBuiltinActors(ctx, nodeAPI); err != nil {
		return err
	}
	utils.ReloadMethodsMap()

	return nil
}
