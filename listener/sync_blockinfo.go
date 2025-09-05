package listener

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"math/big"
	"staking-interaction/common/config"
	"staking-interaction/common/redis"
	"staking-interaction/dto"
	"staking-interaction/model"
	"staking-interaction/repository"
	"staking-interaction/utils"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// SyncBlockInfo 区块同步服务
type SyncBlockInfo struct {
	client        *ethclient.Client
	isSyncRunning int32         // 原子操作控制同步状态
	mu            sync.RWMutex  // 保护doneBlock的读写
	doneBlock     uint64        // 当前已同步的区块号
	workerPool    chan struct{} // 工作池控制并发数量
	blockManager  *repository.BlockSyncManager
	workerWg      sync.WaitGroup // 等待所有交易处理Goroutine退出
	lockManager   *redis.LockManager
	// 添加配置和指标
	config  *SyncConfig
	metrics *SyncMetrics
	//logger        *logrus.Logger
}

type SyncConfig struct {
	WorkerCount     int           `json:"worker_count"`
	BlockTimeout    time.Duration `json:"block_timeout"`
	RetryDelay      time.Duration `json:"retry_delay"`
	MaxBlocksBehind uint64        `json:"max_blocks_behind"`
}

type SyncMetrics struct {
	ProcessedBlocks uint64
	ProcessedTxs    uint64
	FailedBlocks    uint64
	FailedTxs       uint64
	LastProcessTime time.Time
	mu              sync.RWMutex
}

var conf = config.Get()
var blockChainConf = conf.BlockchainConfig

// NewSyncBlockInfo 创建新的区块同步服务
func NewSyncBlockInfo(client *ethclient.Client, config *SyncConfig, lockManager *redis.LockManager) *SyncBlockInfo {
	if config == nil {
		config = &SyncConfig{
			WorkerCount:     blockChainConf.Sync.Workers,
			BlockTimeout:    blockChainConf.Sync.SyncInterval,
			RetryDelay:      blockChainConf.Sync.RetryDelay,
			MaxBlocksBehind: blockChainConf.Sync.BlockBuffer,
		}
	}

	return &SyncBlockInfo{
		client:       client,
		blockManager: repository.NewBlockSyncManager("last_synced_block.txt"),
		workerPool:   make(chan struct{}, config.WorkerCount), // 限制并发处理数量
		config:       config,
		metrics:      &SyncMetrics{},
		lockManager:  lockManager,
		//logger:       logger,
	}
}

// Start 启动区块同步
func (s *SyncBlockInfo) Start() {
	if !atomic.CompareAndSwapInt32(&s.isSyncRunning, 0, 1) {
		fmt.Println("sync service is already running")
		return
	}
	fmt.Println("sync service is running...")
	// 初始化起始区块
	if err := s.initializeStartBlock(); err != nil {
		fmt.Printf("initialize start block: %v\n", err)
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("initialize start block: %v\n", r)
				return
			}
		}()
		s.syncLoop()
	}()

	fmt.Println("Sync service started successfully", "initial_done_block", s.getDoneBlock())
}

func (s *SyncBlockInfo) initializeStartBlock() error {
	//if err := s.loadLastSyncedBlock(); err != nil || s.getDoneBlock() == 0 {
	//fmt.Println("Failed to load last synced block, starting from current", "error", err)
	if s.getDoneBlock() == 0 {
		ctx, cancel := context.WithTimeout(context.Background(), s.config.BlockTimeout)
		defer cancel()

		currentBlock, e := s.client.BlockNumber(ctx)
		if e != nil {
			return fmt.Errorf("failed to get current block: %w", e)
		}
		s.setDoneBlock(currentBlock)
	}
	return nil
}

// Stop 停止区块同步
func (s *SyncBlockInfo) Stop() {
	atomic.StoreInt32(&s.isSyncRunning, 0)
	fmt.Println("Sync service stopping")

	//s.loopWg.Wait()
	s.workerWg.Wait()
	fmt.Println("Sync service stopped")
}

