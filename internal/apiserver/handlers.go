package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bolshagin/xsolla-be-2020/model"
	"github.com/google/uuid"
	"net/http"
	"time"
)

var (
	zeroDate = time.Date(1000, 1, 1, 0, 0, 0, 0, time.UTC)
)

var (
	errSessionExpired       = errors.New("session expired")
	errSessionAlreadyClosed = errors.New("session already closed")
	errTooLongPurpose       = errors.New("purpose must be less than 210 symbols")
	errInvalidCardNum       = errors.New("invalid card number")
	errInvalidCardDate      = errors.New("invalid card date")
	errInvalidCardCode      = errors.New("invalid card code")
)

func (s *APIServer) handleSessionsCreate() http.HandlerFunc {
	type request struct {
		Amount  float64 `json:"amount"`
		Purpose string  `json:"purpose"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		if len(req.Purpose) > 210 {
			s.logger.Error(errTooLongPurpose)
			s.error(w, r, http.StatusBadRequest, errTooLongPurpose)
			return
		}

		session := &model.Session{
			Amount:  req.Amount,
			Purpose: req.Purpose,
		}

		session.SessionToken = uuid.New().String()
		session.CreatedAt = s.now()

		if err := s.store.Session().Create(session); err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		s.logger.Info(fmt.Sprintf("create session with token %v", session.SessionToken))
		s.respond(w, r, http.StatusCreated, session)
	}
}

func (s *APIServer) handlePayment() http.HandlerFunc {
	type request struct {
		SessionToken string `json:"session_token"`
		CardNumber   string `json:"card_number"`
		Code         string `json:"code"`
		Date         string `json:"date"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		session, err := s.store.Session().FindByToken(req.SessionToken)
		if err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		if !isZeroDate(session.ClosedAt) {
			s.logger.Error(errSessionAlreadyClosed)
			s.error(w, r, http.StatusBadRequest, errSessionAlreadyClosed)
			return
		}

		closedAt := s.now()
		delta := closedAt.Sub(session.CreatedAt)
		if delta.Minutes() > 15 {
			s.logger.Error(fmt.Sprintf("session token %v expired", session.SessionToken))
			s.error(w, r, http.StatusBadRequest, errSessionExpired)
			return
		}

		if !IsCardDate(req.Date) {
			s.logger.Error(errInvalidCardDate)
			s.error(w, r, http.StatusBadRequest, errInvalidCardDate)
			return
		}

		if !IsCardCode(req.Code) {
			s.logger.Error(errInvalidCardDate)
			s.error(w, r, http.StatusBadRequest, errInvalidCardCode)
			return
		}

		if !IsCreditCard(req.CardNumber) {
			s.logger.Error(errInvalidCardNum)
			s.error(w, r, http.StatusBadRequest, errInvalidCardNum)
			return
		}

		if err := s.store.Session().CommitSession(session, closedAt); err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.logger.Info(fmt.Sprintf("session %v successfuly closed", session.SessionToken))
		s.respond(w, r, http.StatusOK, map[string]string{"payment": "successful"})
	}
}

func (s *APIServer) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *APIServer) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (s *APIServer) now() time.Time {
	loc, _ := time.LoadLocation("UTC")
	return time.Now().In(loc)
}

func isZeroDate(t time.Time) bool {
	if t.Equal(zeroDate) {
		return true
	}
	return false
}
