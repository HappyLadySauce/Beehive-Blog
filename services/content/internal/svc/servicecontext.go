package svc

import "github.com/HappyLadySauce/Beehive-Blog/services/content/internal/config"

type ServiceContext struct {
	Config config.Config
	Store  *memoryStore
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		Store:  newMemoryStore(),
	}
}
