package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime"

	"start/config"
	"start/db"
	"start/logger"
)

func main() {
	if err := logger.Init(); err != nil {
		log.Fatal(err)
	}
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	slog.Info("服务器启动", "cpuNum", runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	http.HandleFunc("/ws", wwwGolandLtd)
	http.HandleFunc("/api/github/callback", githubCallback)
	http.HandleFunc("/", indexHandler)

	slog.Info("HTTP服务监听", "addr", ":8080")
	slog.Info("数据库", "host", config.DB.Host, "port", config.DB.Port, "name", config.DB.Name)
	if proxy := os.Getenv("HTTP_PROXY"); proxy != "" {
		slog.Info("HTTP代理", "proxy", proxy)
	}
	if err := http.ListenAndServe(":8080", nil); err != nil {
		slog.Error("网络错误", "error", err)
	}
}
