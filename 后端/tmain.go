package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"start/auth"
	"start/config"
	"start/db"

	"github.com/gorilla/websocket"
)

var oauthStates sync.Map

const oauthStateTTL = 10 * time.Minute

func registerOAuthState(state string) {
	oauthStates.Store(state, time.Now())
}

func validateOAuthState(state string, r *http.Request) bool {
	if state == "" {
		return false
	}

	if v, ok := oauthStates.LoadAndDelete(state); ok {
		if time.Since(v.(time.Time)) <= oauthStateTTL {
			return true
		}
	}

	cookie, err := r.Cookie("oauth_state")
	return err == nil && cookie.Value == state
}

// 创建一个 Upgrader 实例，用于将 HTTP 连接升级为 WebSocket
var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wwwGolandLtd(w http.ResponseWriter, r *http.Request) {
	slog.Info("WebSocket连接请求", "remote", r.RemoteAddr)

	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WebSocket升级失败", "error", err)
		return
	}
	defer conn.Close()

	slog.Info("WebSocket连接成功", "remote", r.RemoteAddr)

	NetDataConntmp := &NetDataConn{
		Connection: conn,
		StrMd5:     "",
	}

	NetDataConntmp.PullFromClient()
}

// 首页 - 显示 HTML 登录页面
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("../前端/homepage.html")
	if err != nil {
		http.Error(w, "模板加载失败", http.StatusInternalServerError)
		return
	}

	state := randomState()
	registerOAuthState(state)
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   600,
		SameSite: http.SameSiteLaxMode,
	})

	authURL := config.GitHub.AuthCodeURL(state)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, map[string]string{"AuthURL": authURL})
}

// GitHub 回调：换取 token → 拉取用户信息 → 写入数据库
func githubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "缺少授权 code", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	if !validateOAuthState(state, r) {
		slog.Error("githubCallback", "error", "state mismatch", "state", state)
		http.Error(w, "state 校验失败，请从首页重新点击 GitHub 登录", http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", Value: "", Path: "/", MaxAge: -1})

	githubUser, err := auth.LoginWithCode(code)
	if err != nil {
		slog.Error("githubCallback", "error", err)
		http.Error(w, "GitHub 授权失败", http.StatusInternalServerError)
		return
	}

	player, err := db.SaveGitHubUser(githubUser)
	if err != nil {
		slog.Error("githubCallback", "error", err)
		http.Error(w, "保存用户信息失败", http.StatusInternalServerError)
		return
	}

	slog.Info("githubCallback", "msg", "登录成功", "uid", player.UID, "name", player.PlayerName, "openID", player.OpenID)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html><head><meta charset="UTF-8"><title>登录成功</title></head>
<body>
<h2>登录成功</h2>
<p>UID: %d</p>
<p>用户名: %s</p>
<p>GitHub ID: %s</p>
<p><a href="/">返回首页</a></p>
</body></html>`, player.UID, player.PlayerName, player.OpenID)
}

func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