// 同步循环
func (s *SyncBlockInfo) syncLoop() {

	fmt.Println("Starting block synchronization")
	defer fmt.Println("Block synchronization stopped")

	chainID, err := s.client.ChainID(context.Background())
	if err != nil {
		fmt.Println("failed to get chain ID: ", err)
		return
	}

	for atomic.LoadInt32(&s.isSyncRunning) == 1 {
		blockCtx, cancel := context.WithTimeout(context.Background(), s.config.BlockTimeout)

		currentBlock, err := s.client.BlockNumber(blockCtx)
		if err != nil {
			fmt.Println("failed to get current block number: ", "lastSyncedBlock", s.getDoneBlock())
			cancel()
			time.Sleep(10 * time.Second)
			continue
		}

		doneBlock := s.getDoneBlock()
		// 控制同步距离，避免超前太多
		if currentBlock < doneBlock+s.config.MaxBlocksBehind {
			fmt.Println("Waiting for new blocks", "current", currentBlock, "done", doneBlock)
			cancel()
			time.Sleep(10 * time.Second)
			continue
		}

		// 处理当前区块
		if err := s.processBlock(blockCtx, chainID, doneBlock); err != nil {
			fmt.Println("Failed to process block", "block", doneBlock, "error", err)
			cancel()
			// 处理区块错误时重试延迟加倍
			time.Sleep(20 * time.Second)
			continue
		}

		// 更新已完成区块号并持久化
		newDoneBlock := doneBlock + 1
		s.setDoneBlock(newDoneBlock)
		// 保存区块到文件中
		if err := s.saveSyncedBlock(newDoneBlock); err != nil {
			fmt.Println("Failed to save synced block", "block", newDoneBlock, "error", err)
		} else {
			fmt.Println("Processed block", "block", doneBlock, "newDone", newDoneBlock)
		}
		cancel()
	}
}

// 处理单个区块
func (s *SyncBlockInfo) processBlock(blockCtx context.Context, chainID *big.Int, blockNumber uint64) error {
	// 获取区块详情
	block, err := s.client.BlockByNumber(blockCtx, big.NewInt(int64(blockNumber)))
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return fmt.Errorf("block %d not found: %w", blockNumber, err)
		}
		return fmt.Errorf("get block failed: %w", err)
	}

	fmt.Println("Start processing block", "block_number", blockNumber, "transaction_count", len(block.Transactions()))

	var txWg sync.WaitGroup
	for _, tx := range block.Transactions() {
		txWg.Add(1)
		s.workerPool <- struct{}{} // 使用工作池控制并发
		s.workerWg.Add(1)          // 绑定全局交易WaitGroup

		go func(tx *types.Transaction, blockNum uint64) {
			txCtx, txCancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer func() {
				txCancel()
				txWg.Done()
				<-s.workerPool
				s.workerWg.Done()

				// 捕获panic，防止单个交易处理崩溃整个服务
				if r := recover(); r != nil {
					fmt.Println("Recovered from panic in transaction processing",
						"hash", tx.Hash().Hex(), "panic", r)
				}
			}()

			if err := s.processTransaction(txCtx, chainID, tx); err != nil {
				fmt.Println("Failed to process transaction",
					"tx_hash", tx.Hash().Hex(),
					"block_number", blockNum,
					"error", err)
			}
		}(tx, blockNumber)
	}

	txWg.Wait()
	fmt.Println("Processed block success", "block_number", blockNumber, "transaction_count", len(block.Transactions()))
	return nil
}

// 处理单个交易
func (s *SyncBlockInfo) processTransaction(txCtx context.Context, chainID *big.Int, tx *types.Transaction) error {
	// 获取交易回执
	receipt, err := s.client.TransactionReceipt(txCtx, tx.Hash())
	if err != nil {
		return fmt.Errorf("get transaction receipt: %w", err)
	}

	// 跳过合约创建交易
	if tx.To() == nil {
		fmt.Println("Skipping contract creation transaction", "hash", tx.Hash().Hex())
		return nil
	}

	toAddr := *tx.To()
	// 检查接收地址是否为平台地址
	if !isToAddrValid(toAddr) {
		fmt.Println("Transaction not to our address", "hash", tx.Hash().Hex(), "to", toAddr.Hex())
		return nil
	}

	// 只处理成功的交易
	if receipt.Status != types.ReceiptStatusSuccessful {
		fmt.Println("Skipping failed transaction", "hash", tx.Hash().Hex())
		return nil
	}

	// 获取发送者地址
	signer := types.NewLondonSigner(chainID)
	fromAddr, err := types.Sender(signer, tx)
	if err != nil {
		return fmt.Errorf("get sender address: %w", err)
	}

	// 检查发送者是否为平台用户
	isValidAccount, accountId := isValidAccount(fromAddr)
	if !isValidAccount {
		fmt.Println("Sender is not our customer", "hash", tx.Hash().Hex(), "from", fromAddr.Hex())
		return nil
	}

	// 判断是否为合约交易
	isContract, err := s.isContractTx(txCtx, toAddr)
	if err != nil {
		return fmt.Errorf("check contract tx: %w", err)
	}

	// 处理不同类型的交易
	if isContract {
		return s.handleERC20Tx(tx, receipt, *accountId)
	} else if len(tx.Data()) == 0 {
		// 处理普通BNB交易
		return s.handleTokenTransaction(
			tx,
			receipt,
			*accountId,
			fromAddr,
			*tx.To(),
			tx.Value(),
			config.TokenTypeBNB,
		)
	}

	fmt.Println("Unsupported transaction type", "hash", tx.Hash().Hex())
	return nil
}

