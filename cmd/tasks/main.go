package main

import (
	"os"

	"github.com/GuitarWag/clitasks/internal/cli"
)

var version = "dev"

func main() {
	os.Exit(cli.Execute(version))
}
