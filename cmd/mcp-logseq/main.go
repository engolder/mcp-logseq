package main

import (
	"context"
	"log"

	"github.com/engolder/mcp-logseq/internal/config"
	"github.com/engolder/mcp-logseq/internal/logseqfx"
	"github.com/engolder/mcp-logseq/internal/mcpfx"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"
)

var mcpImplementationOption = fx.Options(
	fx.Provide(func() *mcp.Implementation {
		return &mcp.Implementation{
			Name:    "mcp-logseq",
			Version: "0.0.1",
		}
	}),
)

var configOption = fx.Options(
	fx.Provide(func() *config.Config {
		return &config.Config{}
	}),
)

func main() {
	fx.New(
		configOption,
		mcpImplementationOption,

		fx.Invoke(func(
			server *mcp.Server,
			lc fx.Lifecycle,
		) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						log.Println("Starting stdio server")
						if err := server.Run(context.Background(), mcp.NewStdioTransport()); err != nil {
							log.Println("Failed to start stdio server:", err)
						}
					}()
					return nil
				},
			})
		}),

		logseqfx.Module,
		mcpfx.ServerModule,
	).Run()
}
