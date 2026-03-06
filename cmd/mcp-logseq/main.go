package main

import (
	"context"
	"flag"
	"log"

	"github.com/engolder/mcp-logseq/internal/config"
	"github.com/engolder/mcp-logseq/internal/httpfx"
	"github.com/engolder/mcp-logseq/internal/logseqfx"
	"github.com/engolder/mcp-logseq/internal/mcpfx"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"
)

var args struct {
	httpEnable  bool
	httpPort    int
	stdioEnable bool
}

var configOption = fx.Options(
	fx.Provide(func() *config.Config {
		return &config.Config{
			HTTP: struct {
				Enable     bool
				ServerPort int
			}{
				Enable:     args.httpEnable,
				ServerPort: args.httpPort,
			},
			Stdio: struct {
				Enable bool
			}{
				Enable: args.stdioEnable,
			},
		}
	}),
)

var mcpImplementationOption = fx.Options(
	fx.Provide(func() *mcp.Implementation {
		return &mcp.Implementation{
			Name:    "mcp-logseq",
			Version: "0.0.1",
		}
	}),
)

var mcpStdioServerOption = fx.Options(
	fx.Invoke(func(
		cfg *config.Config,
		server *mcp.Server,
		lc fx.Lifecycle,
	) {
		if !cfg.Stdio.Enable {
			return
		}

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
)

func main() {
	flag.BoolVar(&args.httpEnable, "http", false, "enable HTTP server")
	flag.IntVar(&args.httpPort, "http-port", 8080, "HTTP port to listen on")
	flag.BoolVar(&args.stdioEnable, "stdio", true, "enable stdio server")
	flag.Parse()

	fx.New(
		configOption,
		mcpImplementationOption,
		mcpStdioServerOption,

		logseqfx.Module,
		mcpfx.ServerModule,
		httpfx.ServerModule,
	).Run()
}
