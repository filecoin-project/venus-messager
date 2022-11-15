package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
	"modernc.org/mathutil"

	"github.com/filecoin-project/venus/pkg/crypto"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	gatewayAPI "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
	logging "github.com/ipfs/go-log/v2"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/metrics"
	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/publisher"
	"github.com/filecoin-project/venus-messager/utils"
)

const (
	gasEstimate = "gas estimate: "
	signMsg     = "sign msg: "
)

var msgSelectLog = logging.Logger("msg-select")

func logWithAddress(addr address.Address) *zap.SugaredLogger {
	return msgSelectLog.With("address", addr.String())
}

type messageSelector struct {
	ctx            context.Context
	repo           repo.Repo
	cfg            *config.MessageServiceConfig
	fullNode       v1.FullNode
	addressService *AddressService
	sps            *SharedParamsService
	walletClient   gatewayAPI.IWalletClient

	works       map[address.Address]*work
	msgReceiver publisher.MessageReceiver
}

func newMessageSelector(ctx context.Context,
	repo repo.Repo,
	cfg *config.MessageServiceConfig,
	fullNode v1.FullNode,
	addressService *AddressService,
	sps *SharedParamsService,
	walletClient gatewayAPI.IWalletClient,
	msgReceiver publisher.MessageReceiver,
) (*messageSelector, error) {
	ms := &messageSelector{
		ctx:            ctx,
		repo:           repo,
		cfg:            cfg,
		fullNode:       fullNode,
		addressService: addressService,
		sps:            sps,
		walletClient:   walletClient,

		msgReceiver: msgReceiver,
		works:       make(map[address.Address]*work),
	}

	addrInfos, err := ms.addressService.ListActiveAddress(ctx)
	if err != nil {
		return nil, err
	}

	return ms, ms.tryUpdateWorks(addressMap(addrInfos))
}

// SelectMessage not concurrency safe
func (ms *messageSelector) SelectMessage(ctx context.Context, ts *venusTypes.TipSet) error {
	sharedParams, err := ms.sps.GetSharedParams(ctx)
	if err != nil {
		return err
	}

	activeAddrs, err := ms.addressService.ListActiveAddress(ctx)
	if err != nil {
		return err
	}
	addrSelMsgNum := addrSelectMsgNum(activeAddrs, sharedParams.SelMsgNum)
	addrInfos := addressMap(activeAddrs)
	if err := ms.tryUpdateWorks(addrInfos); err != nil {
		msgSelectLog.Warnf("failed to update work %v", err)
	}

	appliedNonce, err := ms.getNonceInTipset(ctx, ts)
	if err != nil {
		return err
	}

	for _, w := range ms.works {
		if w.getState() == working {
			msgSelectLog.Infof("%s is already selecting message, had took %v", w.addr, time.Since(w.start))
			continue
		}
		w.setState(working)
		w.start = time.Now()
		go func(w *work) {
			ctx, cancel := context.WithTimeout(ctx, (ms.cfg.SignMessageTimeout+ms.cfg.EstimateMessageTimeout)*time.Second)
			defer cancel()
			defer w.setState(free)

			log := logWithAddress(w.addr)
			selectResult, err := w.selectMessage(ctx, appliedNonce, addrInfos[w.addr], ts, addrSelMsgNum[w.addr], sharedParams)
			if err != nil {
				log.Errorf("select message failed %v", err)
				return
			}
			log.Infof("select message result | SelectMsg: %d | ToPushMsg: %d | ErrMsg: %d | took: %v", len(selectResult.SelectMsg),
				len(selectResult.ToPushMsg), len(selectResult.ErrMsg), time.Since(w.start))

			recordMetric(ctx, w.addr, selectResult)

			if err := ms.saveSelectedMessages(ctx, selectResult); err != nil {
				log.Errorf("failed to save selected messages to db %v", err)
				return
			}

			for _, msg := range selectResult.SelectMsg {
				selectResult.ToPushMsg = append(selectResult.ToPushMsg, &venusTypes.SignedMessage{
					Message:   msg.Message,
					Signature: *msg.Signature,
				})
			}
			sort.Slice(selectResult.ToPushMsg, func(i, j int) bool {
				return selectResult.ToPushMsg[i].Message.Nonce < selectResult.ToPushMsg[j].Message.Nonce
			})

			// send messages to push
			select {
			case ms.msgReceiver <- selectResult.ToPushMsg:
			default:
				log.Errorf("message receiver channel is full, skip message %v %v", w.addr, len(selectResult.ToPushMsg))
			}
		}(w)
	}

	return nil
}

