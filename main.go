package main

// @title Echo API
// @version 1.0
// @description This is a sample Echo server with user management endpoints.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// openssl genrsa -out new_private.pem 2048

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"go-echo-101/auth"
	_ "go-echo-101/docs"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq" // PostgreSQL driver
	echoSwagger "github.com/swaggo/echo-swagger"
)

// object atau class User
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	// Age     int    `json:"age"`
	// Address string `json:"address"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

var users = []User{}
var db *sql.DB

func main() {
	db = connectToDatabase()
	redis := connectToRedis()

	e := echo.New()

	e.GET("/generate-token", auth.GenerateTokenJWT) // login
	e.GET("/validate-token", auth.ValidateTokenJWT)
	e.GET("/refresh-token", auth.ValidateRefreshToken)
	e.POST("/uploads", upload)

	group := e.Group("/admin")
	group.Use(auth.AuthMiddleware)
	group.Use(auth.ValidateAdminRole)

	user := e.Group("/user")
	user.Use(auth.AuthMiddleware)
	user.Use(auth.ValidateUserRole)

	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	group.GET("/swagger/*", echoSwagger.WrapHandler)
	group.GET("/users", getUsers)
	group.GET("/users/:id", getUserByID)
	group.POST("/users", createUser)

	go startRedisConsumer()

	e.Start(":8000")
}

func startRedisConsumer() {
	for {
		redis.BRpop(ctx, 0, queuname).Result()
	}
}

func upload(c echo.Context) error {
	fmt.Println("start upload")
	// http://localhost:8000/temp/Screenshot%202025-07-19%20at%2015.07.02.png
	// http://localhost:8000/uploads/Screenshot%202025-07-19%20at%2015.07.02.png

	// c.Bind(&request)
	// if strings.Contains(request.ImageURL, "temp") {
	// 	// Handle the case where the image is in the temp directory
	// }

	// file, err := os.Open(request.ImageURL) // Opens "example.txt" for reading
	// if err != nil {
	// 	log.Fatal("err", err) // Handle potential errors (e.g., file not found)
	// }
	// defer file.Close()

	// filename := filepath.Base(file.Name())

	// Destination
	// dst, err := os.Create("uploads/" + filename) // Creates "example_copy.txt" for writing
	// if err != nil {
	// 	log.Fatal("err", err)
	// }
	// defer dst.Close()

	// // Copy
	// if _, err = io.Copy(dst, file); err != nil {
	// 	return err
	// }

	// os.Remove("temp/" + filename) // Remove the file from the temp directory
	// os.Remove("uploads/" + filename)

	// file.
	// file, err := c.FormFile("file")
	// if err != nil {
	// 	return err
	// }

	// fmt.Println("Uploaded file:", file.Filename)
	// fmt.Println("File size:", file.Size)

	// src, err := file.Open()
	// if err != nil {
	// 	return err
	// }
	// defer src.Close()

	// // Destination
	// dst, err := os.Create("temp/" + file.Filename)
	// if err != nil {
	// 	return err
	// }
	// defer dst.Close()

	// // Copy
	// if _, err = io.Copy(dst, src); err != nil {
	// 	return err
	// }
	return nil
}

func connectToDatabase() *sql.DB {
	connStr := "postgres://milhamsuryapratama:@localhost:5432/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// @Summary Endpoint create a new user
// @Description Create a new user with name, age, and address
// @Tags users
// @Accept json
// @Produce json
// @Param user body User true "User object"
// @Success 201 {object} User
// @Failure 400 {object} ErrorResponse
// @Router /users [post]
func createUser(c echo.Context) error {
	var newUser User
	err := c.Bind(&newUser)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload"})
	}

	redis.LPush(ctx, "queuname", newUser)

	tx, err := db.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}

	_, err = tx.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", "Test Transaction 3", "Test Transaction 3")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}

	_, err = tx.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", "Test Transaction 3", "Test Transaction 3")
	if err != nil {
		tx.Rollback()
		fmt.Println("Error inserting user:", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}

	tx.Commit()
	fmt.Println("User created successfully:", newUser)

	return c.JSON(http.StatusCreated, newUser)
}

// @Summary Get user by ID
// @Description Get user details by user ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} User
// @Failure 404 {object} ErrorResponse
// @Failure 400 {object} ErrorResponse
// @Router /users/{id} [get]
func getUserByID(c echo.Context) error {
	id := c.Param("id")
	IDInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid user ID"})
	}

	for _, user := range users {
		if user.ID == IDInt {
			return c.JSON(http.StatusOK, user)
		}
	}
	return c.JSON(http.StatusNotFound, ErrorResponse{Message: "User not found"})
}

// @Summary Get all users
// @Description Retrieve a list of all users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} User
// @Failure 404 {object} ErrorResponse
// @Router /users [get]
func getUsers(c echo.Context) error {
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Error retrieving users"})
	}
	// defer db.Close()
	defer rows.Close()

	var usersFromDB []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Error scanning user data"})
		}
		usersFromDB = append(usersFromDB, user)
	}
	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Error processing user data"})
	}

	return c.JSON(http.StatusOK, usersFromDB)
}
