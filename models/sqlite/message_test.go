package sqlite

import (
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/filecoin-project/venus-messager/utils"
	venustypes "github.com/filecoin-project/venus/venus-shared/types"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/venus-messager/testhelper"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

func TestSaveAndGetMessage(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	msgCount := 100
	msgs := testhelper.NewMessages(msgCount)
	for _, msg := range msgs {
		assert.NoError(t, messageRepo.CreateMessage(msg))
	}

	msg := msgs[0]

	// tes get message by uid
	result, err := messageRepo.GetMessageByUid(msg.ID)
	assert.NoError(t, err)
	compareMsg(t, msg, result)

	_, err = messageRepo.GetMessageByUid(uuid.NewString())
	assert.Error(t, err)

	// test has message by uid
	has, err := messageRepo.HasMessageByUid(msg.ID)
	assert.NoError(t, err)
	assert.True(t, has)
	has, err = messageRepo.HasMessageByUid(uuid.NewString())
	assert.NoError(t, err)
	assert.False(t, has)

	// test list message
	allMsg, err := messageRepo.ListMessage()
	assert.NoError(t, err)
	assert.Equal(t, msgCount, len(allMsg))

	// test save message
	oneMsg := testhelper.NewMessage()
	assert.NoError(t, messageRepo.SaveMessage(oneMsg))
	res, err := messageRepo.GetMessageByUid(oneMsg.ID)
	assert.NoError(t, err)
	compareMsg(t, oneMsg, res)
	// save again, we expect CreateAt not change and UpdateAt changed
	assert.NoError(t, messageRepo.SaveMessage(oneMsg))
	res2, err := messageRepo.GetMessageByUid(oneMsg.ID)
	assert.NoError(t, err)
	assert.Equal(t, res.CreatedAt, res2.CreatedAt)
	assert.True(t, res.UpdatedAt.Before(res2.UpdatedAt))

	// save changed message
	msg.Nonce = 100
	msg.GasLimit = 1000
	msg.GasPremium = abi.NewTokenAmount(10000)
	msg.GasFeeCap = abi.NewTokenAmount(1000001)
	msg.Height = 10000
	msg.State = types.OnChainMsg
	msg.Receipt = &venustypes.MessageReceipt{
		ExitCode: -1,
		Return:   []byte("return"),
		GasUsed:  1000011,
	}
	msgCid := msg.Cid()
	msg.SignedCid = &msgCid
	msg.UnsignedCid = &msgCid
	assert.NoError(t, messageRepo.SaveMessage(msg))
	res, err = messageRepo.GetMessageByUid(msg.ID)
	assert.NoError(t, err)
	compareMsg(t, msg, res)

	// test batch save message
	msgs2 := testhelper.NewMessages(msgCount)
	assert.NoError(t, messageRepo.BatchSaveMessage(msgs2))
	for _, msg := range msgs2 {
		res, err := messageRepo.GetMessageByUid(msg.ID)
		assert.NoError(t, err)
		compareMsg(t, msg, res)
	}
}

func TestGetMessageByFromAndNonce(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	msgs := testhelper.NewSignedMessages(100)
	for _, msg := range msgs {
		assert.NoError(t, messageRepo.CreateMessage(msg))

		res, err := messageRepo.GetMessageByFromAndNonce(msg.From, msg.Nonce)
		assert.NoError(t, err)
		compareMsg(t, msg, res)

		res, err = messageRepo.GetMessageByFromNonceAndState(msg.From, msg.Nonce, msg.State)
		assert.NoError(t, err)
		compareMsg(t, msg, res)
	}
}

func TestExpireMessage(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	msg := testhelper.NewSignedMessages(1)[0]

	err := messageRepo.CreateMessage(msg)
	assert.NoError(t, err)

	err = messageRepo.ExpireMessage([]*types.Message{msg})
	assert.NoError(t, err)

	msg2, err := messageRepo.GetMessageByUid(msg.ID)
	assert.NoError(t, err)
	assert.Equal(t, types.FailedMsg, msg2.State)
}

func TestGetMessageState(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	msg := testhelper.NewMessage()
	err := messageRepo.CreateMessage(msg)
	assert.NoError(t, err)
	state, err := messageRepo.GetMessageState(msg.ID)
	assert.NoError(t, err)
	assert.Equal(t, state, types.UnFillMsg)

	for _, state := range []types.MessageState{types.UnFillMsg, types.FillMsg, types.OnChainMsg, types.FailedMsg} {
		msg.State = state
		err = messageRepo.SaveMessage(msg)
		assert.NoError(t, err)
		msgState, err := messageRepo.GetMessageState(msg.ID)
		assert.NoError(t, err)
		assert.Equal(t, state, msgState)
	}
}

func TestGetMessageByCid(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	msgs := testhelper.NewSignedMessages(100)
	for _, msg := range msgs {
		assert.NoError(t, messageRepo.CreateMessage(msg))

		res, err := messageRepo.GetMessageByCid(*msg.UnsignedCid)
		assert.NoError(t, err)
		assert.Equal(t, msg.ID, res.ID)

		res, err = messageRepo.GetMessageBySignedCid(*msg.SignedCid)
		assert.NoError(t, err)
		assert.Equal(t, msg.ID, res.ID)
	}
}

func TestGetSignedMessageByTime(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	msg := testhelper.NewMessage()
	err := messageRepo.CreateMessage(msg)
	assert.NoError(t, err)

	signedMsgs := testhelper.NewSignedMessages(10)
	for _, msg := range signedMsgs {
		err := messageRepo.CreateMessage(msg)
		assert.NoError(t, err)
	}
	startTime := time.Now().Add(-time.Second * 3600)
	msgs, err := messageRepo.GetSignedMessageByTime(startTime)
	assert.NoError(t, err)
	assert.LessOrEqual(t, 10, len(msgs))
}

func TestGetSignedMessageByHeight(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	msg := testhelper.NewMessage()
	err := messageRepo.CreateMessage(msg)
	assert.NoError(t, err)

	signedMsgs := testhelper.NewSignedMessages(10)
	for i, msg := range signedMsgs {
		msg.Height = int64(i)
		err := messageRepo.CreateMessage(msg)
		assert.NoError(t, err)
	}
	height := abi.ChainEpoch(5)
	msgs, err := messageRepo.GetSignedMessageByHeight(height)
	assert.NoError(t, err)
	assert.LessOrEqual(t, 5, len(msgs))
}

func TestGetSignedMessageFromFailedMsg(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	signedMsgs := testhelper.NewSignedMessages(10)
	addrs := make([]address.Address, len(signedMsgs))
	for i, msg := range signedMsgs {
		if i%2 == 0 {
			msg.State = types.FailedMsg
		}
		addrs[i] = msg.From
		assert.NoError(t, messageRepo.CreateMessage(msg))
	}
	for i, addr := range addrs {
		msgs, err := messageRepo.GetSignedMessageFromFailedMsg(addr)
		assert.NoError(t, err)
		if i%2 == 0 {
			assert.Len(t, msgs, 1)
		} else {
			assert.Len(t, msgs, 0)
		}
	}
}

func TestListMessageByFromState(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	addr, err := address.NewActorAddress(uuid.New().NodeID())
	assert.NoError(t, err)

	msgList, err := messageRepo.ListMessageByFromState(addr, types.UnFillMsg, false, 1, 100)
	assert.NoError(t, err)
	assert.Len(t, msgList, 0)

	msgList, err = messageRepo.ListMessageByFromState(addr, types.UnFillMsg, false, 0, 100)
	assert.NoError(t, err)
	assert.Len(t, msgList, 0)

	msgCount := 100
	onChainMsgCount := 0
	isAsc := true
	msgs := testhelper.NewMessages(msgCount)
	for _, msg := range msgs {
		msg.State = types.MessageState(rand.Intn(7))
		if msg.State == types.OnChainMsg {
			msg.From = addr
			onChainMsgCount++
		}
		assert.NoError(t, messageRepo.CreateMessage(msg))
	}

	msgList, err = messageRepo.ListMessageByFromState(addr, types.OnChainMsg, isAsc, 1, onChainMsgCount)
	assert.NoError(t, err)
	assert.Equal(t, onChainMsgCount, len(msgList))
	sorted := sort.SliceIsSorted(msgList, func(i, j int) bool {
		return msgList[i].CreatedAt.Before(msgList[j].CreatedAt)
	})
	assert.True(t, sorted)

	msgList, err = messageRepo.ListMessageByFromState(addr, types.OnChainMsg, isAsc, 1, onChainMsgCount/2)
	assert.NoError(t, err)
	assert.Equal(t, onChainMsgCount/2, len(msgList))
}

func TestListMessageByAddress(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	addr, err := address.NewActorAddress(uuid.New().NodeID())
	assert.NoError(t, err)

	msgList, err := messageRepo.ListMessageByAddress(addr)
	assert.NoError(t, err)
	assert.Len(t, msgList, 0)

	msgCount := 100
	count := 0
	msgs := testhelper.NewMessages(msgCount)
	for _, msg := range msgs {
		if rand.Intn(msgCount)%2 == 0 {
			msg.From = addr
			count++
		}
		assert.NoError(t, messageRepo.CreateMessage(msg))
	}

	msgList, err = messageRepo.ListMessageByAddress(addr)
	assert.NoError(t, err)
	assert.Equal(t, count, len(msgList))
}

func TestListFailedMessage(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	msgList, err := messageRepo.ListFailedMessage()
	assert.NoError(t, err)
	assert.Len(t, msgList, 0)

	msgCount := 100
	failedMsgCount := 0
	msgs := testhelper.NewMessages(msgCount)
	for _, msg := range msgs {
		msg.State = types.MessageState(rand.Intn(7))
		if msg.State == types.UnFillMsg {
			msg.Receipt.Return = []byte("gas over limit")
			failedMsgCount++
		}
		assert.NoError(t, messageRepo.CreateMessage(msg))
	}

	msgList, err = messageRepo.ListFailedMessage()
	assert.NoError(t, err)
	assert.Equal(t, failedMsgCount, len(msgList))

	sorted := sort.SliceIsSorted(msgList, func(i, j int) bool {
		return msgList[i].CreatedAt.Before(msgList[j].CreatedAt)
	})
	assert.True(t, sorted)
}

func TestListBlockedMessage(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	msgs := testhelper.NewMessages(3)
	msgs[1].State = types.FillMsg
	assert.NoError(t, messageRepo.CreateMessage(msgs[0]))
	assert.NoError(t, messageRepo.CreateMessage(msgs[1]))

	time.Sleep(5 * time.Second)

	msgList, err := messageRepo.ListBlockedMessage(msgs[0].From, time.Second*2)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(msgList))

	msgList, err = messageRepo.ListBlockedMessage(msgs[1].From, time.Second*2)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(msgList))
}

