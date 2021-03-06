package store

import (
	"database/sql"
	"errors"
	"github.com/bolshagin/xsolla-be-2020/model"
	"time"
)

var (
	errNoSession = errors.New("there is no session with given token")
	errNoStats   = errors.New("there is no created sessions with given period")
)

type SessionRepo struct {
	store *Store
}

func (r *SessionRepo) Create(s *model.Session) error {
	_, err := r.store.db.Exec(
		"INSERT INTO sessions (SessionToken, Amount, Purpose, CreatedAt) VALUES (?, ?, ?, ?)",
		s.SessionToken,
		s.Amount,
		s.Purpose,
		s.CreatedAt)

	if err != nil {
		return err
	}

	if err := r.store.db.QueryRow("SELECT LAST_INSERT_ID()").Scan(&s.SessionID); err != nil {
		return err
	}

	return nil
}

func (r *SessionRepo) FindByToken(token string) (*model.Session, error) {
	s := &model.Session{}

	if err := r.store.db.QueryRow(
		`SELECT SessionID, SessionToken, Amount, Purpose, CreatedAt, ClosedAt FROM sessions WHERE SessionToken = ?`,
		token).Scan(
		&s.SessionID,
		&s.SessionToken,
		&s.Amount,
		&s.Purpose,
		&s.CreatedAt,
		&s.ClosedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errNoSession
		}
		return nil, err
	}

	return s, nil
}

func (r *SessionRepo) CommitSession(s *model.Session, closedAt time.Time) error {
	_, err := r.store.db.Exec(
		"UPDATE sessions SET ClosedAt = ? WHERE SessionToken = ?",
		closedAt,
		s.SessionToken,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *SessionRepo) GetStats(begin, end time.Time) ([]model.Session, error) {
	rows, err := r.store.db.Query(
		"SELECT Amount, Purpose, CreatedAt, ClosedAt FROM sessions WHERE CreatedAt BETWEEN ? AND ? ORDER BY CreatedAt DESC",
		begin,
		end,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []model.Session
	for rows.Next() {
		var s model.Session
		if err := rows.Scan(&s.Amount, &s.Purpose, &s.CreatedAt, &s.ClosedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	if sessions == nil {
		return nil, errNoStats
	}

	return sessions, nil
}
