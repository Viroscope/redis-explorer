package main

import (
	"redis-explorer/internal/ui"
)

func main() {
	app := ui.NewApp()
	app.Run()
}
