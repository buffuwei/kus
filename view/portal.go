package view

import (
	"buffuwei/kus/tools"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

// portal of kus-app : input top-info toptable
type PortalF struct {
	*tview.Flex
	kusApp                 *KusApp
	topInfo                *TopInfo
	inputCmd               *tview.InputField
	inputFilter            *tview.InputField
	filterChangedMicorTime int64
	podTable               *PodTable
	layout                 *PortalLayout
	prevLayout             PortalLayout // previous layout before update
	cluster                string
	namespace              string
}

func (kusApp *KusApp) SetPortal() *KusApp {
	conf := tools.GetConfig()
	cluster, ns := conf.Selected.Cluster, conf.Selected.Namespace
	if cluster != "" && ns != "" {
		go kusApp.Cacher.CacheKuPods(cluster, ns, true)
	}

	portal := &PortalF{
		Flex:        tview.NewFlex().SetDirection(tview.FlexRow),
		kusApp:      kusApp,
		inputCmd:    tview.NewInputField(),
		inputFilter: tview.NewInputField(),
		cluster:     cluster,
		namespace:   ns,
	}
	kusApp.Portal = portal

	kusApp.Portal.
		SetTopInfo().
		setPodtable().
		setInputCmd().
		setInputFilter().
		setDefaultLayout().
		SetInputCapture(portalInputCapture(portal))

	return kusApp
}

func portalInputCapture(portal *PortalF) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		layout := portal.layout
		if event.Rune() == ':' {
			if layout.input != portal.inputCmd || layout.foucus != portal.inputCmd {
				layout.input = portal.inputCmd
				layout.foucus = portal.inputCmd
			} else {
				layout.input = nil
				layout.foucus = layout.body
			}
			portal.refreshLayout()
			return nil
		} else if event.Rune() == '/' {
			if layout.input != portal.inputFilter || layout.foucus != portal.inputFilter {
				layout.input = portal.inputFilter
				layout.foucus = portal.inputFilter
			} else {
				layout.input = nil
				layout.foucus = layout.body
			}
			portal.refreshLayout()
			return nil
		}
		return event
	}
}

// portal input //

// reset InputFilter and podtable.filterkey
// return wheteher changed after resetting
func (portal *PortalF) resetFilter(key string) bool {
	before := portal.podTable.filterKey
	if before == key {
		return false
	}
	portal.inputFilter.SetText(key)
	portal.podTable.filterKey = key
	return true
}

func (portal *PortalF) setInputCmd() *PortalF {
	inputCmd := tview.NewInputField().
		SetLabel("🐮 CMD: ").SetLabelColor(LOGO_COLOR).
		SetFieldWidth(0).
		SetPlaceholder(" ctx / cluster / p / pipeline ...").
		SetPlaceholderTextColor(tcell.ColorGreen).
		SetPlaceholderStyle(tcell.StyleDefault.Background(tcell.ColorBlack)).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorGreen).
		SetAcceptanceFunc(
			func(textToCheck string, lastChar rune) bool {
				// zap.S().Infof("%s , %v \n", textToCheck, lastChar)
				return lastChar != 47 && lastChar != 58 && lastChar != 59 // discard '/' ':' ';'
			})
	inputCmd.SetBorder(true).SetBorderColor(CYAN_COLOR)

	inputCmd.SetDoneFunc(
		func(key tcell.Key) {
			if key == tcell.KeyEnter {
				cmd := strings.ToLower(strings.Trim(inputCmd.GetText(), " "))
				switch cmd {
				case "cluster", "ctx":
					portal.kusApp.Root.SwitchToPage("cluster")
				case "pipeline", "p":
					portal.kusApp.Root.SwitchToPage("pipeline")
				default:
				}
			}
		}).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				portal.podTable.OnlyFocus()
			}
			return event
		})

	portal.inputCmd = inputCmd
	return portal
}

