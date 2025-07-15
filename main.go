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
	"net/http"
	"strconv"

	_ "go-echo-101/docs"

	"github.com/labstack/echo/v4"
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
	e := echo.New()

	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/users", getUsers)
	e.GET("/users/:id", getUserByID)
	e.POST("/users", createUser)
	e.PUT("/users/:id", updateUser)
	e.DELETE("/users/:id", deleteUser)
	e.Start(":8080")
}



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
	id := c.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid user ID"})
	}
	var updateUser User 
	if err := c.Bind(&updateUser); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload"})
	}
	for i, user:= range users {
		if user.ID == userID{
			updateUser.ID = userID
			users[i] = updateUser
			return c.JSON(http.StatusOK, updateUser)
		}
	}
	return c.JSON(http.StatusNotFound, ErrorResponse{Message: "User not found"})
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

	if newUser.Name == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Name is required"})
	}

	users = append(users, newUser)
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
	if len(users) == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Message: "No users found"})
	}
	return c.JSON(http.StatusOK, users)
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
	id := c.Param("id") 
	userID, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid user ID"})
	}
	var checkUser User 
	if err := c.Bind(&checkUser); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload"})
	}
	
	for i, user := range users {
		if user.ID == userID {
			users = append(users[:i], users[i+1:]...)
			return c.JSON(http.StatusOK, map[string]string{"message": "User deleted successfully"})
		}
	}
	return c.JSON(http.StatusNotFound, ErrorResponse{Message: "User not found"})
}


