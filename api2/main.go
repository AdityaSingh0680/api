package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type Movie struct {
	Tconst         string `json:"tconst"`
	TitleType      string `json:"titleType"`
	PrimaryTitle   string `json:"primaryTitle"`
	RuntimeMinutes int    `json:"runtimeMinutes"`
	Genres         string `json:"genres"`
}

func main() {
	// Define the database configuration
	config := &DBConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "2838",
		DBName:   "postgres",
	}

	// Format the connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName)

	// Open a connection to the database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Create a new router using Gorilla Mux
	r := mux.NewRouter()

	// Define a handler function for a POST request to create a new movie
	r.HandleFunc("/api/v1/new-movie", func(w http.ResponseWriter, r *http.Request) {
		var movie Movie
		err := json.NewDecoder(r.Body).Decode(&movie)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = db.Exec(`
			INSERT INTO movies (tconst, titleType, primaryTitle, runtimeMinutes, genres)
			VALUES ($1, $2, $3, $4, $5)
		`, movie.Tconst, movie.TitleType, movie.PrimaryTitle, movie.RuntimeMinutes, movie.Genres)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("success"))
	}).Methods("POST")

	// Start the server on port 8080
	log.Fatal(http.ListenAndServe(":8080", r))
}
