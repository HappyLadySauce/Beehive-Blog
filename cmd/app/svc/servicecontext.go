package svc

import (
	"github.com/HappyLadySauce/Beehive-Blog/pkg/config"
)


type ServiceContext struct {
	Config *config.Config
}

func NewServiceContext(config *config.Config) *ServiceContext {
	return &ServiceContext{
		Config: config,
	}
}