package store

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type Store struct {
	config      *Config
	db          *sql.DB
	sessionRepo *SessionRepo
}

func New(config *Config) *Store {
	return &Store{
		config: config,
	}
}

func (s *Store) Open(cs string) error {
	db, err := sql.Open("mysql", cs)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	s.db = db
	return nil
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) Session() *SessionRepo {
	if s.sessionRepo != nil {
		return s.sessionRepo
	}

	s.sessionRepo = &SessionRepo{
		store: s,
	}

	return s.sessionRepo
}
