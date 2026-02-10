package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
)

type APIIndex struct {
	Artists   string `json:"artists"`
	Locations string `json:"locations"`
	Dates     string `json:"dates"`
	Relation  string `json:"relation"`
}

type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Locations    string   `json:"locations"`
	ConcertDates string   `json:"concertDates"`
	Relations    string   `json:"relations"`
}

type LocationData struct {
	Index []struct {
		ID        int      `json:"id"`
		Locations []string `json:"locations"`
		Dates     string   `json:"dates"`
	} `json:"index"`
}

type DateData struct {
	Index []struct {
		ID    int      `json:"id"`
		Dates []string `json:"dates"`
	} `json:"index"`
}

type RelationData struct {
	Index []struct {
		ID             int                 `json:"id"`
		DatesLocations map[string][]string `json:"datesLocations"`
	} `json:"index"`
}

const apiBase = "https://groupietrackers.herokuapp.com/api"

func fetchArtists() ([]Artist, error) {
	resp, err := http.Get(apiBase + "/artists")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var artists []Artist
	if err := json.NewDecoder(resp.Body).Decode(&artists); err != nil {
		return nil, err
	}
	return artists, nil
}

func fetchLocations() (*LocationData, error) {
	resp, err := http.Get(apiBase + "/locations")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data LocationData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func fetchDates() (*DateData, error) {
	resp, err := http.Get(apiBase + "/dates")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data DateData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func fetchRelations() (*RelationData, error) {
	resp, err := http.Get(apiBase + "/relation")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data RelationData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.Handle("/image/", http.StripPrefix("/image/", http.FileServer(http.Dir("./image"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("./templates/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/api/artists", func(w http.ResponseWriter, r *http.Request) {
		artists, err := fetchArtists()
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des artistes", http.StatusInternalServerError)
			return
		}
		query := r.URL.Query().Get("search")
		if query != "" {
			artists = filterArtists(artists, query)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(artists)
	})

	fmt.Println("Serveur lancé sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func filterArtists(artists []Artist, query string) []Artist {
	var result []Artist
	q := strings.ToLower(query)
	for _, a := range artists {
		if strings.Contains(strings.ToLower(a.Name), q) {
			result = append(result, a)
			continue
		}
		for _, m := range a.Members {
			if strings.Contains(strings.ToLower(m), q) {
				result = append(result, a)
				break
			}
		}
	}
	return result
}
