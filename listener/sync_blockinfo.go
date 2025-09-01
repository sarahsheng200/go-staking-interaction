package listener

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	constant "staking-interaction/common"
	"staking-interaction/service"
)

var (
	isSyncRunning bool
)

func SyncBlockInfo() {
	client, err := ethclient.Dial(constant.RPC_URL)
	if err != nil {
		log.Fatalf("SyncBlockInfo: Failed to connect to the BSC network: %v", err)
	}
	defer client.Close()

	syncBlockService := service.NewSyncBlockService(client)
	syncBlockService.SyncBlocks()
}
