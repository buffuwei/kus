package view

import (
	"buffuwei/kus/kuboard"
	"buffuwei/kus/tools"
	"buffuwei/kus/wings"
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

type PodTable struct {
	*tview.Table
	portal      *PortalF
	refresCh    chan Action
	filterKey   string
	sortKey     SortKey
	terminating []peaID
	selectedRow *Pea
}

type Pea struct {
	cluster, ns, pod, container string
}

func (v Pea) ID() peaID {
	return peaID(v.cluster + ":" + v.ns + ":" + v.pod + ":" + v.container)
}

type peaID string
type SortKey int

const (
	SortKeyNameAsc SortKey = iota
	SortKeyNameDesc
	SortKeyAgeAsc
	SortKeyAgeDesc
)

func (portal *PortalF) setPodtable() *PortalF {
	podTable := &PodTable{
		Table:       tview.NewTable(),
		portal:      portal,
		refresCh:    make(chan Action, 5),
		filterKey:   "",
		sortKey:     SortKeyNameAsc,
		selectedRow: nil,
	}
	portal.podTable = podTable

	podTable.Table.SetSelectable(true, false).
		SetFixed(1, 0).
		SetSelectedStyle(tcell.Style{}).
		SetEvaluateAllRows(true).SetTitle(" Pods ").
		SetTitleColor(LOGO_COLOR).
		SetBorder(true).
		SetBorderPadding(0, 0, 2, 0).
		SetBorderStyle(tcell.StyleDefault.Foreground(CYAN_COLOR))

	// podTable.actionModal = podTable.NewPodActionModal()

	fistCapture := func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := podTable.GetSelection()
		vessel, ok := podTable.GetCell(row, 0).GetReference().(*Pea)
		if ok {
			podTable.selectedRow = vessel
		} else {
			zap.S().Errorf("Failed get pod ref vessel at row %d \n", row)
		}
		return event
	}

	podTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		fistCapture(event)
		return podTableInputCapture(podTable)(event)
	})

	go func() {
		for {
			select {
			case p := <-podTable.refresCh:
				podTable.refresh(p.selectWhichRow, p.scrollToBeginning, false)
			case v := <-portal.kusApp.Cacher.podsReCached:
				podTable.refresh(-1, v.Changed, true)
			}
		}
	}()

	return portal
}

func podTableInputCapture(podTable *PodTable) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		zap.S().Debugf("podTable capture input: %v \n", event)
		portal := podTable.portal
		if portal.layout.body == podTable {
			row, _ := podTable.GetSelection()
			vessel, ok := podTable.GetCell(row, 0).GetReference().(*Pea)
			if !ok {
				zap.S().Errorln("Failed get pod ref vessel")
				return event
			}
			// cluster, ns, pod, container := vessel.cluster, vessel.ns, vessel.pod, vessel.container

			if event.Rune() == 'N' {
				if podTable.sortKey == SortKeyNameAsc {
					podTable.sortKey = SortKeyNameDesc
				} else {
					podTable.sortKey = SortKeyNameAsc
				}
				podTable.Refresh(1, true, true, false)
			} else if event.Rune() == 'A' {
				if podTable.sortKey == SortKeyAgeAsc {
					podTable.sortKey = SortKeyAgeDesc
				} else {
					podTable.sortKey = SortKeyAgeAsc
				}
				podTable.Refresh(1, true, true, false)
			} else if event.Key() == tcell.KeyCtrlK {
				podTable.KillPod(vessel)
			} else if event.Rune() == 'L' {
				portal.kusApp.Logger.OpenLogBackground(vessel)
			} else if event.Rune() == 'l' {
				portal.kusApp.Logger.LoggingInTmux(vessel)
			} else if event.Key() == tcell.KeyCtrlL {
				portal.kusApp.Logger.CloseLogger(vessel)
			} else if event.Key() == tcell.KeyEsc {
				changed := podTable.portal.resetFilter("")
				zap.S().Infoln("reset filter", changed)
				if changed {
					podTable.Refresh(1, true, true, false)
				}
			} else if event.Rune() == 'S' {
				portal.kusApp.Shell.Open(vessel)
			} else if event.Rune() == 's' {
				openTerm(vessel)
			} else if event.Key() == tcell.KeyCtrlC {
				portal.kusApp.Stop()
			} else if event.Key() == tcell.KeyEnter {
				newPodBoard(podTable, PagePortal).show()
				return nil
			}
		}
		return event
	}
}

