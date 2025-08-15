package router

import (
	"github.com/gin-gonic/gin"
	"log"
	"staking-interaction/service"
)

func InitRouter() *gin.Engine {
	log.Println("Initializing router")
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	group := router.Group("/staking")
	group.POST("/stake", service.Stake)
	group.POST("/withdraw", service.Withdraw)
	group.GET("stake/:address", service.GetAllStakesByFromAddress)

	return router
}
