package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

// Structure pour les données d'erreur envoyées aux templates
type ErrorData struct {
	Code    int
	Message string
	Details string
}

// Variables globales pour les templates
var templates *template.Template

// Initialisation des templates au démarrage
func init() {
	var err error
	templates, err = template.ParseGlob("./templates/*.html")
	if err != nil {
		log.Fatalf("Erreur lors du chargement des templates: %v", err)
	}
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

	// Exécuter le template
	err := templates.ExecuteTemplate(w, "index.html", nil)
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
