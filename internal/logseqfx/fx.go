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
		logseqcontroller.NewDeletePageTool,
		fx.ResultTags(`group:"mcp.tools"`),
	)),

	fx.Provide(fx.Annotate(
		logseqcontroller.NewReadPageTool,
		fx.ResultTags(`group:"mcp.tools"`),
	)),

	fx.Provide(fx.Annotate(
		logseqcontroller.NewWritePageTool,
		fx.ResultTags(`group:"mcp.tools"`),
	)),

	fx.Provide(fx.Annotate(
		logseqcontroller.NewEditPageTool,
		fx.ResultTags(`group:"mcp.tools"`),
	)),

	fx.Provide(fx.Annotate(
		logseqcontroller.NewSearchPagesTool,
		fx.ResultTags(`group:"mcp.tools"`),
	)),

	fx.Provide(fx.Annotate(
		logseqcontroller.NewSearchTool,
		fx.ResultTags(`group:"mcp.tools"`),
	)),

	fx.Provide(fx.Annotate(
		logseqcontroller.NewListJournalPagesTool,
		fx.ResultTags(`group:"mcp.tools"`),
	)),
)
