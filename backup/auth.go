package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

//Clave Secreta Unica
var jwtKey = []byte("mi_clave_secreta_jony_123")

//Login

func loginHandler(w http.ResponseWriter, r *http.Request){
	var creds Credential
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil{
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Valido todo
	if creds.Username != "admin" || creds.Password != "1234"{
		w.WriteHeader(http.StatusUnauthorized)
		return
	}	
	experationTime := time.Now().Add(5 *time.Minute)
	claims := &Claims{
		Username: creds.Username, 
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(experationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey :=[]byte(os.Getenv("JWT_SECRET"))
	tokenString, err := token.SignedString(jwtKey)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// Middleware: Guardia de seguridad
func authMiddleware(next http.HandlerFunc) http.HandlerFunc{
	return func (w http.ResponseWriter, r *http.Request){
		tokenString := r.Header.Get("Autorizacion")
	// Limpio el Bearer
		tokenString = strings.TrimPrefix(tokenString, "Bearer")
		if tokenString == ""{
			http.Error(w, "Token requirido", http.StatusUnauthorized)
			return 
		}
		jwtKey := []byte(os.Getenv("JWT_SECRET"))
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func (token *jwt.Token) (interface{}, error){
			return  jwtKey, nil
		})
		if err != nil || !token.Valid{
			http.Error(w, "Token Invalido", http.StatusUnauthorized)
			return 
		}
		fmt.Println("Acceso Autorizado:", claims.Username)
		next(w, r)
	}
}