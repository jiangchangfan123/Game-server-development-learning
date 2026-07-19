package main

import (
	"encoding/json"
	"log/slog"

	"start/auth"
	"start/db"
	Proto "start/Protocol"
	Proto2 "start/Protocol/Protocol2"

	"github.com/gorilla/websocket"
)

type NetDataConn struct {
	Connection *websocket.Conn
	StrMd5     string
}

type RequestBody struct {
	req string
}

func (this *NetDataConn) PullFromClient() {

	for {
		msgType, data, err := this.Connection.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				slog.Info("客户端正常关闭连接")
			} else {
				slog.Error("读取消息失败", "error", err)
			}
			break
		}

		if msgType != websocket.TextMessage {
			continue
		}

		content := string(data)
		if len(content) == 0 {
			break
		}
		slog.Info("收到客户端消息", "content", content)
		go this.SyncMessageFun(content)
	}
}

func (this *NetDataConn) SyncMessageFun(content string) {
	var r RequestBody
	r.req = content

	if ProtoData, err := r.Json2Map(); err != nil {
		slog.Info("Json2Map", "error", err)
	} else {
		this.HandleCltProtocol(ProtoData["Protocol"], ProtoData["Protocol2"], ProtoData)
	}
}

func (this *NetDataConn) HandleCltProtocol(Protocol interface{}, Protocol2 interface{}, ProtoData map[string]interface{}) {
	proto, ok := toInt(Protocol)
	if !ok {
		slog.Error("HandleCltProtocol", "error", "invalid Protocol")
		return
	}

	switch proto {
	case Proto.GameData_Proto:
		this.HandleCltProtocol2(ProtoData["Protocol2"], ProtoData)
	case Proto.GameData_DB_Proto:
		// DB 协议预留
	default:
		slog.Error("HandleCltProtocol", "error", "主协议不存在", "Protocol", proto)
	}
}

func (this *NetDataConn) HandleCltProtocol2(Protocol2 interface{}, ProtoData map[string]interface{}) {
	proto2, ok := toInt(Protocol2)
	if !ok {
		slog.Error("HandleCltProtocol2", "error", "invalid Protocol2")
		return
	}

	switch proto2 {
	case Proto2.CS2_PlayerLoginProto2:
		this.PlayerLogin(ProtoData)
	default:
		slog.Error("HandleCltProtocol2", "error", "子协议不存在", "Protocol2", proto2)
	}
}

// 用户登录：客户端通过 WebSocket 发送 GitHub code，服务器认证并入库
func (this *NetDataConn) PlayerLogin(ProtoData map[string]interface{}) {
	code, _ := ProtoData["Code"].(string)
	if code == "" {
		slog.Error("PlayerLogin", "error", "missing Code")
		return
	}

	githubUser, err := auth.LoginWithCode(code)
	if err != nil {
		slog.Error("PlayerLogin", "error", err)
		return
	}

	player, err := db.SaveGitHubUser(githubUser)
	if err != nil {
		slog.Error("PlayerLogin", "error", err)
		return
	}

	resp := map[string]interface{}{
		"Protocol":   Proto.GameData_Proto,
		"Protocol2":  Proto2.SC2_PlayerLoginProto2,
		"PlayerData": player,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		slog.Error("PlayerLogin", "error", err)
		return
	}
	this.Connection.WriteMessage(websocket.TextMessage, data)
}

func (r *RequestBody) Json2Map() (s map[string]interface{}, err error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(r.req), &result); err != nil {
		slog.Info("Json2Map", "error", err)
		return nil, err
	}
	return result, nil
}

func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	default:
		return 0, false
	}
}
