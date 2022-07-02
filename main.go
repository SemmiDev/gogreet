package main

import (
	"fmt"
	"net/http"
)

func greet(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "kamuu"
	}
	fmt.Fprintf(w, "hi %s!! have a nice day", name)
}

func main() {
	http.HandleFunc("/api/greet", greet)
	fmt.Println("server running on port 3030")
	http.ListenAndServe(":3030", nil)
}
