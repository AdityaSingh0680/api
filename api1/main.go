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

type Movie struct {
	Tconst         string `json:"tconst"`
	PrimaryTitle   string `json:"primaryTitle"`
	RuntimeMinutes int    `json:"runtimeMinutes"`
	Genres         string `json:"genres"`
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func main() {
	dbConfig := &DBConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "2838",
		DBName:   "postgres",
	}

	db, err := sql.Open("postgres", getDBConnectionString(dbConfig))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/longest-duration-movies", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
            SELECT m.tconst, m.primaryTitle, m.runtimeMinutes, m.genres
            FROM movies m
            ORDER BY m.runtimeMinutes DESC
            LIMIT 10
        `)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var movies []Movie
		for rows.Next() {
			var movie Movie
			err := rows.Scan(&movie.Tconst, &movie.PrimaryTitle, &movie.RuntimeMinutes, &movie.Genres)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			movies = append(movies, movie)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(movies)
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}

func getDBConnectionString(config *DBConfig) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName)
}
