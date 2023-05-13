package main

import (
	"api/middleware"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	// Connect to the database
	db, err := middleware.ConnectToDatabase()
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}
	defer db.Close()

	// Create a new router
	r := mux.NewRouter()

	// Define the handler function
	r.HandleFunc("/api/v1/longest-duration-movies", middleware.LongestDurationMoviesHandler(db)).Methods("GET")
	r.HandleFunc("/api/v1/new-movie", middleware.NewMovieHandler(db)).Methods("POST")
	r.HandleFunc("/api/v1/top-rated-movies", middleware.HandleTopRatedMovies(db)).Methods("GET")
	r.HandleFunc("/api/v1/genre-movies-with-subtotals", middleware.GenreMoviesWithSubtotalsHandler(db)).Methods("GET")
	r.HandleFunc("/api/v1/update-runtime-minutes", middleware.UpdateRuntimeMinutesHandler(db)).Methods("POST")

	// Start the HTTP server
	log.Fatal(http.ListenAndServe(":8080", r))
}