func TestListUnChainMessageByAddress(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	addr, err := address.NewActorAddress(uuid.New().NodeID())
	assert.NoError(t, err)

	msgList, err := messageRepo.ListUnChainMessageByAddress(addr, 10)
	assert.NoError(t, err)
	assert.Len(t, msgList, 0)

	msgCount := 100
	unChainMsgCount := 0
	msgs := testhelper.NewMessages(msgCount)
	for _, msg := range msgs {
		msg.Message.From = addr
		msg.State = types.MessageState(rand.Intn(7))
		if msg.State == types.UnFillMsg {
			unChainMsgCount++
		}
		assert.NoError(t, messageRepo.CreateMessage(msg))
	}

	msgList, err = messageRepo.ListUnChainMessageByAddress(addr, unChainMsgCount/2)
	assert.NoError(t, err)
	assert.Equal(t, unChainMsgCount/2, len(msgList))

	sorted := sort.SliceIsSorted(msgList, func(i, j int) bool {
		return msgList[i].CreatedAt.Before(msgList[j].CreatedAt)
	})
	assert.True(t, sorted)
}

func TestListFilledMessageByAddress(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	uid, err := uuid.NewUUID()
	assert.NoError(t, err)
	addr, err := address.NewActorAddress(uid[:])
	assert.NoError(t, err)

	msgs, err := messageRepo.ListFilledMessageByAddress(addr)
	assert.NoError(t, err)
	assert.Len(t, msgs, 0)

	count := 10
	signedMsgs := testhelper.NewSignedMessages(count)
	for i, msg := range signedMsgs {
		if i%2 == 0 {
			msg.State = types.FillMsg
		}
		msg.From = addr
		err := messageRepo.CreateMessage(msg)
		assert.NoError(t, err)
	}

	msgs, err = messageRepo.ListFilledMessageByAddress(signedMsgs[0].From)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count/2, len(msgs))
}

