package listener

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"staking-interaction/common/config"
	"staking-interaction/common/redis"
	"staking-interaction/model"
	"staking-interaction/repository"
	"staking-interaction/utils"
	"strconv"
	"sync/atomic"
	"time"
)

type SyncWithdrawHandler struct {
	client                   *ethclient.Client
	isWithDrawHandlerRunning int32
	lockManager              *redis.LockManager
}

var withdrawConf = config.Get().BlockchainConfig

func NewSyncWithdrawHandler(client *ethclient.Client, lockManager *redis.LockManager) *SyncWithdrawHandler {
	return &SyncWithdrawHandler{
		client:      client,
		lockManager: lockManager,
	}
}

func (s *SyncWithdrawHandler) Start() {
	atomic.StoreInt32(&s.isWithDrawHandlerRunning, 1)
	s.syncWithdraw()
}

func (s *SyncWithdrawHandler) Stop() {
	atomic.StoreInt32(&s.isWithDrawHandlerRunning, 0)
	fmt.Println("stop SyncWithdrawHandler...")
}

func (s *SyncWithdrawHandler) syncWithdraw() {
	for atomic.LoadInt32(&s.isWithDrawHandlerRunning) == 1 {
		withDrawList, err := repository.GetWithdrawalInfoByStatus(config.WithdrawStatusPending)
		if err != nil {
			fmt.Printf("SyncWithdrawHandler: GetWithdrawInfo failed: %v\n", err)
			continue
		}
		for _, withdrawInfo := range withDrawList {
			err := s.processWithdraw(withdrawInfo)
			if err != nil {
				fmt.Printf("SyncWithdrawHandler: processWithdraw failed: %v\n", err)
			}
			fmt.Println("success processWithdraw")
		}
		time.Sleep(5 * time.Second)
	}
}

func (s *SyncWithdrawHandler) processWithdraw(withdrawInfo model.Withdrawal) error {
	fmt.Println("withdrawInfo.Hash:", withdrawInfo.Hash)
	hash := common.HexToHash(withdrawInfo.Hash)
	ctx, cancel := context.WithTimeout(context.Background(), withdrawConf.Sync.SyncInterval)
	defer cancel()

	receipt, err := s.client.TransactionReceipt(ctx, hash)
	if err != nil {
		return fmt.Errorf("SyncWithdrawHandler: GetTransactionReceipt failed: %v\n", err)

	}
	tx, isPending, err := s.client.TransactionByHash(ctx, hash)
	if err != nil || isPending {
		return fmt.Errorf("SyncWithdrawHandler: GetTransactionByHash failed: %v\n", err)
	}

	//更新withdraw信息
	gasUsed := utils.Uint64ToBigInt(receipt.GasUsed)
	gasPrice := tx.GasPrice()
	fee := new(big.Int).Mul(gasPrice, gasUsed) // 手续费 = gasPrice × gasUsed
	value, err := utils.StringToBigInt(withdrawInfo.Value)
	if err != nil {
		return fmt.Errorf("SyncWithdrawHandler: StringToBigInt failed: %v\n", err)
	}
	amount := new(big.Int).Add(fee, value)
	withdrawInfo.Fee = fee.String()
	withdrawInfo.GasPrice = gasPrice.String()
	withdrawInfo.Amount = amount.String()

	return s.executeWithdrawWithLock(ctx, withdrawInfo, receipt, amount)
}

