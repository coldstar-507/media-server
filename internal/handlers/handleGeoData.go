package handlers

import (
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
)

// polygon      is [][][]float
// multipolygon is [][][][]float
// one list is redundant
//   [][]float would suffice for polygon
//   [][][]float would suffice for multipolygon

//go:embed assets/my-countries.json
var countriesJSON []byte

type countryname = string

// var countriesMap map[string]any
var countriesData map[countryname]*countrydata
var countries []*country

type countrydata struct {
	Geometry geometry `json:"geometry"`
	Country  *country `json:"country"`
}

type country struct {
	Name string `json:"name"`
	Iso3 string `json:"iso3"`
	Dial string `json:"dial"`
}

type geometry struct {
	Type        string  `json:"type"`
	Coordinates [][]any `json:"coordinates"`
}

type geodata struct {
	// Type     string `json:"type"`
	Features []struct {
		// Type       string `json:"type"`
		// Id         string `json:"id"`
		Properties struct {
			Name      string `json:"name"`
			Continent string `json:"continent"`
			Iso3      string `json:"iso3"`
			Region    string `json:"region"`
			Dial      string `json:"dial"`
			French    string `json:"french_short"`
		} `json:"properties"`
		Geometry geometry `json:"geometry"`
	} `json:"features"`
}

func InitCoords() {
	log.Println("InitCoords")
	var geoData geodata
	err := json.Unmarshal(countriesJSON, &geoData)
	if err != nil {
		log.Fatal("initCoords: error unmarshaling JSON:", err)
	}
	countriesData = make(map[countryname]*countrydata, len(geoData.Features))
	countries = make([]*country, 0, len(geoData.Features))
	for _, x := range geoData.Features {
		// log.Printf("%s, %s, %s\n",
		// 	x.Properties.Iso3, x.Properties.Dial, x.Properties.Name)
		if len(x.Properties.Dial) > 0 {
			country := &country{
				Dial: x.Properties.Dial,
				Iso3: x.Properties.Iso3,
				Name: x.Properties.Name,
			}
			countries = append(countries, country)
			countriesData[x.Properties.Name] = &countrydata{
				Geometry: x.Geometry,
				Country:  country,
			}
		}
	}
}

func HandleGetCoords(w http.ResponseWriter, r *http.Request) {
	countryName := r.PathValue("country")
	if val := countriesData[countryName]; val == nil {
		w.WriteHeader(500)
	} else if err := json.NewEncoder(w).Encode(val.Geometry); err != nil {
		w.WriteHeader(501)
	}
}

func HandleGetCountries(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(countries); err != nil {
		w.WriteHeader(500)
	}
}

func HandleGetCountriesPretty(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "   ")
	if err := enc.Encode(countries); err != nil {
		w.WriteHeader(500)
	}
}
