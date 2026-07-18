package protocol

const (
	INIT_PROTO        = iota
	GameData_Proto    //GameData_Proto=1 游戏主协议
	GameData_DB_Proto //GameData_DB_Proto=2 游戏的DB的主协议
)
