package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
)

type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
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
		query := r.URL.Query().Get("search")
		artists := getArtists(query)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(artists)
	})

	fmt.Println("Serveur lancé sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func getArtists(search string) []Artist {
	allArtists := []Artist{
		{ID: 1, Name: "Queen", Members: []string{"Freddie Mercury", "Brian May"}, CreationDate: 1970},
		{ID: 2, Name: "SOJA", Members: []string{"Jacob Hemphill"}, CreationDate: 1997},
	}
	return allArtists
}
