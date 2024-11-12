package task

import (
	"be-golang-todo/models"
	database "be-golang-todo/src/helper/db"
	config "be-golang-todo/src/helper/redis"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"

	"fmt"

	_ "github.com/lib/pq"
)

func CreateTaskHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req models.Task
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the request data
	errors := validateCreateTaskRequest(req)
	if len(errors) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": errors,
		})
		return
	}

	// Insert into the database
	_, err := database.DB.Exec("INSERT INTO task (title, description, due_date) VALUES ($1, $2, $3)", req.Title, req.Description, req.DueDate)
	if err != nil {
		http.Error(w, "Failed to create todo", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

func GetAllTaskHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	cacheKey := "tasks:all"
	fmt.Printf("request header: %s\n", r.Header.Get("Username"))

	// Check if data is cached in Redis
	cachedtasks, err := config.RDB.Get(config.CTX, cacheKey).Result()
	if err == nil {
		// Cache hit, return cached data
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedtasks))
		return
	}

	// Cache miss, query the database
	rows, err := database.DB.Query("SELECT id, title, description, due_date FROM task")
	if err != nil {
		http.Error(w, "Failed to retrieve tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate); err != nil {
			http.Error(w, "Failed to scan todo", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	// Convert tasks to JSON
	tasksJSON, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, "Failed to encode tasks", http.StatusInternalServerError)
		return
	}

	// Cache result in Redis with a 2-minute expiration
	err = config.RDB.Set(config.CTX, cacheKey, tasksJSON, 30*time.Second).Err()
	if err != nil {
		log.Println("Failed to cache data in Redis:", err)
	}

	// Return the response
	w.Header().Set("Content-Type", "application/json")
	w.Write(tasksJSON)
}

func GetAllTaskPaginationHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse query parameters
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// Default page and limit
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Cache key with filters
	cacheKey := fmt.Sprintf("tasks:%s:%s:%d:%d", status, search, page, limit)
	fmt.Printf("request header: %s\n", r.Header.Get("Username"))

	// Check if data is cached in Redis
	cachedtasks, err := config.RDB.Get(config.CTX, cacheKey).Result()
	if err == nil {
		// Cache hit, return cached data
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedtasks))
		return
	}

	// Build the database query with filters
	query := "SELECT id, title, description, due_date FROM task WHERE 1=1 AND deleted_at IS NULL"
	args := []interface{}{}
	argID := 1

	// Add status filter if provided
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argID)
		args = append(args, status)
		argID++
	}

	// Add search filter if provided
	if search != "" {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argID, argID+1)
		args = append(args, "%"+search+"%", "%"+search+"%")
		argID += 2
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY due_date LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, limit, offset)

	// Query the database
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to retrieve tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate); err != nil {
			http.Error(w, "Failed to scan todo", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	// Get total count for pagination
	var totalTasks int
	countQuery := "SELECT COUNT(*) FROM task WHERE 1=1"
	if status != "" {
		countQuery += " AND status = $1"
	}
	if search != "" {
		countQuery += " AND (title ILIKE $2 OR description ILIKE $3)"
	}
	database.DB.QueryRow(countQuery, args...).Scan(&totalTasks)

	// Calculate total pages
	totalPages := (totalTasks + limit - 1) / limit

	response := map[string]interface{}{
		"tasks": tasks,
		"pagination": map[string]interface{}{
			"current_page": page,
			"total_pages":  totalPages,
			"total_tasks":  totalTasks,
		},
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode tasks", http.StatusInternalServerError)
		return
	}

	// Cache result in Redis with a half-minute expiration
	err = config.RDB.Set(config.CTX, cacheKey, responseJSON, 30*time.Second).Err()
	if err != nil {
		log.Println("Failed to cache data in Redis:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func GetDetailTaskHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi((ps.ByName("id")))

	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
	}

	var task models.Task
	row := database.DB.QueryRow("SELECT id, title, description, status FROM task WHERE ID = $1", id)

	if err := row.Scan(&task.ID, &task.Title, &task.Description, &task.Status); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		fmt.Println(err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var task models.Task
	err = json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	currentTime := time.Now()

	query := `UPDATE task SET title = $1, description = $2, status = $3, updated_at = $4, updated_by = $5 WHERE id = $6`
	res, err := database.DB.Exec(query, task.Title, task.Description, task.Status, currentTime, r.Header.Get("Username"), id)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Todo not found or no changes made", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	currentTime := time.Now()

	query := `UPDATE task SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	res, err := database.DB.Exec(query, currentTime, id)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Task not found or already deleted", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
