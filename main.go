package main

import (
	"embed"

	"goterm/backend"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	backend.Run(assets)
}
