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

type Favorite struct {
	Id      int64     `json:"id"`
	Uid     int64     `json:"-"`
	Title   string    `json:"title"`
	Url     string    `json:"url"`
	Img     string    `json:"img"`
	Created time.Time `json:"created"`
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

func (f Favorite) MarshalJSON() ([]byte, error) {
	type Alias Favorite
	return json.Marshal(&struct {
		Alias
		Created int64 `json:"created"`
	}{
		Alias:   (Alias)(f),
		Created: f.Created.UnixNano() / 1e6,
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

func (u User) addFavorite(db *sql.DB, favorite *Favorite) error {
	result, err := db.Exec("INSERT INTO favorite(uid, title, url, img) VALUES ($1, $2, $3, $4)",
		u.Id, favorite.Title, favorite.Url, favorite.Img)
	if err != nil {
		return err
	}
	favorite.Id, _ = result.LastInsertId()
	favorite.Created = time.Now()
	return nil
}

func (u User) getFavorites(db *sql.DB, start, count int) ([]Favorite, error) {
	rows, err := db.Query(
		"SELECT id, uid, title, url, img, created FROM favorite WHERE uid = $1 LIMIT $2 OFFSET $3",
		u.Id, count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	favorites := make([]Favorite, 0)

	for rows.Next() {
		var f Favorite
		if err := rows.Scan(&f.Id, &f.Uid, &f.Title, &f.Url, &f.Img, &f.Created); err != nil {
			return nil, err
		}
		favorites = append(favorites, f)
	}

	return favorites, nil
}

func (u User) deleteFavorite(db *sql.DB, f Favorite) error {
	_, err := db.Exec("DELETE FROM favorite WHERE id=$1 and uid=$2", f.Id, u.Id)

	return err
}
