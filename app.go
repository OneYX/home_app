package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//go:generate go-bindata-assetfs static/...
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(user, password, dbName string) {
	connectionString :=
		fmt.Sprintf("%s.db?User=%s&pass=%s", dbName, user, password)
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
	uid INTEGER NOT NULL,
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
	router := a.Router.PathPrefix("/api").Subrouter()
	router.HandleFunc("/login", a.login).Methods("POST")
	router.HandleFunc("/favorite", ValidateMiddleware(a.getFavorite)).Methods("GET")
	router.HandleFunc("/favorite", ValidateMiddleware(a.addFavorite)).Methods("POST")
	router.HandleFunc("/favorite/{id:[0-9]+}", ValidateMiddleware(a.deleteFavorite)).Methods("DELETE")
	a.Router.PathPrefix("/").Handler(http.FileServer(assetFS()))

	err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		methods, err := route.GetMethods()
		if err != nil {
			return err
		}
		s := fmt.Sprintf("%s, methods=[%s]", pathTemplate, strings.Join(methods, ","))
		fmt.Println(s)
		return nil
	})

	if err != nil {
		fmt.Println(err)
	}
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
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	user := context.Get(r, "user").(User)
	favorites, err := user.getFavorites(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, favorites)
}

func (a *App) addFavorite(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(User)
	var f Favorite
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&f); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	if err := user.addFavorite(a.DB, &f); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, f)
}

func (a *App) deleteFavorite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Favorite ID")
		return
	}

	f := Favorite{Id: id}
	user := context.Get(r, "user").(User)
	if err := user.deleteFavorite(a.DB, f); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
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
