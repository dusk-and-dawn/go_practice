// main.go
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

type ExecuteRequest struct {
	Model    string         `json:"model,omitempty"`
	Action   string         `json:"action"`
	Input    map[string]any `json:"input"`
	Options  map[string]any `json:"options,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type ExecuteResponse struct {
	Action  string         `json:"action"`
	Result  map[string]any `json:"result"`
	Usage   map[string]any `json:"usage,omitempty"`
	TraceID string         `json:"trace_id"`
}

// ---- Router / Handler pattern ----

type Handler func(ctx context.Context, req ExecuteRequest) (ExecuteResponse, error)

type Gateway struct {
	client   *openai.Client
	model    string
	handlers map[string]Handler
	apiKey   string
}

func NewGateway(apiKey string) *Gateway {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	g := &Gateway{
		client: &client,
		//model:    selectedModel, //"gpt-5.2",
		apiKey:   apiKey,
		handlers: map[string]Handler{},
	}

	g.handlers["optimize_team_description"] = g.handleOptimizeTeamDescription

	return g
}

func (g *Gateway) Execute(ctx context.Context, req ExecuteRequest) (ExecuteResponse, error) {
	h, ok := g.handlers[req.Action]
	if !ok {
		return ExecuteResponse{}, fmt.Errorf("unknown action: %s", req.Action)
	}
	return h(ctx, req)
}

func (g *Gateway) handleOptimizeTeamDescription(ctx context.Context, req ExecuteRequest) (ExecuteResponse, error) {
	raw, ok := req.Input["text"].(string)
	if !ok || raw == "" {
		return ExecuteResponse{}, fmt.Errorf("input.text must be a non-empty string")
	}

	system := "Imagine you are part of a HR software in which a user is currently entering text describing a team they want to add. " +
		"It is your task to return a text which includes all facts from their prompt, but optimize it as regards business language and details that they omitted. " +
		"Please always create a basic role description for each mentioned role, and elaborate on general categories, if given. " +
		"Do not exceed 250 words and do not treat it as a dialogue, treat your answer like it would be the final team description to go into the software."

	input := system + "\n\nUser prompt:\n" + raw

	modelToUse := req.Model
	if modelToUse == "" {
		modelToUse = "gpt-5.2" // Fallback default
	}

	resp, err := g.client.Responses.New(ctx, responses.ResponseNewParams{
		Model: modelToUse,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(input),
		},
	})
	if err != nil {
		return ExecuteResponse{}, err
	}

	out := ExecuteResponse{
		Action: req.Action,
		Result: map[string]any{
			"text": resp.OutputText(),
		},
		Usage: map[string]any{
			"input_tokens":  resp.Usage.InputTokens,
			"output_tokens": resp.Usage.OutputTokens,
		},
		TraceID: newTraceID(),
	}
	return out, nil
}

func (g *Gateway) handleExecuteHTTP(w http.ResponseWriter, r *http.Request) {
	traceID := newTraceID()
	start := time.Now()

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	resp, err := g.Execute(r.Context(), req)
	if err != nil {
		log.Printf("[%s] action=%s error=%v", traceID, req.Action, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp.TraceID = traceID
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)

	log.Printf("[%s] action=%s ms=%d", traceID, req.Action, time.Since(start).Milliseconds())
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "static/main.html")
}

func (g *Gateway) handleSubmitDemo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	text := r.FormValue("optimizable_text")

	req := ExecuteRequest{
		Action: "optimize_team_description",
		Input:  map[string]any{"text": text},
		Metadata: map[string]any{
			"app": "demo-html",
		},
		Options: map[string]any{
			"context_level": "low",
		},
	}

	resp, err := g.Execute(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	out, _ := resp.Result["text"].(string)

	tmpl := template.Must(template.ParseFiles("static/result.html"))
	_ = tmpl.Execute(w, struct{ Output string }{Output: out})
}

// ---- HTTP layer ----

func main() {
	_ = godotenv.Load() // Error handling might be good later xD

	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		log.Fatal("OPENAI_API_KEY not set")
	}

	gw := NewGateway(key)

	mux := http.NewServeMux()

	mux.HandleFunc("/", serveHome)
	mux.HandleFunc("/submit", gw.handleSubmitDemo)

	mux.HandleFunc("/v1/ai/execute", gw.handleExecuteHTTP)

	// mux.HandleFunc("/v1/ai/execute", func(w http.ResponseWriter, r *http.Request) {
	// 	traceID := newTraceID()
	// 	start := time.Now()

	// 	if r.Method != http.MethodPost {
	// 		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	// 		return
	// 	}

	// 	var req ExecuteRequest
	// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	// 		http.Error(w, "invalid JSON", http.StatusBadRequest)
	// 		return
	// 	}

	// 	ctx := r.Context()
	// 	resp, err := gw.Execute(ctx, req)
	// 	if err != nil {
	// 		log.Printf("[%s] action=%s error=%v", traceID, req.Action, err)
	// 		http.Error(w, err.Error(), http.StatusBadRequest)
	// 		return
	// 	}

	// 	resp.TraceID = traceID
	// 	w.Header().Set("Content-Type", "application/json")
	// 	_ = json.NewEncoder(w).Encode(resp)

	// 	log.Printf("[%s] action=%s ms=%d", traceID, req.Action, time.Since(start).Milliseconds())
	// })

	addr := ":8080"
	log.Printf("AI Gateway listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func newTraceID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
