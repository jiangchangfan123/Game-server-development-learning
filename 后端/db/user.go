package db

import (
	"database/sql"
	"fmt"

	"start/auth"
	"start/config"
	Proto2 "start/Protocol/Protocol2"

	_ "github.com/go-sql-driver/mysql"
)

var conn *sql.DB

func Init() error {
	var err error
	conn, err = sql.Open("mysql", config.DB.DSN())
	if err != nil {
		return err
	}

	if err = conn.Ping(); err != nil {
		return fmt.Errorf("connect mysql: %w", err)
	}

	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			github_id VARCHAR(64) NOT NULL UNIQUE,
			username VARCHAR(128) NOT NULL,
			email VARCHAR(256),
			avatar_url VARCHAR(512),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`)
	return err
}

func Close() {
	if conn != nil {
		conn.Close()
	}
}

// SaveGitHubUser 保存或更新 GitHub 用户，返回玩家结构
func SaveGitHubUser(user *auth.GitHubUser) (*Proto2.PlayerST, error) {
	githubID := fmt.Sprintf("%d", user.ID)

	_, err := conn.Exec(`
		INSERT INTO users (github_id, username, email, avatar_url)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			username = VALUES(username),
			email = VALUES(email),
			avatar_url = VALUES(avatar_url)
	`, githubID, user.Login, user.Email, user.AvatarURL)
	if err != nil {
		return nil, err
	}

	var uid int
	var username, openID string
	err = conn.QueryRow(
		"SELECT id, username, github_id FROM users WHERE github_id = ?",
		githubID,
	).Scan(&uid, &username, &openID)
	if err != nil {
		return nil, err
	}

	return &Proto2.PlayerST{
		UID:        uid,
		PlayerName: username,
		OpenID:     openID,
	}, nil
}
