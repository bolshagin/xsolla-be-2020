package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bolshagin/xsolla-be-2020/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

var (
	zeroDate             = time.Date(1000, 1, 1, 0, 0, 0, 0, time.UTC)
	sessDuration float64 = 60 * 15
	layout               = "2006-01-02"
	secretKey            = []byte("secretKey")
)

var (
	errSessionExpired       = errors.New("session expired")
	errSessionAlreadyClosed = errors.New("session already closed")
	errTooLongPurpose       = errors.New("purpose must be less than 210 symbols")
	errInvalidCardNum       = errors.New("invalid card number")
	errInvalidCardDate      = errors.New("invalid card date")
	errInvalidCardCode      = errors.New("invalid card code")
	errNotAuthorized        = errors.New("not authorized")
	errNotValidToken        = errors.New("not valid jwt-token")
	errTokenIsExpired       = errors.New("jwt-token is expired")
	errCantHandleToken      = errors.New("cant handle jwt-token")
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
		if delta.Seconds() > sessDuration {
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

func (s *APIServer) handleSessionsStats() http.HandlerFunc {
	type request struct {
		DateBegin string `json:"date_begin"`
		DateEnd   string `json:"date_end"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		dateB, dateE, err := s.parseDates(req.DateBegin, req.DateEnd)
		if err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		var sessions []model.Session
		sessions, err = s.store.Session().GetStats(dateB, dateE)
		if err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, sessions)
	}
}

func (s *APIServer) handleTokenCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := &jwt.MapClaims{
			"exp":  time.Now().Add(time.Hour * time.Duration(1)).Unix(),
			"user": "Default User",
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenS, err := token.SignedString(secretKey)
		if err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.logger.Info(fmt.Sprintf("create jwt token %v", tokenS))
		s.respond(w, r, http.StatusCreated, map[string]string{"jwt_token": tokenS})
	}
}

func checkJWTToken(s *APIServer, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenH := r.Header.Get("Authorization")
		s.logger.Info(fmt.Sprintf("try authorizate with token header `%v`", tokenH))

		auth := strings.SplitN(tokenH, " ", 2)
		if len(auth) != 2 {
			s.logger.Error(errNotAuthorized)
			s.error(w, r, http.StatusUnauthorized, errNotAuthorized)
			return
		}

		claims := &jwt.MapClaims{
			"exp":  time.Now().Add(time.Hour * time.Duration(1)).Unix(),
			"user": "Default User",
		}

		var tokenS = auth[1]
		token, err := jwt.ParseWithClaims(tokenS, claims, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})

		if token.Valid {
			next(w, r)
		} else if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				s.logger.Error(errNotValidToken)
				s.error(w, r, http.StatusUnauthorized, errNotValidToken)
				return
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				s.logger.Error(errTokenIsExpired)
				s.error(w, r, http.StatusUnauthorized, errTokenIsExpired)
				return
			} else {
				s.logger.Error(errCantHandleToken)
				s.error(w, r, http.StatusUnauthorized, errCantHandleToken)
				return
			}
		} else {
			s.logger.Error(errCantHandleToken)
			s.error(w, r, http.StatusUnauthorized, errCantHandleToken)
			return
		}
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

func (s *APIServer) parseDates(begin, end string) (time.Time, time.Time, error) {
	dateB, err := time.Parse(layout, begin)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	dateE, err := time.Parse(layout, end)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return dateB, dateE, nil
}

func isZeroDate(t time.Time) bool {
	if t.Equal(zeroDate) {
		return true
	}
	return false
}
