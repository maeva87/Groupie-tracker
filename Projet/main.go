package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	BaseAPIURL        = "https://groupietrackers.herokuapp.com/api"
	ArtistsEndpoint   = BaseAPIURL + "/artists"
	LocationsEndpoint = BaseAPIURL + "/locations"
	DatesEndpoint     = BaseAPIURL + "/dates"
	RelationEndpoint  = BaseAPIURL + "/relation"
)

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

type LocationsData struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
	DatesURL  string   `json:"dates"`
}

type LocationsIndex struct {
	Index []LocationsData `json:"index"`
}

type DatesData struct {
	ID    int      `json:"id"`
	Dates []string `json:"dates"`
}

type DatesIndex struct {
	Index []DatesData `json:"index"`
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
	LocationsList  []string            `json:"locationsList"`
	DatesList      []string            `json:"datesList"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

type APIResponse struct {
	Artists   string `json:"artists"`
	Locations string `json:"locations"`
	Dates     string `json:"dates"`
	Relation  string `json:"relation"`
}

type APIClient struct {
	httpClient *http.Client
	cache      map[string]cacheEntry
	cacheMutex sync.RWMutex
	cacheTTL   time.Duration
}

type cacheEntry struct {
	data      []byte
	timestamp time.Time
}

func NewAPIClient() *APIClient {
	return &APIClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache:    make(map[string]cacheEntry),
		cacheTTL: 5 * time.Minute,
	}
}

func (c *APIClient) fetchURL(url string) ([]byte, error) {
	c.cacheMutex.RLock()
	if entry, exists := c.cache[url]; exists {
		if time.Since(entry.timestamp) < c.cacheTTL {
			c.cacheMutex.RUnlock()
			return entry.data, nil
		}
	}
	c.cacheMutex.RUnlock()

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("erreur requête HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("statut HTTP: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur lecture réponse: %w", err)
	}

	c.cacheMutex.Lock()
	c.cache[url] = cacheEntry{
		data:      body,
		timestamp: time.Now(),
	}
	c.cacheMutex.Unlock()

	return body, nil
}

func (c *APIClient) GetArtists() ([]Artist, error) {
	data, err := c.fetchURL(ArtistsEndpoint)
	if err != nil {
		return nil, err
	}

	var artists []Artist
	if err := json.Unmarshal(data, &artists); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	return artists, nil
}

func (c *APIClient) GetArtistByID(id int) (*Artist, error) {
	url := fmt.Sprintf("%s/%d", ArtistsEndpoint, id)
	data, err := c.fetchURL(url)
	if err != nil {
		return nil, err
	}

	var artist Artist
	if err := json.Unmarshal(data, &artist); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	return &artist, nil
}

func (c *APIClient) GetLocations() (*LocationsIndex, error) {
	data, err := c.fetchURL(LocationsEndpoint)
	if err != nil {
		return nil, err
	}

	var locations LocationsIndex
	if err := json.Unmarshal(data, &locations); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	return &locations, nil
}

func (c *APIClient) GetLocationsByArtistID(id int) (*LocationsData, error) {
	url := fmt.Sprintf("%s/%d", LocationsEndpoint, id)
	data, err := c.fetchURL(url)
	if err != nil {
		return nil, err
	}

	var locations LocationsData
	if err := json.Unmarshal(data, &locations); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	return &locations, nil
}

func (c *APIClient) GetDates() (*DatesIndex, error) {
	data, err := c.fetchURL(DatesEndpoint)
	if err != nil {
		return nil, err
	}

	var dates DatesIndex
	if err := json.Unmarshal(data, &dates); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	return &dates, nil
}

func (c *APIClient) GetDatesByArtistID(id int) (*DatesData, error) {
	url := fmt.Sprintf("%s/%d", DatesEndpoint, id)
	data, err := c.fetchURL(url)
	if err != nil {
		return nil, err
	}

	var dates DatesData
	if err := json.Unmarshal(data, &dates); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	return &dates, nil
}

func (c *APIClient) GetRelations() (*RelationIndex, error) {
	data, err := c.fetchURL(RelationEndpoint)
	if err != nil {
		return nil, err
	}

	var relations RelationIndex
	if err := json.Unmarshal(data, &relations); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	return &relations, nil
}

func (c *APIClient) GetRelationByArtistID(id int) (*Relation, error) {
	url := fmt.Sprintf("%s/%d", RelationEndpoint, id)
	data, err := c.fetchURL(url)
	if err != nil {
		return nil, err
	}

	var relation Relation
	if err := json.Unmarshal(data, &relation); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	return &relation, nil
}

func (c *APIClient) GetArtistComplete(id int) (*ArtistComplete, error) {
	artist, err := c.GetArtistByID(id)
	if err != nil {
		return nil, err
	}

	complete := &ArtistComplete{
		Artist: *artist,
	}

	locations, err := c.GetLocationsByArtistID(id)
	if err == nil {
		complete.LocationsList = locations.Locations
	}

	dates, err := c.GetDatesByArtistID(id)
	if err == nil {
		complete.DatesList = dates.Dates
	}

	relation, err := c.GetRelationByArtistID(id)
	if err == nil {
		complete.DatesLocations = relation.DatesLocations
	}

	return complete, nil
}

func (c *APIClient) GetAllArtistsComplete() ([]ArtistComplete, error) {
	artists, err := c.GetArtists()
	if err != nil {
		return nil, err
	}

	relations, err := c.GetRelations()
	if err != nil {
		return nil, err
	}

	relationsMap := make(map[int]Relation)
	for _, r := range relations.Index {
		relationsMap[r.ID] = r
	}

	result := make([]ArtistComplete, len(artists))
	for i, artist := range artists {
		result[i] = ArtistComplete{
			Artist: artist,
		}
		if relation, exists := relationsMap[artist.ID]; exists {
			result[i].DatesLocations = relation.DatesLocations
		}
	}

	return result, nil
}

func (c *APIClient) ClearCache() {
	c.cacheMutex.Lock()
	c.cache = make(map[string]cacheEntry)
	c.cacheMutex.Unlock()
}

var apiClient *APIClient

type ErrorData struct {
	Code    int
	Message string
	Details string
}

var templates *template.Template

func init() {
	var err error
	templates, err = template.ParseGlob("./templates/*.html")
	if err != nil {
		log.Fatalf("Erreur chargement templates: %v", err)
	}
	apiClient = NewAPIClient()
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s - %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func handleError(w http.ResponseWriter, code int, message string, details string) {
	w.WriteHeader(code)

	data := ErrorData{
		Code:    code,
		Message: message,
		Details: details,
	}

	err := templates.ExecuteTemplate(w, "error.html", data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur %d: %s", code, message), code)
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	handleError(w, http.StatusNotFound, "Page non trouvée",
		fmt.Sprintf("La page '%s' n'existe pas.", r.URL.Path))
}

func badRequestHandler(w http.ResponseWriter, _ *http.Request, details string) {
	handleError(w, http.StatusBadRequest, "Requête invalide", details)
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	handleError(w, http.StatusMethodNotAllowed, "Méthode non autorisée",
		fmt.Sprintf("La méthode '%s' n'est pas autorisée.", r.Method))
}

func internalServerErrorHandler(w http.ResponseWriter, err error) {
	handleError(w, http.StatusInternalServerError, "Erreur serveur",
		"Une erreur inattendue s'est produite.")
	log.Printf("Erreur: %v", err)
}

type PageData struct {
	Artists []ArtistComplete
	Artist  *ArtistComplete
	Error   string
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		notFoundHandler(w, r)
		return
	}

	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		methodNotAllowedHandler(w, r)
		return
	}

	artists, err := apiClient.GetAllArtistsComplete()
	if err != nil {
		internalServerErrorHandler(w, err)
		return
	}

	data := PageData{
		Artists: artists,
	}

	err = templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		internalServerErrorHandler(w, err)
		return
	}
}

func artistHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		methodNotAllowedHandler(w, r)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		badRequestHandler(w, r, "L'ID de l'artiste est requis")
		return
	}

	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || id < 1 {
		badRequestHandler(w, r, "L'ID doit être un nombre valide")
		return
	}

	artist, err := apiClient.GetArtistComplete(id)
	if err != nil {
		notFoundHandler(w, r)
		return
	}

	data := PageData{
		Artist: artist,
	}

	err = templates.ExecuteTemplate(w, "artist.html", data)
	if err != nil {
		internalServerErrorHandler(w, err)
		return
	}
}

func safeFileServer(dir string, prefix string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := dir + r.URL.Path[len(prefix)-1:]
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			notFoundHandler(w, r)
			return
		}
		if err != nil {
			internalServerErrorHandler(w, err)
			return
		}
		http.StripPrefix(prefix, fs).ServeHTTP(w, r)
	})
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	mux := http.NewServeMux()

	mux.Handle("/static/", safeFileServer("./static", "/static/"))
	mux.Handle("/image/", safeFileServer("./image", "/image/"))
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/artist", artistHandler)

	handler := loggingMiddleware(mux)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Println("Serveur lancé sur http://localhost:8080")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Erreur serveur: %v", err)
	}
}
