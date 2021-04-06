package mysql

import (
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/crypto"
	venustypes "github.com/filecoin-project/venus/pkg/types"
	"github.com/ipfs/go-cid"
	"gorm.io/gorm"

	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/ipfs-force-community/venus-messager/types"
	"github.com/ipfs-force-community/venus-messager/utils"
)

type mysqlMessage struct {
	ID      string `gorm:"column:id;type:varchar(256);primary_key"`
	Version uint64 `gorm:"column:version;type:bigint unsigned"`

	From  string `gorm:"column:from_addr;type:varchar(256);NOT NULL;index:msg_from;index:idx_from_nonce;index:msg_from_state;"`
	Nonce uint64 `gorm:"column:nonce;type:bigint unsigned;index:msg_nonce;index:idx_from_nonce"`
	To    string `gorm:"column:to;type:varchar(256);NOT NULL"`

	Value types.Int `gorm:"column:value;type:varchar(256);"`

	GasLimit   int64     `gorm:"column:gas_limit;type:bigint"`
	GasFeeCap  types.Int `gorm:"column:gas_fee_cap;type:varchar(256);"`
	GasPremium types.Int `gorm:"column:gas_premium;type:varchar(256);"`

	Method int `gorm:"column:method;type:int"`

	Params []byte `gorm:"column:params;type:blob;"`

	Signature *repo.SqlSignature `gorm:"column:signed_data;type:blob;"`

	UnsignedCid string `gorm:"column:unsigned_cid;type:varchar(256);index:msg_unsigned_cid;"`
	SignedCid   string `gorm:"column:signed_cid;type:varchar(256);index:msg_signed_cid"`

	Height    int64               `gorm:"column:height;type:bigint;index:msg_height"`
	Receipt   *repo.SqlMsgReceipt `gorm:"embedded;embeddedPrefix:receipt_"`
	TipsetKey string              `gorm:"column:tipset_key;type:varchar(1024);"`

	Meta *MsgMeta `gorm:"embedded;embeddedPrefix:meta_"`

	WalletName string `gorm:"column:wallet_name;type:varchar(256)"`

	State types.MessageState `gorm:"column:state;type:int;index:msg_state;index:msg_from_state;"`

	IsDeleted int       `gorm:"column:is_deleted;index;default:-1;NOT NULL"` // 是否删除 1:是  -1:否
	CreatedAt time.Time `gorm:"column:created_at;index;NOT NULL"`            // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;index;NOT NULL"`            // 更新时间
}

func (sqlMsg *mysqlMessage) TableName() string {
	return "messages"
}

func (sqlMsg *mysqlMessage) Message() *types.Message {
	var destMsg = &types.Message{
		ID: sqlMsg.ID,
		UnsignedMessage: venustypes.UnsignedMessage{
			Version:    sqlMsg.Version,
			Nonce:      sqlMsg.Nonce,
			Value:      big.NewFromGo(sqlMsg.Value.Int),
			GasLimit:   sqlMsg.GasLimit,
			GasFeeCap:  big.NewFromGo(sqlMsg.GasFeeCap.Int),
			GasPremium: big.NewFromGo(sqlMsg.GasPremium.Int),
			Method:     abi.MethodNum(sqlMsg.Method),
			Params:     sqlMsg.Params,
		},
		Height:     sqlMsg.Height,
		Receipt:    sqlMsg.Receipt.MsgReceipt(),
		Signature:  (*crypto.Signature)(sqlMsg.Signature),
		Meta:       sqlMsg.Meta.Meta(),
		WalletName: sqlMsg.WalletName,
		State:      sqlMsg.State,
	}
	destMsg.From, _ = address.NewFromString(sqlMsg.From)
	destMsg.To, _ = address.NewFromString(sqlMsg.To)
	if len(sqlMsg.UnsignedCid) > 0 {
		unsignedCid, _ := cid.Decode(sqlMsg.UnsignedCid)
		destMsg.UnsignedCid = &unsignedCid
	}
	if len(sqlMsg.SignedCid) > 0 {
		signedCid, _ := cid.Decode(sqlMsg.SignedCid)
		destMsg.SignedCid = &signedCid
	}
	if len(sqlMsg.TipsetKey) > 0 {
		destMsg.TipSetKey, _ = utils.StringToTipsetKey(sqlMsg.TipsetKey)
	}

	return destMsg
}