// 刷新podtable, selectRow<0时将忽略,
func (podTable *PodTable) Refresh(selectRow int, scrollToBeginning, tableFocus, tidyPrev bool) {
	if tidyPrev {
		podTable.Table.Clear()
		podTable.setPodTableTitle("")
		podTable.portal.inputFilter.SetText("")
		podTable.filterKey = ""
	}
	podTable.refresCh <- Action{selectRow, scrollToBeginning}
	if tableFocus {
		podTable.OnlyFocus()
	}
}

func (podTable *PodTable) OnlyFocus() {
	podTable.portal.layout.foucus = podTable
	podTable.portal.kusApp.SetFocus(podTable)
	podTable.portal.refreshLayout()
}

// 加载pod数据到table, selectRow表示选中第几行
// scrollToBeginning 是否滚动到头部
// onFocus 只有当focus时才刷新, 否则跳过
func (podTable *PodTable) refresh(selectRow int, scrollToBeginning, onlyHasFocus bool) {
	kusApp := podTable.portal.kusApp
	cluster := podTable.portal.cluster
	ns := podTable.portal.namespace
	if onlyHasFocus && !podTable.portal.HasFocus() {
		return
	}
	// zap.S().Infof("will renew podtable: %s %s \n", cluster, ns)

	pods, time, err := kusApp.Cacher.GetKuPods(cluster, ns)
	if err != nil {
		return
	}

	filterKey := podTable.filterKey
	sortKey := podTable.sortKey
	pods = fitlerAndSort(pods, filterKey, sortKey)

	kusApp.QueueUpdateDraw(func() {
		podTable.Clear()
		if scrollToBeginning {
			podTable.ScrollToBeginning()
		}
		podTable.setPodTableContent(pods)
	})
	kusApp.QueueUpdateDraw(func() {
		podTable.setPodTableTitle(time)
	})
	if selectRow >= 0 {
		kusApp.QueueUpdateDraw(func() {
			podTable.Select(selectRow, 0)
		})
	}

	podTable.portal.ReTopInfo()
}

func fitlerAndSort(pods []*kuboard.KuPod, filterKey string, sortKey SortKey) []*kuboard.KuPod {
	// zap.S().Infof("FilterSort pods: [%s] [%v] \n", filterKey, sortKey)
	result := []*kuboard.KuPod{}
	if filterKey == "" {
		result = pods
	} else {
		for _, p := range pods {
			if strings.Contains(p.Name, filterKey) {
				result = append(result, p)
			}
		}
	}

	if sortKey == SortKeyNameAsc {
		sort.Slice(result, func(i, j int) bool {
			return result[i].Name < result[j].Name
		})
	} else if sortKey == SortKeyNameDesc {
		sort.Slice(result, func(i, j int) bool {
			return result[i].Name > result[j].Name
		})
	} else if sortKey == SortKeyAgeAsc {
		sort.Slice(result, func(i, j int) bool {
			return result[i].AgeSeconds < result[j].AgeSeconds
		})
	} else if sortKey == SortKeyAgeDesc {
		sort.Slice(result, func(i, j int) bool {
			return result[i].AgeSeconds > result[j].AgeSeconds
		})
	}

	return result
}

func (podTable *PodTable) setPodTableTitle(time string) {
	cluster := podTable.portal.cluster
	ns := podTable.portal.namespace
	filter := podTable.filterKey
	podNum := podTable.GetRowCount() - 1
	tmpl := " %s %s - Pods[%d] </%s> - %s "
	title := fmt.Sprintf(tmpl, cluster, ns, podNum, filter, time)
	podTable.SetTitle(title)
}

// NAME↓ ↑
var podTableHeaders = [...]string{"No.", "Name", "Ready", "Restarts", "Phase", "Addr", "Node", "Tag", "StartTime", "Age", "💡"}

