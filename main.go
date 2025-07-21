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
	_ "go-echo-101/docs"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// object atau class User
type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Address string `json:"address"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

var users = []User{
	{ID: 1, Name: "John Doe", Age: 30, Address: "123 Main St"},
	{ID: 2, Name: "Jane Smith", Age: 25, Address: "456 Elm St"},
	{ID: 3, Name: "Alice Johnson", Age: 28, Address: "789 Oak St"},
	{ID: 4, Name: "Bob Brown", Age: 35, Address: "101 Pine St"},
}

func main() {
	db := connectToDatabase()
	defer db.Close()
	e := echo.New()
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/users", getUsers)
	e.GET("/users/:id", getUserByID)
	e.POST("/users", createUser)
	e.PUT("/users/:id", updateUser)
	e.DELETE("/users/:id", deleteUser)
	e.Start(":8080")
}

func connectToDatabase() *sql.DB{
	connStr := "user=postgres dbname=postgres password=polkmn1234 host=localhost port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	err = db.Ping()
	if err != nil {
		log.Fatal("Database Connection Failed: ", err)
	}
	fmt.Println("Database Connected Successfully")
	return db
}

// func createUser(db *sql.DB, users User) error {
// 	query := "INSERT INTO public.users (name, age, address) values ($1, $2, $3) RETURNING id"
// 	_, err := db.Exec(query, users.Name, users.Age, users.Address)
// 	return err
// }



// @Summary Update user by ID
// @Description Update an existing user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body User true "Updated user object"
// @Success 200 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /users/{id} [put]
func updateUser(c echo.Context) error {
	db := connectToDatabase()
	defer db.Close()
	id := c.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid user ID"})
	}
	var updateUser User
	if err := c.Bind(&updateUser); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload"})
	}
	query := "UPDATE users SET name = $1, age = $2, address = $3 WHERE id = $4 RETURNING id"
	err = db.QueryRow(query, updateUser.Name, updateUser.Age, updateUser.Address, userID).Scan(&updateUser.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, ErrorResponse{Message: "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to update user"})
	}
	return c.JSON(http.StatusOK, updateUser)
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
	db:=connectToDatabase()
	defer db.Close()

	var newUser User
	err := c.Bind(&newUser)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload"})
	}
	_, err = db.Exec("INSERT INTO users (name,age,address) values ($1, $2, $3)",newUser.Name, newUser.Age, newUser.Address)
	if err != nil {
		log.Fatal("Error inserting user : %v", err)
	}
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
	db := connectToDatabase()
	defer db.Close()

	id := c.Param("id")
	IDInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid user ID"})
	}
	var user User
	query:= "SELECT id, name, age, address FROM users WHERE id = $1"
	err = db.QueryRow(query, IDInt).Scan(&user.ID, &user.Name, &user.Age, &user.Address)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, ErrorResponse{Message: "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to fetch user"})
	}
	return c.JSON(http.StatusOK, user)
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
	db := connectToDatabase()
	rows,err := db.Query("Select * from users")
	if err != nil {
		return err
	}
	defer rows.Close()
	var usersDb []User
	for rows.Next(){
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Age, &user.Address); err != nil {
			return err
		}
		usersDb = append(usersDb, user)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, usersDb)
}

// @Summary Delete user by ID
// @Description Delete a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /users/{id} [delete]
func deleteUser(c echo.Context) error {
	db:=connectToDatabase()
	defer db.Close()

	id := c.Param("id") 
	userID, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid user ID"})
	}
	query := "Delete from users where id = $1"
	result, err := db.Exec(query,userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to Delete User" })
	}
	rowsAffected, err :=  result.RowsAffected()
	fmt.Println("Row Affected",rowsAffected, err)
	if rowsAffected == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Message: "User not found"})
	}
	return c.JSON(http.StatusOK, ErrorResponse{Message: "User successfully deleted"})
}


