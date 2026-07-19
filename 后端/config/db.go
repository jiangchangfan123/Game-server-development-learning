package config

import "fmt"

type DBConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

func (d *DBConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name)
}