func TestListFilledMessageByHeight(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	randHeight := rand.Uint64() / 2
	for _, msg := range testhelper.NewSignedMessages(10) {
		msg.Height = int64(randHeight)
		msg.State = types.FillMsg
		err := messageRepo.CreateMessage(msg)
		assert.NoError(t, err)
	}

	result, err := messageRepo.ListFilledMessageByHeight(abi.ChainEpoch(randHeight))
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, 10, len(result))
}

func TestListUnFilledMessage(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	addr, err := address.NewActorAddress(uuid.New().NodeID())
	assert.NoError(t, err)

	msgList, err := messageRepo.ListUnFilledMessage(addr)
	assert.NoError(t, err)
	assert.Len(t, msgList, 0)

	msgCount := 100
	unFillMsgCount := 0
	msgs := testhelper.NewMessages(msgCount)
	for _, msg := range msgs {
		msg.Message.From = addr
		msg.State = types.MessageState(rand.Intn(7))
		if msg.State == types.UnFillMsg {
			unFillMsgCount++
		}
		assert.NoError(t, messageRepo.CreateMessage(msg))
	}

	msgList, err = messageRepo.ListUnFilledMessage(addr)
	assert.NoError(t, err)
	assert.Equal(t, unFillMsgCount, len(msgList))
}

