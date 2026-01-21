package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

// func main() {
// 	fmt.Println(quote.Go())
// }

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		fmt.Println("OPENAI_API_KEY not set")
	}

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

		context2 := "Imagine you are part of a HR software in which a user is currently entering text describing a team they want to add. It is your task to return a text which includes all facts from their prompt, but optimize it as regards business language and details that they omitted. Please always create a basic role description for each mentioned role, and elaborate on general categories, if given. Do not exceed 250 words and do not treat it as a dialogue, treat your answer like it would be the final team description to go into the software. Their prompt: "
		text := r.FormValue("optimizable_text")
		// fmt.Fprintf(w, "got: %q\n", text)
		fmt.Println(text)

		client := openai.NewClient(
			option.WithAPIKey(key),
		)

		resp, err := client.Responses.New(
			context.TODO(),
			responses.ResponseNewParams{
				Model: "gpt-5-nano",
				Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(context2 + "" + text)},
			})
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(resp.OutputText())
		// fmt.Fprintf(w, "got: %q\n", resp.OutputText())

		var tmpl = template.Must(
			template.ParseFiles("static/result.html"),
		)

		tmpl.Execute(w, struct {
			Output string
		}{
			Output: resp.OutputText(),
		})

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
