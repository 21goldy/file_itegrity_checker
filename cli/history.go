package cli

import (
	"strings"

	"github.com/21goldy/file_itegrity_checker.git/utils"
)

type HistoryCmd struct {
	Filepath string `arg:"" help:"Enter filepath to watch hash history"`
}

func (h *HistoryCmd) Run() error {
	utils.PrintHashHistory(strings.TrimSpace(h.Filepath))
	return nil
}
