package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type NetDataConn struct {
	Connection *websocket.Conn
	StrMd5     string
}

func (this *NetDataConn) PullFromClient() {

	for {
		// 核心修改：替换原来的 Receive 读取逻辑
		msgType, data, err := this.Connection.ReadMessage()
		if err != nil {
			// 可选：区分正常关闭和异常断开，打印日志
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				fmt.Println("客户端正常关闭连接")
			} else {
				fmt.Println("读取消息失败：", err)
			}
			break
		}

		// 只处理文本消息，过滤二进制、ping/pong帧
		if msgType != websocket.TextMessage {
			continue
		}

		content := string(data)
		if len(content) == 0 {
			break
		}
		fmt.Println("收到客户端消息：", content)
	}
}
