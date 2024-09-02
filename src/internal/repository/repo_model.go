package repository

import (
	"github.com/timberly/Go_Day03-1/src/internal/models"
)

type searchRequestParams struct {
	Took    float64 `json:"took"`
	Timeout bool    `json:"timed_out"`
	Shards  struct {
		Total      int64 `json:"total"`
		Successful int64 `json:"successful"`
		Skipped    int64 `json:"skipped"`
		Failed     int64 `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int64  `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			Index  string             `json:"_index"`
			Id     string             `json:"_id"`
			Score  float64            `json:"_score"`
			Source *models.Restaurant `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