func (ms *messageSelector) saveSelectedMessages(ctx context.Context, selectResult *MsgSelectResult) error {
	startSaveDB := time.Now()
	log := msgSelectLog.With("address", selectResult.Address.Addr.String())
	log.Infof("start save messages to database")
	err := ms.repo.Transaction(func(txRepo repo.TxRepo) error {
		if len(selectResult.SelectMsg) > 0 {
			if err := txRepo.MessageRepo().BatchSaveMessage(selectResult.SelectMsg); err != nil {
				return err
			}

			addrInfo := selectResult.Address
			if err := txRepo.AddressRepo().UpdateNonce(ctx, addrInfo.Addr, addrInfo.Nonce); err != nil {
				return err
			}
		}

		for _, m := range selectResult.ErrMsg {
			msgSelectLog.Infof("update message %s return value with error %s", m.id, m.err)
			if err := txRepo.MessageRepo().UpdateErrMsg(m.id, m.err); err != nil {
				return err
			}
		}
		return nil
	})
	log.Infof("end save messages to database, took %v, err %v", time.Since(startSaveDB), err)

	return err
}

func (ms *messageSelector) getNonceInTipset(ctx context.Context, ts *venusTypes.TipSet) (*utils.NonceMap, error) {
	applied := utils.NewNonceMap()
	// todo change with venus/lotus message for tipset
	selectMsg := func(m *venusTypes.Message) error {
		// The first match for a sender is guaranteed to have correct nonce -- the block isn't valid otherwise
		if _, ok := applied.Get(m.From); !ok {
			applied.Add(m.From, m.Nonce)
		}

		val, _ := applied.Get(m.From)
		if val != m.Nonce {
			return nil
		}
		val++
		applied.Add(m.From, val)
		return nil
	}

	msgs, err := ms.fullNode.ChainGetMessagesInTipset(ctx, ts.Key())
	if err != nil {
		return nil, fmt.Errorf("failed to get message in tipset %v", err)
	}
	for _, msg := range msgs {
		err := selectMsg(msg.Message)
		if err != nil {
			return nil, fmt.Errorf("failed to decide whether to select message for block: %w", err)
		}
	}

	return applied, nil
}

func (ms *messageSelector) tryUpdateWorks(addrInfos map[address.Address]*types.Address) error {
	ws := make(map[address.Address]*work, len(addrInfos))
	for _, addrInfo := range addrInfos {
		w, ok := ms.works[addrInfo.Addr]
		if !ok {
			msgSelectLog.Infof("add a work %v", addrInfo.Addr)
			ws[addrInfo.Addr] = newWork(addrInfo.Addr, ms.cfg, ms.fullNode, ms.repo, ms.addressService, ms.walletClient)
		} else {
			ws[addrInfo.Addr] = w
			delete(ms.works, addrInfo.Addr)
		}
	}
	for addr, w := range ms.works {
		if _, ok := ws[addr]; !ok {
			if w.state == working {
				ws[addr] = w
			} else {
				msgSelectLog.Infof("remove a work %v", addr)
				delete(ms.works, addr)
			}
		}
	}
	ms.works = ws

	return nil
}

func addressMap(addrList []*types.Address) map[address.Address]*types.Address {
	addrs := make(map[address.Address]*types.Address, len(addrList))
	for i, addr := range addrList {
		addrs[addr.Addr] = addrList[i]
	}

	return addrs
}

func addrSelectMsgNum(addrList []*types.Address, defSelMsgNum uint64) map[address.Address]uint64 {
	selMsgNum := make(map[address.Address]uint64)
	for _, addr := range addrList {
		if num, ok := selMsgNum[addr.Addr]; ok && addr.SelMsgNum > 0 && num < addr.SelMsgNum {
			selMsgNum[addr.Addr] = addr.SelMsgNum
		} else if !ok {
			if addr.SelMsgNum == 0 {
				selMsgNum[addr.Addr] = defSelMsgNum
			} else {
				selMsgNum[addr.Addr] = addr.SelMsgNum
			}
		}
	}

	return selMsgNum
}

func recordMetric(ctx context.Context, addr address.Address, selectResult *MsgSelectResult) {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.WalletAddress, addr.String()))

	stats.Record(ctx, metrics.SelectedMsgNumOfLastRound.M(int64(len(selectResult.SelectMsg))))
	stats.Record(ctx, metrics.ToPushMsgNumOfLastRound.M(int64(len(selectResult.ToPushMsg))))
	stats.Record(ctx, metrics.ErrMsgNumOfLastRound.M(int64(len(selectResult.ErrMsg))))
}

