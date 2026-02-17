package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

const BaseAPIURL = "https://groupietrackers.herokuapp.com/api"

type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
}

type Relation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

type RelationIndex struct {
	Index []Relation `json:"index"`
}

type ArtistComplete struct {
	Artist
	DatesLocations map[string][]string
}

type ErrorData struct {
	Code    int
	Message string
}

type PageData struct {
	Artists []ArtistComplete
	Artist  *ArtistComplete
}

var templates *template.Template

func init() {
	var err error
	templates, err = template.ParseGlob("./templates/*.html")
	if err != nil {
		log.Fatal(err)
	}
}

func fetchJSON(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func getArtists() ([]Artist, error) {
	var artists []Artist
	err := fetchJSON(BaseAPIURL+"/artists", &artists)
	return artists, err
}

func getArtistByID(id int) (*Artist, error) {
	var artist Artist
	err := fetchJSON(fmt.Sprintf("%s/artists/%d", BaseAPIURL, id), &artist)
	return &artist, err
}

func getRelations() (*RelationIndex, error) {
	var relations RelationIndex
	err := fetchJSON(BaseAPIURL+"/relation", &relations)
	return &relations, err
}

func getRelationByID(id int) (*Relation, error) {
	var relation Relation
	err := fetchJSON(fmt.Sprintf("%s/relation/%d", BaseAPIURL, id), &relation)
	return &relation, err
}

func getAllArtists() ([]ArtistComplete, error) {
	artists, err := getArtists()
	if err != nil {
		return nil, err
	}

	relations, err := getRelations()
	if err != nil {
		return nil, err
	}

	relMap := make(map[int]Relation)
	for _, r := range relations.Index {
		relMap[r.ID] = r
	}

	result := make([]ArtistComplete, len(artists))
	for i, a := range artists {
		result[i] = ArtistComplete{Artist: a}
		if rel, ok := relMap[a.ID]; ok {
			result[i].DatesLocations = rel.DatesLocations
		}
	}
	return result, nil
}

func getArtistComplete(id int) (*ArtistComplete, error) {
	artist, err := getArtistByID(id)
	if err != nil {
		return nil, err
	}

	relation, _ := getRelationByID(id)

	complete := &ArtistComplete{Artist: *artist}
	if relation != nil {
		complete.DatesLocations = relation.DatesLocations
	}
	return complete, nil
}

func showError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	templates.ExecuteTemplate(w, "error.html", ErrorData{code, msg})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		showError(w, 404, "Page non trouvée")
		return
	}
	if r.Method != "GET" {
		showError(w, 405, "Méthode non autorisée")
		return
	}

	artists, err := getAllArtists()
	if err != nil {
		showError(w, 500, "Erreur serveur")
		return
	}

	templates.ExecuteTemplate(w, "index.html", PageData{Artists: artists})
}

func artistHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		showError(w, 405, "Méthode non autorisée")
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		showError(w, 400, "ID manquant")
		return
	}

	var id int
	fmt.Sscanf(idStr, "%d", &id)
	if id < 1 {
		showError(w, 400, "ID invalide")
		return
	}

	artist, err := getArtistComplete(id)
	if err != nil {
		showError(w, 404, "Artiste non trouvé")
		return
	}

	templates.ExecuteTemplate(w, "artist.html", PageData{Artist: artist})
}

func fileServer(dir, prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := dir + r.URL.Path[len(prefix)-1:]
		if _, err := os.Stat(path); os.IsNotExist(err) {
			showError(w, 404, "Fichier non trouvé")
			return
		}
		http.StripPrefix(prefix, http.FileServer(http.Dir(dir))).ServeHTTP(w, r)
	})
}

func main() {
	http.Handle("/static/", fileServer("./static", "/static/"))
	http.Handle("/image/", fileServer("./image", "/image/"))
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/artist", artistHandler)

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	fmt.Println("Serveur lancé sur http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
