package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	Proto "start/Protocol"
	Proto2 "start/Protocol/Protocol2"
	"start/auth"
	"start/db"

	"github.com/gorilla/websocket"
)

type NetDataConn struct {
	Connection *websocket.Conn
	StrMd5     string
	Player     *Proto2.PlayerST
	writeMu    sync.Mutex
}

type RequestBody struct {
	req string
}

func (this *NetDataConn) PullFromClient() {
	defer UnregisterPlayer(this)

	for {
		msgType, data, err := this.Connection.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				fmt.Println("客户端正常关闭连接")
			} else {
				fmt.Println("读取消息失败：", err)
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
		fmt.Println("收到客户端消息：", content)
		this.SyncMessageFun(content)
	}
}

func (this *NetDataConn) SyncMessageFun(content string) {
	fmt.Println(content)

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
	case Proto2.CS2_PlayerBindProto2:
		this.PlayerBind(ProtoData)
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

	RegisterPlayer(player, this)
	this.PlayerSendServerMessage(player)
}

// HTTP OAuth 成功后，客户端通过 WebSocket 绑定已登录玩家
func (this *NetDataConn) PlayerBind(ProtoData map[string]interface{}) {
	//获取openid
	openID := protoString(ProtoData["OpenID"])
	if openID == "" {
		slog.Error("PlayerBind", "error", "missing OpenID", "data", ProtoData)
		return
	}

	slog.Info("PlayerBind", "msg", "开始绑定", "openID", openID)

	//根据openid从数据库获取玩家信息
	player, err := db.GetPlayerByOpenID(openID)
	if err != nil {
		slog.Error("PlayerBind", "error", err, "openID", openID)
		return
	}

	//注册该玩家
	RegisterPlayer(player, this)
	this.SendMessage(Proto.GameData_Proto, Proto2.SC2_PlayerBindProto2, player)
}

func (this *NetDataConn) SendMessage(proto, proto2 int, payload interface{}) {
	resp := map[string]interface{}{
		"Protocol":  proto,
		"Protocol2": proto2,
	}
	if payload != nil {
		resp["PlayerData"] = payload
	}

	data, err := json.Marshal(resp)
	if err != nil {
		slog.Error("SendMessage", "error", err)
		return
	}

	this.writeMu.Lock()
	defer this.writeMu.Unlock()
	if err := this.Connection.WriteMessage(websocket.TextMessage, data); err != nil {
		slog.Error("SendMessage", "error", err, "strMd5", this.StrMd5, "proto", proto, "proto2", proto2)
		return
	}

	slog.Info("SendMessage", "msg", "发送成功", "strMd5", this.StrMd5, "proto", proto, "proto2", proto2, "bytes", len(data))
}

// 发送给客户端的数据信息函数
func (this *NetDataConn) PlayerSendServerMessage(player *Proto2.PlayerST) {
	this.SendMessage(Proto.GameData_Proto, Proto2.SC2_PlayerLoginProto2, player)
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

func protoString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case float64:
		return fmt.Sprintf("%.0f", s)
	default:
		return fmt.Sprint(v)
	}
}
