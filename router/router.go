package router

import (
	"github.com/gin-gonic/gin"
	"log"
	"staking-interaction/controller"
)

func InitRouter() *gin.Engine {
	log.Println("Initializing router")
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	group := router.Group("/v1")

	staking := group.Group("/staking")
	{
		staking.POST("/stake", controller.Stake)
		staking.POST("/withdraw", controller.Withdraw)
		staking.GET("stake/:address", controller.GetAllStakesByFromAddress)
	}

	airdrop := group.Group("/airdropping")
	{
		airdrop.POST("/airdroperc20", controller.AirdropERC20)
		airdrop.POST("/airdropbnb", controller.AirdropBNB)
		airdrop.POST("/generateWallet", controller.GenerateMultiWallets)
	}

	transfer := group.Group("/transfer")
	{
		transfer.POST("/transferERC20", controller.SendErc20)
		transfer.POST("/transferBNB", controller.SendBNB)
	}

	return router
}
