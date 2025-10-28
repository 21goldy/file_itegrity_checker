package cli

import (
	"strings"

	"github.com/21goldy/file_itegrity_checker.git/utils"
)

type WatchCmd struct {
	Filepath string `arg:"" help:"Enter the filepath to watch"`
}

func (w *WatchCmd) Run() error {
	utils.WatchFile(strings.TrimSpace(w.Filepath))
	return nil
}
