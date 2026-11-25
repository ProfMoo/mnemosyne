// Command frontend runs the go-app UI in WebAssembly.
package main

import (
	"github.com/maxence-charriere/go-app/v11/pkg/app"

	"github.com/profmoo/mnemosyne/internal/ui"
)

func main() {
	ui.RegisterRoutes()
	app.RunWhenOnBrowser()
}
