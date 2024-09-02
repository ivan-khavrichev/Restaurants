package transport

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/timberly/Go_Day03-1/src/internal/models"
	"github.com/timberly/Go_Day03-1/src/internal/services"
)

type StoreHandler interface {
	GetPlaces(limit int, offset int) ([]models.Restaurant, int, error)
	GetClosest(limit int, lat, lon float64) ([]models.Restaurant, error)
}

type Handler struct {
	handler StoreHandler
}

func NewHandler(h StoreHandler) *Handler {
	return &Handler{handler: h}
}

func (h *Handler) GetPlaces(limit int, offset int) ([]models.Restaurant, int, error) {
	places, total, err := h.handler.GetPlaces(limit, offset)
	if err != nil {
		slog.Error("Cannot get places")
		return []models.Restaurant{}, 0, err
	}

	return places, total, nil
}

func (h *Handler) GetClosest(limit int, lat, lon float64) ([]models.Restaurant, error) {
	places, err := h.handler.GetClosest(limit, lat, lon)
	if err != nil {
		slog.Error("Cannot get clocest places")
		return []models.Restaurant{}, err
	}

	return places, nil
}

func (h *Handler) GetPlacesHandlerHTML(c *gin.Context) {
	page, err := strconv.Atoi(c.Query("page"))
	places, total, err1 := h.handler.GetPlaces(10, (page-1)*10)
	pageLast := total / 10
	if err != nil || err1 != nil || page < 1 || page > pageLast {
		c.String(http.StatusBadRequest, "Invalid 'page' value: 'foo'")
	} else {
		pagePrev := page - 1
		pageNext := page + 1
		if page == 1 {
			pagePrev = 0
		} else if page == pageLast {
			pageNext = 0
		}
		c.HTML(http.StatusOK, "index.html", gin.H{
			"total":       total,
			"restaurants": places,
			"prev":        pagePrev,
			"next":        pageNext,
			"last":        pageLast,
		},
		)
	}
}

func (h *Handler) GetPlacesHandlerJSON(c *gin.Context) {
	page, err := strconv.Atoi(c.Query("page"))
	places, total, err1 := h.handler.GetPlaces(10, (page-1)*10)
	pageLast := total / 10
	if err != nil || err1 != nil || page < 1 || page > pageLast {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid 'page' value: 'foo'"})
	} else {
		pagePrev := page - 1
		pageNext := page + 1
		if page == 1 {
			pagePrev = 0
		} else if page == pageLast {
			pageNext = 0
		}
		var result = APIjson{
			Name:   "PLaces",
			Total:  total,
			Places: places,
			Prev:   pagePrev,
			Next:   pageNext,
			Last:   pageLast,
		}
		c.IndentedJSON(http.StatusOK, result)
	}
}

func (h *Handler) GetClocestHandlerJSON(c *gin.Context) {
	lat, err1 := strconv.ParseFloat(c.Query("lat"), 64)
	lon, err2 := strconv.ParseFloat(c.Query("lon"), 64)
	places, err := h.handler.GetClosest(3, lat, lon)
	if err != nil || err1 != nil || err2 != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid 'recommend' value: 'foo'"})
	} else {
		c.IndentedJSON(http.StatusOK, gin.H{
			"name":   "Recommendation",
			"places": places,
		},
		)
	}
}

func GetTokenHandler(c *gin.Context) {
	username := "user"
	password := "qwerty"
	if username == "user" && password == "qwerty" {
		tokenString, err := services.GetToken(c, username)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Cannot create token"})
			return
		} else {
			c.IndentedJSON(http.StatusOK, gin.H{
				"token": tokenString,
			},
			)
		}
	} else {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	}
}
