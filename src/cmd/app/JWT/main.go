package main

import (
	"github.com/gin-gonic/gin"

	"github.com/timberly/Go_Day03-1/src/internal/repository"
	"github.com/timberly/Go_Day03-1/src/internal/services"
	"github.com/timberly/Go_Day03-1/src/internal/transport"
)

func main() {
	elastic, err := repository.New([]string{"http://localhost:8888/"}, "places")
	if err != nil {
		return
	}

	router := gin.Default()

	server := transport.NewHandler(elastic)

	router.GET("/api/places", server.GetPlacesHandlerJSON)
	router.GET("/api/recommend", services.AuthenticateMiddleware, server.GetClocestHandlerJSON)
	router.GET("/api/get_token", transport.GetTokenHandler)

	router.Run("127.0.0.1:8888")
}
