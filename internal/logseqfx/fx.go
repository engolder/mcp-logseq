package logseqfx

import (
	"github.com/engolder/mcp-logseq/internal/logseq"
	"github.com/engolder/mcp-logseq/internal/logseq/logseqcontroller"
	"github.com/engolder/mcp-logseq/internal/logseq/logseqsvc"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"logseq",

	fx.Provide(logseq.NewClient),

	fx.Provide(logseqsvc.NewBlockSvc),
	fx.Provide(logseqsvc.NewPageSvc),
	fx.Provide(logseqsvc.NewSearchSvc),

	fx.Provide(fx.Annotate(
		logseqcontroller.NewGetBlockTool,
		fx.ResultTags(`group:"mcp.tools"`),
	)),

	fx.Provide(fx.Annotate(
		logseqcontroller.NewSearchBlocksTool,
		fx.ResultTags(`group:"mcp.tools"`),
	)),

	fx.Provide(fx.Annotate(
		logseqcontroller.NewGetPageTool,
		fx.ResultTags(`group:"mcp.tools"`),
	)),

	fx.Provide(fx.Annotate(
		logseqcontroller.NewCreatePageTool,
		fx.ResultTags(`group:"mcp.tools"`),
	)),

	fx.Provide(fx.Annotate(
		logseqcontroller.NewListPagesTool,
		fx.ResultTags(`group:"mcp.tools"`),
	)),

	fx.Provide(fx.Annotate(
		logseqcontroller.NewListJournalPagesTool,
		fx.ResultTags(`group:"mcp.tools"`),
	)),

	fx.Provide(fx.Annotate(
		logseqcontroller.NewInsertBlockTool,
		fx.ResultTags(`group:"mcp.tools"`),
	)),

	fx.Provide(fx.Annotate(
		logseqcontroller.NewUpdateBlockTool,
		fx.ResultTags(`group:"mcp.tools"`),
	)),
)
