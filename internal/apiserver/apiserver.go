package apiserver

import (
	"fmt"
	"github.com/bolshagin/xsolla-be-2020/store"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
)

type APIServer struct {
	config *Config
	logger *logrus.Logger
	router *mux.Router
	store  *store.Store
}

func New(config *Config) *APIServer {
	return &APIServer{
		config: config,
		logger: logrus.New(),
		router: mux.NewRouter(),
	}
}

func (s *APIServer) Start() error {
	if err := s.configureLogger(); err != nil {
		return err
	}
	s.configureRouter()

	if err := s.configureStore(); err != nil {
		return err
	}

	s.logger.Info("starting api server")

	return http.ListenAndServe(s.config.BindAddr, s.router)
}

func (s *APIServer) configureLogger() error {
	level, err := logrus.ParseLevel(s.config.LogLevel)
	if err != nil {
		return err
	}
	s.logger.SetLevel(level)
	return nil
}

func (s *APIServer) configureStore() error {
	st := store.New(s.config.Store)
	cs := getConnectionString(s.config)

	if err := st.Open(cs); err != nil {
		return err
	}
	s.store = st
	return nil
}

func (s *APIServer) configureRouter() {
	s.router.HandleFunc("/session", s.handleSessionsCreate()).Methods("POST")
	s.router.HandleFunc("/pay", s.handlePayment()).Methods("POST")
	s.router.HandleFunc("/stat", checkJWTToken(s, s.handleSessionsStats())).Methods("GET")
	s.router.HandleFunc("/get-token", s.handleTokenCreate()).Methods("GET")
}

func getConnectionString(config *Config) string {
	return fmt.Sprintf(
		"%s:%s@/%s?parseTime=true",
		config.Store.User,
		config.Store.Password,
		config.Store.DBName,
	)
}