const (
	free workState = iota
	working
)

var errSingMessage = errors.New("sign message faield")

type workState int

type MsgSelectResult struct {
	Address   *types.Address
	SelectMsg []*types.Message
	ToPushMsg []*venusTypes.SignedMessage
	ErrMsg    []msgErrInfo
}

type msgErrInfo struct {
	id  string
	err string
}

type work struct {
	addr           address.Address
	cfg            *config.MessageServiceConfig
	fullNode       v1.FullNode
	repo           repo.Repo
	addressService *AddressService
	walletClient   gatewayAPI.IWalletClient

	state workState
	start time.Time

	lk sync.Mutex
}

func newWork(addr address.Address,
	cfg *config.MessageServiceConfig,
	fullNode v1.FullNode,
	repo repo.Repo,
	addressService *AddressService,
	walletClient gatewayAPI.IWalletClient,
) *work {
	return &work{
		addr:           addr,
		cfg:            cfg,
		addressService: addressService,
		fullNode:       fullNode,
		repo:           repo,
		walletClient:   walletClient,
		state:          free,
	}
}

func (w *work) selectMessage(ctx context.Context, appliedNonce *utils.NonceMap, addrInfo *types.Address, ts *venusTypes.TipSet, maxAllowPendingMessage uint64, sharedParams *types.SharedSpec) (*MsgSelectResult, error) {
	log := logWithAddress(addrInfo.Addr)

	// 没有绑定账号肯定无法签名
	accounts, err := w.addressService.GetAccountsOfSigner(ctx, addrInfo.Addr)
	if err != nil {
		log.Errorf("get account failed %v", err)
		return nil, err
	}

	// 判断是否需要推送消息
	nonceInLatestTs, actorNonce, err := w.getNonce(ctx, ts, appliedNonce)
	if err != nil {
		return nil, err
	}
	if nonceInLatestTs > addrInfo.Nonce {
		log.Warnf("nonce in db %d is smaller than nonce on chain %d, update to latest", addrInfo.Nonce, nonceInLatestTs)
		addrInfo.Nonce = nonceInLatestTs
		addrInfo.UpdatedAt = time.Now()
		err := w.repo.AddressRepo().UpdateNonce(ctx, addrInfo.Addr, addrInfo.Nonce)
		if err != nil {
			return nil, fmt.Errorf("update nonce failed %v", err)
		}
	}

	toPushMessage := w.getFilledMessage(nonceInLatestTs)

	// calc the message needed
	nonceGap := addrInfo.Nonce - nonceInLatestTs
	if nonceGap >= maxAllowPendingMessage {
		log.Errorf("there are %d message not to be package", len(toPushMessage), nonceGap)
		return &MsgSelectResult{
			ToPushMsg: toPushMessage,
			Address:   addrInfo,
		}, nil
	}
	wantCount := maxAllowPendingMessage - nonceGap
	log.Infof("state actor nonce %d, latest nonce in ts %d, assigned nonce %d, nonce gap %d, want %d", actorNonce, nonceInLatestTs, addrInfo.Nonce, nonceGap, wantCount)

	// get unfill message
	selectCount := mathutil.MinUint64(wantCount*2, 100)
	messages, err := w.repo.MessageRepo().ListUnChainMessageByAddress(addrInfo.Addr, int(selectCount))
	if err != nil {
		return nil, fmt.Errorf("list unfill message error %v", err)
	}

	if len(messages) == 0 {
		log.Infof("have no unfill message")
		return &MsgSelectResult{
			ToPushMsg: toPushMessage,
			Address:   addrInfo,
		}, nil
	}

	var errMsg []msgErrInfo
	count := uint64(0)
	selectMsg := make([]*types.Message, 0, len(messages))

	estimateResult, candidateMessages, err := w.estimateMessage(ctx, ts, messages, sharedParams, addrInfo)
	if err != nil {
		return nil, err
	}

	// sign
	for index, msg := range candidateMessages {
		// if error print error message
		if len(estimateResult[index].Err) != 0 {
			errMsg = append(errMsg, msgErrInfo{id: msg.ID, err: gasEstimate + estimateResult[index].Err})
			log.Errorf("estimate message %s fail %s", msg.ID, estimateResult[index].Err)
			continue
		}
		estimateMsg := estimateResult[index].Msg
		if count >= wantCount {
			break
		}

		// 分配nonce
		msg.Nonce = addrInfo.Nonce
		msg.GasFeeCap = estimateMsg.GasFeeCap
		msg.GasPremium = estimateMsg.GasPremium
		msg.GasLimit = estimateMsg.GasLimit

		unsignedCid := msg.Message.Cid()
		msg.UnsignedCid = &unsignedCid

		// 签名
		sig, err := w.signMessage(ctx, msg, accounts)
		if err != nil {
			if errors.Is(err, errSingMessage) {
				errMsg = append(errMsg, msgErrInfo{id: msg.ID, err: fmt.Sprintf("%v%v", signMsg, errors.Unwrap(err))})
				log.Errorf("sign message %s failed %v", msg.ID, err)
				break
			}
			log.Error(err)
			continue
		}

		msg.Signature = sig
		msg.State = types.FillMsg

		// signed cid for t1 address
		signedMsg := venusTypes.SignedMessage{
			Message:   msg.Message,
			Signature: *msg.Signature,
		}
		signedCid := signedMsg.Cid()
		msg.SignedCid = &signedCid

		selectMsg = append(selectMsg, msg)
		addrInfo.Nonce++
		count++
	}

	return &MsgSelectResult{
		SelectMsg: selectMsg,
		ToPushMsg: toPushMessage,
		Address:   addrInfo,
		ErrMsg:    errMsg,
	}, nil
}

