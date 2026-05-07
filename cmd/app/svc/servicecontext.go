package svc

import (
	"github.com/HappyLadySauce/Beehive-Blog/pkg/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config *config.Config
	DB *gorm.DB
	Cache *redis.Client
}

func NewServiceContext(config *config.Config) *ServiceContext {
	return &ServiceContext{
		Config: config,
	}
}