func (s *SyncWithdrawHandler) executeWithdrawWithLock(ctx context.Context, withdrawInfo model.Withdrawal, receipt *types.Receipt, amount *big.Int) error {
	//获取锁
	account, err := repository.GetAccount(withdrawInfo.WalletAddress)
	if err != nil {
		return fmt.Errorf("get wallet account failed: %w", err)
	}
	assetLock, err := s.lockManager.AcquireAssetLock(context.Background(), account.AccountID, withdrawInfo.TokenType)
	if err != nil {
		return fmt.Errorf("acquire assetLock failed: %w ,accountid:%d, tx_hash:%s", err, account.AccountID, withdrawInfo.Hash)
	}
	withdrawLock, err := s.lockManager.AcquireWithdrawLock(context.Background(), withdrawInfo.ID)
	if err != nil {
		return fmt.Errorf("acquire assetLock failed: %w ,withdrawid:%d, tx_hash:%s", err, withdrawInfo.ID, withdrawInfo.Hash)
	}

	//释放锁
	defer func() {
		unlockCtx, unlockCancel := context.WithTimeout(context.Background(), withdrawConf.Sync.RetryDelay)
		defer unlockCancel()

		if err := assetLock.Unlock(unlockCtx); err != nil {
			fmt.Printf("unlock assetLock failed: %w ,accountid:%d, tx_hash:%s\n", err, account.AccountID, withdrawInfo.Hash)
		}
		if err := withdrawLock.Unlock(unlockCtx); err != nil {
			fmt.Printf("unlock withdrawLock failed: %w ,withdrawid:%d, tx_hash:%s\n", err, withdrawInfo.ID, withdrawInfo.Hash)
		}
		fmt.Printf("lock  releasing: blocknumber:%s, tx_hash:%s\n", receipt.BlockNumber.String(), withdrawInfo.Hash)
	}()

	return repository.SwWithTransaction(func(wd *repository.SwRepo) error {
		if receipt.Status != types.ReceiptStatusSuccessful {
			withdrawInfo.Status = config.WithdrawStatusFailed
			if err := wd.UpdateWithdrawalInfo(withdrawInfo); err != nil {
				fmt.Printf("SyncWithdrawHandler: UpdateWithdrawInfo failed: %v\n", err)
			}
			return fmt.Errorf("SyncWithdrawHandler: GetTransactionReceipt failed id: %d, hash:%s\n", withdrawInfo.ID, withdrawInfo.Hash)
		}

		withdrawInfo.Status = config.WithdrawStatusSuccess
		if err := wd.UpdateWithdrawalInfo(withdrawInfo); err != nil {
			return fmt.Errorf("SyncWithdrawHandler: UpdateWithdrawInfo failed: %v, withdrawid:%d\n", err, withdrawInfo.ID)
		}

		blockNumber := receipt.BlockNumber
		currentHeight, err := s.client.BlockNumber(ctx)
		if err != nil {
			return fmt.Errorf("SyncWithdrawHandler: GetBlockNumber failed: %v\n", err)
		}
		expectBlockNumber := new(big.Int).Add(blockNumber, big.NewInt(int64(withdrawConf.Sync.BlockBuffer)))
		if big.NewInt(int64(currentHeight)).Cmp(expectBlockNumber) < 0 { //current height< block number+30
			return fmt.Errorf("SyncWithdrawHandler: current block height is not enough, currentHeight: %d, receipt block number:%d \n", currentHeight, blockNumber)
		}
		// add bill info
		asset, err := wd.GetAssetByAccountIdWithLock(account.AccountID)
		if err != nil {
			return fmt.Errorf("SyncWithdrawHandler: GetAccountAsset failed: %v, accountid:%d\n", err, account.AccountID)
		}
		var pre string
		switch withdrawInfo.TokenType {
		case config.TokenTypeBNB:
			pre = asset.BnbBalance
		case config.TokenTypeMTK:
			pre = asset.MtkBalance
		}
		preBalance, err := utils.StringToBigInt(pre)
		if err != nil {
			return fmt.Errorf("SyncWithdrawHandler: preBalance parse failed: %v, preBalance:%s\n", err, pre)
		}
		if preBalance.Cmp(amount) == -1 {
			return fmt.Errorf("SyncWithdrawHandler: balance is not enough, preBalance:%s, txamount:%s\n", preBalance, withdrawInfo.Amount)
		}
		nextBalance := new(big.Int).Sub(preBalance, amount)
		bill := model.Bill{
			AccountID:   account.AccountID,
			TokenType:   withdrawInfo.TokenType,
			BillType:    config.BillTypeWithdrawal,
			Amount:      amount.String(),
			Fee:         strconv.FormatUint(receipt.GasUsed, 10),
			PreBalance:  pre,
			NextBalance: nextBalance.String(),
		}
		if err := wd.AddBill(&bill); err != nil {
			return fmt.Errorf("AddBill failed: %w", err)
		}
		fmt.Println("update bill success")

		if err := wd.UpdateAssetWithOptimisticLock(asset, nextBalance.String(), withdrawInfo.TokenType); err != nil {
			return fmt.Errorf("update asset: %w", err)
		}
		return nil
	})
}
