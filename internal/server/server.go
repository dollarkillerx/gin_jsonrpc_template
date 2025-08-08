package server

import (
	"github.com/google/feitian/internal/api"
	"github.com/google/feitian/internal/conf"
	"github.com/google/feitian/internal/storage"
)

type Server struct {
	storage   *storage.Storage
	apiServer *api.ApiServer
	conf      conf.Config
}

func NewServer(storage *storage.Storage, conf conf.Config) *Server {
	return &Server{storage: storage, apiServer: api.NewApiServerWithDeps(storage, conf), conf: conf}
}

func (s *Server) Run() error { return s.apiServer.Run() }