func FromMessage(srcMsg *types.Message) *mysqlMessage {
	destMsg := &mysqlMessage{
		ID:         srcMsg.ID,
		Version:    srcMsg.Version,
		To:         srcMsg.To.String(),
		From:       srcMsg.From.String(),
		Nonce:      srcMsg.Nonce,
		GasLimit:   srcMsg.GasLimit,
		Method:     int(srcMsg.Method),
		Params:     srcMsg.Params,
		Signature:  (*repo.SqlSignature)(srcMsg.Signature),
		Height:     srcMsg.Height,
		Receipt:    repo.FromMsgReceipt(srcMsg.Receipt),
		Meta:       FromMeta(srcMsg.Meta),
		WalletName: srcMsg.WalletName,
		State:      srcMsg.State,
		IsDeleted:  -1,
	}

	if srcMsg.UnsignedCid != nil {
		destMsg.UnsignedCid = srcMsg.UnsignedCid.String()
	}

	if srcMsg.SignedCid != nil {
		destMsg.SignedCid = srcMsg.SignedCid.String()
	}

	if srcMsg.Value.Int != nil {
		destMsg.Value = types.Int{Int: srcMsg.Value.Int}
	}

	if srcMsg.GasFeeCap.Int != nil {
		destMsg.GasFeeCap = types.Int{Int: srcMsg.GasFeeCap.Int}
	}

	if srcMsg.GasPremium.Int != nil {
		destMsg.GasPremium = types.Int{Int: srcMsg.GasPremium.Int}
	}

	if !srcMsg.TipSetKey.IsEmpty() {
		destMsg.TipsetKey = srcMsg.TipSetKey.String()
	}

	return destMsg
}

type MsgMeta struct {
	ExpireEpoch       abi.ChainEpoch `gorm:"column:expire_epoch;type:bigint;"`
	GasOverEstimation float64        `gorm:"column:gas_over_estimation;type:decimal(10,2);"`
	MaxFee            types.Int      `gorm:"column:max_fee;type:varchar(256);"`
	MaxFeeCap         types.Int      `gorm:"column:max_fee_cap;type:varchar(256);"`
}

func (meta *MsgMeta) Meta() *types.MsgMeta {
	return &types.MsgMeta{
		ExpireEpoch:       meta.ExpireEpoch,
		GasOverEstimation: meta.GasOverEstimation,
		MaxFee:            big.NewFromGo(meta.MaxFee.Int),
		MaxFeeCap:         big.NewFromGo(meta.MaxFeeCap.Int),
	}
}

func FromMeta(srcMeta *types.MsgMeta) *MsgMeta {
	if srcMeta == nil {
		return &MsgMeta{
			ExpireEpoch:       0,
			GasOverEstimation: 0,
			MaxFee:            types.Int{},
			MaxFeeCap:         types.Int{},
		}
	}
	meta := &MsgMeta{
		ExpireEpoch:       srcMeta.ExpireEpoch,
		GasOverEstimation: srcMeta.GasOverEstimation,
	}

	if srcMeta.MaxFee.Int != nil {
		meta.MaxFee = types.Int{Int: srcMeta.MaxFee.Int}
	}

	if srcMeta.MaxFeeCap.Int != nil {
		meta.MaxFeeCap = types.Int{Int: srcMeta.MaxFeeCap.Int}
	}
	return meta
}

var _ repo.MessageRepo = (*mysqlMessageRepo)(nil)

type mysqlMessageRepo struct {
	*gorm.DB
}

func newMysqlMessageRepo(db *gorm.DB) *mysqlMessageRepo {
	return &mysqlMessageRepo{DB: db}
}

