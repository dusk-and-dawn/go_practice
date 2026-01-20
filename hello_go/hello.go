package main

import (
	"fmt"

	"net/http"
)

// func main() {
// 	fmt.Println(quote.Go())
// }

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "illegal method - send ICE", http.StatusMethodNotAllowed)
			return
		}
		http.ServeFile(w, r, "static/main.html")
	})

	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		text := r.FormValue("optimizable_text")
		fmt.Fprintf(w, "got: %q\n", text)
		fmt.Println(text)
	})

	http.ListenAndServe(":8080", nil)
}

// func main() {
// 	for i := 0; i < 3; i++ {
// 		fmt.Println(quote.Go())
// 	}

// 	// func test() {

// 	// }
// }
