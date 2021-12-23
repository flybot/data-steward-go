package database

import "github.com/jmoiron/sqlx"

type Maker interface {
	Connect(host string, port int, user string, password string, dbname string) error
	GetConnection() *sqlx.DB
	Close() error
}
