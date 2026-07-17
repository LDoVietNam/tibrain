package db

import (
	"database/sql"

	"github.com/ti/router/tibrain/internal/mcp"
	_ "modernc.org/sqlite"
)

type Hub struct {
	db     *sql.DB
	mcpSrv *mcp.MCPServer
}

func NewHub(db *sql.DB) *Hub {
	h := &Hub{db: db, mcpSrv: mcp.NewMCPServer()}
	mcp.RegisterFileTools(h.mcpSrv)
	return h
}

func (h *Hub) GetDB() *sql.DB {
	return h.db
}

func (h *Hub) GetMCPServer() *mcp.MCPServer {
	return h.mcpSrv
}