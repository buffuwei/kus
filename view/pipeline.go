package view

import (
	"buffuwei/kus/wings"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

type PipelineF struct {
	*tview.Flex
	kusApp *KusApp
	table  *tview.Table
}

func (kusApp *KusApp) SetPipeline() *KusApp {
	pipelineF := &PipelineF{
		Flex:   tview.NewFlex().SetDirection(tview.FlexRow),
		kusApp: kusApp,
		table:  tview.NewTable(),
	}
	pipelineF.configPipelineTable()

	kusApp.Pipeline = pipelineF
	return kusApp
}

func (pipelineF *PipelineF) configPipelineTable() {
	pipelineF.AddItem(pipelineF.table, 0, 1, true)
	pipelineF.table.SetSelectable(true, false).
		SetFixed(1, 0).
		SetSelectedStyle(tcell.Style{}).
		SetEvaluateAllRows(true).
		SetTitle(" Pipelines ").
		SetTitleColor(LOGO_COLOR).
		SetBorder(true).
		SetBorderPadding(0, 0, 2, 0).
		SetBorderStyle(tcell.StyleDefault.Foreground(CYAN_COLOR))

	pipelineF.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pipelineF.kusApp.Root.SwitchToPage("portal")
			return event
		} else if event.Key() == tcell.KeyEnter {
			zap.S().Debugln("pipeline table caputre enter pressed")
			// row, _ := pipelineF.table.GetSelection()
			// pr, ok := pipelineF.table.GetCell(row, 0).GetReference().(*PipelineRef)
			// if !ok {
			// 	zap.S().Errorln("Failed get pipeline ref")
			// 	return event
			// }
			// pipelineF.showModal()

			return nil
		}
		return event
	})

	// pipelineActionMomal未消失的情况下 pipelineTable获得焦点
	pipelineF.table.SetFocusFunc(func() {
		if pipelineF.kusApp.Root.HasPage(modalName) {
			pipelineF.kusApp.Root.RemovePage(modalName)
		}
	})

	go func() {
		for range pipelineF.kusApp.Cacher.pipelinesReCached {
			ps := pipelineF.kusApp.Cacher.GetPipelines()
			setPipelineTableData(pipelineF.table, ps)
		}
	}()
}

const modalName string = "pipelineActionModal"

func (pipelineF *PipelineF) showModal() {
	row, _ := pipelineF.table.GetSelection()
	pipelineRef := pipelineF.table.GetCell(row, 0).GetReference().(*PipelineRef)
	tag := pipelineRef.branch + "-" + pipelineRef.commitId
	cellName := "comm-java-template-test-heyuan"

	modal := tview.NewModal()
	modal.AddButtons([]string{"Deploy", "Rerun"}).SetText(tag).SetTitle(cellName).SetBorder(true)
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		zap.S().Infof("buttonIndex: %d, buttonLabel: %s", buttonIndex, buttonLabel)
		if buttonLabel == "Deploy" {

			// TODO: deploy from a image
			// wings.Deploy(cellName, pipelineRef.app, tag)

		}
		pipelineF.kusApp.Root.RemovePage(modalName)
	})
	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pipelineF.kusApp.Root.RemovePage(modalName)
		}
		return event
	})

	pipelineF.kusApp.Root.AddPage(modalName, modal, true, true)
}

func setPipelineTableData(table *tview.Table, ps []*wings.Pipeline) {
	for i, col := range tableHeaders {
		table.SetCell(0, i,
			&tview.TableCell{
				Text:          col,
				Align:         tview.AlignLeft,
				NotSelectable: true,
			})
	}

	for i, pod := range ps {
		setRow(table, i+1, pod)
	}
}

var tableHeaders = [...]string{"No.", "App", "Branch", "Commit", "User", "Time", "Status", "💡"}

func setRow(table *tview.Table, idx int, p *wings.Pipeline) {
	// TODO: update param
	color := getColor(true, 99999999)

	idCell := newCell(fmt.Sprintf("%d", idx), color)
	idCell.SetReference(&PipelineRef{p.ApplicationName, p.Commits.Branch, p.Commits.CommitId})
	table.SetCell(idx, 0, idCell)
	table.SetCell(idx, 1, newCell(p.ApplicationName, color))
	table.SetCell(idx, 2, newCell(p.Commits.Branch, color))
	table.SetCell(idx, 3, newCell(p.Commits.CommitId, color))
	table.SetCell(idx, 4, newCell(p.Commits.UserName, color))
	table.SetCell(idx, 5, newCell(p.CreateTime, color))
	table.SetCell(idx, 6, newCell(p.Runners[0].Status, color))

}

type PipelineRef struct {
	app, branch, commitId string
}
