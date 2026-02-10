# Groupie Tracker

Un site web interactif permettant de visualiser et d'explorer les informations sur des artistes musicaux à travers une API.

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

## Table des matières

- [À propos](#à-propos)
- [Fonctionnalités](#fonctionnalités)
- [Technologies utilisées](#technologies-utilisées)
- [Prérequis](#prérequis)
- [Installation](#installation)
- [Utilisation](#utilisation)
- [Structure du projet](#structure-du-projet)
- [API](#api)
- [Tests](#tests)
- [Bonnes pratiques](#bonnes-pratiques)
- [Contributeurs](#contributeurs)

## À propos

Groupie Tracker est un projet qui consiste à recevoir et manipuler des données provenant d'une API pour créer un site web convivial affichant des informations sur des artistes musicaux, leurs concerts et leurs tournées.

### Objectifs du projet

- Manipulation et stockage de données
- Traitement de fichiers JSON
- Création d'interfaces HTML dynamiques
- Gestion des événements client-serveur
- Communication request-response

## Fonctionnalités

- **Affichage des artistes** : Visualisation des informations sur des artistes (nom, Géolocalisation, premier album)
- **Localisation des concerts** : Carte interactive des lieux de concerts passés et à venir
- **Dates de concerts** : Calendrier des événements passés et futurs
- **Relations** : Liens entre artistes, dates et lieux
- **Système de recherche** : Recherche par nom d'artiste, genre, pays
- **Responsive Design** : Interface adaptée à tous les appareils
- **Événements dynamiques** : Interactions client-serveur en temps réel

## Technologies utilisées

### Backend
- **Go** (Golang) - Langage principal
- Packages standard Go uniquement

### Frontend
- **HTML5** - Structure des pages
- **CSS3** - Stylisation
- **JavaScript** - Interactivité

### API
- **RESTful API** - Récupération des données
- **JSON** - Format d'échange de données

## Prérequis

- Go 1.21 ou supérieur
- Un navigateur web moderne
- Git (pour cloner le projet)

## Installation

1. **Cloner le repository**
```bash
git clone https://github.com/maeva87/Groupie-tracker.git
cd groupie-tracker
```

2. **Vérifier l'installation de Go**
```bash
go version
```

3. **Installer les dépendances** (si nécessaire)
```bash
go mod download
```

## Utilisation

1. **Lancer le serveur**
```bash
go run main.go
```

2. **Accéder au site**
Ouvrez votre navigateur et accédez à :
```
http://localhost:8080
```

3. **Arrêter le serveur**
Appuyez sur `Ctrl+C` dans le terminal

## Structure du projet
```
groupie-tracker/
|── projet/
    ├── main.go
    ├── image/
    |   ├── logo.jpg
    ├── static/
    |   ├── style.css
    ├── templates/
    |   ├── about.html
    |   ├── contact.html
    |   ├── index.html
    ├── tmp/
    |   ├── main.exe
    ├── go.mod
    ├── README.md


## API

L'application consomme une API externe composée de 4 endpoints :

### 1. Artists
```
GET /artists
```
Retourne les informations sur les artistes :
- Nom(s)
- Image
- Année de début d'activité
- Date du premier album

### 2. Locations
```
GET /locations
```
Retourne les lieux de concerts (passés et à venir)

### 3. Dates
```
GET /dates
```
Retourne les dates de concerts (passées et à venir)

### 4. Relations
```
GET /relation
```
Lie les artistes, dates et lieux ensemble

### Exemple de requête
```go
resp, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

var artists []Artist
json.NewDecoder(resp.Body).Decode(&artists)
```

## Tests

Lancer les tests unitaires :
```bash
# Tous les tests
go test ./...

# Tests avec couverture
go test -cover ./...

# Tests avec rapport détaillé
go test -v ./...

# Tests d'un package spécifique
go test ./handlers
```

## Bonnes pratiques

### Code
- Gestion des erreurs appropriée
- Code commenté et documenté
- Respect des conventions Go (gofmt, golint)
- Pas de crash serveur
- Validation des données utilisateur

### Sécurité
- Protection contre les injections
- Validation des entrées
- Gestion appropriée des erreurs HTTP

### Performance
- Cache des données API
- Optimisation des requêtes
- Compression des ressources

## Fonctionnalités implémentées

- Récupération et affichage des artistes
- Affichage des lieux de concerts
- Affichage des dates de concerts
-  Système de recherche
- Filtres par genre/pays
- Interface responsive
- Gestion des erreurs
- Tests unitaires

## Gestion des erreurs

Le serveur gère les erreurs suivantes :
- 404 - Page non trouvée
- 500 - Erreur serveur interne
- Erreurs de connexion API
- Erreurs de parsing JSON

## Exemples d'événements client-serveur

### Recherche en temps réel
```javascript
// Client - JavaScript
fetch('/api/search?query=' + searchTerm)
    .then(response => response.json())
    .then(data => displayResults(data));
```
```go
// Server - Go
func SearchHandler(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("query")
    results := searchArtists(query)
    json.NewEncoder(w).Encode(results)
}
```

## Contributeurs

- **Jonathan HASARD** - Développeur principal
- **Kinsley SUEDILE COLLET** - Développeur principal
- **Maéva Neveu** - Développeur principal

## Licence

Ce projet est sous licence MIT.

## Liens utiles

- [Documentation Go](https://golang.org/doc/)
- [RESTful API Guide](https://restfulapi.net/)
- [JSON Format](https://www.json.org/)
- [Client-Server Architecture](https://en.wikipedia.org/wiki/Client%E2%80%93server_model)
