package data

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
)

type MovieModel struct {
	DB *sql.DB
}

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"` // Use the - directive
	Title     string    `json:"title" validate:"required,min=1,max=100"`
	Year      int32     `json:"year,omitempty" validate:"required,min=1888"`
	Runtime   int32     `json:"runtime,omitempty" validate:"required,min=20"`
	Genres    []string  `json:"genres,omitempty" validate:"required,min=1,max=10,unique"`
	Version   int32     `json:"version" validate:"omitempty,min=1"`
}

func maxCurrentYear(fl validator.FieldLevel) bool {
	currentYear := (int64)(time.Now().Year())
	year := fl.Field().Int()
	return year <= currentYear
}

func (m MovieModel) Insert(movie *Movie) error {

	query := "INSERT INTO movies (title, year, runtime, genres) VALUES($1, $2, $3, $4) RETURNING id, created_at, version"

	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := "SELECT * FROM movies WHERE id=$1"

	var movie Movie

	if err := m.DB.QueryRow(query, id).Scan(&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version); err != nil {
		return nil, err
	}

	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {

	query := `UPDATE movies SET title=$1, year=$2, runtime=$3, genres=$4 WHERE id = $5 RETURNING id, created_at, title, year ,runtime,genres, version`

	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.ID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)
	if err != nil {
		log.Print(err.Error())
		return err
	}

	return nil
}

func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := "DELETE from movies WHERE id=$1"

	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m MovieModel) List(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {

	// 	query := `SELECT id, created_at, title, year, runtime, genres, version
	// FROM movies WHERE (Lower(title)=Lower($1) OR $1='') AND (genres @>$2 OR $2='{}')
	// ORDER BY id`

	query := fmt.Sprintf(`SELECT count(*) OVER(), id, created_at, title, year, runtime, genres, version
FROM movies WHERE (to_tsvector('simple',title) @@ plainto_tsquery('simple',$1) OR $1='') AND (genres @>$2 OR $2='{}')
ORDER BY %s %s,id ASC LIMIT $3 OFFSET $4`, strings.TrimPrefix(filters.Sort, "-"), filters.sortDirection())

	log.Print(query)

	// ctx, cancel := con.WithTimeout(context.Background(), 3*time.Second)
	// defer cancel()

	rows, err := m.DB.Query(query, title, pq.Array(genres), filters.limit(), filters.offset())
	if err != nil {
		log.Print(err)
		return nil, Metadata{}, err
	}

	defer rows.Close()

	movies := []*Movie{}

	totalRecords := 0

	for rows.Next() {

		var movie Movie

		err := rows.Scan(
			&totalRecords,
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		movies = append(movies, &movie)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	meta := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return movies, meta, nil

}

func (m *MovieModel) GetAverageRating(movie_ID int64) (*AverageRating, error) {

	//query := `SELECT AVG(rating) AS average_rating,count(*) AS rating_count FROM ratings WHERE movie_id = $1`
	query := `SELECT COALESCE(AVG(rating), 0) AS average_rating,COALESCE(count(*), 0) AS rating_count FROM ratings
	WHERE movie_id = $1;`

	var averageRating AverageRating

	if err := m.DB.QueryRow(query, movie_ID).Scan(&averageRating.AverageRating, &averageRating.RatingCount); err != nil {
		if err == sql.ErrNoRows {
			// No ratings found for the movie, set averageRating to 0
			averageRating.AverageRating = 0
			averageRating.RatingCount = 0
			return &averageRating, nil
		}
		return nil, err
	}

	return &averageRating, nil
}
