package run

import (
	"github.com/franchb/htmgo/cli/htmgo/v2/tasks/process"
)

func Setup() {
	process.RunOrExit(process.NewRawCommand("", "go mod download"))
	process.RunOrExit(process.NewRawCommand("", "go mod tidy"))
	MakeBuildable()
	EntGenerate()
}
