package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
	"staking-interaction/common"
)

var DB *gorm.DB
var err error

func MysqlConn() error {
	dsn := common.MYSQL_USERNAME + ":" + common.MYSQL_PASSWORD + "@tcp(" + common.MYSQL_URL + ")/" + common.MYSQL_DATABASE + "?" + common.MYSQL_CONFIG

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})

	if err != nil {
		log.Fatalf("Failed to connect to the MySQL db: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Printf("Failed to connect to the DB: %v", err)
		return err
	}
	if err := sqlDB.Ping(); err != nil {
		log.Printf("Failed to ping to the MySQL db: %v", err)
		return err
	}
	log.Println("MySQL connect successfully")
	return nil
}

func CloseConn() error {
	if DB != nil {
		return nil
	}
	mysqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return mysqlDB.Close()
}
