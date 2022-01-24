package models

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	shared "github.com/filecoin-project/venus/venus-shared/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/venus-messager/config"
	"github.com/filecoin-project/venus-messager/models/repo"
	"github.com/filecoin-project/venus-messager/models/sqlite"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

func NewSignedMessages(count int) []*types.Message {
	msgs := make([]*types.Message, 0, count)
	for i := 0; i < count; i++ {
		msg := NewMessage()
		msg.Nonce = uint64(i)
		msg.Signature = &crypto.Signature{Type: crypto.SigTypeSecp256k1, Data: []byte(uuid.New().String())}
		unsignedCid := msg.Message.Cid()
		msg.UnsignedCid = &unsignedCid
		signedCid := (&shared.SignedMessage{
			Message:   msg.Message,
			Signature: *msg.Signature,
		}).Cid()
		msg.SignedCid = &signedCid
		msgs = append(msgs, msg)
	}

	return msgs
}

func NewMessages(count int) []*types.Message {
	msgs := make([]*types.Message, count)
	for i := 0; i < count; i++ {
		msgs[i] = NewMessage()
	}

	return msgs
}

func NewMessage() *types.Message {
	return &types.Message{
		ID:      shared.NewUUID().String(),
		Message: NewUnsignedMessage(),
		Meta: &types.SendSpec{
			ExpireEpoch:       100,
			MaxFee:            big.NewInt(10),
			GasOverEstimation: 0.5,
		},
		Receipt: &shared.MessageReceipt{ExitCode: -1},
		State:   types.UnFillMsg,
	}
}

func NewUnsignedMessage() shared.Message {
	rand.Seed(time.Now().Unix())
	uid, _ := uuid.NewUUID()
	from, _ := address.NewActorAddress(uid[:])
	uid, _ = uuid.NewUUID()
	to, _ := address.NewActorAddress(uid[:])
	return shared.Message{
		From:       from,
		To:         to,
		Value:      big.NewInt(rand.Int63n(1024)),
		GasLimit:   rand.Int63n(100),
		GasFeeCap:  abi.NewTokenAmount(2000),
		GasPremium: abi.NewTokenAmount(1024),
	}
}

func ObjectToString(i interface{}) string {
	res, _ := json.MarshalIndent(i, "", " ")
	return string(res)
}

func setupRepo(t *testing.T) (repo.Repo, repo.Repo) {
	sqliteRepo, err := sqlite.OpenSqlite(&config.SqliteConfig{File: "./test_sqlite_db", Debug: true})
	assert.NoError(t, err)

	//mysqlRepo, err := mysql.OpenMysql(&config.MySqlConfig{
	//	ConnectionString: "root:Root1234@(localhost:3306)/messager?parseTime=true&loc=Local",
	//	MaxOpenConn:      1,
	//	MaxIdleConn:      1,
	//	ConnMaxLifeTime:  time.Second * 1,
	//	Debug:            true,
	//})
	assert.NoError(t, err)
	assert.NoError(t, sqliteRepo.AutoMigrate())
	//assert.NoError(t, mysqlRepo.AutoMigrate())
	return sqliteRepo, nil
}
