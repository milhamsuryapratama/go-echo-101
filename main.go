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
}

type ErrorResponse struct {
	Message string `json:"message"`
}

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
	// connStr := "postgres://milhamsuryapratama:@localhost:5432/postgres?sslmode=disable"
	connStr := "postgres://ighfarhasbiash:@localhost:5432/meeting_rooms_sk?sslmode=disable" // Koneksi ke db di local Hasbi
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

	tx, err := db.Begin() // Mulai transaksi
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}

	// insert data ke dalam tabel user_try dan kembalikan ID yang dihasilkan
	var lastInsertID int
	err = tx.QueryRow("INSERT INTO user_try (name, email) VALUES ($1, $2) RETURNING id", newUser.Name, newUser.Email).Scan(&lastInsertID)
	// Ambil data user yang baru saja dimasukkan
	err = tx.QueryRow("SELECT id, name, email FROM user_try WHERE id = $1", lastInsertID).Scan(&newUser.ID, &newUser.Name, &newUser.Email)

	// Jika terjadi error saat insert atau mengambil data, rollback transaksi
	if err != nil {
		tx.Rollback()                                                                      // Rollback jika terjadi error
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()}) // Kembalikan pesan error ke client
	}

	tx.Commit() // Commit transaksi

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

	// Query untuk mengambil data user berdasarkan ID
	data := db.QueryRow("SELECT id, name, email FROM user_try WHERE id = $1", IDInt)

	var user User
	err = data.Scan(&user.ID, &user.Name, &user.Email) // Scan data ke dalam struct User
	if err != nil {
		// Jika tidak ditemukan, kembalikan 404 Not Found
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, ErrorResponse{Message: "User not found"})
		}
		// Jika terjadi error lain, kembalikan 500 Internal Server Error
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
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
	rows, err := db.Query("SELECT id, name, email FROM user_try")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Error retrieving users"})
	}

	var usersFromDB []User
	for rows.Next() { // Iterasi setiap baris hasil query
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
// @Failure 404 {object} ErrorResponse
// @Failure 400 {object} ErrorResponse
// @Router /users/{id} [put]
func updateUser(c echo.Context) error {
	var updatedUser User
	if err := c.Bind(&updatedUser); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload"})
	}
	id := c.Param("id")
	IDInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid user ID"})
	}
	// Mulai transaksi
	tx, err := db.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	// Update data user berdasarkan ID
	result, err := tx.Exec("UPDATE user_try SET name = $1, email = $2 WHERE id = $3", updatedUser.Name, updatedUser.Email, IDInt)
	// Ambil data user yang baru saja diupdate
	err = tx.QueryRow("SELECT id, name, email FROM user_try WHERE id = $1", IDInt).Scan(&updatedUser.ID, &updatedUser.Name, &updatedUser.Email)

	// Cek apakah ada baris yang terpengaruh
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		// Jika tidak ada baris yang terpengaruh, berarti user tidak ditemukan
		return c.JSON(http.StatusNotFound, ErrorResponse{Message: "User not found"})
	}

	// Jika terjadi error saat update atau mengambil data, rollback transaksi
	if err != nil {
		tx.Rollback() // Rollback jika terjadi error
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	tx.Commit() // Commit transaksi
	return c.JSON(http.StatusOK, updatedUser)
}

// @Summary Delete user by ID
// @Description Delete user by user ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Failure 400 {object} ErrorResponse
// @Router /users/{id} [delete]
func deleteUser(c echo.Context) error {
	id := c.Param("id")
	IDInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid user ID"})
	}

	// Mulai transaksi
	tx, err := db.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}

	// Hapus data user berdasarkan ID
	result, err := tx.Exec("DELETE FROM user_try WHERE id = $1", IDInt)
	// Cek apakah ada baris yang terpengaruh
	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		// Jika tidak ada baris yang terpengaruh, berarti user tidak ditemukan
		return c.JSON(http.StatusNotFound, ErrorResponse{Message: "User not found"})
	}

	// Jika terjadi error saat delete, rollback transaksi
	if err != nil {
		tx.Rollback() // Rollback jika terjadi error
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	tx.Commit() // Commit transaksi

	return c.NoContent(http.StatusNoContent)
}
