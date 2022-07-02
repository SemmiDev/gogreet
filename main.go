package main

import (
	"fmt"
	"net/http"
)

func greet(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "Kamuu"
	}
	name += " 😊"
	resp := fmt.Sprintf(`<div style='text-align: center;'> <h1>Hii %s</h1> <p><i>have a nice day.</i></p> </div>`, name)
	fmt.Fprint(w, resp)
}

func main() {
	http.HandleFunc("/greet", greet)
	fmt.Println("server running on port 8080")
	http.ListenAndServe(":8080", nil)
}
