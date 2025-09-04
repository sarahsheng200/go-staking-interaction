package listener

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	constant "staking-interaction/common"
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
}

func NewSyncWithdrawHandler(client *ethclient.Client) *SyncWithdrawHandler {
	return &SyncWithdrawHandler{
		client: client,
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
		withDrawList, err := repository.GetWithdrawalInfoByStatus(constant.WithdrawStatusPending)
		if err != nil {
			fmt.Printf("SyncWithdrawHandler: GetWithdrawInfo failed: %v\n", err)
			continue
		}
		for _, withdrawInfo := range withDrawList {
			s.processWithdraw(withdrawInfo)
			fmt.Println("success processWithdraw")
		}
		time.Sleep(5 * time.Second)
	}
}

func (s *SyncWithdrawHandler) processWithdraw(withdrawInfo model.Withdrawal) {
	fmt.Println("withdrawInfo.Hash:", withdrawInfo.Hash)
	hash := common.HexToHash(withdrawInfo.Hash)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	receipt, err := s.client.TransactionReceipt(ctx, hash)
	if err != nil {
		fmt.Printf("SyncWithdrawHandler: GetTransactionReceipt failed: %v\n", err)
		return
	}
	tx, isPending, err := s.client.TransactionByHash(ctx, hash)
	if err != nil || isPending {
		fmt.Printf("SyncWithdrawHandler: GetTransactionByHash failed: %v\n", err)
		return
	}
	//更新withdraw信息
	gasUsed := utils.Uint64ToBigInt(receipt.GasUsed)
	gasPrice := tx.GasPrice()
	fee := new(big.Int).Mul(gasPrice, gasUsed) // 手续费 = gasPrice × gasUsed
	value, err := utils.StringToBigInt(withdrawInfo.Value)
	if err != nil {
		fmt.Printf("SyncWithdrawHandler: StringToBigInt failed: %v\n", err)
		return
	}
	amount := new(big.Int).Add(fee, value)
	withdrawInfo.Fee = fee.String()
	withdrawInfo.GasPrice = gasPrice.String()
	withdrawInfo.Amount = amount.String()

	e := repository.WdWithTransaction(func(wd *repository.WdRepo) error {
		if receipt.Status != types.ReceiptStatusSuccessful {
			withdrawInfo.Status = constant.WithdrawStatusFailed
			if err := wd.UpdateWithdrawalInfo(withdrawInfo); err != nil {
				fmt.Printf("SyncWithdrawHandler: UpdateWithdrawInfo failed: %v\n", err)
			}
			return fmt.Errorf("SyncWithdrawHandler: GetTransactionReceipt failed id: %d, hash:%s\n", withdrawInfo.ID, withdrawInfo.Hash)
		}

		withdrawInfo.Status = constant.WithdrawStatusSuccess
		if err := wd.UpdateWithdrawalInfo(withdrawInfo); err != nil {
			return fmt.Errorf("SyncWithdrawHandler: UpdateWithdrawInfo failed: %v, withdrawid:%d\n", err, withdrawInfo.ID)
		}

		blockNumber := receipt.BlockNumber
		currentHeight, err := s.client.BlockNumber(ctx)
		if err != nil {
			return fmt.Errorf("SyncWithdrawHandler: GetBlockNumber failed: %v\n", err)
		}
		expectBlockNumber := new(big.Int).Add(blockNumber, big.NewInt(30))
		if big.NewInt(int64(currentHeight)).Cmp(expectBlockNumber) < 0 { //current height< block number+30
			return fmt.Errorf("SyncWithdrawHandler: current block height is not enough, currentHeight: %d, receipt block number:%d \n", currentHeight, blockNumber)
		}
		// add bill info
		account := wd.GetAccount(withdrawInfo.WalletAddress)
		asset, err := wd.GetAccountAsset(account.AccountID)
		if err != nil {
			return fmt.Errorf("SyncWithdrawHandler: GetAccountAsset failed: %v, accountid:%d\n", err, account.AccountID)
		}
		var pre string
		switch withdrawInfo.TokenType {
		case constant.TokenTypeBNB:
			pre = asset.BnbBalance
		case constant.TokenTypeMTK:
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
			BillType:    constant.BillTypeWithdrawal,
			Amount:      amount.String(),
			Fee:         strconv.FormatUint(receipt.GasUsed, 10),
			PreBalance:  pre,
			NextBalance: nextBalance.String(),
		}
		if err := wd.AddBill(&bill); err != nil {
			return fmt.Errorf("AddBill failed: %w", err)
		}
		fmt.Println("update bill success")

		switch withdrawInfo.TokenType {
		case constant.TokenTypeBNB:
			asset.BnbBalance = nextBalance.String()
		case constant.TokenTypeMTK:
			asset.MtkBalance = nextBalance.String()
		}
		if err := wd.UpdateAsset(asset); err != nil {
			return fmt.Errorf("UpdateAsset failed: %v", err)
		}
		return nil
	})
	if e != nil {
		fmt.Printf("SyncWithdrawHandler: transaction failed: %v\n", err)
		return
	}
}
