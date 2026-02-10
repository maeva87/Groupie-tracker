package main

import (
	"fmt"
	"net/http"
	"text/template"
)

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

	http.HandleFunc("/about.html", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("./templates/about.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/contact.html", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("./templates/contact.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/recherche", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("./templates/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	fmt.Println("Serveur lancé sur http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
