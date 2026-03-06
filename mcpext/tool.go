package mcpext

import "github.com/modelcontextprotocol/go-sdk/mcp"

type ToolRegistrar interface {
	Register(server *mcp.Server)
}
