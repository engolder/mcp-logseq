package httpfx

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/engolder/mcp-logseq/internal/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"
)

var ServerModule = fx.Module(
	"http.server",

	fx.Provide(func(mcpServer *mcp.Server, cfg *config.Config) *http.Server {
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return mcpServer
		}, nil)
		return &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.HTTP.ServerPort),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 120 * time.Second,
			Handler:      handler,
		}
	}),

	fx.Invoke(func(
		cfg *config.Config,
		server *http.Server,
		lc fx.Lifecycle,
	) {
		if !cfg.HTTP.Enable {
			return
		}

		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					log.Println("Starting HTTP server on", server.Addr)
					if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						log.Println("Failed to start HTTP server:", err)
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return server.Shutdown(ctx)
			},
		})
	}),
)
