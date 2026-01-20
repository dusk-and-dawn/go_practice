package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
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

		client := openai.NewClient(
			option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
		)

		resp, err := client.Responses.New(context.TODO(), openai.ResponseNewParams{
			Model: "gpt-5-nano",
			Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(text)},
		})
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(resp.OutputText())
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
