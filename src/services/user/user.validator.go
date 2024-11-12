package user

import (
	"be-golang-todo/models"
)

func validateCreateUserRequest(req models.User) map[string]string {
	errors := make(map[string]string)
	if len(*req.Username) == 0 {
		errors["username"] = "Username is required"
	} else if len(*req.Username) < 5 || len(*req.Username) > 20 {
		errors["username"] = "Username must be between 5 and 20 characters"
	}
	if len(*req.Password) == 0 {
		errors["password"] = "password is required"
	} else if len(*req.Username) < 5 || len(*req.Password) > 20 {
		errors["password"] = "Password must be between 5 and 20 characters"
	}
	return errors
}

func validateLoginRequest(req models.User) map[string]string {
	errors := make(map[string]string)
	if len(*req.Username) == 0 {
		errors["username"] = "Username is required"
	}
	if len(*req.Password) == 0 {
		errors["password"] = "password is required"
	}
	return errors
}
