package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type User struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// create db is not exist
func initDB() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}
}

// INTERFACE LAYER, FACILITATING COMMUNICATION BETWEEN DIFFERENT COMPONENTS IN THE SYSTEM
func routeRest(router *gin.Engine) {
	router.GET("/users", getUsersHandler)
	router.GET("/users/:id", getUserHandler)
	router.POST("/users", createUserHandler)
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "users.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize database
	initDB()

	router := gin.Default()

	// set rest route
	routeRest(router)

	port := ":6001"
	log.Printf("Starting user service. PORT: %s\n", port)
	router.Run(port)
}

// =========== INTERFACE HANDLER, HANDLING REQUEST RESPONSE API DEPEND INTERFACE ===========

// handler request response list users
func getUsersHandler(c *gin.Context) {
	pageNum, err := strconv.Atoi(c.DefaultQuery("page_num", "1"))
	if err != nil {
		log.Println("error handler: code error 008, ", "Invalid page_num param")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page_num param"})
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		log.Println("error handler: code error 007, ", "Invalid page_size param")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page_size param"})
		return
	}

	users, err := getUsersUsecase(pageNum, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": true, "users": users})
}

// handler request response detail user
func getUserHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println("error handler: code error 006, ", "Invalid user ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	users, err := getUserUsecase(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": true, "user": users})
}

// handler request response create user
func createUserHandler(c *gin.Context) {
	var body User
	if err := c.ShouldBind(&body); err != nil {
		log.Println("error handler: code error 005, ", "Invalid body request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body request"})
		return
	}

	user, err := createUserUsecase(body.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"result": true, "user": user})
}

// =========== USECASE LAYER, SERVES AS AN INTERMEDIARY BETWEEN THE PRESENTATION LAYER AND THE DATA LAYER ===========

// get list data user by params
func getUsersUsecase(pageNum, pageSize int) ([]User, error) {
	// call users find repository
	users, err := find(pageNum, pageSize)
	if err != nil {
		return nil, errors.New("database error: get list users error database")
	}

	return users, err
}

// get detail data user by id
func getUserUsecase(userID int) (*User, error) {
	// call users find repository
	user, err := findByID(userID)
	if err != nil {
		return nil, errors.New("database error: get detail user error database")
	}

	return user, err
}

// create user
func createUserUsecase(name string) (*User, error) {
	// call users find repository
	user, err := create(name)
	if err != nil {
		return nil, errors.New("database error: create user error database")
	}

	return user, err
}

// =========== REPOSITORY LAYER, ABSTRACTION OVER THE DATA PERSISTENCE (databases, file systems, or external APIs) ===========

// Function to get list users data
func find(pageNum, pageSize int) ([]User, error) {
	// set offset position
	offset := (pageNum - 1) * pageSize

	rows, err := db.Query("SELECT id, name, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?", pageSize, offset)
	if err != nil {
		log.Println("error handler: code error 004, ", err)
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.CreatedAt, &user.UpdatedAt); err != nil {
			log.Println("error handler: code error 003, ", err)
			return nil, err
		}
		users = append(users, user)
	}

	return users, err
}

// Function to get user by id
func findByID(id int) (*User, error) {
	var user User
	err := db.QueryRow("SELECT id, name, created_at, updated_at FROM users WHERE id = ?", id).Scan(&user.ID, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		log.Println("error handler: code error 002, ", err)
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}

		return nil, err
	}

	return &user, nil
}

// Function to create user
func create(name string) (*User, error) {
	var user User
	user.Name = name
	user.CreatedAt = time.Now().UnixNano() / int64(time.Microsecond)
	user.UpdatedAt = user.CreatedAt

	result, err := db.Exec("INSERT INTO users (name, created_at, updated_at) VALUES (?, ?, ?)", user.Name, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		log.Println("error handler: code error 001, ", err)
		return nil, err
	}

	userID, _ := result.LastInsertId()
	user.ID = int(userID)

	return &user, nil
}
