package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"time"
)

type MovieModel struct {
	Db *sql.DB
}

func (m MovieModel) Insert(movie *Movie) error {
	query := `INSERT INTO movies (title, year, runtime, genres) VALUES ($1, $2, $3, $4) RETURNING id, created_at, version`
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.Db.QueryRowContext(ctx, query, args...).Scan(&movie.Id, &movie.CreatedAt, &movie.Version)
}
func (m MovieModel) Read(id int64) (*Movie, error) {
	if id < 1 {

		return nil, ErrRecordNotFound
	}
	query := `SELECT pg_sleep(10), id, created_at, title, year, runtime, genres, version FROM movies WHERE id = $1`
	var movie Movie
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.Db.QueryRowContext(ctx, query, id).Scan(&movie.Id, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err

		}
	}
	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {
	query := `UPDATE movies SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1 WHERE id = $5 AND version=$6 RETURNING version`
	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.Id,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.Db.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil

}
func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `DELETE FROM movies WHERE id = $1`
	result, err := m.Db.Exec(query, id)
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
func (m MovieModel) GetAll(title string, genres []string, filter Filters) ([]*Movie, Metadata, error) {
	query := fmt.Sprintf(`SELECT  count(*) OVER(), id, created_at, title, year, runtime, genres, version FROM movies WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '') AND (genres @> $2 OR $2 = '{}') ORDER BY %s %s, id ASC LIMIT $3 OFFSET $4`, filter.SortColumn(), filter.SortDirection())
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []any{title, pq.Array(genres), filter.limit(), filter.offset()}
	rows, err := m.Db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()
	totalRecords := 0
	movies := []*Movie{}
	for rows.Next() {
		var movie Movie

		err := rows.Scan(
			&totalRecords,
			&movie.Id,
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
	metadata := calculateMetadata(totalRecords, filter.Page, filter.PageSize)
	return movies, metadata, nil

}
