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

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

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

	e := echo.New()
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/users", getUsers)
	e.POST("/users", createUser)
	e.GET("/users/:id", getUserByID)
	e.PUT("/users/:id", updateUser)
	e.DELETE("/users/:id", deleteUser)

	e.Start(":8080")
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

// @Summary Update user by ID
// @Description Update user details by user ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body User true "User object"
// @Success 200 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /users/{id} [put]
func updateUser(c echo.Context) error {
	id := c.Param("id")            // ambil parameter id yang akan diupdate
	IDInt, err := strconv.Atoi(id) // konversi string ke integer
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid user ID"})
	}

	var updatedUser User       // buat variabel untuk menyimpan data user yang akan diupdate
	err = c.Bind(&updatedUser) // bind data dari request body ke variabel updatedUser
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload"})
	}

	// cari user dengan ID yang sesuai
	for i, user := range users {
		if user.ID == IDInt {
			users[i] = updatedUser // update data user
			return c.JSON(http.StatusOK, updatedUser)
		}
	}
	return c.JSON(http.StatusNotFound, ErrorResponse{Message: "User not found"})
}

// @Summary Delete user by ID
// @Description Delete user by user ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /users/{id} [delete]
func deleteUser(c echo.Context) error {
	id := c.Param("id")            // ambil parameter id yang akan dihapus
	IDInt, err := strconv.Atoi(id) // konversi string ke integer
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid user ID"})
	}

	// cari user dengan ID yang sesuai
	// jika ditemukan, hapus user tersebut dari slice users
	for i, user := range users {
		if user.ID == IDInt {
			users = append(users[:i], users[i+1:]...) // hapus user dari slice
			// menghapus user dengan cara menggabungkan slice sebelum dan sesudah index yang dihapus
			// sehingga user dengan ID tersebut tidak ada lagi di slice users
			// mengembalikan status 204 No Content sebagai respons
			// karena tidak ada konten yang dikembalikan setelah penghapusan
			return c.NoContent(http.StatusNoContent)
		}
	}
	return c.JSON(http.StatusNotFound, ErrorResponse{Message: "User not found"})
}
