package models

type Location struct {
	Lat float64 `json:"lat" es:"lat"`
	Lon float64 `json:"lon" es:"lon"`
}

type Restaurant struct {
	ID      int      `json:"id" es:"id"`
	Name    string   `json:"name" es:"name"`
	Address string   `json:"address" es:"address"`
	Phone   string   `json:"phone" es:"phone"`
	Locat   Location `json:"location" es:"location"`
}