func (m *mysqlMessageRepo) GetMessageState(id string) (types.MessageState, error) {
	type Result struct {
		State int
	}

	var result Result
	err := m.DB.Table("messages").
		Select("state").
		Where("id = ?", id).
		Scan(&result).Error
	if err != nil {
		return types.UnKnown, err
	}

	return types.MessageState(result.State), nil
}

func (m *mysqlMessageRepo) ExpireMessage(msgs []*types.Message) error {
	for _, msg := range msgs {
		err := m.DB.Table("messages").Where("id=?", msg.ID).UpdateColumn("state", types.ExpireMsg).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *mysqlMessageRepo) ListFilledMessageByAddress(addr address.Address) ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	err := m.DB.Find(&sqlMsgs, "from_addr=? AND state=?", addr.String(), types.FillMsg).Error
	if err != nil {
		return nil, err
	}
	result := make([]*types.Message, len(sqlMsgs))
	for index, sqlMsg := range sqlMsgs {
		result[index] = sqlMsg.Message()
	}
	return result, nil
}

func (m *mysqlMessageRepo) ListFilledMessageBelowNonce(addr address.Address, nonce uint64) ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	err := m.DB.Find(&sqlMsgs, "from_addr=? AND state=? AND nonce <", addr.String(), types.FillMsg, nonce).Error
	if err != nil {
		return nil, err
	}
	result := make([]*types.Message, len(sqlMsgs))
	for index, sqlMsg := range sqlMsgs {
		result[index] = sqlMsg.Message()
	}
	return result, nil
}

func (m *mysqlMessageRepo) ListFilledMessageByHeight(height abi.ChainEpoch) ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	err := m.DB.Find(&sqlMsgs, "height=? AND state=?", height, types.FillMsg).Error
	if err != nil {
		return nil, err
	}
	result := make([]*types.Message, len(sqlMsgs))
	for index, sqlMsg := range sqlMsgs {
		result[index] = sqlMsg.Message()
	}
	return result, nil
}

func (m *mysqlMessageRepo) ListUnChainMessageByAddress(addr address.Address) ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	err := m.DB.Find(&sqlMsgs, "from_addr=? AND state=?", addr.String(), types.UnFillMsg).Order("created_at").Error
	if err != nil {
		return nil, err
	}
	result := make([]*types.Message, len(sqlMsgs))
	for index, sqlMsg := range sqlMsgs {
		result[index] = sqlMsg.Message()
	}
	return result, nil
}

//todo better batch update
func (m *mysqlMessageRepo) BatchSaveMessage(msgs []*types.Message) error {
	for _, msg := range msgs {
		err := m.SaveMessage(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *mysqlMessageRepo) CreateMessage(msg *types.Message) error {
	sqlMsg := FromMessage(msg)
	sqlMsg.CreatedAt = time.Now()
	sqlMsg.UpdatedAt = time.Now()
	return m.DB.Create(sqlMsg).Error
}

func (m *mysqlMessageRepo) SaveMessage(msg *types.Message) error {
	sqlMsg := FromMessage(msg)
	sqlMsg.UpdatedAt = time.Now()
	return m.DB.Omit("created_at").Save(sqlMsg).Error
}

func (m *mysqlMessageRepo) GetMessageByUid(id string) (*types.Message, error) {
	var msg mysqlMessage
	if err := m.DB.Where("id = ?", id).First(&msg).Error; err != nil {
		return nil, err
	}
	return msg.Message(), nil
}

func (m *mysqlMessageRepo) GetMessageByCid(unsignedCid cid.Cid) (*types.Message, error) {
	var msg mysqlMessage
	if err := m.DB.Where("unsigned_cid = ?", unsignedCid.String()).First(&msg).Error; err != nil {
		return nil, err
	}
	return msg.Message(), nil
}

func (m *mysqlMessageRepo) GetMessageBySignedCid(signedCid cid.Cid) (*types.Message, error) {
	var msg mysqlMessage
	if err := m.DB.Where("signed_cid = ?", signedCid.String()).First(&msg).Error; err != nil {
		return nil, err
	}
	return msg.Message(), nil
}

func (m *mysqlMessageRepo) GetSignedMessageByTime(start time.Time) ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	if err := m.DB.Where("created_at >= ? and signed_data is not null", start).Find(&sqlMsgs).Error; err != nil {
		return nil, err
	}
	result := make([]*types.Message, len(sqlMsgs))
	for idx, msg := range sqlMsgs {
		result[idx] = msg.Message()
	}

	return result, nil
}

func (m *mysqlMessageRepo) GetSignedMessageByHeight(height abi.ChainEpoch) ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	if err := m.DB.Where("height >= ? and signed_data is not null", uint64(height)).Find(&sqlMsgs).Error; err != nil {
		return nil, err
	}
	result := make([]*types.Message, len(sqlMsgs))
	for idx, msg := range sqlMsgs {
		result[idx] = msg.Message()
	}

	return result, nil
}

