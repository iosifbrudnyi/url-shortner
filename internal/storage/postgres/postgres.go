package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/iosifbrudnyi/url-shortner/internal/config"
	"github.com/iosifbrudnyi/url-shortner/internal/storage"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(db_config config.Db) (*Storage, error) {
	const op = "storage.postgres.New"
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s",
		db_config.Host, db_config.Port, db_config.User, db_config.Pass, db_config.Path)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(url string, alias string) (int64, error) {
	const op = "storage.postgres.SaveURL"
	stmt, err := s.db.Prepare(`INSERT INTO url(url, alias) VALUES ($1, $2) RETURNING id`)
	if err != nil {
		return 0, fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var id int64
	err = stmt.QueryRow(url, alias).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgres.GetURL"

	query, err := s.db.Prepare("SELECT url FROM url WHERE alias = $1")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resURL string

	err = query.QueryRow(alias).Scan(&resURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrURLNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: query row: %w", op, err)
	}

	return resURL, nil
}
