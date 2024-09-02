package main

import (
	"github.com/gin-gonic/gin"

	"github.com/timberly/Go_Day03-1/src/internal/repository"
	"github.com/timberly/Go_Day03-1/src/internal/transport"
)

func main() {
	elastic, err := repository.New([]string{"http://localhost:8888/"}, "places")
	if err != nil {
		return
	}

	router := gin.Default()

	server := transport.NewHandler(elastic)

	router.GET("/api/recommend", server.GetClocestHandlerJSON)

	router.Run("127.0.0.1:8888")
}
