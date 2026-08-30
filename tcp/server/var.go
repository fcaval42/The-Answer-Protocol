package main

import (
	"sync"
	"net"
)

var (
    pendingInvites     = make(map[string]chan string)
    pendingInvitesLock sync.Mutex
)

type Group struct {
	ID      string
	Members map[string]net.Conn
	messages []string
	Mu      sync.RWMutex
}

var (
	groups     = make(map[string]*Group)
	groupsLock sync.Mutex
)

var (
	ConnectedClients     = make(map[string]net.Conn)
	ConnectedLock sync.Mutex
)

var (
	globaleMessage     []string
	globaleMessageLock sync.Mutex
)
var (inGroupChat	 map[string][]net.Conn
	inGroupChatLock sync.Mutex
)
var (
	inGlobal     []net.Conn
	inGlobalLock sync.Mutex
)
var (
	group     [][]net.Conn
	groupLock sync.Mutex
)

