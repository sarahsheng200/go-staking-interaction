package main

import (
	"log"
	"staking-interaction/database"
	"staking-interaction/listener"
)

func main() {
	err := database.MysqlConn()
	if err != nil {
		log.Fatal("MySQL database connect failed: ", err)
		return
	}
	defer func() {
		err := database.CloseConn()
		if err != nil {
			log.Fatal("Close database failed: ", err)
		}
	}()

	listener.SyncBlockInfo()
}
