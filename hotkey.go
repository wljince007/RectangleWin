//go:build linux

package main

import (
	"sync"
)

var (
	hotkeyRegistrations = make(map[int]*HotKey)
	hotkeyMu            sync.Mutex
)

// HotKey结构体复用
type HotKey struct {
	Key     string
	Handler func()
}
