package middleware

import (
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

var (
	log  *logrus.Logger
	once sync.Once
)

// InitLogger 初始化日志（项目启动时调用一次即可）
func InitLogger(level logrus.Level, jsonFormat bool) {
	once.Do(func() {
		log = logrus.New()
		log.SetOutput(os.Stdout)
		log.SetLevel(level)
		if jsonFormat {
			log.SetFormatter(&logrus.JSONFormatter{})
		} else {
			log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
		}
	})
}

// GetLogger 提供全局 logger
func GetLogger() *logrus.Logger {
	if log == nil {
		InitLogger(logrus.InfoLevel, false) // 默认配置
	}
	return log
}