func TestListFilledMessageBelowNonce(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	addr, err := address.NewActorAddress(uuid.New().NodeID())
	assert.NoError(t, err)

	msgs, err := messageRepo.ListFilledMessageBelowNonce(addr, 10)
	assert.NoError(t, err)
	assert.Len(t, msgs, 0)

	count := 100
	maxNonce := 1000
	aimNonce := uint64(500)
	belowNonceCount := 0
	signedMsgs := testhelper.NewSignedMessages(count)
	for _, msg := range signedMsgs {
		msg.Nonce = uint64(rand.Intn(maxNonce))
		if msg.Nonce%2 == 0 {
			msg.State = types.FillMsg
			msg.From = addr
			if msg.Nonce < aimNonce {
				belowNonceCount++
			}
		}
		err := messageRepo.CreateMessage(msg)
		assert.NoError(t, err)
	}

	msgs, err = messageRepo.ListFilledMessageBelowNonce(addr, aimNonce)
	assert.NoError(t, err)
	assert.Equal(t, belowNonceCount, len(msgs))
}

func TestUpdateMessageInfoByCid(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	msg := testhelper.NewSignedMessages(1)[0]
	unsignedCid := msg.UnsignedCid

	err := messageRepo.CreateMessage(msg)
	assert.NoError(t, err)

	rec := &venustypes.MessageReceipt{
		ExitCode: 0,
		Return:   []byte{'g', 'd'},
		GasUsed:  34,
	}
	tsKeyStr := "{ bafy2bzacec7ymsvmwjgew5whbhs4c3gg5k76pu6y7tun67lqw6unt6xo2nn62 bafy2bzacediq3wdlglhbc6ezlmnks46hdl2kyc3alghiov3c6jpt5qcf76s32 bafy2bzacebjjsg2vqadraxippg46rkysbyucgl27qzu6p6bgepcn7ybgjmqxs }"
	tsKey, err := utils.StringToTipsetKey(tsKeyStr)
	assert.NoError(t, err)

	height := abi.ChainEpoch(10)
	state := types.OnChainMsg
	err = messageRepo.UpdateMessageInfoByCid(unsignedCid.String(), rec, height, state, tsKey)
	assert.NoError(t, err)

	msg2, err := messageRepo.GetMessageByCid(*unsignedCid)
	assert.NoError(t, err)
	assert.Equal(t, int64(height), msg2.Height)
	assert.Equal(t, rec, msg2.Receipt)
	assert.Equal(t, state, msg2.State)
	assert.Equal(t, tsKeyStr, msg2.TipSetKey.String())
}