func (w *work) getNonce(ctx context.Context, ts *venusTypes.TipSet, appliedNonce *utils.NonceMap) (uint64, uint64, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, w.cfg.DefaultTimeout)
	defer cancel()
	actorI, err := handleTimeout(timeoutCtx, w.fullNode.StateGetActor, []interface{}{w.addr, ts.Key()})
	if err != nil {
		return 0, 0, err
	}
	actor := actorI.(*venusTypes.Actor)
	nonceInLatestTs := actor.Nonce
	// todo actor nonce maybe the latest ts. not need appliedNonce
	if nonceInTs, ok := appliedNonce.Get(w.addr); ok {
		msgSelectLog.Infof("update address %s nonce in ts %d  nonce in actor %d", w.addr, nonceInTs, nonceInLatestTs)
		nonceInLatestTs = nonceInTs
	}

	return nonceInLatestTs, actor.Nonce, nil
}

func (w *work) getFilledMessage(nonceInLatestTs uint64) []*venusTypes.SignedMessage {
	filledMessage, err := w.repo.MessageRepo().ListFilledMessageByAddress(w.addr)
	if err != nil {
		msgSelectLog.Warnf("list filled message %v", err)
	}
	msgs := make([]*venusTypes.SignedMessage, 0, len(filledMessage))
	for _, msg := range filledMessage {
		if nonceInLatestTs > msg.Nonce {
			continue
		}
		msgs = append(msgs, &venusTypes.SignedMessage{
			Message:   msg.Message,
			Signature: *msg.Signature,
		})
	}

	return msgs
}

func (w *work) estimateMessage(ctx context.Context,
	ts *venusTypes.TipSet,
	msgs []*types.Message,
	sharedParams *types.SharedSpec,
	addrInfo *types.Address,
) ([]*venusTypes.EstimateResult, []*types.Message, error) {
	candidateMessages := make([]*types.Message, 0, len(msgs))
	estimateMesssages := make([]*venusTypes.EstimateMessage, 0, len(msgs))

	for _, msg := range msgs {
		// global msg meta
		newMsgMeta := mergeMsgSpec(sharedParams, msg.Meta, addrInfo, msg)

		if msg.GasFeeCap.NilOrZero() && !newMsgMeta.GasFeeCap.NilOrZero() {
			msg.GasFeeCap = newMsgMeta.GasFeeCap
		}

		baseFee := ts.At(0).ParentBaseFee
		if !newMsgMeta.BaseFee.NilOrZero() && baseFee.GreaterThan(newMsgMeta.BaseFee) {
			msgSelectLog.Infof("skip msg %v, base fee too height %v(local) < %v(chain), height %v", msg.ID, newMsgMeta.BaseFee, baseFee, ts.Height())
			continue
		}

		candidateMessages = append(candidateMessages, msg)
		estimateMesssages = append(estimateMesssages, &venusTypes.EstimateMessage{
			Msg: &msg.Message,
			Spec: &venusTypes.MessageSendSpec{
				MaxFee:            newMsgMeta.MaxFee,
				GasOverEstimation: newMsgMeta.GasOverEstimation,
				GasOverPremium:    newMsgMeta.GasOverPremium,
			},
		})

		msgSelectLog.Infof("estimate message %s, gas fee cap %s, gas limit %v, gas premium: %s, "+
			"meta maxfee %s, over estimation %f, gas over premium %f", msg.ID, msg.GasFeeCap, msg.GasLimit, msg.GasPremium,
			newMsgMeta.MaxFee, newMsgMeta.GasOverEstimation, newMsgMeta.GasOverPremium)
	}

	estimateMsgCtx, estimateMsgCancel := context.WithTimeout(ctx, w.cfg.EstimateMessageTimeout)
	defer estimateMsgCancel()

	estimateResult, err := w.fullNode.GasBatchEstimateMessageGas(estimateMsgCtx, estimateMesssages, addrInfo.Nonce, ts.Key())

	return estimateResult, candidateMessages, err
}

