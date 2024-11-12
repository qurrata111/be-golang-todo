package task

import (
	"be-golang-todo/models"
)

func validateCreateTaskRequest(req models.Task) map[string]string {
	errors := make(map[string]string)
	if len(*req.Title) == 0 {
		errors["title"] = "Title is required"
	} else if len(*req.Title) < 3 || len(*req.Title) > 255 {
		errors["title"] = "Title must be between 3 and 255 characters"
	}
	if len(*req.Description) == 0 {
		errors["description"] = "Description is required"
	}
	return errors
}
