package main

import (
	database "be-golang-todo/src/helper/db"
	config "be-golang-todo/src/helper/redis"
	"be-golang-todo/src/middlewares"
	"be-golang-todo/src/services/task"
	"be-golang-todo/src/services/user"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"

	_ "github.com/lib/pq"
)

// Initialize PostgreSQL and Redis clients
func init() {
	database.Init()
	log.Println("Connected to postgresql")

	config.InitRedis()
	log.Println("Connected to redis")
}

func main() {
	router := httprouter.New()
	router.POST("/login", user.LoginUserHandler)
	router.POST("/register", user.CreateUserHandler)
	// router.GET("/tasks/all", middlewares.ProtectedHandler(todo.GetAllTodosHandler))
	router.GET("/tasks", middlewares.ProtectedHandler(task.GetAllTaskPaginationHandler))
	router.GET("/tasks/:id", middlewares.ProtectedHandler(task.GetDetailTaskHandler))
	router.PUT("/tasks/:id", middlewares.ProtectedHandler(task.UpdateTaskHandler))
	router.DELETE("/tasks/:id", middlewares.ProtectedHandler(task.DeleteTaskHandler))
	router.POST("/tasks", middlewares.ProtectedHandler(task.CreateTaskHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server is running on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
