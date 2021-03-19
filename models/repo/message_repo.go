package repo

import (
	"time"

	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/go-address"
	venustypes "github.com/filecoin-project/venus/pkg/types"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs-force-community/venus-messager/types"
)

type MessageRepo interface {
	ExpireMessage(msg []*types.Message) error
	BatchSaveMessage(msg []*types.Message) error
	CreateMessage(msg *types.Message) error
	SaveMessage(msg *types.Message) (types.UUID, error)

	GetMessageByFromAndNonce(from address.Address, nonce uint64) (*types.Message, error)
	GetMessageByUid(uuid types.UUID) (*types.Message, error)
	GetMessageState(uuid types.UUID) (types.MessageState, error)
	GetMessageByCid(unsignedCid cid.Cid) (*types.Message, error)
	GetMessageBySignedCid(signedCid cid.Cid) (*types.Message, error)
	GetSignedMessageByTime(start time.Time) ([]*types.Message, error)
	GetSignedMessageByHeight(height abi.ChainEpoch) ([]*types.Message, error)
	ListMessage() ([]*types.Message, error)
	ListUnChainMessageByAddress(addr address.Address) ([]*types.Message, error)
	ListFilledMessageByAddress(addr address.Address) ([]*types.Message, error)
	ListFilledMessageByHeight(height abi.ChainEpoch) ([]*types.Message, error)
	ListUnchainedMsgs() ([]*types.Message, error)
	ListSignedMsgs() ([]*types.Message, error)
	ListFilledMessageBelowNonce(addr address.Address, nonce uint64) ([]*types.Message, error)

	UpdateMessageInfoByCid(unsignedCid string, receipt *venustypes.MessageReceipt, height abi.ChainEpoch, state types.MessageState, tsKey venustypes.TipSetKey) (string, error)
	UpdateMessageStateByCid(unsignedCid string, state types.MessageState) (string, error)
	UpdateMessageStateByID(uuid types.UUID, state types.MessageState) (types.UUID, error)
}
