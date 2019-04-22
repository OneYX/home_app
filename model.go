package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type User struct {
	Id       int64     `json:"id"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	Token    string    `json:"token,omitempty"`
	Created  time.Time `json:"created"`
}

func (u User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(&struct {
		Alias
		Password string `json:"password,omitempty"`
		Created  int64  `json:"created"`
	}{
		Alias:   (Alias)(u),
		Created: u.Created.UnixNano() / 1e6,
	})
}

func (u *User) login(db *sql.DB) error {
	password := u.Password
	err := db.QueryRow("SELECT id, password, created FROM user_info WHERE username=$1",
		u.Username).Scan(&u.Id, &u.Password, &u.Created)

	switch err {
	case nil:
		if u.Password != password {
			return errors.New("password error")
		}
		break
	case sql.ErrNoRows:
		result, err := db.Exec("INSERT INTO user_info(username, password) VALUES ($1, $2)",
			u.Username, u.Password)
		if err != nil {
			return err
		}
		u.Id, _ = result.LastInsertId()
		u.Created = time.Now()
		break
	default:
		return err
	}

	return nil
}
