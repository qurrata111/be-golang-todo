package user

import (
	"be-golang-todo/models"
	database "be-golang-todo/src/helper/db"
	"be-golang-todo/src/helper/utils"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
)

func CreateUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req models.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the request data
	errors := validateCreateUserRequest(req)
	if len(errors) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": errors,
		})
		return
	}

	// Check if username is already taken
	var existingUser models.User
	if err := database.DB.QueryRow("SELECT username FROM \"user\" WHERE username = $1", req.Username).Scan(&existingUser.Username); err != sql.ErrNoRows {
		http.Error(w, "Username is already taken", http.StatusConflict)
		return
	}

	// Hash the password before storing it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Insert the new user into the database
	_, err = database.DB.Exec("INSERT INTO \"user\" (username, password) VALUES ($1, $2)", req.Username, hashedPassword)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

func LoginUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req models.User

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the request data
	errors := validateLoginRequest(req)
	if len(errors) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": errors,
		})
		return
	}

	// Retrieve the user from the database by username
	var storedUser models.User
	err := database.DB.QueryRow("SELECT id, username, password FROM \"user\" WHERE username = $1", req.Username).Scan(
		&storedUser.ID, &storedUser.Username, &storedUser.Password)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Compare the provided password with the stored hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(*storedUser.Password), []byte(*req.Password)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT
	token, err := utils.GenerateToken(*req.Username)
	if err != nil {
		http.Error(w, "Failed to generate JWT", http.StatusInternalServerError)
		return
	}

	claims, err := utils.DecodeToken(token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Successfully authenticated
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Login successful",
		"token":   token,
		"decode":  claims["username"].(string),
	})
}
