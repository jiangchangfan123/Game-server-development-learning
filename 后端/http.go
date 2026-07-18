package main

import (
	"fmt"
	"net/http"
	"runtime"
)

// 游戏服务器初始化
func init() {

}

func main() {

	fmt.Println("本机几核:", runtime.NumCPU())
	//设置同时执行 Go 代码的最大 OS 线程数为 n。
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	http.HandleFunc("/ws", wwwGolandLtd)
	http.HandleFunc("/api/github/callback", githubCallback)
	http.HandleFunc("/", indexHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("网络错误", err)
		return
	}

	//消息处理，先处理主协议，再处理子协议
}
