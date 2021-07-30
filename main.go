package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var tpl = template.Must(template.ParseFiles("index.html"))

var apiKey *string

type Search struct {
	Status int
	SearchKey string
	Response Response

}

type Response struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Visibility int `json:"visibility"`
	Wind       struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Dt  int `json:"dt"`
	Sys struct {
		Type    int     `json:"type"`
		ID      int     `json:"id"`
		Message float64 `json:"message"`
		Country string  `json:"country"`
		Sunrise int     `json:"sunrise"`
		Sunset  int     `json:"sunset"`
	} `json:"sys"`
	Timezone int    `json:"timezone"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
//	Cod      int    `json:"cod"`
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request)  {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	params := u.Query()
	searchKey := params.Get("q")

	search := &Search{}
	search.SearchKey = searchKey

	endpoint := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&units=metric&appid=%s", url.QueryEscape(search.SearchKey), *apiKey)
	resp, err := http.Get(endpoint)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println(resp.StatusCode)

	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		search.Status = 404
	} else if resp.StatusCode != 200 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}



	err = json.NewDecoder(resp.Body).Decode(&search.Response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, search)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func (r *Response) SunriseTime() string {
	var UnixTimestamp int64 = int64(r.Sys.Sunrise)
	tm := time.Unix(UnixTimestamp, 0)
	return fmt.Sprintf("%s", tm.Format("15:04"))
}

func (r *Response) SunsetTime() string {
	var UnixTimestamp int64 = int64(r.Sys.Sunset)
	tm := time.Unix(UnixTimestamp, 0)
	return fmt.Sprintf("%s", tm.Format("15:04"))
}

func main() {
	//d4cfac109eda02b68c597b7defd4cbf9

	apiKey = flag.String("apikey", "", "https://openweathermap.org")
	flag.Parse()

	if *apiKey == "" {
		log.Fatal("apikey must me entered by flag")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("assets"))

	mux.HandleFunc("/", indexHandler)
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	mux.HandleFunc("/search/", searchHandler)

	http.ListenAndServe(":"+port, mux)
}
