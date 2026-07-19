package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"start/config"
	"start/db"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("本机几核:", runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	http.HandleFunc("/ws", wwwGolandLtd)
	http.HandleFunc("/api/github/callback", githubCallback)
	http.HandleFunc("/", indexHandler)

	fmt.Println("服务器启动: http://localhost:8080")
	fmt.Printf("数据库: %s:%s/%s\n", config.DB.Host, config.DB.Port, config.DB.Name)
	if proxy := os.Getenv("HTTP_PROXY"); proxy != "" {
		fmt.Printf("HTTP 代理: %s\n", proxy)
	}
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("网络错误", err)
	}
}