// 通用代币交易处理逻辑
func (s *SyncBlockInfo) handleTokenTransaction(
	tx *types.Transaction,
	receipt *types.Receipt,
	accountId int,
	fromAddr, toAddr common.Address,
	amount *big.Int,
	tokenType int,
) error {
	//获取锁
	assetLock, err := s.lockManager.AcquireAssetLock(context.Background(), accountId, tokenType)
	if err != nil {
		return fmt.Errorf("acquire assetLock failed: %w ,accountid:%d, tx_hash:%s", err, accountId, tx.Hash().Hex())
	}

	//释放锁
	defer func() {
		unlockCtx, unlockCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer unlockCancel()

		if err := assetLock.Unlock(unlockCtx); err != nil {
			fmt.Printf("unlock assetLock failed: %w ,accountid:%d, tx_hash:%s\n", err, accountId, tx.Hash().Hex())
		} else {
			fmt.Printf("lock relase success, tx_hash:%s, blocknumber:%d\n", tx.Hash().Hex(), receipt.BlockNumber.Int64())
		}

	}()

	return s.executeTransactionWithLock(receipt, accountId, fromAddr, toAddr, amount, tokenType)
}

func (s *SyncBlockInfo) executeTransactionWithLock(
	receipt *types.Receipt,
	accountId int,
	fromAddr, toAddr common.Address,
	amount *big.Int,
	tokenType int) error {
	return repository.TxWithTransaction(func(txRepo *repository.TxRepository) error {
		hash := receipt.TxHash.String()
		// 确认交易记录是否已存在
		isExistTx, err := txRepo.TransactionExists(hash)
		if err != nil {
			return fmt.Errorf("get TransactionExists failed: %w, hash:%s", err, hash)
		}
		if isExistTx {
			return fmt.Errorf("this transaction is existed, hash:%s", hash)
		}

		// 获取账户资产
		asset, err := txRepo.GetAssetByAccountIdWithLock(accountId)
		if err != nil {
			return fmt.Errorf("get account asset: %w", err)
		}

		// 计算new balance
		preBalance, nextBalance, err := s.calculateBalance(tokenType, asset, amount)
		if err != nil {
			return fmt.Errorf("unsupported token type: %d", tokenType)
		}

		// 创建账单记录
		bill := model.Bill{
			AccountID:   accountId,
			TokenType:   tokenType,
			BillType:    config.BillTypeRecharge,
			Amount:      amount.String(),
			Fee:         strconv.FormatUint(receipt.GasUsed, 10),
			PreBalance:  preBalance,
			NextBalance: nextBalance,
			CreatedAt:   time.Now(),
		}
		if err := txRepo.AddBill(&bill); err != nil {
			return fmt.Errorf("add bill: %w", err)
		}

		// 创建交易日志
		transLog := model.TransactionLog{
			AccountID:   accountId,
			TokenType:   tokenType,
			Hash:        hash,
			Amount:      amount.String(),
			FromAddress: fromAddr.Hex(),
			ToAddress:   toAddr.Hex(),
			BlockNumber: receipt.BlockNumber.String(),
			CreatedAt:   time.Now(),
		}
		if err := txRepo.AddTransactionLog(&transLog); err != nil {
			return fmt.Errorf("add transaction log: %w", err)
		}

		// 更新资产余额（使用乐观锁）
		if err := txRepo.UpdateAssetWithOptimisticLock(asset, nextBalance, tokenType); err != nil {
			return fmt.Errorf("update asset: %w", err)
		}

		fmt.Println("Successfully processed transaction",
			"hash", hash,
			"accountId", accountId,
			"tokenType", tokenType,
			"amount", amount.String())
		return nil
	})
}

