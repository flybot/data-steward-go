package main

import (
	"database/sql"
	"fmt"
	"log"
)

const usersTableName string = "extractor.users"

// User is the data type for user object
type User struct {
	ID        int          `json:"id" sql:"id"`
	Email     string       `json:"email" validate:"required" db:"email"`
	Password  string       `json:"password" validate:"required" db:"pass"`
	Username  string       `json:"username" db:"username"`
	CreatedAt sql.NullTime `json:"created_at" db:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at" db:"updated_at"`
	State     int          `json:"state" db:"state"`
}

func (user *User) Create() (int, error) {
	newID := 0
	query := fmt.Sprintf("INSERT INTO "+usersTableName+"(email, pass, username, state) VALUES('%s', '%s', '%s', %d) RETURNING id", user.Email, user.Password, user.Username, user.State)
	err := srv.db.GetConnection().QueryRow(query).Scan(&newID)
	if err != nil {
		return 0, err
	}
	return newID, nil
}
func (user *User) Get(whereStr string) error {
	err := srv.db.GetConnection().Get(user, "select * from "+usersTableName+" where "+whereStr)
	return err
}
func (user *User) Save() error {
	_, err := srv.db.GetConnection().NamedExec("UPDATE "+usersTableName+" SET email=:email, username=:username, updated_at=NOW(), state=:state", user)
	return err
}
func UserExists(column string, searchedValue string) bool {
	var counter int
	query := fmt.Sprintf("SELECT count(*) FROM %s WHERE %s = '%s'", usersTableName, column, searchedValue)
	err := srv.db.GetConnection().Get(&counter, query)
	if err != nil {
		log.Printf(err.Error())
	}
	if counter > 0 {
		return true
	}
	return false
}
