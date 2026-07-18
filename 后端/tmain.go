package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

// 创建一个 Upgrader 实例，用于将 HTTP 连接升级为 WebSocket
var upgrade = websocket.Upgrader{}

func wwwGolandLtd(w http.ResponseWriter, r *http.Request) {
	fmt.Println("绝区零社区欢迎你 !")

	//  将 HTTP 连接升级为 WebSocket 连接
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 从原始的 http.Request 中获取 query 参数
	data := r.URL.Query().Get("data")
	fmt.Println("data:", data)

	//网络信息
	NetDataConntmp := &NetDataConn{
		Connection: conn,
		StrMd5:     "",
	}

	NetDataConntmp.PullFromClient()

	//处理结构信息
}

// 首页 - 显示 HTML 登录页面
func indexHandler(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("../前端/homepage.html")
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}

// GitHub 回调
func githubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	fmt.Fprintf(w, "授权成功,code: %s", code)
}