func (m *mysqlMessageRepo) GetMessageByFromAndNonce(from address.Address, nonce uint64) (*types.Message, error) {
	var msg mysqlMessage
	if err := m.DB.Where("from_addr = ? and nonce = ?", from.String(), nonce).First(&msg).Error; err != nil {
		return nil, err
	}
	return msg.Message(), nil
}

func (m *mysqlMessageRepo) ListMessage() ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	if err := m.DB.Find(&sqlMsgs).Error; err != nil {
		return nil, err
	}

	result := make([]*types.Message, len(sqlMsgs))
	for idx, msg := range sqlMsgs {
		result[idx] = msg.Message()
	}
	return result, nil
}

func (m *mysqlMessageRepo) ListUnchainedMsgs() ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	if err := m.DB.Model((*mysqlMessage)(nil)).
		Where("height=0 and signed_data is null").
		Find(&sqlMsgs).Error; err != nil {
		return nil, err
	}

	var result = make([]*types.Message, len(sqlMsgs))

	for idx, msg := range sqlMsgs {
		result[idx] = msg.Message()
	}
	return result, nil
}

func (m *mysqlMessageRepo) ListSignedMsgs() ([]*types.Message, error) {
	var sqlMsgs []*mysqlMessage
	if err := m.DB.Model((*mysqlMessage)(nil)).
		Where("height=0 and signed_data is not null").
		Find(&sqlMsgs).Error; err != nil {
		return nil, err
	}

	var result = make([]*types.Message, len(sqlMsgs))

	for idx, msg := range sqlMsgs {
		result[idx] = msg.Message()
	}
	return result, nil
}

func (m *mysqlMessageRepo) UpdateMessageInfoByCid(unsignedCid string,
	receipt *venustypes.MessageReceipt,
	height abi.ChainEpoch,
	state types.MessageState,
	tsKey venustypes.TipSetKey) error {
	rcp := repo.FromMsgReceipt(receipt)
	updateClause := map[string]interface{}{
		"height":               uint64(height),
		"receipt_exit_code":    rcp.ExitCode,
		"receipt_return_value": rcp.ReturnValue,
		"receipt_gas_used":     rcp.GasUsed,
		"state":                state,
		"tipset_key":           tsKey.String(),
	}
	return m.DB.Model(&mysqlMessage{}).
		Where("unsigned_cid = ?", unsignedCid).
		UpdateColumns(updateClause).Error
}

func (m *mysqlMessageRepo) UpdateMessageStateByCid(cid string, state types.MessageState) error {
	return m.DB.Model(&mysqlMessage{}).
		Where("unsigned_cid = ?", cid).UpdateColumn("state", state).Error
}

func (m *mysqlMessageRepo) UpdateMessageStateByID(id string, state types.MessageState) error {
	return m.DB.Debug().Model(&mysqlMessage{}).
		Where("id = ?", id).UpdateColumn("state", state).Error
}

func (m *mysqlMessageRepo) UpdateUnFilledMessageStateByAddress(addr address.Address, state types.MessageState) error {
	return m.DB.Debug().Model(&mysqlMessage{}).Where("from_addr = ? and state = ?", addr.String(), types.UnFillMsg).
		UpdateColumn("state", state).Error
}
