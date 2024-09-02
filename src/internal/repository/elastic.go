package repository

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/timberly/Go_Day03-1/src/internal/models"
)

var errorIncorCoord error = errors.New("incorrectCoordinate")

const (
	mapping = `
	{
		"mappings":{
			"properties": {
				"name": {
						"type":  "text"
				},
				"address": {
						"type":  "text"
				},
				"phone": {
						"type":  "text"
				},
				"location": {
					"type": "geo_point"
				}
			}
		}
	}
	`

	settings = `{ "index.max_result_window" : 20000
	}`
)

type ElasticSearch struct {
	client *es.Client
	index  string
}

func New(addresses []string, index string) (*ElasticSearch, error) {
	cfg := es.Config{
		Addresses: addresses,
	}

	client, err := es.NewClient(cfg)
	if err != nil {
		slog.Error("Error creating client: ", err)
		return nil, err
	}

	slog.Info("Create client")

	return &ElasticSearch{
		client: client,
		index:  index,
	}, nil
}

func (e *ElasticSearch) CreateIndex() error {
	res, err := e.client.Indices.Exists([]string{e.index})
	if err != nil {
		slog.Error("Cannot check index existence: ", err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		slog.Info("Index already exists")
		e.Info()
		return nil
	}
	if res.StatusCode != 404 {
		slog.Error("Error in index existence response")
		return err
	}

	_, err = e.client.Indices.Create(
		e.index,
		e.client.Indices.Create.WithBody(strings.NewReader(mapping)),
	)
	if err != nil {
		slog.Error("Cannot create index: ", err)
		return err
	}
	e.putSettings()
	e.Info()

	slog.Info("Create index")

	return nil
}

func (e *ElasticSearch) putSettings() error {
	res, err := esapi.IndicesPutSettingsRequest{
		Index: []string{e.index},  Body:  strings.NewReader(settings),
	}.Do(context.Background(), e.client.Transport)
	if err != nil {
		slog.Error("Cannot set settings: ", err)
		return err
	}
	defer res.Body.Close()

	return nil
}

func (e *ElasticSearch) Info() error {
	res, err := e.client.Info()
	if err != nil {
		slog.Error("Error creating client: ", err)
		return err
	}
	defer res.Body.Close()
	fmt.Println(res)

	slog.Info("Get info")

	return nil
}

func (e *ElasticSearch) InsertIndex() error {
	restaurants, err := readCSV("../materials/data.csv")
	if err != nil {
		slog.Error("Cannot import data: ", err)
		return err
	}

	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: e.client,
		Index:  e.index,
	})
	if err != nil {
		slog.Error("Error creating the indexer: ", err)
		return err
	}

	start := time.Now().UTC()
	var countSuccessful uint64
	for i, rest := range restaurants {
		indexBulk(rest, bi, countSuccessful, i)
	}

	if err := bi.Close(context.Background()); err != nil {
		slog.Error("Unexpected error: ", err)
		return err
	}

	biStats := bi.Stats()
	dur := time.Since(start)
	printInfo(biStats, dur)

	return nil
}

func indexBulk(rest models.Restaurant, bi esutil.BulkIndexer, countSuccessful uint64, index int) error {
	data, err := json.Marshal(rest)
	if err != nil {
		fmt.Println("Cannot encode article ", err)
		return err
	}

	err = bi.Add(
		context.Background(),
		esutil.BulkIndexerItem{
			Action:     "index",
			DocumentID: strconv.Itoa(index),
			Body:       bytes.NewReader(data),

			OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
				atomic.AddUint64(&countSuccessful, 1)
			},

			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
				if err != nil {
					log.Printf("ERROR: %s", err)
				} else {
					log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
				}
			},
		},
	)
	if err != nil {
		fmt.Println("Unexpected error: ", err)
		return err
	}

	return nil
}

func printInfo(biStats esutil.BulkIndexerStats, dur time.Duration) {
	if biStats.NumFailed > 0 {
		fmt.Printf(
			"Indexed [%s] documents with [%s] errors in %s (%s docs/sec)",
			humanize.Comma(int64(biStats.NumFlushed)),
			humanize.Comma(int64(biStats.NumFailed)),
			dur.Truncate(time.Millisecond),
			humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed))),
		)
	} else {
		fmt.Printf(
			"Sucessfuly indexed [%s] documents in %s (%s docs/sec)\n",
			humanize.Comma(int64(biStats.NumFlushed)),
			dur.Truncate(time.Millisecond),
			humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed))),
		)
	}
}

func (e *ElasticSearch) GetPlaces(limit int, offset int) ([]models.Restaurant, int, error) {
	res, err := e.client.Search(
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(e.index),
		e.client.Search.WithSort("id"),
		e.client.Search.WithSize(limit),
		e.client.Search.WithFrom(offset),
		e.client.Search.WithTrackTotalHits(true),
		e.client.Search.WithPretty(),
	)
	if err != nil {
		slog.Error("!Unexpected error: ", err)
		return nil, 0, err
	}
	defer res.Body.Close()

	var result searchRequestParams
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		slog.Error("Unexpected error: ", err)
		return nil, 0, err
	}

	var places []models.Restaurant
	for _, hit := range result.Hits.Hits {
		places = append(places, *hit.Source)
	}

	return places, int(result.Hits.Total.Value), nil
}


func (e *ElasticSearch) GetClosest(limit int, lat, lon float64) ([]models.Restaurant, error) {
	query := `{
    "query": {
	    "match_all": {}
    },
    "sort": [
      {
        "_geo_distance": {
          "location": {
            "lat": %f,
            "lon": %f
          },
          "order": "asc",
          "unit": "km",
          "mode": "min",
          "distance_type": "arc",
          "ignore_unmapped": true
        }
      }
    ]
  }
	`
	query = fmt.Sprintf(query, lat, lon)

	res, err := e.client.Search(
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(e.index),
		e.client.Search.WithBody(strings.NewReader(query)),
		e.client.Search.WithSize(limit),
		e.client.Search.WithTrackTotalHits(true),
		e.client.Search.WithPretty(),
	)
	if err != nil {
		slog.Error("!Unexpected error: ", err)
		return nil, err
	}
	defer res.Body.Close()

	var result searchRequestParams
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		slog.Error("Unexpected error: ", err)
		return nil, err
	}

	var places []models.Restaurant
	for _, hit := range result.Hits.Hits {
		places = append(places, *hit.Source)
	}

	return places, nil
}


func readCSV(filename string) ([]models.Restaurant, error) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Cannot open file: ", err)
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t'

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading records")
		return nil, err
	}

	var Restaurants []models.Restaurant
	for i, elem := range records {
		if i > 0 {
			id, err := strconv.Atoi(elem[0])
			name := elem[1]
			address := elem[2]
			phone := elem[3]
			lon, err1 := strconv.ParseFloat(elem[4], 64)
			lat, err2 := strconv.ParseFloat(elem[5], 64)
			if err != nil || err1 != nil || err2 != nil {
				fmt.Println("Incorrect coordinate")
				return nil, errorIncorCoord
			}

			Restaurants = append(Restaurants, models.Restaurant{
				ID:      id + 1,
				Name:    name,
				Address: address,
				Phone:   phone,
				Locat: models.Location{
					Lat: lat,
					Lon: lon,
				},
			})
		}
	}

	return Restaurants, nil
}
