package main

import (
	"github.com/atlas-depin/cygnus/cmd"
)

func main() {
	root := cmd.RootCmd()

	cmd.Execute(root)
}
