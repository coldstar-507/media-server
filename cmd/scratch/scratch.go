package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

//go:embed assets/countries.geo.json
var countriesJSON []byte

var geoData GeoData

// type Point struct {
// 	Lat float64 `json:"latitude"`
// 	Lon float64 `json:"longitude"`
// }

// var coords map[string][][]Point
var coords_ map[string]any

type GeoData struct {
	Type     string `json:"type"`
	Features []struct {
		Type       string `json:"type"`
		Id         string `json:"id"`
		Properties struct {
			Name string `json:"name"`
		} `json:"properties"`
		Geometry struct {
			Type        string    `json:"type"`
			Coordinates [][][]any `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
}

func InitCoords() {
	err := json.Unmarshal(countriesJSON, &geoData)
	if err != nil {
		log.Fatal("initCoords: error unmarshaling JSON:", err)
	}
	coords_ = make(map[string]any, len(geoData.Features))
	for _, x := range geoData.Features {
		fmt.Println(x.Properties.Name)
		coords_[x.Properties.Name] = x.Geometry
	}
}

func HandleGetCoords(w http.ResponseWriter, r *http.Request) {
	countryName := r.PathValue("country")
	if val := coords_[countryName]; val == nil {
		w.WriteHeader(500)
	} else if err := json.NewEncoder(w).Encode(val); err != nil {
		w.WriteHeader(501)
	}
}

func main() {
	InitCoords()
	country := "Canada"
	c := coords_[country]
	e, err := json.MarshalIndent(c, "", "   ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(e))
}

// func main() {
// 	err := json.Unmarshal(countriesJSON, &geoData)
// 	if err != nil {
// 		log.Fatal("Error unmarshaling JSON:", err)
// 	}

// 	// coords = make(map[string][][]Point, len(geoData.Features))
// 	for _, x := range geoData.Features {
// 		fmt.Println(x.Properties.Name)
// 		// fmt.Println(x.Geometry.Coordinates[0][0]...)
// 		// polys := make([][]Point, 0, len(x.Geometry.Coordinates))
// 		// for _, poly := range x.Geometry.Coordinates {
// 		// 	pol := make([]Point, 0, len(poly))
// 		// 	for _, point := range poly {
// 		// 		p := Point{Lat: point[0].(float64), Lon: point[1].(float64)}
// 		// 		pol = append(pol, p)
// 		// 	}
// 		// 	polys = append(polys, pol)
// 		// }
// 		coords_[x.Properties.Name] = x.Geometry.Coordinates
// 	}
// }
