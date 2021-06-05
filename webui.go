package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"plugmeter/web"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func start_webui(port int) {
	fmt.Println("Starting Web UI on port ", port)
	// mux := http.NewServeMux()
	r := mux.NewRouter()
	handler := web.AssetHandler("/static/", "public")

	// Add handler for static files
	r.PathPrefix("/static/").Handler(handler)

	// r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	// })

	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("", get).Methods(http.MethodGet)
	api.HandleFunc("/user/{userID}/comment/{commentID}", params).Methods(http.MethodGet)
	api.HandleFunc("/plugs", api_plugs).Methods(http.MethodGet)
	api.HandleFunc("/plugs/{plugID}", api_plug).Methods(http.MethodGet)

	// r.HandleFunc(0)

	h := handlers.CORS(handlers.AllowedMethods([]string{"POST", "GET"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedHeaders([]string{"X-Requested-With"}))(r)

	srv := &http.Server{
		Handler: h,
		Addr:    "127.0.0.1:3000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	srv.ListenAndServe()
	// mux.Handle("/", clientHandler())
	// http.ListenAndServe(":3000", r)
}

func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "get called"}`))
}

func params(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

	userID := -1
	var err error
	if val, ok := pathParams["userID"]; ok {
		userID, err = strconv.Atoi(val)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a number"}`))
			return
		}
	}

	commentID := -1
	if val, ok := pathParams["commentID"]; ok {
		commentID, err = strconv.Atoi(val)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "need a number"}`))
			return
		}
	}

	query := r.URL.Query()
	location := query.Get("location")

	w.Write([]byte(fmt.Sprintf(`{"userID": %d, "commentID": %d, "location": "%s" }`, userID, commentID, location)))
}

// API
//   /plugs
//   /plugs/<plugID>
//   /power/<plugID>

// Handler
func api_plugs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	plugs := get_plugs()
	encoded, err := json.Marshal(plugs)
	if err != nil {
		fmt.Println("Error marshalling plugs", err)
	}
	w.Write([]byte(encoded))

	// w.Write([]byte(fmt.Sprintf(
	// 	`[{"plugId": "%s"}, {"plugId": "%s"}]`, nowFormat, nowFormat)))
}

// Handler
func api_plug(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

	plugID := pathParams["plugID"]
	plug := get_plug(plugID)
	encoded, err := json.Marshal(plug)
	if err != nil {
		fmt.Println("Error marshalling plug", plug, err)
	}
	w.Write([]byte(encoded))

	// w.Write([]byte(fmt.Sprintf(`{"plugId": "%s"}`, plugID)))
}
