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
	"errors"
	"time"
	"strconv"
	"github.com/dgrijalva/jwt-go"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// object atau class User
type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Address string `json:"address"`
	Password string `json:"password"`
}

type UserWithoutPassword struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Address string `json:"address"`
}
type SuccessResponseLogin struct {
	Data    Token   `json:"data"`
	Message string `json:"message"`
}
type Token struct {
	Access_Token  string `json:"accessToken"`
	Refresh_Token string `json:"refreshToken"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}
type LoginUser struct {
	Name         string `json:"name"`
	Password         string `json:"password"`
}

func main() {
	db := connectToDatabase()
	err := CreatedUser(db)
	if err != nil {
		log.Fatal("Error creating table: ", err)
	}
	defer db.Close()
	e := echo.New()
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.POST("/users/login", UserLogin)
	e.POST("/users/refresh-token", ValidateRefreshToken)
	e.POST("/users", createUser)
	group := e.Group("/users")
	group.Use(AuthMiddleware)
	group.GET("", getUsers)
	group.GET("/:id", getUserByID)
	group.PUT("/:id", updateUser)
	group.DELETE("/:id", deleteUser)
	e.Start(":8080")
}

func CreatedUser(db *sql.DB) error {
	query := `
  CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY NOT NULL,
    name VARCHAR(100) DEFAULT '',  -- Kolom ini boleh NULL dan diberi string kosong default
    age INT NOT NULL,
    password VARCHAR(100) NOT NULL,
    address VARCHAR(100) DEFAULT ''  -- Kolom ini boleh NULL dan diberi string kosong default
);
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}
	fmt.Println("Table 'users' created successfully!")
	return nil
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
var jwtKey = []byte("my_secret_key")
func ValidateRefreshToken(c echo.Context) error {
	tokenString := c.Request().Header.Get("Authorization")
	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "authorization header is missing"})
	}
	if tokenString[:7] != "Bearer " {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid token format"})
	}

	tokenString = tokenString[7:]
	fmt.Println("INI AUTH HEADER", tokenString)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})
	fmt.Println("Token String with bearer:", token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "failed to parse token: " + err.Error()})
	}
	if !token.Valid {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "token is not valid"})
	}
	if _, ok := token.Claims.(jwt.MapClaims); !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid token"})
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid token claims"})
	}
	name, ok := claims["name"].(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "name claim is missing"})
	}
	userId, ok := claims["userId"].(float64)
	if !ok {	

		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "userId claim is missing"})
	}	

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name":     name,
		"userId":   userId,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	accessTokenString, err := accessToken.SignedString(jwtKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error generating access token: " + err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"accessToken":  accessTokenString,
	})
}

// UserLogin godoc
// @Summary Endpoint for user login
// @Description Login user with name and password
// @Tags users
// @Accept json
// @Produce json
// @Param user body LoginUser true "User login object"
// @Success 200 {object} SuccessResponseLogin
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /users/login [post]
func UserLogin(c echo.Context) error {
	db := connectToDatabase()
	defer db.Close()
	var loginUser LoginUser
	var token Token
	err := c.Bind(&loginUser)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Login Failed"})
	}

	var storedPassword string
	var userId int
	err = db.QueryRow("SELECT id, password FROM users WHERE name = $1", loginUser.Name).Scan(&userId, &storedPassword)
	if err != nil {
		fmt.Println("Error fetching user:", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Login Failed"})
	}
	fmt.Println("Stored Password:", storedPassword)
	fmt.Println("Login Password:", loginUser.Password)
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(loginUser.Password))
	if err != nil {
		fmt.Println("Error fetching user:", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Login Failed"})
	}
	accessClaimsMap := jwt.MapClaims{
		"userId":   userId,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	accessToken, err := generateJWTToken(accessClaimsMap)
	if err != nil {
		log.Println("Error generating access token:", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Error generating access token"})
	}

	refreshClaims := jwt.MapClaims{
		"userId":   userId,
		"exp":      time.Now().Add(30 * 24 * time.Hour).Unix(), 
	}

	refreshToken, err := generateJWTToken(refreshClaims)
	if err != nil {
		log.Println("Error generating refresh token:", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Error generating refresh token"})
	}

	token.Access_Token = accessToken
	token.Refresh_Token = refreshToken

	return c.JSON(http.StatusOK, SuccessResponseLogin{
		Data:    token,
		Message: "Login successful",
	})
}
func generateJWTToken(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	fmt.Println("JWT KEY", string(jwtKey))
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", fmt.Errorf("error generating token: %v", err)
	}
	return tokenString, nil
}


func ValidateTokenJWT(c echo.Context) (jwt.MapClaims, error) {
	tokenString := c.Request().Header.Get("Authorization")
	if tokenString == "" {
		return nil, errors.New("authorization header is missing")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}


func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, err := ValidateTokenJWT(c)
	if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": err.Error()})
		}
		c.Set("userClaims", claims)
		return next(c)
	}
}



// @Summary Update user by ID
// @Description Update an existing user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body User true "Updated user object"
// @Param Authorization header string true "Bearer <JWT Token>"
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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Error hashing password"})
	}
	_, err = db.Exec("INSERT INTO users (name,age,address,password) values ($1, $2, $3, $4)", newUser.Name, newUser.Age, newUser.Address, hashedPassword)
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
// @Param Authorization header string true "Bearer <JWT Token>"
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
// @Param Authorization header string true "Bearer <JWT Token>"
// @Success 200 {array} UserWithoutPassword
// @Failure 404 {object} ErrorResponse
// @Router /users [get]
func getUsers(c echo.Context) error {
	db := connectToDatabase()
	rows,err := db.Query("Select id, name, age, address from users")
	fmt.Println("Rows")
	if err != nil {
		fmt.Println("Error querying users:", err)
		return err
	}
	defer rows.Close()
	var usersDb []UserWithoutPassword
	for rows.Next(){
		var user UserWithoutPassword
		if err := rows.Scan(&user.ID, &user.Name, &user.Age, &user.Address); err != nil {
			fmt.Println("Error scanning user:", err)
			return err
		}
		usersDb = append(usersDb, user)
	}
	if err = rows.Err(); err != nil {
		fmt.Println("Error reading rows:", err)
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
// @Param Authorization header string true "Bearer <JWT Token>"
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


