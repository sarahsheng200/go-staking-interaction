package router

import (
	"github.com/gin-gonic/gin"
	"log"
	"staking-interaction/service"
	airdropService "staking-interaction/service/airdrop"
)

func InitRouter() *gin.Engine {
	log.Println("Initializing router")
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	group := router.Group("/v1")

	staking := group.Group("/staking")
	{
		staking.POST("/stake", service.Stake)
		staking.POST("/withdraw", service.Withdraw)
		staking.GET("stake/:address", service.GetAllStakesByFromAddress)
	}

	airdrop := group.Group("/airdropping")
	{
		airdrop.POST("/airdroperc20", airdropService.AirdropERC20)
		airdrop.POST("/generateWallet", airdropService.GenerateMultiWallets)
	}

	return router
}
