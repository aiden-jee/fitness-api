package main

import (
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"

	"fitness.com/exercisestore"
	"fitness.com/middleware"
	"golang.org/x/time/rate"
)

type exerciseServer struct {
	store *exercisestore.ExerciseStore
}

func NewExerciseServer() *exerciseServer {
	store := exercisestore.New()
	return &exerciseServer{store: store}
}

func jsonEncoder(w http.ResponseWriter, v any, code int) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(js)
}

func (es *exerciseServer) createExerciseHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling create exercise at %s\n", req.URL.Path)

	type RequestExercise struct {
		Name   string   `json:"name"`
		Sets   int      `json:"sets"`
		Reps   int      `json:"reps"`
		Weight float64  `json:"weight"`
		Tags   []string `json:"tags"`
	}

	type ResponseID struct {
		ID int `json:"id"`
	}

	content := req.Header.Get("Content-Type")
	media, _, err := mime.ParseMediaType(content)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if media != "application/json" {
		http.Error(w, "expected application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()

	var re RequestExercise
	if err := dec.Decode(&re); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := es.store.CreateExercise(re.Name, re.Sets, re.Reps, re.Weight, re.Tags)
	jsonEncoder(w, ResponseID{ID: id}, http.StatusCreated)
}

func (es *exerciseServer) getExerciseHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handing get exercise at %s\n", req.URL.Path)

	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	exercise, err := es.store.GetExerciseByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	jsonEncoder(w, exercise, http.StatusOK)
}

func (es *exerciseServer) getAllExercisesHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling get all exercises at %s\n", req.URL.Path)

	exercises := es.store.GetAllExercises()

	jsonEncoder(w, exercises, http.StatusOK)
}

func (es *exerciseServer) deleteExerciseByIDHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling delete exercise at %s\n", req.URL.Path)

	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = es.store.DeleteExercise(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (es *exerciseServer) deleteAllExercisesHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling delete all exerices at %s\n", req.URL.Path)

	type DeleteAllResponse struct {
		Deleted int `json:"deleted"`
	}

	deleted := es.store.DeleteAllExercises()

	res := DeleteAllResponse{
		Deleted: deleted,
	}

	jsonEncoder(w, res, http.StatusOK)
}

func main() {
	mux := http.NewServeMux()
	server := NewExerciseServer()
	limiter := rate.NewLimiter(1, 1)

	mux.HandleFunc("POST /exercise/", server.createExerciseHandler)
	mux.HandleFunc("GET /exercise/", server.getAllExercisesHandler)
	mux.HandleFunc("GET /exercise/{id}/", server.getExerciseHandler)
	mux.HandleFunc("DELETE /exercise/{id}/", server.deleteExerciseByIDHandler)
	mux.HandleFunc("DELETE /exercise/", server.deleteAllExercisesHandler) // does changing the order of this affect the routing?

	handler := middleware.RateLimit(limiter)(mux)
	handler = middleware.Recovery(handler)
	handler = middleware.Logging(handler)

	log.Fatal(http.ListenAndServe("localhost:"+os.Getenv("SERVERPORT"), handler))

}
