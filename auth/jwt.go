package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

var SecretKey = []byte("my_secret_key")

func ValidateRefreshToken(c echo.Context) error {
	tokenString := c.Request().Header.Get("Authorization")
	if tokenString == "" {
		return errors.New("authorization header is missing")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return SecretKey, nil
	})
	if err != nil {
		return errors.New("failed to parse token: " + err.Error())
	}

	if !token.Valid {
		return errors.New("invalid token")
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "milhamsuryapratama",
		"email":    "ilham@gmail.com",
		"exp":      time.Now().Add(1 * time.Minute).Unix(),
	})

	accessTokenStr, err := accessToken.SignedString(SecretKey)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"accessToken": accessTokenStr,
	})
}

func GenerateTokenJWT(c echo.Context) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "milhamsuryapratama",
		"email":    "ilham@gmail.com",
		"role":     "admin",
		"exp":      time.Now().Add(1 * time.Minute).Unix(),
	})

	accessToken, err := token.SignedString(SecretKey)
	if err != nil {
		return err
	}

	refreshTokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})

	refreshToken, err := refreshTokenClaims.SignedString(SecretKey)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func ValidateTokenJWT(c echo.Context) error {
	tokenString := c.Request().Header.Get("Authorization")
	if tokenString == "" {
		return errors.New("authorization header is missing")
	}

	// fmt.Println("Token String with bearer:", tokenString)

	// Remove "Bearer " prefix if present
	// if tokenString[:7] != "Bearer " {
	// 	return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid token format"})
	// }

	// tokenString = tokenString[7:]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return SecretKey, nil
	})
	if err != nil {
		return errors.New("failed to parse token: " + err.Error())
	}

	if !token.Valid {
		return errors.New("invalid token")
	}

	if user, ok := token.Claims.(jwt.MapClaims); ok {
		c.Set("user", user)
		return nil
	}

	return errors.New("invalid token")
}

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := ValidateTokenJWT(c); err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": err.Error()})
		}
		return next(c)
	}
}

func ValidateAdminRole(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("admin").(jwt.MapClaims)
		if user["role"].(string) != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "forbidden"})
		}
		return next(c)
	}
}

func ValidateUserRole(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(jwt.MapClaims)
		if user["role"].(string) != "user" {
			return c.JSON(http.StatusForbidden, map[string]string{"message": "forbidden"})
		}
		return next(c)
	}
}
