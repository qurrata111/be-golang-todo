package test

import (
	"be-golang-todo/src/services/task"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestGetAllTaskPaginateHandler(t *testing.T) {
	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/tasks", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Use httptest to create a response recorder
	rr := httptest.NewRecorder()

	// Set up httprouter and attach the handler to the route
	router := httprouter.New()
	router.GET("/tasks", task.GetAllTaskPaginationHandler)

	// Serve the request
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Optionally, check the response body
	expected := `{
    "pagination": {
        "current_page": 1,
        "total_pages": 0,
        "total_tasks": 0
    },
    "tasks": [
        {
            "ID": 3,
            "Title": "tes",
            "Description": "panjang penjelasannya",
            "Status": null,
            "DueDate": null,
            "CreatedAt": null,
            "CreatedBy": null,
            "UpdatedAt": null,
            "UpdatedBy": null,
            "DeletedAt": null
        },
        {
            "ID": 4,
            "Title": "tes",
            "Description": "panjang penjelasannya UPDATED",
            "Status": null,
            "DueDate": null,
            "CreatedAt": null,
            "CreatedBy": null,
            "UpdatedAt": null,
            "UpdatedBy": null,
            "DeletedAt": null
        }
    ]
}`

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