func (podTable *PodTable) setPodTableContent(pods []*kuboard.KuPod) {
	sortKey := podTable.sortKey

	for i, col := range podTableHeaders {
		if col == "Name" && sortKey == SortKeyNameAsc {
			col = "Name↑"
		} else if col == "Name" && sortKey == SortKeyNameDesc {
			col = "Name↓"
		} else if col == "Age" && sortKey == SortKeyAgeAsc {
			col = "Age↑"
		} else if col == "Age" && sortKey == SortKeyAgeDesc {
			col = "Age↓"
		}

		maxWidth := 0
		if col == "Addr" {
			maxWidth = 5
			zap.S().Infof("maxWidth %d \n", maxWidth)
		}

		podTable.SetCell(0, i,
			&tview.TableCell{
				Text:     col,
				Align:    tview.AlignLeft,
				MaxWidth: maxWidth,
				// Color:         tcell.ColorWhite,
				NotSelectable: true,
			})
	}

	for i, pod := range pods {
		podTable.setRow(i+1, pod)
	}

	podTable.portal.kusApp.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		width, _ := screen.Size()
		zap.S().Infof("screen width %d \n", width)
		return false
	})

}

func (podTable *PodTable) setRow(idx int, p *kuboard.KuPod) {
	cluster := podTable.portal.cluster
	ns := podTable.portal.namespace
	table := podTable.Table
	color := getColor(p.IsReady, p.AgeSeconds)

	vessel := &Pea{cluster, ns, p.Name, p.Container}

	table.SetCell(idx, 0, newIdxCell(idx, color, vessel))
	table.SetCell(idx, 1, newCell(vessel.pod, color))
	table.SetCell(idx, 2, newCell(p.ReadyRatio, color))
	table.SetCell(idx, 3, newCell(fmt.Sprintf("%d", p.Restarts), color))
	table.SetCell(idx, 4, newCell(p.Phase, color))
	table.SetCell(idx, 5, newCell2(fmt.Sprintf("%s:%s", p.Ip, p.Port), color, 25))
	table.SetCell(idx, 6, newCell(p.Node, color))
	table.SetCell(idx, 7, newCell2(p.ImageTag, color, 25))
	table.SetCell(idx, 8, newCell(p.StartTime, color))
	table.SetCell(idx, 9, newCell(p.Age, color))

	light := ""
	if podTable.portal.kusApp.Logger.PickLogger(vessel) != nil {
		light += "Log "
	}
	if _, ok := podTable.portal.kusApp.Shell.PickExecutor(vessel); ok {
		light += "SH "
	}
	// podTable.termicating
	if tools.Contains[peaID](podTable.terminating, vessel.ID()) {
		light += "Terminating "
	}

	table.SetCell(idx, 10, newCell(light, color))
}

func getColor(ready bool, elapsed int64) tcell.Color {
	if !ready {
		return tcell.GetColor("#FF4400")
	}
	if elapsed < 60*10 {
		return LOGO_COLOR
	}
	return CYAN_COLOR
}

type Action struct {
	selectWhichRow    int
	scrollToBeginning bool
}

func newCell(text string, textColor tcell.Color) *tview.TableCell {
	c := &tview.TableCell{Text: text, Color: textColor, BackgroundColor: BGColor}
	c.SetExpansion(1)
	return c
}

func newCell2(text string, textColor tcell.Color, maxWidth int) *tview.TableCell {
	c := &tview.TableCell{Text: text, Color: textColor, BackgroundColor: BGColor, MaxWidth: maxWidth}
	return c
}

var BGColor = tcell.GetColor("#000000")

func newIdxCell(idx int, textColor tcell.Color, vessel *Pea) *tview.TableCell {
	c := newCell(fmt.Sprintf("%d", idx), textColor)
	c.SetReference(vessel)
	return c
}

func (podTable *PodTable) KillPod(v *Pea) {
	podTable.terminating = append(podTable.terminating, v.ID())
	go func() {
		kuboard.KillPod(v.cluster, v.ns, v.pod)
		// refresh cache
		podTable.portal.kusApp.Cacher.CacheKuPods(v.cluster, v.ns, false)
	}()
}

type pipelineModal struct {
	*tview.Modal
	podTable *PodTable
	kusApp   *KusApp
}

func newPipelineModal(podTable *PodTable) *pipelineModal {
	pm := &pipelineModal{
		Modal:    tview.NewModal(),
		podTable: podTable,
		kusApp:   podTable.portal.kusApp,
	}

	v := podTable.selectedRow
	ps := wings.PipelinePage("", "", v.container, "", 10)
	tags := make([]string, 0, 10)
	for _, p := range ps {
		tags = append(tags, p.GetTag())
	}

	pm.AddButtons(tags).
		// SetText(" deploy ").
		SetBorder(true).
		SetTitle(" deploying " + v.container + " ")

	return pm
}
