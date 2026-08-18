package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("TAP")
	w.SetContent(widget.NewLabel("Ça compile !"))
	w.ShowAndRun()
}
