package mcp

import (
	"github.com/jlentink/twinmind-mcp/internal/api"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	api *api.Client
	mcp *server.MCPServer
}

func New(apiClient *api.Client) *Server {
	s := &Server{api: apiClient}
	s.mcp = server.NewMCPServer(
		"twinmind-mcp",
		"1.0.0",
	)
	s.registerTools()
	return s
}

func (s *Server) Serve() error {
	return server.ServeStdio(s.mcp)
}
