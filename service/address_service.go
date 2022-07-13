package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus-messager/gateway"
	"github.com/filecoin-project/venus-messager/log"
	"github.com/filecoin-project/venus-messager/models/repo"
	v1 "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	venusTypes "github.com/filecoin-project/venus/venus-shared/types"
	types "github.com/filecoin-project/venus/venus-shared/types/messager"
)

var errAddressNotExists = errors.New("address not exists")

type AddressService struct {
	repo repo.Repo
	log  *log.Logger

	sps          *SharedParamsService
	nodeClient   v1.FullNode
	walletClient *gateway.IWalletCli
}

func NewAddressService(repo repo.Repo,
	logger *log.Logger,
	sps *SharedParamsService,
	walletClient *gateway.IWalletCli,
	nodeClient v1.FullNode) *AddressService {
	addressService := &AddressService{
		repo: repo,
		log:  logger,

		sps:          sps,
		nodeClient:   nodeClient,
		walletClient: walletClient,
	}

	return addressService
}

func (addressService *AddressService) SaveAddress(ctx context.Context, address *types.Address) (venusTypes.UUID, error) {
	err := addressService.repo.Transaction(func(txRepo repo.TxRepo) error {
		has, err := txRepo.AddressRepo().HasAddress(ctx, address.Addr)
		if err != nil {
			return err
		}
		if has {
			return fmt.Errorf("address already exists")
		}
		return txRepo.AddressRepo().SaveAddress(ctx, address)
	})

	return address.ID, err
}

func (addressService *AddressService) UpdateNonce(ctx context.Context, addr address.Address, nonce uint64) error {
	return addressService.repo.AddressRepo().UpdateNonce(ctx, addr, nonce)
}

func (addressService *AddressService) GetAddress(ctx context.Context, addr address.Address) (*types.Address, error) {
	return addressService.repo.AddressRepo().GetAddress(ctx, addr)
}

func (addressService *AddressService) WalletHas(ctx context.Context, account string, addr address.Address) (bool, error) {
	return addressService.walletClient.WalletHas(ctx, account, addr)
}

func (addressService *AddressService) HasAddress(ctx context.Context, addr address.Address) (bool, error) {
	return addressService.repo.AddressRepo().HasAddress(ctx, addr)
}

func (addressService *AddressService) ListAddress(ctx context.Context) ([]*types.Address, error) {
	return addressService.repo.AddressRepo().ListAddress(ctx)
}

func (addressService *AddressService) ListActiveAddress(ctx context.Context) ([]*types.Address, error) {
	return addressService.repo.AddressRepo().ListActiveAddress(ctx)
}

func (addressService *AddressService) DeleteAddress(ctx context.Context, addr address.Address) error {
	return addressService.repo.AddressRepo().DelAddress(ctx, addr)
}

func (addressService *AddressService) ForbiddenAddress(ctx context.Context, addr address.Address) error {
	if err := addressService.repo.AddressRepo().UpdateState(ctx, addr, types.AddressStateForbbiden); err != nil {
		return err
	}
	addressService.log.Infof("forbidden address %v success", addr.String())

	return nil
}

func (addressService *AddressService) ActiveAddress(ctx context.Context, addr address.Address) error {
	if err := addressService.repo.AddressRepo().UpdateState(ctx, addr, types.AddressStateAlive); err != nil {
		return err
	}
	addressService.log.Infof("active address %v success", addr.String())

	return nil
}

func (addressService *AddressService) SetSelectMsgNum(ctx context.Context, addr address.Address, num uint64) error {
	if err := addressService.repo.AddressRepo().UpdateSelectMsgNum(ctx, addr, num); err != nil {
		return err
	}
	addressService.log.Infof("set select msg num: %s %d", addr.String(), num)

	return nil
}

func (addressService *AddressService) SetFeeParams(ctx context.Context, addr address.Address, gasOverEstimation, gasOverPremium float64, maxFeeStr, maxFeeCapStr string) error {
	has, err := addressService.repo.AddressRepo().HasAddress(ctx, addr)
	if err != nil {
		return err
	}
	if !has {
		return errAddressNotExists
	}

	var needUpdate bool
	var maxFee, maxFeeCap big.Int
	if len(maxFeeStr) != 0 {
		maxFee, err = venusTypes.BigFromString(maxFeeStr)
		if err != nil {
			return fmt.Errorf("parsing max-spend: %v", err)
		}
		needUpdate = true
	}
	if len(maxFeeCapStr) != 0 {
		maxFeeCap, err = venusTypes.BigFromString(maxFeeCapStr)
		if err != nil {
			return fmt.Errorf("parsing max-feecap: %v", err)
		}
		needUpdate = true
	}
	if gasOverEstimation > 0 {
		needUpdate = true
	}
	if gasOverPremium > 0 {
		needUpdate = true
	}
	if !needUpdate {
		return nil
	}

	return addressService.repo.AddressRepo().UpdateFeeParams(ctx, addr, gasOverEstimation, gasOverPremium, maxFee, maxFeeCap)
}

func (addressService *AddressService) ActiveAddresses() map[address.Address]struct{} {
	addrs := make(map[address.Address]struct{})
	addrList, err := addressService.ListActiveAddress(context.Background())
	if err != nil {
		addressService.log.Errorf("list address %v", err)
		return addrs
	}

	for _, addr := range addrList {
		addrs[addr.Addr] = struct{}{}
	}

	return addrs
}
