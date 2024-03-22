package view

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var Version string = "0.1"
var Commit string

type TopInfo struct {
	*tview.Flex
	podtable *PortalF
	info     *tview.TextView
	helps    *tview.Flex
	logo     *tview.TextView
	C        chan string
}

func (podtable *PortalF) SetTopInfo() *PortalF {
	t := &TopInfo{
		podtable: podtable,
		info:     tview.NewTextView().SetTextColor(tcell.ColorYellowGreen),
		helps:    newHelps(),
		logo:     tview.NewTextView().SetText(fmt.Sprintf(LOGO, Version, Commit)).SetTextColor(LOGO_COLOR),
		Flex:     tview.NewFlex().SetDirection(tview.FlexColumn),
		C:        make(chan string, 10),
	}
	t.AddItem(t.info, 0, 1, false).
		AddItem(t.helps, 0, 4, false).
		AddItem(t.logo, 0, 1, false).
		SetBorderPadding(0, 0, 2, 0)

	go func() {
		for msg := range t.C {
			t.handle(msg)
		}
	}()

	podtable.topInfo = t
	return podtable
}

func (topInfo *TopInfo) handle(msg string) {
	if msg == "refresh" {
		topInfo.podtable.kusApp.QueueUpdate(func() {
			topInfo.refresh()
		})
	}
}

func newHelps() *tview.Flex {
	f := tview.NewFlex().SetDirection(tview.FlexColumn)
	f.AddItem(tview.NewTextView().SetTextColor(tcell.ColorYellowGreen).SetText(help1), 0, 1, false)
	f.AddItem(tview.NewTextView().SetTextColor(tcell.ColorYellowGreen).SetText(help2), 0, 1, false)
	f.AddItem(tview.NewTextView().SetTextColor(tcell.ColorYellowGreen).SetText(help3), 0, 1, false)
	f.AddItem(tview.NewTextView().SetTextColor(tcell.ColorYellowGreen).SetText(help4), 0, 1, false)
	return f
}

func (topInfo *TopInfo) refresh() {
	clu := topInfo.podtable.cluster
	ns := topInfo.podtable.namespace
	logPer := topInfo.podtable.kusApp.Logger.percentLogger(clu, ns)
	shellPer := topInfo.podtable.kusApp.Shell.percentShell(clu, ns)
	i := fmt.Sprintf(infoTmpl, clu, ns, logPer, shellPer)
	topInfo.info.SetText(i)
}

var LOGO string = `_  ___   _ ___ 
| |/ / | | / __|
| ' <| |_| \__ \  Version: %s
|_|\_\\___/|___/  Commit: %s`

var infoTmpl string = `Cluster: %s
Namespace: %s
Logger: %s
Shell: %s
`

var help1 string = `<:> Command input
</> Fileter input
<s/S> Sh embed / web
<l/L> Log viewer`

var help2 string = `<:ctx>  Show cluster-ns page
<ctrl-q> Quit`

var help3 string = ``

var help4 string = `<j/k>     Up / Down
<g/G>     Top / Bottom
<ctrl-b>  Page Up
<ctrl-f>  Page Down`
