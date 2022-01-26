package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main()  {
	a := app.New()
	w := a.NewWindow("hello")
	hello := widget.NewLabel("hello Fyne!")
	w.SetContent(container.NewVBox(
		hello,
		widget.NewButton("hi", func() {
			hello.SetText("welcome")
		}),
	))
	w.ShowAndRun()
}


