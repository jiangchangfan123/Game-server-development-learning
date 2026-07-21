package main

import (
	"log/slog"
	"time"

	Proto "start/Protocol"
	Proto2 "start/Protocol/Protocol2"
)

func G_timer() {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	//每20秒循环一次
	for range ticker.C {
		count := OnlinePlayerCount()
		slog.Info("G_timer", "online", count)
		if count == 0 {
			continue
		}

		//让每个玩家都向服务器发送一个消息(测试)
		RangeOnlinePlayers(func(key string, conn *NetDataConn) {
			if conn.Player == nil {
				return
			}
			conn.SendMessage(Proto.GameData_Proto, Proto2.SC2_PlayerLoginProto2, conn.Player)
		})
	}
}
