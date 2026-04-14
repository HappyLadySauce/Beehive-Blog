package svc

import "github.com/HappyLadySauce/Beehive-Blog/services/identity/internal/config"

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
