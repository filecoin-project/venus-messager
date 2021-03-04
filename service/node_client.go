package service

import (
	"context"
	"net/http"
	"net/url"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/venus/pkg/chain"
	"github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	"github.com/ipfs-force-community/venus-messager/config"
)

type NodeClient struct {
	// ChainNotify returns channel with chain head updates.
	// First message is guaranteed to be of len == 1, and type == 'current'.
	ChainNotify func(context.Context) (<-chan []*chain.HeadChange, error)

	// MpoolBatchPush batch pushes a signed message to mempool.
	MpoolBatchPush func(context.Context, []*types.SignedMessage) ([]cid.Cid, error)

	ChainGetReceipts func(context.Context, cid.Cid) ([]types.MessageReceipt, error)

	ChainGetParentMessages func(ctx context.Context, cid cid.Cid) ([]types.Message, error)

	StateGetActor func(context.Context, address.Address, types.TipSetKey) (*types.Actor, error)

	StateSearchMsgLimited func(ctx context.Context, msg cid.Cid, limit abi.ChainEpoch) (*chain.MsgLookup, error)
}

func NewNodeClient(ctx context.Context, cfg *config.NodeConfig) (NodeClient, jsonrpc.ClientCloser, error) {
	headers := http.Header{}
	if len(cfg.Token) != 0 {
		headers.Add("Authorization", "Bearer "+string(cfg.Token))
	}
	addr, err := DialArgs(cfg.Url)
	if err != nil {
		return NodeClient{}, nil, err
	}
	var res NodeClient
	closer, err := jsonrpc.NewMergeClient(ctx, addr, "Filecoin", []interface{}{&res}, headers)
	return res, closer, err
}

func DialArgs(addr string) (string, error) {
	ma, err := multiaddr.NewMultiaddr(addr)
	if err == nil {
		_, addr, err := manet.DialArgs(ma)
		if err != nil {
			return "", err
		}

		return "ws://" + addr + "/rpc/v0", nil
	}

	_, err = url.Parse(addr)
	if err != nil {
		return "", err
	}
	return addr + "/rpc/v0", nil
}
