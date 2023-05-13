package middleware

import (
	"api/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// below function connects to the PostgreSQL database by reading the database URL from the .env file
func ConnectToDatabase() (*sql.DB, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		return nil, err
	}

	return db, nil
}

func LongestDurationMoviesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		// Execute the SQL query to retrieve the top 10 movies with the longest runtime
		query := `
			SELECT tconst, primaryTitle, runtimeMinutes, genres
			FROM movies
			ORDER BY runtimeMinutes DESC
			LIMIT 10
		`
		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Create a slice of Movie objects to store the results
		movies := make([]models.Movie, 0)

		// Iterate through the rows and add each Movie object to the slice
		for rows.Next() {
			var movie models.Movie
			err := rows.Scan(&movie.Tconst, &movie.PrimaryTitle, &movie.RuntimeMinutes, &movie.Genres)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			movies = append(movies, movie)
		}

		// Check for errors during iteration
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Convert the slice of Movie objects to JSON and write it to the HTTP response
		jsonBytes, err := json.Marshal(movies)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonBytes)
	}
}

func NewMovieHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the JSON request body into a NewMovie struct
		var newMovie models.Movie
		err := json.NewDecoder(r.Body).Decode(&newMovie)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Execute the SQL query to insert the new movie into the movies table
		query := `
			INSERT INTO movies (tconst, titleType, primaryTitle, runtimeMinutes, genres)
			VALUES ($1, $2, $3, $4, $5)
		`
		_, err = db.Exec(query, newMovie.Tconst, newMovie.TitleType, newMovie.PrimaryTitle, newMovie.RuntimeMinutes, newMovie.Genres)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write a success message to the HTTP response
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "success")
	}
}

func HandleTopRatedMovies(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		rows, err := db.Query(`
			SELECT m.tconst, m.primaryTitle, m.genres, r.averageRating
			FROM movies m
			JOIN ratings r ON m.tconst = r.tconst
			WHERE r.averageRating > $1
			ORDER BY r.averageRating DESC
		`, 6.0)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var movies []models.Rating
		for rows.Next() {
			var movie models.Rating
			err = rows.Scan(&movie.Tconst, &movie.PrimaryTitle, &movie.Genres, &movie.AverageRating)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			movies = append(movies, movie)
		}
		if err = rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		output, err := json.Marshal(movies)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(output)
	}
}

func GenreMoviesWithSubtotalsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		results, err := GetGenreMoviesWithSubtotals(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Generate the output as a tabular format with dashed border lines
		output := "+----------+------------------------+----------+\n"
		output += "|  Genre   |      Primary Title      | NumVotes |\n"
		output += "+----------+------------------------+----------+\n"
		for _, result := range results {
			if result.Title == "TOTAL" {
				output += fmt.Sprintf("|          |        %-16s | %-8d |\n", result.Title, result.NumVotes)
			} else if result.Genre == "" {
				output += fmt.Sprintf("|          | %-22s |          |\n", result.Title)
			} else {
				output += fmt.Sprintf("| %-8s | %-22s | %-8d |\n", result.Genre, result.Title, result.NumVotes)
			}
			output += "+----------+------------------------+----------+\n"
		}

		// Set the response headers
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Write the output to the HTTP response
		fmt.Fprint(w, output)

		// Print the output to the console
		fmt.Print(output)
	}
}

func GetGenreMoviesWithSubtotals(db *sql.DB) ([]models.GenreMovieResult, error) {
	// Execute the SQL query
	rows, err := db.Query(`
		SELECT genres, primaryTitle, numVotes
		FROM movies
		JOIN ratings ON movies.tconst = ratings.tconst
		ORDER BY genres, numVotes DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse the results
	var (
		prevGenre string
		total     int
		results   []models.GenreMovieResult
	)
	for rows.Next() {
		var (
			genre    string
			title    string
			numVotes int
		)
		if err := rows.Scan(&genre, &title, &numVotes); err != nil {
			return nil, err
		}

		if prevGenre == "" {
			prevGenre = genre
			results = append(results, models.GenreMovieResult{Genre: genre, Title: title, NumVotes: numVotes})
			total += numVotes
		} else if genre == prevGenre {
			results = append(results, models.GenreMovieResult{Genre: "", Title: title, NumVotes: numVotes})
			total += numVotes
		} else {
			results = append(results, models.GenreMovieResult{Genre: "", Title: "TOTAL", NumVotes: total})
			results = append(results, models.GenreMovieResult{Genre: genre, Title: title, NumVotes: numVotes})
			prevGenre = genre
			total = numVotes
		}
	}
	results = append(results, models.GenreMovieResult{Genre: "", Title: "TOTAL", NumVotes: total})
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil

}
func UpdateRuntimeMinutesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		// Execute the SQL query to update the runtimeMinutes field of all movies
		err := UpdateRuntimeMinutes(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write a success message to the HTTP response
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Successfully updated runtimeMinutes for all movies")
	}
}

func UpdateRuntimeMinutes(db *sql.DB) error {
	// Execute a SQL query to update the runtimeMinutes field of all movies
	query := `
		UPDATE movies SET runtimeMinutes = 
		CASE 
			WHEN genres LIKE '%Documentary%' THEN runtimeMinutes + 15
			WHEN genres LIKE '%Animation%' THEN runtimeMinutes + 30
			ELSE runtimeMinutes + 45
		END
	`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}
