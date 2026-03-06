package mcpfx

import (
	"github.com/engolder/mcp-logseq/mcpext"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"
)

var ServerModule = fx.Module(
	"mcp.server",

	fx.Provide(fx.Annotate(
		func(
			impl *mcp.Implementation,
			registrars []mcpext.ToolRegistrar,
		) *mcp.Server {
			server := mcp.NewServer(impl, nil)
			for _, registrar := range registrars {
				registrar.Register(server)
			}
			return server
		},
		fx.ParamTags(``, `group:"mcp.tools"`),
	)),
)
