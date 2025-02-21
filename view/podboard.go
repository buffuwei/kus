package view

import (
	"buffuwei/kus/tools"
	"buffuwei/kus/wings"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

type PodBoard struct {
	*tview.Flex
	kusApp           *KusApp
	info             *tview.TextView
	imageTable       *tview.Table
	imageTableStopCh chan struct{}
	pea              *Pea
	previousPage     string
}

func newPodBoard(podTable *PodTable, previousPage string) *PodBoard {
	pb := &PodBoard{
		Flex:             tview.NewFlex(),
		kusApp:           podTable.portal.kusApp,
		info:             tview.NewTextView(),
		imageTable:       tview.NewTable(),
		imageTableStopCh: make(chan struct{}, 10),
		pea:              podTable.selectedRow,
		previousPage:     previousPage,
	}

	pb.Flex.SetDirection(tview.FlexRow).
		AddItem(pb.imageTable, 0, 2, true).
		AddItem(pb.info, 0, 1, false)

	pb.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			pb.hide()
		}
		return event
	})

	go pb.setImageTable()
	go pb.setInfo()

	return pb
}

func (pb *PodBoard) show() {
	pb.kusApp.Root.AddAndSwitchToPage(PagePodBoard, pb, true)
	pb.kusApp.SetFocus(pb)
}

func (pb *PodBoard) hide() {
	zap.S().Infof("hide podboard \n")
	pb.kusApp.Root.SwitchToPage(pb.previousPage)
	pb.imageTableStopCh <- struct{}{}
	zap.S().Infof("hide podboard end\n")
}

func (pb *PodBoard) setImageTable() *PodBoard {
	pb.imageTable.
		SetSelectable(true, false).
		SetTitle(" Images ").
		SetTitleColor(LOGO_COLOR).
		SetBorder(true).
		SetBorderColor(CYAN_COLOR).
		SetBorderPadding(0, 0, 2, 2)

	pb.imageTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			tag := tagOfSelection(pb.imageTable)
			wsp := tools.GetConfig().GetSelectedAsset().Wingsplatform
			if wsp != nil && tag != "" {
				ok := wings.Deploy(pb.pea.container, wsp, tag)
				if ok {
					pb.kusApp.toastMsg("deploy success")
				} else {
					pb.kusApp.toastMsg("deploy failed")
				}
			}
		}
		return event
	})

	wsp := tools.GetConfig().GetSelectedAsset().Wingsplatform
	if wsp != nil {
		drawImageTableData(pb, wsp)
	}

	ticker := time.NewTicker(3 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				if wsp != nil {
					drawImageTableData(pb, wsp)
				}
			case <-pb.imageTableStopCh:
				zap.S().Infof("stop image table refresh\n")
				ticker.Stop()
			}
		}
	}()

	zap.S().Infof("set image table: %v, %v\n", pb.pea, wsp)
	return pb
}

// Set image table data and draw
func drawImageTableData(pb *PodBoard, wsp *tools.Wingsplatform) {
	ps := wings.AppPipelines(pb.pea.container, wsp, 15)

	setImageTableContent := func() {
		table := pb.imageTable
		table.Clear()
		table.SetTitle(" Images - " + tools.CurrDateTime())

		table.SetCell(0, 0, &tview.TableCell{Text: "No.", Expansion: 1, NotSelectable: true})
		table.SetCell(0, 1, &tview.TableCell{Text: "Branch", Expansion: 1, NotSelectable: true})
		table.SetCell(0, 2, &tview.TableCell{Text: "Tag", Expansion: 4, NotSelectable: true})
		table.SetCell(0, 3, &tview.TableCell{Text: "Status", Expansion: 1, NotSelectable: true})
		table.SetCell(0, 4, &tview.TableCell{Text: "User", Expansion: 1, NotSelectable: true})
		table.SetCell(0, 5, &tview.TableCell{Text: "CreateTime", Expansion: 2, NotSelectable: true})
		table.SetCell(0, 6, &tview.TableCell{Text: "Elapsed", Expansion: 1, NotSelectable: true})

		for i, p := range ps {
			row := i + 1
			tag := p.Commits.Branch + "-" + p.Commits.CommitId
			table.SetCellSimple(row, 0, fmt.Sprintf("%d", row))
			table.SetCellSimple(row, 1, p.Commits.Branch)
			table.SetCellSimple(row, 2, tag)
			table.SetCellSimple(row, 3, p.Runners[0].Status)
			table.SetCellSimple(row, 4, p.Commits.UserName)
			table.SetCellSimple(row, 5, p.CreateTime)
			table.SetCellSimple(row, 6, tools.GetTimeElapse(p.CreateTime))
		}
	}

	pb.kusApp.QueueUpdateDraw(setImageTableContent)
	// setImageTableContent()
}

func tagOfSelection(table *tview.Table) string {
	row, _ := table.GetSelection()
	zap.S().Debugf("row: %v \n", row)
	ref := table.GetCell(row, 2).Text
	zap.S().Infof("ref: %v\n", ref)
	tag := table.GetCell(row, 2).Text
	zap.S().Debugf("tag: %v\n", tag)
	return tag
}

func refreshImageTableData(pb *PodBoard, wsp *tools.Wingsplatform) {
	// pb.imageTable.Clear()
	// pb.imageTable.SetTitle(title(pb.pea))
	drawImageTableData(pb, wsp)
}

func (pb *PodBoard) setInfo() *PodBoard {

	return pb
}
