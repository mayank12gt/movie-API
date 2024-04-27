package data

import (
	"context"
	"database/sql"
	"time"
)

type Rating struct {
	User_id    int64     `json:"user_id"`
	Movie_id   int64     `json:"movie_id"`
	Rating     float64   `json:"rating"`
	Created_at time.Time `json:"created_at"`
	Version    int32     `json:"version"`
}

type AverageRating struct {
	AverageRating float64 `json:"average_rating"`
	RatingCount   int64   `json:"rating_count"`
}

type RatingModel struct {
	DB *sql.DB
}

func (m *RatingModel) AddRating(rating *Rating) error {

	query := `INSERT INTO ratings (user_id, movie_id, rating) VALUES ($1,$2,$3) RETURNING user_id,movie_id,rating,created_at,version`

	args := []interface{}{rating.User_id, rating.Movie_id, rating.Rating}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&rating.User_id, &rating.Movie_id, &rating.Rating, &rating.Created_at, &rating.Version)

	return err
}

// func (m *RatingModel) UpdateRating(rating *Rating) error {

// 	query := `UPDATE users
// 	SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
// 	WHERE id = $5 AND version = $6
// 	RETURNING version`

// 	args := []interface{}{rating.User_id, rating.Movie_id, rating.Rating}

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&rating.User_id, &rating.Movie_id, &rating.Rating, &rating.Created_at, &rating.Version)

// 	return err
// }
