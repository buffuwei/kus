package view

import (
	"buffuwei/kus/kuboard"
	"buffuwei/kus/tools"
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

var CYAN_COLOR = tcell.GetColor("#87CDFA")
var LOGO_COLOR = tcell.GetColor("#FBA202")

const (
	PageLogger  string = "logger"
	PageShell   string = "shell"
	PagePortal  string = "portal"
	PageCluster string = "cluster"
)

type KusApp struct {
	*tview.Application
	Root    *tview.Pages
	Cluster *ClusterF
	Portal  *PortalF
	Shell   *ShellF
	Logger  *LoggerF
	Err     *Toast
	Cacher  *GlobalCacher
}

func StartApplication() {
	prerequisite()
	kusApp := newKusApp()
	go kusApp.serv()
	if err := kusApp.SetRoot(kusApp.Root, true).SetFocus(kusApp.Root).Run(); err != nil {
		zap.S().Error(err)
		kusApp.Stop()
	}
}

func newKusApp() *KusApp {
	kusApp := &KusApp{
		Application: tview.NewApplication().EnableMouse(true).EnablePaste(true),
		Root:        tview.NewPages(),
	}

	kusApp.SetCacher().
		SetPortal().
		SetCluster().
		SetShell().
		SetLogger().
		SetToast()

	conf := tools.GetConfig()

	if conf.Selected.Cluster != "" && conf.Selected.Namespace != "" {
		kusApp.Root.AddPage(PagePortal, kusApp.Portal, true, true).
			AddPage(PageCluster, kusApp.Cluster, true, false)
	} else {
		kusApp.Root.AddPage(PagePortal, kusApp.Portal, true, false).
			AddPage(PageCluster, kusApp.Cluster, true, true)
	}
	kusApp.Root.AddPage(PageShell, kusApp.Shell, true, false).
		AddPage(PageLogger, kusApp.Logger, true, false)

	kusApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlQ {
			kusApp.Stop()
		} else if event.Key() == tcell.KeyCtrlC {
			// prevent ctrl-c to stop
			return tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModNone)
		}
		return event
	})

	return kusApp
}

func prerequisite() {
	if kuboard.Cookie() == "" || kuboard.CheckTokenFailed() {
		username := tools.GetConfig().Kuboard.Username
		password := tools.GetConfig().Kuboard.Password
		token, err := kuboard.NewToken(username, password)
		if err != nil {
			fmt.Printf("%s (network or login problem)\n", err.Error())
			fmt.Printf("please check your config file: %s \n", tools.ConfigPath())
			os.Exit(1)
		}

		tools.GetConfig().UpdateToken(token)
		_, err = kuboard.GetSelfName()
		if err != nil {
			fmt.Printf("%s (network or login problem)\n", err.Error())
			fmt.Printf("please check your config file: %s \n", tools.ConfigPath())
			os.Exit(1)
		}
	}

	go tools.Clean(tools.HomeDir())
}
