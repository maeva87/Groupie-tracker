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

// ============================================================================
// CONSTANTES API
// ============================================================================

const (
	BaseAPIURL        = "https://groupietrackers.herokuapp.com/api"
	ArtistsEndpoint   = BaseAPIURL + "/artists"
	LocationsEndpoint = BaseAPIURL + "/locations"
	DatesEndpoint     = BaseAPIURL + "/dates"
	RelationEndpoint  = BaseAPIURL + "/relation"
)

// ============================================================================
// STRUCTURES DE DONNÉES POUR L'API
// ============================================================================

// Artist représente un artiste/groupe musical
type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Locations    string   `json:"locations"`    // URL vers les locations
	ConcertDates string   `json:"concertDates"` // URL vers les dates
	Relations    string   `json:"relations"`    // URL vers les relations
}

// LocationsData représente les données de localisation d'un artiste
type LocationsData struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
	DatesURL  string   `json:"dates"`
}

// LocationsIndex représente la liste de toutes les locations
type LocationsIndex struct {
	Index []LocationsData `json:"index"`
}

// DatesData représente les dates de concert d'un artiste
type DatesData struct {
	ID    int      `json:"id"`
	Dates []string `json:"dates"`
}

// DatesIndex représente la liste de toutes les dates
type DatesIndex struct {
	Index []DatesData `json:"index"`
}

// Relation représente la relation entre un artiste et ses concerts (lieu -> dates)
type Relation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

// RelationIndex représente la liste de toutes les relations
type RelationIndex struct {
	Index []Relation `json:"index"`
}

