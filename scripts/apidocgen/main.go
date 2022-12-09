package main

import (
	"os"

	chidocgen "github.com/go-chi/docgen"

	"github.com/coder/coder/coderd"
)

func main() {
	api := coderd.New(nil)
	os.Setenv("GOPATH", "gopath")
	chidocgen.PrintRoutes(api.RootHandler)
}
