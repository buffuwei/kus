package view

import (
	"buffuwei/kus/kuboard"
	"buffuwei/kus/tools"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

var CYAN_COLOR = tcell.GetColor("#87CDFA")
var LOGO_COLOR = tcell.GetColor("#FBA202")

const (
	PageLogger   string = "logger"
	PageShell    string = "shell"
	PagePortal   string = "portal"
	PageCluster  string = "cluster"
	PagePipeline string = "pipeline"
	PagePodBoard string = "podboard"
)

type event string

const (
	EVENT_PREPARED = "prepared"
)

type KusApp struct {
	*tview.Application
	Root          *tview.Pages
	Cluster       *ClusterF
	Portal        *PortalF
	Shell         *ShellF
	Logger        *LoggerF
	Pipeline      *PipelineF
	Toast         *Toast
	Cacher        *GlobalCacher
	EventCh       chan event
	EventHandlers []func(event)
}

func StartApplication() {
	prerequisite()
	kusApp := newKusApp()
	// kusApp.EventCh <- EVENT_PREPARED

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
		EventCh:     make(chan event, 10),
	}

	// go func() {
	// 	for {
	// 		evt := <-kusApp.EventCh
	// 		for i, handler := range kusApp.EventHandlers {
	// 			zap.S().Debugf("event handler called: %d - %v\n", i, handler)
	// 			handler(evt)
	// 		}
	// 	}
	// }()

	kusApp.SetCacher().
		SetPortal().
		SetCluster().
		SetShell().
		SetLogger().
		SetPipeline().
		setToast()

	fn := kusApp.GetBeforeDrawFunc()
	kusApp.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		// zap.S().Infof("before draw func called\n")
		return fn(screen)
	})

	conf := tools.GetConfig()

	if conf.Selected.Cluster != "" && conf.Selected.Namespace != "" {
		kusApp.Root.AddPage(PagePortal, kusApp.Portal, true, true).
			AddPage(PageCluster, kusApp.Cluster, true, false)
	} else {
		kusApp.Root.AddPage(PagePortal, kusApp.Portal, true, false).
			AddPage(PageCluster, kusApp.Cluster, true, true)
	}
	kusApp.Root.AddPage(PageShell, kusApp.Shell, true, false).
		AddPage(PageLogger, kusApp.Logger, true, false).
		AddPage(PagePipeline, kusApp.Pipeline, true, false)

	kusApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlQ {
			kusApp.Stop()
		} else if event.Key() == tcell.KeyCtrlC {
			// prevent ctrl-c to stop
			// return tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModNone)
		}
		return event
	})

	return kusApp
}

func prerequisite() {
	for i := 0; i < 3; i++ {
		_, err := kuboard.GetSelfName()
		if err != nil {
			refreshToken()
			continue
		}
		return
	}
	fmt.Printf("Failed to checkout token, existed!\n")
	// if kuboard.Cookie() == "" || kuboard.CheckTokenFailed() {
	// 	username := tools.GetConfig().Kuboard.Username
	// 	password := tools.GetConfig().Kuboard.Password
	// 	token, err := kuboard.NewToken(username, password)
	// 	if err != nil {
	// 		fmt.Printf("%s (failed to get kuboard token)\n", err.Error())
	// 		fmt.Printf("please check your config file: %s \n", tools.ConfigPath())
	// 		os.Exit(1)
	// 	}

	// 	tools.GetConfig().UpdateToken(token)
	// 	_, err = kuboard.GetSelfName()
	// 	if err != nil {
	// 		fmt.Printf("%s (network or login problem)\n", err.Error())
	// 		fmt.Printf("please check your config file: %s \n", tools.ConfigPath())
	// 		os.Exit(1)
	// 	}
	// }

	go tools.Clean(tools.HomeDir())
}

func refreshToken() {
	username := tools.GetConfig().Kuboard.Username
	password := tools.GetConfig().Kuboard.Password
	for i := 0; i < 3; i++ {
		token, err := kuboard.NewToken(username, password)
		if err != nil {
			zap.S().Errorf("New token err: %s \n", err.Error())
			continue
		}
		tools.GetConfig().UpdateToken(token)
		return
	}
	fmt.Printf("Failed to refresh the token three times : %s %s \n", username, password)
}
