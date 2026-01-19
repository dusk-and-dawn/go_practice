package main

import (
	"fmt"

	"rsc.io/quote"
)

// func main() {
// 	fmt.Println(quote.Go())
// }

// func main() {
// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
// 	})

// 	http.ListenAndServe(":8080", nil)
// }

func main() {
	for i := 0; i < 3; i++ {
		fmt.Println(quote.Go())
	}

	// func test() {

	// }
}
