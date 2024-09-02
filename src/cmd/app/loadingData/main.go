package main

import (
	"github.com/timberly/Go_Day03-1/src/internal/repository"
)

func main() {
	elastic, err := repository.New([]string{"http://localhost:9200/"}, "places")
	if err != nil {
		return
	}

	err = elastic.CreateIndex()
	if err != nil {
		return
	}

	err = elastic.InsertIndex()
	if err != nil {
		return
	}
}
