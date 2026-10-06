package main

import (
	"os"
	"runtime/debug"

	"github.com/GuitarWag/clitasks/v3/internal/cli"
)

// version is set by `make` through -ldflags.
var version = "dev"

func main() {
	os.Exit(cli.Execute(buildVersion()))
}

// buildVersion falls back to the module version that `go install
// module@version` records, because that path does not pass -ldflags.
func buildVersion() string {
	if version != "dev" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return version
}
