package sqlite

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"url-shortener/internal/storage"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

//go:embed migrations/001_init.sql
var schemaSQL string

func New(storagePath string, driverName string) (*Storage, error) {
	const op = "storage.sqlite.new"

	db, err := sql.Open(driverName, storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(schemaSQL)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUrl(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveUrl"

	stmt, err := s.db.Prepare("INSERT INTO url(url,alias) VALUES(?,?)")

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUrlExist)
		}

		return 0, fmt.Errorf("#{op}: #{err}")
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id:%w", op, err)
	}

	return id, nil
}

func (s *Storage) GetAlias(url string) (string, error) {
	const op = "storage.sqlite.GetAlias"

	var originalAlias string

	stmt, err := s.db.Prepare("SELECT alias FROM url WHERE url=?")

	if err != nil {
		return "", fmt.Errorf("%s: preparestatement: %w", op, err)
	}

	defer stmt.Close()

	err = stmt.QueryRow(url).Scan(&originalAlias)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrUrlNotFound
		}

		return "", fmt.Errorf("%s: executestatement: %w", op, err)
	}

	return originalAlias, nil
}

func (s *Storage) GetURLByAlias(alias string) (string, error) {
	const op = "storage.sqlite.GetURLByAlias"

	var originalUrl string

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias=?")

	if err != nil {
		return "", fmt.Errorf("%s: preparestatement: %w", op, err)
	}

	defer stmt.Close()

	err = stmt.QueryRow(alias).Scan(&originalUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrUrlNotFound
		}

		return "", fmt.Errorf("%s: executestatement: %w", op, err)
	}

	return originalUrl, nil
}

func (s *Storage) DeleteURL(url string) error {
	const op = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE url=?")

	if err != nil {
		return fmt.Errorf("%s: preparestatement: %w", op, err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(url)
	if err != nil {
		return fmt.Errorf("%s: executestatement: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteURLByAlias(alias string) error {
	const op = "storage.sqlite.DeleteURLByAlias"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias=?")

	if err != nil {
		return fmt.Errorf("%s: preparestatement: %w", op, err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: executestatement: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateURL(urlToSave string, alias string) error {
	const op = "storage.sqlite.UpdateURL"

	stmt, err := s.db.Prepare("UPDATE url SET url=? WHERE alias=?")

	if err != nil {
		return fmt.Errorf("%s: preparestatement: %w", op, err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(urlToSave, alias)
	if err != nil {
		return fmt.Errorf("%s: executestatement: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateAlias(url string, aliasToSave string) error {
	const op = "storage.sqlite.UpdateAlias"

	stmt, err := s.db.Prepare("UPDATE url SET alias =? WHERE url =?")
	if err != nil {
		return fmt.Errorf("%s: preparestatement: %w", op, err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(aliasToSave, url)
	if err != nil {
		return fmt.Errorf("%s: executestatement: %w", op, err)
	}

	return nil
}
