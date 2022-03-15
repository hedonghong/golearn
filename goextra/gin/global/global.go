package global

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golearn/goextra/gin/config"
	"gorm.io/gorm"
)

var (
	MINI_DB *gorm.DB
	MINI_REDIS *redis.Client
	MINI_LOG *zap.Logger
	MINI_VP *viper.Viper
	MINI_CONFIG *config.Server

)
