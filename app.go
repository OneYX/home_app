package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(user, password, dbName string) {
	connectionString :=
		fmt.Sprintf("%s.db?User=%s&pass=%s", dbName, user, password)
	println(connectionString)
	var err error
	a.DB, err = sql.Open("sqlite3", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	createSql := `CREATE TABLE IF NOT EXISTS user_info (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username VARCHAR (64) NOT NULL,
	password VARCHAR (64) NOT NULL,
	created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE IF NOT EXISTS favorite (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title VARCHAR (64) NOT NULL,
	url VARCHAR (255) NOT NULL,
	img TEXT NOT NULL,
	created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
`
	_, err = a.DB.Exec(createSql)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, handlers.LoggingHandler(os.Stdout, a.Router)))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/test", ValidateMiddleware(TestEndpoint)).Methods("GET")
	a.Router.HandleFunc("/login", a.login).Methods("POST")
	a.Router.HandleFunc("/favorite", a.getFavorite).Methods("GET")
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var u User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := u.login(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, err.Error())
			//respondWithError(w, http.StatusNotFound, "User Create fail")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       u.Id,
		"username": u.Username,
		"created":  u.Created,
	})
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	u.Token = tokenString
	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) getFavorite(w http.ResponseWriter, r *http.Request) {

}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