func (portal *PortalF) setInputFilter() *PortalF {
	inputFilter := tview.NewInputField()
	portal.inputFilter = inputFilter

	inputFilter.SetLabel("🐶 Filter: ").SetLabelColor(LOGO_COLOR).
		SetFieldWidth(0).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorGreen).
		SetAcceptanceFunc(
			func(textToCheck string, lastChar rune) bool {
				return lastChar != 47 && lastChar != 58 && lastChar != 59 // discard '/' ':' ';'
			}).
		SetBorder(true).SetBorderColor(CYAN_COLOR)

	inputFilter.SetChangedFunc(func(text string) {
		last := portal.filterChangedMicorTime
		if time.Now().UnixMicro()-last > 1000_000 {

		}
		// portal.podTable.filterKey = text
		// portal.podTable.Refresh(1, true, false, false)
	}).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			keyword := strings.Trim(inputFilter.GetText(), " ")
			changed := portal.resetFilter(keyword)
			zap.S().Debugf("filter input %s %v \n", keyword, changed)
			if changed {
				portal.podTable.Refresh(1, true, true, false)
			} else {
				portal.podTable.OnlyFocus()
			}
		} else if event.Key() == tcell.KeyEsc {
			changed := portal.resetFilter("")
			zap.S().Infoln("reset filter", changed)
			if changed {
				portal.podTable.Refresh(1, true, true, false)
			} else {
				portal.podTable.OnlyFocus()
			}
			return nil
		}
		return event
	})

	return portal
}

type PortalLayout struct {
	top             tview.Primitive
	topFixedSize    int
	topProportion   int
	input           tview.Primitive
	inputFixedSize  int
	inputProportion int
	body            tview.Primitive
	bodyFixedSize   int
	bodyProportion  int
	foucus          tview.Primitive
}

func (portal *PortalF) setDefaultLayout() *PortalF {
	portal.layout = &PortalLayout{
		top:             portal.topInfo,
		topFixedSize:    3,
		topProportion:   0,
		input:           portal.inputCmd,
		inputFixedSize:  3,
		inputProportion: 0,
		body:            portal.podTable,
		bodyFixedSize:   0,
		bodyProportion:  3,
		foucus:          portal.podTable,
	}
	portal.ReTopInfo()
	portal.refreshLayout()
	return portal
}

func (portal *PortalF) refreshLayout() *PortalF {
	// TODO compare if changed
	// zap.S().Debugf("refreshLayout will clear flex firstly \n")
	portal.Flex.Clear()

	layout := portal.layout

	b1 := portal.prevLayout.top == layout.top
	b2 := portal.prevLayout.topFixedSize == layout.topFixedSize
	b3 := portal.prevLayout.topProportion == layout.topProportion
	zap.S().Infof("refreshLayout b1=%v b2=%v b3=%v\n", b1, b2, b3)
	c1 := portal.prevLayout.input == layout.input
	c2 := portal.prevLayout.inputFixedSize == layout.inputFixedSize
	c3 := portal.prevLayout.inputProportion == layout.inputProportion
	e := portal.prevLayout.foucus == layout.foucus
	zap.S().Infof("refreshLayout c1=%v c2=%v c3=%v e=%v\n", c1, c2, c3, e)

	if layout.top != nil {
		focus := layout.foucus == layout.top
		portal.Flex.AddItem(layout.top, layout.topFixedSize, layout.topProportion, focus)
	}
	if layout.input != nil {
		focus := layout.foucus == layout.input
		portal.Flex.AddItem(layout.input, layout.inputFixedSize, layout.inputProportion, focus)
	}
	if layout.body != nil {
		focus := layout.foucus == layout.body
		portal.Flex.AddItem(layout.body, layout.bodyFixedSize, layout.bodyProportion, focus)
	}

	portal.prevLayout = *layout

	// TODO ??
	portal.kusApp.SetFocus(portal.kusApp.Portal)
	return portal
}

func (portal *PortalF) ReTopInfo() *PortalF {
	portal.topInfo.C <- "refresh"
	return portal
}
