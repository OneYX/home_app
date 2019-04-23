package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"net/http"
	"strings"

	"./utils"
)

func ValidateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authorizationHeader := req.Header.Get("authorization")
		if authorizationHeader != "" {
			bearerToken := strings.Split(authorizationHeader, " ")
			if len(bearerToken) == 2 {
				token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("there was an error")
					}
					return []byte("secret"), nil
				})
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, err.Error())
					return
				}
				if token.Valid {
					var user User
					utils.Decode(token.Claims.(jwt.MapClaims), &user)
					context.Set(req, "user", user)
					next(w, req)
				} else {
					respondWithError(w, http.StatusUnauthorized, "Invalid authorization token")
				}
			}
		} else {
			respondWithError(w, http.StatusForbidden, "An authorization header is required")
		}
	})
}
