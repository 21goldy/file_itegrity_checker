package main

import (
	"log"

	"github.com/21goldy/file_itegrity_checker.git/cli"
)

func main() {
	if err := cli.RunCli(); err != nil {
		log.Fatalln(err)
	}
}
