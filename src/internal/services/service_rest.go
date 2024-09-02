package services

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/timberly/Go_Day03-1/src/internal/models"
)

var secretKey = []byte("school21")

type Store interface {
	GetPlaces(limit int, offset int) ([]models.Restaurant, int, error)
	GetClosest(limit int, lat, lon float64) ([]models.Restaurant, error)
}

type Places struct {
	repo Store
}

func NewPlaces(s Store) *Places {
	return &Places{repo: s}
}

func (p *Places) GetPlaces(limit int, offset int) ([]models.Restaurant, int, error) {
	places, total, err := p.repo.GetPlaces(limit, offset)
	if err != nil {
		slog.Error("Cannot get places")
		return []models.Restaurant{}, 0, err
	}

	return places, total, nil
}

func (p *Places) GetClosest(limit int, lat, lon float64) ([]models.Restaurant, error) {
	places, err := p.repo.GetClosest(limit, lat, lon)
	if err != nil {
		slog.Error("Cannot get clocest places")
		return []models.Restaurant{}, err
	}

	return places, nil
}

func GetToken(c *gin.Context, username string) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": username,
			"exp": time.Now().Add(time.Minute).Unix(),
			"iat": time.Now().Unix(),
		})

	slog.Info("Token claims added")

	tokenString, err := claims.SignedString(secretKey)
	if err != nil {
		slog.Error("Cannot sign the key")
		return "", err
	}

	c.SetCookie("token", tokenString, 60, "/", "127.0.0.1", true, true)

	return tokenString, nil
}

func AuthenticateMiddleware(c *gin.Context) {
	сookieTokenString, err := c.Request.Cookie("token")
	if err != nil {
		slog.Error("Token verification failed:\n", err)
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "Token verification failed"})
		c.Abort()
		return
	}

	tokenString := strings.Split(сookieTokenString.String(), "=")
	token, err := verifyToken(tokenString[1])
	if err != nil {
		slog.Error("Token verification failed:\n", err)
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "Token verification failed"})
		c.Abort()
		return
	}

	slog.Info("Token verified successfully. Claims: %s\\n", token.Claims)

	c.Next()
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}
