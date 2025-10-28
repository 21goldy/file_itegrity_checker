package cli

import "github.com/alecthomas/kong"

type CLI struct {
	Watch   WatchCmd   `cmd:"" help:"This command starts watching over the given file for any hash changes and gives an alert, when a new hash is found!"`
	History HistoryCmd `cmd:"" help:"This command takes in a filepath as string, and prints the complete hashing history of the file in the '<Timestamp> <Hash>' Format"`
}

func RunCli() error {
	cli := CLI{}

	ctx := kong.Parse(&cli,
		kong.Name("fileintegritytool"),
		kong.Description("This tool allows to compare hashes, watch file in-real time and print the hash history of a file"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	return ctx.Run()
}
