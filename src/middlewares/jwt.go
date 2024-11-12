package middlewares

import (
	"be-golang-todo/src/helper/utils"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func ProtectedHandler(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Retrieve the token from the Authorization header
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized: missing or invalid token", http.StatusUnauthorized)
			return
		}

		// Extract the token string by trimming the "Bearer " prefix
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Decode and validate the JWT token
		claims, err := utils.DecodeToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		// Get username from claims and type assert to a string
		username, ok := claims["username"].(string)
		if !ok {
			http.Error(w, "Unauthorized: invalid token claims", http.StatusUnauthorized)
			return
		}

		// Proceed to the next handler with the response writer, request, and params
		r.Header.Set("username", username)
		next(w, r, ps)
	}
}
