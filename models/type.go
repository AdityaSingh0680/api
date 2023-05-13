package models

// Movie represents a row from the movies table
type Movie struct {
	Tconst         string `json:"tconst"`
	TitleType      string `json:"titleType"`
	PrimaryTitle   string `json:"primaryTitle"`
	RuntimeMinutes int    `json:"runtimeMinutes"`
	Genres         string `json:"genres"`
}
type Rating struct {
	Tconst        string  `json:"tconst"`
	PrimaryTitle  string  `json:"primaryTitle"`
	Genres        string  `json:"genres"`
	AverageRating float64 `json:"averageRating"`
}
type GenreMovieResult struct {
	Genre    string `json:"genre"`
	Title    string `json:"title"`
	NumVotes int    `json:"numVotes"`
}
