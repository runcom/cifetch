package main

import (
	"fmt"

	"github.com/codegangsta/cli"
)

// TODO(runcom): document args and usage
var manifestCommand = cli.Command{
	Name:      "manifest",
	Usage:     "",
	ArgsUsage: ``,
	Action: func(context *cli.Context) {
		fmt.Println("manifest")
	},
}
