package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

var jwtSecret = []byte("your-secret-key") // Replace with an environment variable in production

// GenerateToken generates a JWT token for a given username
func GenerateToken(username string) (string, error) {
	// Define the token expiration time
	expirationTime := time.Now().Add(24 * time.Hour) // Token valid for 24 hours

	// Create JWT claims including standard and custom fields
	claims := jwt.MapClaims{
		"username": username,
		"exp":      expirationTime.Unix(),
	}

	// Create the token with the specified signing method and claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key and return it
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func DecodeToken(tokenString string) (jwt.MapClaims, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token’s signing method is what you expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	// Check if parsing was successful and token is valid
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract the claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return nil, errors.New("failed to parse claims")
}