// ArtistComplete représente un artiste avec toutes ses données combinées
type ArtistComplete struct {
	Artist
	LocationsList  []string            `json:"locationsList"`
	DatesList      []string            `json:"datesList"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

// APIResponse représente la réponse principale de l'API
type APIResponse struct {
	Artists   string `json:"artists"`
	Locations string `json:"locations"`
	Dates     string `json:"dates"`
	Relation  string `json:"relation"`
}

// ============================================================================
// CLIENT API AVEC CACHE
// ============================================================================

// APIClient gère les appels à l'API avec un système de cache
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

// NewAPIClient crée un nouveau client API
func NewAPIClient() *APIClient {
	return &APIClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache:    make(map[string]cacheEntry),
		cacheTTL: 5 * time.Minute, // Cache de 5 minutes
	}
}

// fetchURL récupère les données d'une URL avec gestion du cache
func (c *APIClient) fetchURL(url string) ([]byte, error) {
	// Vérifier le cache
	c.cacheMutex.RLock()
	if entry, exists := c.cache[url]; exists {
		if time.Since(entry.timestamp) < c.cacheTTL {
			c.cacheMutex.RUnlock()
			log.Printf("Cache hit pour: %s", url)
			return entry.data, nil
		}
	}
	c.cacheMutex.RUnlock()

	// Faire la requête HTTP
	log.Printf("Requête API: %s", url)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la requête HTTP: %w", err)
	}
	defer resp.Body.Close()

	// Vérifier le code de statut
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("statut HTTP inattendu: %d %s", resp.StatusCode, resp.Status)
	}

	// Lire le corps de la réponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la lecture de la réponse: %w", err)
	}

	// Mettre en cache
	c.cacheMutex.Lock()
	c.cache[url] = cacheEntry{
		data:      body,
		timestamp: time.Now(),
	}
	c.cacheMutex.Unlock()

	return body, nil
}

// GetArtists récupère la liste de tous les artistes
func (c *APIClient) GetArtists() ([]Artist, error) {
	data, err := c.fetchURL(ArtistsEndpoint)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des artistes: %w", err)
	}

	var artists []Artist
	if err := json.Unmarshal(data, &artists); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing JSON des artistes: %w", err)
	}

	log.Printf("Récupération réussie de %d artistes", len(artists))
	return artists, nil
}

// GetArtistByID récupère un artiste spécifique par son ID
func (c *APIClient) GetArtistByID(id int) (*Artist, error) {
	url := fmt.Sprintf("%s/%d", ArtistsEndpoint, id)
	data, err := c.fetchURL(url)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération de l'artiste %d: %w", id, err)
	}

	var artist Artist
	if err := json.Unmarshal(data, &artist); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing JSON de l'artiste: %w", err)
	}

	return &artist, nil
}

// GetLocations récupère toutes les locations
func (c *APIClient) GetLocations() (*LocationsIndex, error) {
	data, err := c.fetchURL(LocationsEndpoint)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des locations: %w", err)
	}

	var locations LocationsIndex
	if err := json.Unmarshal(data, &locations); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing JSON des locations: %w", err)
	}

	log.Printf("Récupération réussie de %d locations", len(locations.Index))
	return &locations, nil
}

// GetLocationsByArtistID récupère les locations d'un artiste spécifique
func (c *APIClient) GetLocationsByArtistID(id int) (*LocationsData, error) {
	url := fmt.Sprintf("%s/%d", LocationsEndpoint, id)
	data, err := c.fetchURL(url)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des locations de l'artiste %d: %w", id, err)
	}

	var locations LocationsData
	if err := json.Unmarshal(data, &locations); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing JSON des locations: %w", err)
	}

	return &locations, nil
}

// GetDates récupère toutes les dates de concert
func (c *APIClient) GetDates() (*DatesIndex, error) {
	data, err := c.fetchURL(DatesEndpoint)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des dates: %w", err)
	}

	var dates DatesIndex
	if err := json.Unmarshal(data, &dates); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing JSON des dates: %w", err)
	}

	log.Printf("Récupération réussie de %d dates", len(dates.Index))
	return &dates, nil
}

// GetDatesByArtistID récupère les dates de concert d'un artiste spécifique
func (c *APIClient) GetDatesByArtistID(id int) (*DatesData, error) {
	url := fmt.Sprintf("%s/%d", DatesEndpoint, id)
	data, err := c.fetchURL(url)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des dates de l'artiste %d: %w", id, err)
	}

	var dates DatesData
	if err := json.Unmarshal(data, &dates); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing JSON des dates: %w", err)
	}

	return &dates, nil
}

// GetRelations récupère toutes les relations artiste-concerts
func (c *APIClient) GetRelations() (*RelationIndex, error) {
	data, err := c.fetchURL(RelationEndpoint)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des relations: %w", err)
	}

	var relations RelationIndex
	if err := json.Unmarshal(data, &relations); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing JSON des relations: %w", err)
	}

	log.Printf("Récupération réussie de %d relations", len(relations.Index))
	return &relations, nil
}

// GetRelationByArtistID récupère la relation d'un artiste spécifique
func (c *APIClient) GetRelationByArtistID(id int) (*Relation, error) {
	url := fmt.Sprintf("%s/%d", RelationEndpoint, id)
	data, err := c.fetchURL(url)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération de la relation de l'artiste %d: %w", id, err)
	}

	var relation Relation
	if err := json.Unmarshal(data, &relation); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing JSON de la relation: %w", err)
	}

	return &relation, nil
}

// GetArtistComplete récupère un artiste avec toutes ses données (locations, dates, relations)
func (c *APIClient) GetArtistComplete(id int) (*ArtistComplete, error) {
	// Récupérer l'artiste de base
	artist, err := c.GetArtistByID(id)
	if err != nil {
		return nil, err
	}

	// Créer la structure complète
	complete := &ArtistComplete{
		Artist: *artist,
	}

	// Récupérer les locations
	locations, err := c.GetLocationsByArtistID(id)
	if err != nil {
		log.Printf("Avertissement: impossible de récupérer les locations pour l'artiste %d: %v", id, err)
	} else {
		complete.LocationsList = locations.Locations
	}

	// Récupérer les dates
	dates, err := c.GetDatesByArtistID(id)
	if err != nil {
		log.Printf("Avertissement: impossible de récupérer les dates pour l'artiste %d: %v", id, err)
	} else {
		complete.DatesList = dates.Dates
	}

	// Récupérer les relations (map lieu -> dates)
	relation, err := c.GetRelationByArtistID(id)
	if err != nil {
		log.Printf("Avertissement: impossible de récupérer les relations pour l'artiste %d: %v", id, err)
	} else {
		complete.DatesLocations = relation.DatesLocations
	}

	return complete, nil
}

// GetAllArtistsComplete récupère tous les artistes avec leurs données complètes
func (c *APIClient) GetAllArtistsComplete() ([]ArtistComplete, error) {
	artists, err := c.GetArtists()
	if err != nil {
		return nil, err
	}

	relations, err := c.GetRelations()
	if err != nil {
		return nil, err
	}

	// Créer une map des relations par ID pour accès rapide
	relationsMap := make(map[int]Relation)
	for _, r := range relations.Index {
		relationsMap[r.ID] = r
	}

	// Combiner les données
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

// ClearCache vide le cache de l'API
func (c *APIClient) ClearCache() {
	c.cacheMutex.Lock()
	c.cache = make(map[string]cacheEntry)
	c.cacheMutex.Unlock()
	log.Println("Cache API vidé")
}

// ============================================================================
// VARIABLE GLOBALE DU CLIENT API
// ============================================================================

var apiClient *APIClient

// Structure pour les données d'erreur envoyées aux templates
type ErrorData struct {
	Code    int
	Message string
	Details string
}

// Variables globales pour les templates
var templates *template.Template

// Initialisation des templates et du client API au démarrage
func init() {
	var err error
	templates, err = template.ParseGlob("./templates/*.html")
	if err != nil {
		log.Fatalf("Erreur lors du chargement des templates: %v", err)
	}

	// Initialiser le client API
	apiClient = NewAPIClient()
	log.Println("Client API initialisé")
}

// Middleware de logging pour tracer les requêtes HTTP
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[%s] %s %s - Début de la requête", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s - Terminé en %v", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}

// Fonction centralisée pour gérer les erreurs HTTP
func handleError(w http.ResponseWriter, code int, message string, details string) {
	log.Printf("Erreur %d: %s - %s", code, message, details)

	w.WriteHeader(code)

	data := ErrorData{
		Code:    code,
		Message: message,
		Details: details,
	}

	// Essayer de charger le template d'erreur
	err := templates.ExecuteTemplate(w, "error.html", data)
	if err != nil {
		// Si le template d'erreur n'existe pas, renvoyer une réponse texte simple
		log.Printf("Erreur lors du rendu du template d'erreur: %v", err)
		http.Error(w, fmt.Sprintf("Erreur %d: %s", code, message), code)
	}
}

// Gestionnaire pour les erreurs 404 (Page non trouvée)
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	handleError(w, http.StatusNotFound, "Page non trouvée",
		fmt.Sprintf("La page '%s' n'existe pas sur ce serveur.", r.URL.Path))
}

// Gestionnaire pour les erreurs 400 (Mauvaise requête)
func badRequestHandler(w http.ResponseWriter, _ *http.Request, details string) {
	handleError(w, http.StatusBadRequest, "Requête invalide", details)
}

// Gestionnaire pour les erreurs 405 (Méthode non autorisée)
func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	handleError(w, http.StatusMethodNotAllowed, "Méthode non autorisée",
		fmt.Sprintf("La méthode '%s' n'est pas autorisée pour cette ressource.", r.Method))
}

// Gestionnaire pour les erreurs 500 (Erreur interne du serveur)
func internalServerErrorHandler(w http.ResponseWriter, err error) {
	handleError(w, http.StatusInternalServerError, "Erreur interne du serveur",
		"Une erreur inattendue s'est produite. Veuillez réessayer plus tard.")
	log.Printf("Erreur interne: %v", err)
}

// PageData représente les données envoyées aux templates de page
type PageData struct {
	Artists []ArtistComplete
	Artist  *ArtistComplete
	Error   string
}

// Handler principal pour la page d'accueil
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier que c'est bien la route racine
	if r.URL.Path != "/" {
		notFoundHandler(w, r)
		return
	}

	// Vérifier la méthode HTTP (accepter seulement GET et HEAD)
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		methodNotAllowedHandler(w, r)
		return
	}

	// Récupérer tous les artistes depuis l'API
	artists, err := apiClient.GetAllArtistsComplete()
	if err != nil {
		log.Printf("Erreur lors de la récupération des artistes: %v", err)
		internalServerErrorHandler(w, err)
		return
	}

	// Préparer les données pour le template
	data := PageData{
		Artists: artists,
	}

	// Exécuter le template
	err = templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		internalServerErrorHandler(w, err)
		return
	}
}

// Handler pour afficher un artiste spécifique
func artistHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier la méthode HTTP
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		methodNotAllowedHandler(w, r)
		return
	}

	// Récupérer l'ID de l'artiste depuis les paramètres de requête
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		badRequestHandler(w, r, "L'ID de l'artiste est requis")
		return
	}

	// Convertir l'ID en entier
	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil || id < 1 {
		badRequestHandler(w, r, "L'ID de l'artiste doit être un nombre valide")
		return
	}

	// Récupérer les données complètes de l'artiste
	artist, err := apiClient.GetArtistComplete(id)
	if err != nil {
		log.Printf("Erreur lors de la récupération de l'artiste %d: %v", id, err)
		notFoundHandler(w, r)
		return
	}

	// Préparer les données pour le template
	data := PageData{
		Artist: artist,
	}

	// Exécuter le template
	err = templates.ExecuteTemplate(w, "artist.html", data)
	if err != nil {
		internalServerErrorHandler(w, err)
		return
	}
}

// Wrapper pour le FileServer avec gestion d'erreurs
func safeFileServer(dir string, prefix string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Vérifier si le fichier existe
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

		// Servir le fichier
		http.StripPrefix(prefix, fs).ServeHTTP(w, r)
	})
}

func main() {
	// Configuration du logger
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Démarrage du serveur Groupie-Tracker...")

	// Création du multiplexeur de routes
	mux := http.NewServeMux()

	// Routes pour les fichiers statiques avec gestion d'erreurs
	mux.Handle("/static/", safeFileServer("./static", "/static/"))
	mux.Handle("/image/", safeFileServer("./image", "/image/"))

	// Route principale
	mux.HandleFunc("/", homeHandler)

	// Route pour afficher un artiste spécifique
	mux.HandleFunc("/artist", artistHandler)

	// Application du middleware de logging
	handler := loggingMiddleware(mux)

	// Configuration du serveur avec timeouts
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Println("╔════════════════════════════════════════════╗")
	fmt.Println("║   Serveur Groupie-Tracker démarré !        ║")
	fmt.Println("║   URL: http://localhost:8080               ║")
	fmt.Println("╚════════════════════════════════════════════╝")

	// Démarrer le serveur avec gestion d'erreur
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Erreur fatale du serveur: %v", err)
	}
}
