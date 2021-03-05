package service

import (
	"context"
	venusTypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/sirupsen/logrus"
	"sync"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
)

type MessageService struct {
	repo repo.Repo
	log  *logrus.Logger

	nodeClient *NodeClient

	mutx sync.Mutex

	isStateRefreshTaskRunning bool
}

func NewMessageService(repo repo.Repo, nc *NodeClient, logger *logrus.Logger) *MessageService {
	return &MessageService{repo: repo, log: logger, nodeClient: nc}
}

func (ms *MessageService) PushMessage(ctx context.Context, msg *types.Message) (string, error) {
	return ms.repo.MessageRepo().SaveMessage(msg)
}

func (ms *MessageService) GetMessage(ctx context.Context, uuid string) (*types.Message, error) {
	return ms.repo.MessageRepo().GetMessage(uuid)
}
func (ms *MessageService) ListMessage(ctx context.Context) ([]*types.Message, error) {
	return ms.repo.MessageRepo().ListMessage()
}

func (ms *MessageService) ReconnectCheck(ctx context.Context, head *venusTypes.TipSet) error {
	ms.log.Infof("reconnect to node ")
	return nil
}

func (ms *MessageService) ProcessNewHead(ctx context.Context, apply, revert []*venusTypes.TipSet) error {
	ms.log.Infof("receive new head from chain")
	ms.GoRefreshMessageState()
	return nil
}
