package auth

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

var jwtSecret = []byte("Xy123@@@") // untuk signing JWT

// Struct untuk request body login
type ReqLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Struct untuk request body refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// @Summary Generate JWT
// @Description Generate access token and refresh token for user login
// @Tags auth
// @Accept json
// @Produce json
// @Param req_login body ReqLogin true "ReqLogin object"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /login [post]
func GenerateJWT(c echo.Context) error {
	// cek body request apakah ada data yang dikirim
	if c.Request().Body == http.NoBody {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Request body is empty"})
	}
	var req_login ReqLogin
	// Bind request body ke struct req
	if err := c.Bind(&req_login); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
	}

	// validasi email dan password, cek ke database

	// generate token JWT dengan klaim yang diperlukan
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": req_login.Email,
		"role":  "admin",                                // role bisa disesuaikan dengan db nanti
		"user":  "ighfarhasbiash",                       // username bisa disesuaikan dengan db nanti
		"type":  "access",                               // untuk membedakan token akses dan refresh
		"exp":   time.Now().Add(time.Minute * 1).Unix(), // token berlaku selama 1 menit
		"iat":   time.Now().Unix(),                      // waktu token dibuat
	})
	// generate akses token JWT
	accessToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return err
	}

	// generate token refresh JWT dengan klaim yang diperlukan
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": req_login.Email,
		"user":  "ighfarhasbiash",                      // username bisa disesuaikan dengan db nanti
		"role":  "admin",                               // role bisa disesuaikan dengan db nanti
		"type":  "refresh",                             // untuk membedakan token akses dan refresh
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // token berlaku selama 24 jam
		"iat":   time.Now().Unix(),                     // waktu token dibuat
	})
	// generate refresh token JWT
	refreshTokenString, err := refreshToken.SignedString(jwtSecret)
	if err != nil {
		return err
	}

	// kembalikan akses token dan refresh token
	return c.JSON(http.StatusOK, map[string]string{
		"accessToken":  accessToken,
		"refreshToken": refreshTokenString,
	})

}

// Validasi JWT dengan middleware
func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Ambil token JWT dari header Authorization
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.ErrUnauthorized
		}
		tokenString := authHeader[len("Bearer "):] // Mengambil token setelah "Bearer "

		// Validasi token JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			// cek dulu apakah token menggunakan signing method yang benar
			// dalam hal ini kita menggunakan HMAC SHA-256
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, echo.ErrUnauthorized
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Invalid or expired token",
			})
		}

		// cek klaim token JWT apakah tipe token adalah "access"?
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["type"] != "access" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Invalid access token",
			})
		}

		// Jika token valid, lanjutkan ke handler
		return next(c)
	}
}

// @Summary Refresh Access Token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refreshToken body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /refresh_token [post]
func RefreshAccessToken(c echo.Context) error {
	// Ambil refresh token dari body request
	var req RefreshTokenRequest
	// Bind request body ke struct req
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	// Validasi refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, echo.ErrUnauthorized
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid or expired refresh token"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid token claims"})
	}

	// cek klaim token JWT apakah tipe token adalah "refresh"?
	if claims["type"] != "refresh" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid refresh token"})
	}

	// Cek apakah refresh token sudah expired
	exp, ok := claims["exp"].(float64)
	if !ok || int64(exp) < time.Now().Unix() {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Refresh token expired, please login again"})
	}

	// Generate access token baru
	newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": claims["email"],
		"user":  claims["user"],
		"role":  claims["role"],
		"type":  "access", // untuk membedakan token akses dan refresh
		"exp":   time.Now().Add(time.Minute * 1).Unix(),
		"iat":   time.Now().Unix(),
	})
	accessTokenString, err := newAccessToken.SignedString(jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to generate access token"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"accessToken": accessTokenString,
	})
}
