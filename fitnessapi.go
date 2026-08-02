package main

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"fitness.com/authstore"
	"fitness.com/exercisestore"
	"fitness.com/middleware"
	"golang.org/x/time/rate"
)

type apiServer struct {
	exerciseStore *exercisestore.ExerciseStore
	authStore     *authstore.AuthStore
}

/*************** Helper Functions ***************/

func NewAPIServer() *apiServer {
	exerciseStore := exercisestore.New()
	authStore := authstore.New()
	return &apiServer{exerciseStore: exerciseStore, authStore: authStore}
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

func (as *apiServer) validateToken(token string) error {
	decodedTokenBytes, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return err
	}
	tokenHash := sha256.Sum256(decodedTokenBytes)
	if _, ok := as.authStore.Sessions[tokenHash]; !ok {
		return fmt.Errorf("Session not found!")
	}
	return nil
}

/*************** Handlers ***************/
func (as *apiServer) createExerciseHandler(w http.ResponseWriter, req *http.Request) {
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

	id := as.exerciseStore.CreateExercise(re.Name, re.Sets, re.Reps, re.Weight, re.Tags)
	jsonEncoder(w, ResponseID{ID: id}, http.StatusCreated)
}

func (as *apiServer) getExerciseHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handing get exercise at %s\n", req.URL.Path)

	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	exercise, err := as.exerciseStore.GetExerciseByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	jsonEncoder(w, exercise, http.StatusOK)
}

func (as *apiServer) getAllExercisesHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling get all exercises at %s\n", req.URL.Path)

	exercises := as.exerciseStore.GetAllExercises()

	jsonEncoder(w, exercises, http.StatusOK)
}

func (as *apiServer) deleteExerciseByIDHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling delete exercise at %s\n", req.URL.Path)

	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = as.exerciseStore.DeleteExercise(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (as *apiServer) deleteAllExercisesHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling delete all exerices at %s\n", req.URL.Path)

	type DeleteAllResponse struct {
		Deleted int `json:"deleted"`
	}

	deleted := as.exerciseStore.DeleteAllExercises()

	res := DeleteAllResponse{
		Deleted: deleted,
	}

	jsonEncoder(w, res, http.StatusOK)
}

func (as *apiServer) loginHandler(w http.ResponseWriter, req *http.Request) {
	type RequestLogin struct {
		User string `json:"user"`
		Pw   string `json:"pw"`
	}

	type Response struct {
		Token   string    `json:"token"`
		Expires time.Time `json:"expires"`
	}

	content := req.Header.Get("Content-Type")
	media, _, err := mime.ParseMediaType(content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if media != "application/json" {
		http.Error(w, "expected application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()

	var re RequestLogin
	if err := dec.Decode(&re); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	session, token, err := as.authStore.LoginUser(re.User, re.Pw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonEncoder(w, Response{token, session.ExpiresAt}, http.StatusOK)
}

func (as *apiServer) registerHandler(w http.ResponseWriter, req *http.Request) {
	type RequestRegister struct {
		User string `json:"user"`
		Pw   string `json:"pw"`
	}

	content := req.Header.Get("Content-Type")
	media, _, err := mime.ParseMediaType(content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if media != "application/json" {
		http.Error(w, "expected application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()

	var re RequestRegister
	if err := dec.Decode(&re); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := as.authStore.RegisterUser(re.User, re.Pw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonEncoder(w, token, http.StatusCreated)
}

func (as *apiServer) deleteSessionHandler(w http.ResponseWriter, req *http.Request) {
	authorization := req.Header.Get("Authorization")
	scheme, token, ok := strings.Cut(authorization, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		http.Error(w, "invalid authorization format!", http.StatusUnauthorized)
		return
	}

	err := as.authStore.DeleteSession(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	certFile := flag.String("cert", "cert.pem", "TLS cert PEM file")
	keyFile := flag.String("key", "key.pem", "TLS key PEM file")
	flag.Parse()

	rootMux := http.NewServeMux()

	mux := http.NewServeMux()

	// seperate mux for authenticated requests
	authMux := http.NewServeMux()
	server := NewAPIServer()
	limiter := rate.NewLimiter(1, 1)

	authMux.HandleFunc("POST /exercise/", server.createExerciseHandler)
	authMux.HandleFunc("GET /exercise/", server.getAllExercisesHandler)
	authMux.HandleFunc("GET /exercise/{id}/", server.getExerciseHandler)
	authMux.HandleFunc("DELETE /exercise/{id}/", server.deleteExerciseByIDHandler)
	authMux.HandleFunc("DELETE /exercise/", server.deleteAllExercisesHandler)
	authMux.HandleFunc("DELETE /auth/", server.deleteSessionHandler)
	mux.HandleFunc("POST /auth/login/", server.loginHandler)
	mux.HandleFunc("POST /auth/register/", server.registerHandler)

	handler := middleware.RateLimit(limiter)(mux)
	handler = middleware.Recovery(handler)

	authHandler := middleware.Authenticate(server.validateToken)(authMux)
	authHandler = middleware.RateLimit(limiter)(authHandler)
	authHandler = middleware.Recovery(authHandler)

	rootMux.Handle("DELETE /auth/", authHandler)
	rootMux.Handle("/auth/", handler)
	rootMux.Handle("/exercise/", authHandler)

	rootHandler := middleware.Logging(rootMux)
	addr := "localhost:" + os.Getenv("SERVERPORT")
	srv := http.Server{
		Addr:    addr,
		Handler: rootHandler,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	}
	log.Fatal(srv.ListenAndServeTLS(*certFile, *keyFile))

}
