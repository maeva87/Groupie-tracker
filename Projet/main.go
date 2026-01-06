package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type Artist struct {
	Name    string
	Members []string
}

func main() {
	fmt.Println("Starting server...")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		data, _ := ioutil.ReadFile("templates/index.html")
		w.Write(data)
	})
	http.ListenAndServe(":8080", nil)
}