func (w *work) signMessage(ctx context.Context, msg *types.Message, accounts []string) (*crypto.Signature, error) {
	data, err := msg.Message.ToStorageBlock()
	if err != nil {
		return nil, fmt.Errorf("serialize message %s failed %v", msg.ID, err)
	}

	signMsgCtx, signMsgCancel := context.WithTimeout(ctx, w.cfg.SignMessageTimeout)
	sigI, err := handleTimeout(signMsgCtx, w.walletClient.WalletSign, []interface{}{w.addr, accounts, msg.Message.Cid().Bytes(), venusTypes.MsgMeta{
		Type:  venusTypes.MTChainMsg,
		Extra: data.RawData(),
	}})
	signMsgCancel()
	if err != nil {
		return nil, fmt.Errorf("%v %w", err, errSingMessage)
	}

	return sigI.(*crypto.Signature), nil
}

func (w *work) setState(state workState) {
	w.lk.Lock()
	defer w.lk.Unlock()

	w.state = state
}

func (w *work) getState() workState {
	w.lk.Lock()
	defer w.lk.Unlock()

	return w.state
}

func CapGasFee(msg *venusTypes.Message, maxFee abi.TokenAmount) {
	if maxFee.NilOrZero() {
		return
	}

	gl := venusTypes.NewInt(uint64(msg.GasLimit))
	totalFee := venusTypes.BigMul(msg.GasFeeCap, gl)

	if totalFee.LessThanEqual(maxFee) {
		return
	}

	msg.GasFeeCap = big.Div(maxFee, gl)
	msg.GasPremium = big.Min(msg.GasFeeCap, msg.GasPremium) // cap premium at FeeCap
}

type GasSpec struct {
	GasOverEstimation float64
	MaxFee            big.Int
	GasOverPremium    float64
	GasFeeCap         big.Int
	BaseFee           big.Int
}

func mergeMsgSpec(globalSpec *types.SharedSpec, sendSpec *types.SendSpec, addrInfo *types.Address, msg *types.Message) *GasSpec {
	newMsgMeta := &GasSpec{
		GasOverEstimation: sendSpec.GasOverEstimation,
		GasOverPremium:    sendSpec.GasOverPremium,
		MaxFee:            sendSpec.MaxFee,
	}

	if sendSpec.GasOverEstimation == 0 {
		if addrInfo.GasOverEstimation != 0 {
			newMsgMeta.GasOverEstimation = addrInfo.GasOverEstimation
		} else if globalSpec != nil {
			newMsgMeta.GasOverEstimation = globalSpec.GasOverEstimation
		}
	}
	if sendSpec.MaxFee.NilOrZero() {
		if !addrInfo.MaxFee.NilOrZero() {
			newMsgMeta.MaxFee = addrInfo.MaxFee
		} else if globalSpec != nil {
			newMsgMeta.MaxFee = globalSpec.MaxFee
		}
	}

	if msg.GasFeeCap.NilOrZero() {
		if !addrInfo.GasFeeCap.NilOrZero() {
			newMsgMeta.GasFeeCap = addrInfo.GasFeeCap
		} else if globalSpec != nil {
			newMsgMeta.GasFeeCap = globalSpec.GasFeeCap
		}
	}

	if sendSpec.GasOverPremium == 0 {
		if addrInfo.GasOverPremium != 0 {
			newMsgMeta.GasOverPremium = addrInfo.GasOverPremium
		} else if globalSpec.GasOverPremium != 0 {
			newMsgMeta.GasOverPremium = globalSpec.GasOverPremium
		}
	}

	if !addrInfo.BaseFee.NilOrZero() {
		newMsgMeta.BaseFee = addrInfo.BaseFee
	} else if globalSpec != nil {
		newMsgMeta.BaseFee = globalSpec.BaseFee
	}

	return newMsgMeta
}
