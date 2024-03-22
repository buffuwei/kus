package view

import "github.com/rivo/tview"

type Toast struct {
	*tview.Modal
	msg string
}

func (kusApp *KusApp) SetToast() *KusApp {
	kusApp.Err = &Toast{
		Modal: tview.NewModal().AddButtons([]string{"Dismiss"}),
		msg:   "",
	}

	kusApp.Err.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		kusApp.Root.RemovePage("err")
		kusApp.Err.SetText("")
	})

	return kusApp
}

func (kusApp *KusApp) ShowErr(msg string) {
	kusApp.Err.SetText(msg)
	kusApp.Root.AddPage("err", kusApp.Err, true, true)
}
