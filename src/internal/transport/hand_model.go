package transport

import (
	"github.com/timberly/Go_Day03-1/src/internal/models"
)

type APIjson struct {
	Name   string              `json:"name"`
	Total  int                 `json:"total"`
	Places []models.Restaurant `json:"places"`
	Prev   int              `json:"prev_page"`
	Next   int              `json:"next_page"`
	Last   int              `json:"last_page"`
}
