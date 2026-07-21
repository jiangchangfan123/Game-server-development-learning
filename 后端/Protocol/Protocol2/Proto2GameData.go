package Protocol2

const (
	INIT_PROTO2           = iota
	CS2_PlayerLoginProto2 //客户端->服务器 用户登录协议
	SC2_PlayerLoginProto2 //服务器->客户端 用户登录协议

	//选择房间协议
	CS2_ChooseRoomProto2
	SC2_ChooseRoomProto2

	// WebSocket 绑定已登录玩家（HTTP OAuth 成功后使用）
	CS2_PlayerBindProto2
	SC2_PlayerBindProto2
)

type PlayerST struct {
	UID        int
	PlayerName string
	OpenID     string // GitHub 用户唯一 ID
}

type Head_Proto struct {
	Proto  int //主协议 -- 模块化
	Proto2 int //子协议 -- 模块化的功能
}

type CS2_PlayerLogin struct {
	Head_Proto
	Code string //github授权code
}

type SC2_PlayerLogin struct {
	Head_Proto
	PlayerData *PlayerST //玩家结构
}
