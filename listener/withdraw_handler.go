package listener

import (
	"fmt"
	constant "staking-interaction/common"
	"staking-interaction/model"
	"staking-interaction/repository"
	"staking-interaction/service"
	"staking-interaction/utils"
	"sync/atomic"
	"time"
)

type WithdrawHandler struct {
	txService                *service.TransactionService
	isWithDrawHandlerRunning int32 // 原子操作控制同步状态
}

func NewWithdrawHandler(txService *service.TransactionService) *WithdrawHandler {
	return &WithdrawHandler{
		txService: txService,
	}
}

func (w *WithdrawHandler) Start() {
	atomic.StoreInt32(&w.isWithDrawHandlerRunning, 1)

	for atomic.LoadInt32(&w.isWithDrawHandlerRunning) == 1 {
		w.processWithdrawals()
		time.Sleep(5 * time.Second)
	}
}

func (w *WithdrawHandler) processWithdrawals() {
	withDrawList, err := repository.GetWithdrawalInfoByStatus(constant.WithdrawStatusInit)
	if err != nil {
		fmt.Printf("get init withdraw info: %w", err)
		return
	}

	for i, withdraw := range withDrawList {
		if atomic.LoadInt32(&w.isWithDrawHandlerRunning) != 1 {
			fmt.Printf("Service stopping, processed %d/%d withdrawals\n", i, len(withDrawList))
			break
		}
		err := repository.WdWithTransaction(func(wRepo *repository.WdRepo) error {
			var newWithdraw model.Withdrawal
			asset, err := wRepo.GetAssetByAddress(withdraw.WalletAddress)
			if err != nil {
				return fmt.Errorf("get asset by address failed: %w", err)
			}
			switch withdraw.TokenType {
			case constant.TokenTypeBNB:
				res, err := w.transactionBNB(withdraw, asset.BnbBalance)
				if err != nil {
					return fmt.Errorf("transactionBNB failed: %w, withdrawid: %d", err, withdraw.ID)
				}
				newWithdraw = *res
			case constant.TokenTypeMTK:
				res, err := w.transactionERC20(withdraw, asset.MtkBalance)
				if err != nil {
					return fmt.Errorf("transactionERC20 failed:  %w", err)
				}
				newWithdraw = *res
			}
			if err := wRepo.UpdateWithdrawalInfo(newWithdraw); err != nil {
				return fmt.Errorf("update withdraw info failed: %w", err)
			}
			fmt.Println("update withdraw info success")
			return nil
		})

		if err != nil {
			fmt.Println("transaction failed: ", err)
		}
	}
}

func (w *WithdrawHandler) transactionBNB(withdraw model.Withdrawal, bnbBalance string) (*model.Withdrawal, error) {
	amount, err := utils.StringToBigInt(withdraw.Value)
	if err != nil {
		return nil, fmt.Errorf("transactionBNB: parse withdraw amount error: %w", err)
	}
	balance, err := utils.StringToBigInt(bnbBalance)
	if err != nil {
		return nil, fmt.Errorf("transactionBNB: parse withdraw bnb balance error: %w", err)
	}
	if amount.Cmp(balance) == 1 {
		return nil, fmt.Errorf("transactionBNB: balance is not enough, amount: %w, balance:%w", amount, balance)
	}
	tx, err := w.txService.SendBNB(withdraw.WalletAddress, amount)
	if err != nil {
		return nil, fmt.Errorf("transactionBNB: send withdraw amount error: %w", err)
	}

	withdraw.Hash = tx.Hash
	withdraw.Status = constant.WithdrawStatusPending

	return &withdraw, nil
}

func (w *WithdrawHandler) transactionERC20(withdraw model.Withdrawal, mtkBalance string) (*model.Withdrawal, error) {
	amount, err := utils.StringToBigInt(withdraw.Value)
	if err != nil {
		return nil, fmt.Errorf("transactionERC20: parse withdraw amount error: %w, withdraw.Amount:%s", err, withdraw.Amount)
	}
	balance, err := utils.StringToBigInt(mtkBalance)
	if err != nil {
		return nil, fmt.Errorf("transactionERC20: parse withdraw mtk balance error: %w", err)
	}
	if amount.Cmp(balance) == 1 {
		return nil, fmt.Errorf("transactionERC20: balance is not enough, amount: %w, balance:%w", amount, balance)
	}
	tx, err := w.txService.SendErc20(withdraw.WalletAddress, amount)
	if err != nil {
		return nil, fmt.Errorf("transactionERC20:send SendErc20error: %w, walletAddress: %s, withdrawid:%d", err, withdraw.WalletAddress, withdraw.ID)
	}

	withdraw.Hash = tx.Hash
	withdraw.Status = constant.WithdrawStatusPending

	return &withdraw, nil
}

func (w *WithdrawHandler) Stop() {
	atomic.StoreInt32(&w.isWithDrawHandlerRunning, 0)
}
