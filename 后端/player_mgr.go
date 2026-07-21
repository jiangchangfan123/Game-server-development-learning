package main

import (
	"fmt"
	"log/slog"
	"sync"

	Proto2 "start/Protocol/Protocol2"
)

var (
	G_PlayerDate = make(map[string]*NetDataConn)
	playerMu     sync.RWMutex
)

func makeStrMd5(player *Proto2.PlayerST) string {
	return fmt.Sprintf("%s_%d", player.PlayerName, player.UID)
}

// 注册玩家连接
func RegisterPlayer(player *Proto2.PlayerST, conn *NetDataConn) {
	key := makeStrMd5(player)
	conn.StrMd5 = key
	conn.Player = player

	playerMu.Lock()
	defer playerMu.Unlock()

	if old, ok := G_PlayerDate[key]; ok && old != conn {
		old.Connection.Close()
	}
	G_PlayerDate[key] = conn
	slog.Info("RegisterPlayer", "msg", "玩家上线", "strMd5", key, "online", len(G_PlayerDate))
}

// 注销玩家连接(释放连接)
func UnregisterPlayer(conn *NetDataConn) {
	if conn.StrMd5 == "" {
		return
	}

	playerMu.Lock()
	defer playerMu.Unlock()

	if G_PlayerDate[conn.StrMd5] == conn {
		delete(G_PlayerDate, conn.StrMd5)
	}
}

// 辅助函数:对连接服务器的玩家都进行某种操作
func RangeOnlinePlayers(fn func(key string, conn *NetDataConn)) {
	playerMu.RLock()
	defer playerMu.RUnlock()

	for key, conn := range G_PlayerDate {
		fn(key, conn)
	}
}

// 获取连接服务器的玩家个数
func OnlinePlayerCount() int {
	playerMu.RLock()
	defer playerMu.RUnlock()
	return len(G_PlayerDate)
}
