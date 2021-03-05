package repo

import (
	"github.com/ipfs-force-community/venus-messager/types"
)

type MessageRepo interface {
	SaveMessage(msg *types.Message) (string, error)
	GetMessage(uuid string) (*types.Message, error)
	UpdateMessageReceipt(msg *types.Message) (string, error)
	ListMessage() ([]*types.Message, error)
	ListUnchainedMsgs() ([]*types.Message, error)
}