func (s *SyncBlockInfo) calculateBalance(tokenType int, asset *model.AccountAsset, amount *big.Int) (string, string, error) {
	var preBalanceStr string
	var err error
	var preBalance *big.Int

	switch tokenType {
	case config.TokenTypeMTK:
		preBalanceStr = asset.MtkBalance
		preBalance, err = utils.StringToBigInt(preBalanceStr)
		if err != nil {
			return "", "", fmt.Errorf("parse pre balance: %w", err)
		}

	case config.TokenTypeBNB:
		preBalanceStr = asset.BnbBalance
		preBalance, err = utils.StringToBigInt(preBalanceStr)
		if err != nil {
			return "", "", fmt.Errorf("parse pre balance: %w", err)
		}

	default:
		return "", "", fmt.Errorf("unsupported token type: %d", tokenType)
	}

	nextBalance := new(big.Int).Add(preBalance, amount)
	return preBalanceStr, nextBalance.String(), nil
}

// 处理ERC20代币交易
func (s *SyncBlockInfo) handleERC20Tx(tx *types.Transaction, receipt *types.Receipt, accountId int) error {
	transEvent, err := parseERC20TxByReceipt(receipt)
	if err != nil {
		return fmt.Errorf("parse erc20 tx: %w", err)
	}

	return s.handleTokenTransaction(
		tx,
		receipt,
		accountId,
		transEvent.FromAddress,
		transEvent.ToAddress,
		transEvent.Value,
		config.TokenTypeMTK,
	)
}

// 解析ERC20交易事件
func parseERC20TxByReceipt(receipt *types.Receipt) (*dto.TransferEvent, error) {

	tran := dto.TransferEvent{}
	var erc20TransferEventABIJson = `[
		{
			"anonymous": false,
			"inputs": [
				{"indexed": true, "name": "from", "type": "address"},
				{"indexed": true, "name": "to", "type": "address"},
				{"indexed": false, "name": "value", "type": "uint256"}
			],
			"name": "Transfer",
			"type": "event"
		}
	]`

	tokenAddr := common.HexToAddress(blockChainConf.Contracts.Token)
	erc20ABI, err := abi.JSON(strings.NewReader(erc20TransferEventABIJson))
	if err != nil {
		return nil, fmt.Errorf("parse erc20 abi: %w", err)
	}

	transferEventSig := erc20ABI.Events["Transfer"].ID.Hex()

	for _, log := range receipt.Logs {
		if log.Address != tokenAddr {
			continue
		}

		if common.BytesToHash(log.Topics[0].Bytes()).Hex() != transferEventSig {
			continue
		}

		tran.FromAddress = common.HexToAddress(log.Topics[1].String())
		tran.ToAddress = common.HexToAddress(log.Topics[2].String())

		if err := erc20ABI.UnpackIntoInterface(&tran, "Transfer", log.Data); err != nil {
			return nil, fmt.Errorf("unpack transfer data: %w", err)
		}
		return &tran, nil
	}
	return nil, fmt.Errorf("no erc20 transfer event found in receipt")
}

// 检查是否为合约地址
func (s *SyncBlockInfo) isContractTx(blockCtx context.Context, to common.Address) (bool, error) {
	// 合约地址有字节码，普通地址无
	code, err := s.client.CodeAt(blockCtx, to, nil)
	if err != nil {
		return false, fmt.Errorf("get code at address: %w", err)
	}
	return len(code) > 0, nil
}

// 检查接收地址是否有效
func isToAddrValid(toAddr common.Address) bool {
	for _, addr := range blockChainConf.Owners {
		if common.HexToAddress(addr) == toAddr {
			return true
		}
	}
	return false
}

// 检查账户是否有效
func isValidAccount(fromAddr common.Address) (bool, *int) {
	account, err := repository.GetAccount(fromAddr.Hex())
	if err != nil {
		return false, nil
	}
	if account.AccountID > 0 {
		return true, &account.AccountID
	}
	return false, nil
}

func getTokenTypeName(tokenType int) string {
	switch tokenType {
	case config.TokenTypeMTK:
		return "MTK"
	case config.TokenTypeBNB:
		return "BNB"
	default:
		return strconv.Itoa(tokenType)
	}
}

// 加载上次同步的区块号
func (s *SyncBlockInfo) loadLastSyncedBlock() error {
	block, err := s.blockManager.GetLastSyncedBlock()
	if err != nil {
		return err
	}
	s.setDoneBlock(block)
	return nil
}

// 保存已同步的区块号
func (s *SyncBlockInfo) saveSyncedBlock(block uint64) error {
	return s.blockManager.SaveSyncedBlock(block)
}

// 获取当前已同步的区块号
func (s *SyncBlockInfo) getDoneBlock() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.doneBlock
}

// 设置已同步的区块号
func (s *SyncBlockInfo) setDoneBlock(block uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.doneBlock = block
}
