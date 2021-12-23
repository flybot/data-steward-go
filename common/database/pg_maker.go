package database

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PgDb struct {
	Conn *sqlx.DB
}

func InitPostgresql() Maker {
	return &PgDb{}
}

func (db *PgDb) Connect(host string, port int, user string, password string, dbname string) error {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s", host, port, user, password, dbname)
	conn, err := sqlx.Open("postgres", connectionString)
	conn.Ping()

	if err != nil {
		log.Fatal("Error opening Database connection")
	}
	db.Conn = conn
	log.Println("Database connection opened")
	return nil
}

func (db *PgDb) GetConnection() *sqlx.DB {
	return db.Conn
}

func (db *PgDb) Close() error {
	err := db.Conn.Close()
	if err != nil {
		log.Printf("Error closing DB: %v", err)
	}

	log.Println("Database closed")

	return err
}
