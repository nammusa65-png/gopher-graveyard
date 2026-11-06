// Command graveyard is the web gateway: it serves the Gopher
// Graveyard UI and bridges browser requests to the Gopher protocol
// over raw TCP, which browsers cannot speak natively.
package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nammusa65-png/gopher-graveyard/internal/gopher"
)

//go:embed web
var webFS embed.FS

type fetchResponse struct {
	URL   string        `json:"url"`
	Kind  string        `json:"kind"` // "menu" or "text"
	Items []gopher.Item `json:"items,omitempty"`
	Text  string        `json:"text,omitempty"`
	Raw   string        `json:"raw"`
}

func main() {
	addr := flag.String("addr", ":8070", "listen address (host:port)")
	flag.Parse()

	mux := http.NewServeMux()
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("/api/fetch", handleFetch)
	mux.HandleFunc("/api/check", handleCheck)

	log.Printf("Gopher Graveyard gateway listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func handleFetch(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("url")
	u, err := gopher.ParseURL(raw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if q := r.URL.Query().Get("search"); q != "" {
		u.Search = q
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	data, err := gopher.Fetch(ctx, u)
	if err != nil {
		// The caller renders this as a tombstone: the host is dead
		// (or unreachable), which is part of the experience.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{
			"url": u.String(), "error": err.Error(),
		})
		return
	}
	resp := fetchResponse{URL: u.String(), Raw: string(data)}
	if u.Type == '1' || u.Type == '7' {
		resp.Kind = "menu"
		resp.Items = gopher.ParseMenu(data)
	} else {
		resp.Kind = "text"
		resp.Text = strings.TrimSuffix(strings.TrimRight(string(data), "\r\n"), "\n.")
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleCheck(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("url")
	u, err := gopher.ParseURL(raw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	u.Selector = ""
	u.Type = '1'
	u.Search = ""
	// One retry: intermediaries occasionally stall a first attempt,
	// and a single transient failure should not bury a live host.
	alive := false
	for attempt := 0; attempt < 2 && !alive; attempt++ {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		_, ferr := gopher.Fetch(ctx, u)
		cancel()
		alive = ferr == nil
		if !alive && attempt == 0 {
			time.Sleep(time.Second)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"url":   u.String(),
		"alive": alive,
	})
}