func TestUpdateMessageStateByCid(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	msg := testhelper.NewSignedMessages(1)[0]
	msg.State = types.FillMsg
	cid := msg.Message.Cid()
	msg.UnsignedCid = &cid

	err := messageRepo.CreateMessage(msg)
	assert.NoError(t, err)

	err = messageRepo.UpdateMessageStateByCid(cid.String(), types.OnChainMsg)
	assert.NoError(t, err)

	msg2, err := messageRepo.GetMessageByUid(msg.ID)
	assert.NoError(t, err)
	assert.Equal(t, types.OnChainMsg, msg2.State)
}

func TestUpdateMessageStateByID(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	msg := testhelper.NewSignedMessages(1)[0]
	msg.State = types.FillMsg
	assert.NoError(t, messageRepo.CreateMessage(msg))

	err := messageRepo.UpdateMessageStateByID(msg.ID, types.OnChainMsg)
	assert.NoError(t, err)

	res, err := messageRepo.GetMessageByUid(msg.ID)
	assert.NoError(t, err)
	assert.Equal(t, types.OnChainMsg, res.State)
}

func TestMarkBadMessage(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	msgs := testhelper.NewMessages(1)
	for _, msg := range msgs {
		err := messageRepo.CreateMessage(msg)
		assert.NoError(t, err)
	}

	err := messageRepo.MarkBadMessage(msgs[0].ID)
	assert.NoError(t, err)

	msg, err := messageRepo.GetMessageByUid(msgs[0].ID)
	assert.NoError(t, err)
	assert.Equal(t, types.FailedMsg, msg.State)
}

func TestUpdateReturnValue(t *testing.T) {
	messageRepo := setupRepo(t).MessageRepo()

	msgs := testhelper.NewMessages(2)
	for _, msg := range msgs {
		err := messageRepo.CreateMessage(msg)
		assert.NoError(t, err)
	}
	failedInfo := "gas estimate failed"
	err := messageRepo.UpdateReturnValue(msgs[0].ID, failedInfo)
	assert.NoError(t, err)
	msg, err := messageRepo.GetMessageByUid(msgs[0].ID)
	assert.NoError(t, err)
	assert.Equal(t, failedInfo, string(msg.Receipt.Return))

	failedMsgs, err := messageRepo.ListFailedMessage()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(failedMsgs), 1)
}

func compareMsg(t *testing.T, expect, actual *types.Message) {
	expect.UpdatedAt = actual.UpdatedAt
	expect.CreatedAt = actual.CreatedAt
	expectStr := testhelper.ObjectToString(expect)
	actualStr := testhelper.ObjectToString(actual)
	assert.Equal(t, expectStr, actualStr)
}