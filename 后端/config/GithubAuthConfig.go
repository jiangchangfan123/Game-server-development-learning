package config

import "net/url"

type GitHubAuthConfig struct {
	ClientId     string // github APP 的 ClientID
	ClientSecret string // github APP 密钥
	RedirectURL  string // 重定向地址
	AuthURL      string // github 授权认证 URL
	TokenURL     string // access_token URL
}

//获取授权URL
func (g *GitHubAuthConfig) AuthCodeURL(state string) string {
	u, _ := url.Parse(g.AuthURL)
	q := u.Query()
	q.Set("client_id", g.ClientId)
    q.Set("redirect_uri", g.RedirectURL)
    q.Set("scope", "read:user user:email")
    q.Set("state", state)
    u.RawQuery = q.Encode()
    return u.String()
}