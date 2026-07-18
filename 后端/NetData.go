package main

import (
	"encoding/json"
	"fmt"

	"log/slog"
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
		go this.SyncMessageFun(content)
	}
}

// 处理消息
func (this *NetDataConn) SyncMessageFun(content string) {
	fmt.Println(content)

	var r RequestBody
	r.req = content

	if ProtoData, err := r.Json2Map(); err != nil {
		slog.Info("Json2Map", "error", err)
	} else {
		this.HandleCltProtocol(ProtoData["Protocol"], ProtoData["Protocol2"], ProtoData)
	}
	//实现格式的处理函数：主协议、子协议、struct

}

func (this *NetDataConn) HandleCltProtocol(Protocol interface{}, Protocol2 interface{}, ProtoData map[string]interface{}) {
	//分发处理	首先判断主协议存在，再判断子协议存在
	switch Protocol {
	case Proto.GameData_Proto:
		{
			//子协议处理
			this.HandleCltProtocol2(Protocol2, ProtoData)
		}
	case Proto.GameData_DB_Proto:
		{

		}
	default:
		panic("主协议不存在!!!")
	}

}

func (this *NetDataConn) HandleCltProtocol2(Protocol2 interface{}, ProtoData map[string]interface{}) {
	//处理子协议
	switch Protocol2 {
	case Proto2.CS2_PlayerLoginProto2:
		{
			//功能函数处理 -- 用户登录协议
			this.PlayerLogin(ProtoData)
		}
	default:
		panic("子协议不存在!!!")
	}

}

// 用户登录的协议
func (this *NetDataConn) PlayerLogin(ProtoData map[string]interface{}) {
	//服务器的逻辑处理
	//获取我们从client传过来的code
	//通过微信提供的接口-- 获取微信玩家的个人信息
	//将用户数据信息存到我们的数据库里(异步处理)
	//返回给客户端数据
}

// 将json字符串转换为map
func (r *RequestBody) Json2Map() (s map[string]interface{}, err error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(r.req), &result); err != nil {
		slog.Info("Json2Map", "error", err)
		return nil, err
	}
	return result, nil
}
