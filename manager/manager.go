package manager

import (
	"sync"

	"github.com/patrickmn/go-cache"
)

type Manager struct {
	// antiflood variables
	cache *cache.Cache

	// parser variables
	stateLock         sync.RWMutex
	registeredParsers []Parser
}

func NewManager() *Manager {
	m := new(Manager)
	m.initAntiflood()
	return m
}
