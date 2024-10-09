package view

import "github.com/rivo/tview"

type Toast struct {
	*tview.Modal
	msg string
}

func (kusApp *KusApp) setToast() *KusApp {
	kusApp.Toast = &Toast{
		Modal: tview.NewModal().AddButtons([]string{"Dismiss"}),
		msg:   "",
	}

	kusApp.Toast.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		kusApp.Root.RemovePage("toast")
		kusApp.Toast.SetText("")
	})

	return kusApp
}

func (kusApp *KusApp) toastMsg(msg string) {
	kusApp.Toast.SetText(msg)
	kusApp.Root.AddPage("toast", kusApp.Toast, true, true)
}
