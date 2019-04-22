package main

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"time"
)

type JSONTime time.Time

func (t JSONTime)MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("\"%d\"", time.Time(t).UnixNano() / 1e6)
	return []byte(stamp), nil
}

type user struct {
	Id       int64     `json:"id"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	Token    string    `json:"token"`
	Created  JSONTime `json:"created"`
}

func (u *user) login(db *sql.DB) error {
	password := u.Password
	err := db.QueryRow("SELECT id, password, created FROM user_info WHERE username=$1",
		u.Username).Scan(&u.Id, &u.Password, &u.Created)

	switch err {
	case nil:
		if u.Password != password {
			return errors.New("password error.")
		}
		break
	case sql.ErrNoRows:
		result, err := db.Exec("INSERT INTO user_info(username, password) VALUES ($1, $2)",
			u.Username, u.Password)
		if err != nil {
			return err
		}
		u.Id, _ = result.LastInsertId()
		u.Created = JSONTime(time.Now())
		break
	default:
		return err
	}

	return nil
}
