package logseqfx

import (
	"github.com/engolder/mcp-logseq/internal/logseq"
	"github.com/engolder/mcp-logseq/internal/logseq/logseqcontroller"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"logseq",

	fx.Provide(logseq.NewClient),

	fx.Provide(fx.Annotate(
		logseqcontroller.NewGetBlockTool,
		fx.ResultTags(`group:"mcp.tools"`),
	)),
)
