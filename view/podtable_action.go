package view

// import (
// 	"buffuwei/kus/tools"
// 	"buffuwei/kus/wings"
// 	"fmt"
// 	"time"

// 	"github.com/gdamore/tcell/v2"
// 	"github.com/rivo/tview"
// 	"go.uber.org/zap"
// )

// const actionModalName = "podActionModalPage"

// // 对于podTable的每行都需要新建PodActionModal
// type PodActionModal struct {
// 	actionTable *tview.Table
// 	podTable    *PodTable
// 	kusApp      *KusApp
// 	vessel      *Pea
// 	stop        chan struct{}
// }

// func (podTable *PodTable) NewPodActionModal() *PodActionModal {
// 	actionTable := tview.NewTable()
// 	actionTable.SetSelectable(true, false).SetFixed(1, 0).
// 		SetSelectedStyle(tcell.Style{}).SetEvaluateAllRows(true).
// 		SetBorder(true).SetBorderPadding(0, 1, 1, 1).
// 		SetTitleColor(LOGO_COLOR).SetBorderPadding(0, 0, 2, 0).
// 		SetBorderStyle(tcell.StyleDefault.Foreground(CYAN_COLOR)).
// 		SetTitle(" Deploy ")

// 	// setTableData(actionTable, vessel.container)

// 	pam := &PodActionModal{
// 		actionTable: actionTable,
// 		podTable:    podTable,
// 		kusApp:      podTable.portal.kusApp,
// 		stop:        make(chan struct{}),
// 	}

// 	return pam
// }

// // 配置一个空的action模态框List
// func (podTable *PodTable) setPodActionModal2(vessel *Pea) {
// 	actionTable := tview.NewTable()
// 	actionTable.SetSelectable(true, false).SetFixed(1, 0).
// 		SetSelectedStyle(tcell.Style{}).SetEvaluateAllRows(true).
// 		SetBorder(true).
// 		SetBorderPadding(0, 1, 1, 1).
// 		SetTitleColor(LOGO_COLOR).
// 		SetBorderPadding(0, 0, 2, 0).
// 		SetBorderStyle(tcell.StyleDefault.Foreground(CYAN_COLOR)).
// 		SetTitle(vessel.container)

// 	wsp := tools.GetConfig().GetSelectedAsset().Wingsplatform

// 	setTableData(actionTable, vessel.container, wsp)
// 	actionTable.Select(1, 0)

// 	actionTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
// 		zap.S().Infof("actionTable capture input %+v \n", event)
// 		if event.Key() == tcell.KeyEsc {
// 			closePodActionModal(podTable.portal.kusApp)
// 			return nil
// 		} else if event.Key() == tcell.KeyEnter {
// 			r, _ := actionTable.GetSelection()
// 			if r > 0 {
// 				tag := actionTable.GetCell(r, 1).Text
// 				app := vessel.container
// 				// TODO:
// 				wings.Deploy(app, nil, tag)
// 			}
// 		}
// 		return event
// 	})

// 	pam := &PodActionModal{
// 		actionTable: actionTable,
// 		podTable:    podTable,
// 		kusApp:      podTable.portal.kusApp,
// 		vessel:      vessel,
// 		stop:        make(chan struct{}),
// 	}

// 	go func() {
// 		ticker := time.NewTicker(time.Millisecond * 2000)
// 		for {
// 			select {
// 			case <-ticker.C:
// 				// pam.refresh()
// 			case <-pam.stop:
// 				zap.S().Infof("stop refresh action table %s \n", vessel.pod)
// 				return
// 			}
// 		}
// 	}()

// 	podTable.actionModal = pam
// 	podTable.portal.kusApp.Root.AddPage(actionModalName, wrapModal(actionTable), true, true)
// 	zap.S().Infof("podActionModal created\n")
// }

// func wrapModal(actionTable *tview.Table) *tview.Flex {
// 	return tview.NewFlex().
// 		AddItem(nil, 0, 1, false).
// 		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
// 			AddItem(nil, 0, 1, false).
// 			AddItem(actionTable, 10, 1, true).
// 			AddItem(nil, 0, 1, false), 115, 1, true).
// 		AddItem(nil, 0, 1, false)
// }

// func setTableData(actionTable *tview.Table, app string, wsp *tools.Wingsplatform) {
// 	start := time.Now().UnixNano()
// 	setActionTableHeader(actionTable)
// 	ps := wings.AppPipelines(app, wsp, 10)
// 	for i, p := range ps {
// 		idx := i + 1
// 		setActionTableData(actionTable, idx, p)
// 	}
// 	// time.Sleep(184000 * time.Nanosecond)
// 	end := time.Now().UnixNano()
// 	zap.S().Debugf("after set action table data cost %d ns\n", end-start)
// }

// func setActionTableHeader(actionTable *tview.Table) {
// 	actionTable.SetCellSimple(0, 0, "No.")
// 	actionTable.SetCellSimple(0, 1, "Tag")
// 	actionTable.SetCellSimple(0, 2, "Status")
// 	actionTable.SetCellSimple(0, 3, "User")
// 	actionTable.SetCellSimple(0, 5, "Elapsed")
// 	actionTable.SetCellSimple(0, 4, "CreateTime")
// }

// func setActionTableData(actionTable *tview.Table, row int, p *wings.Pipeline) {
// 	actionTable.SetCellSimple(row, 0, fmt.Sprintf("%d", row))
// 	if p == nil {
// 		actionTable.SetCellSimple(row, 1, "")
// 		actionTable.SetCellSimple(row, 2, "")
// 		actionTable.SetCellSimple(row, 3, "")
// 		actionTable.SetCellSimple(row, 4, "")
// 		actionTable.SetCellSimple(row, 5, "")
// 	} else {
// 		tag := p.Commits.Branch + "-" + p.Commits.CommitId
// 		actionTable.SetCellSimple(row, 1, tag)
// 		actionTable.SetCellSimple(row, 2, p.Runners[0].Status)
// 		actionTable.SetCellSimple(row, 3, p.Commits.UserName)
// 		actionTable.SetCellSimple(row, 4, tools.GetTimeElapse(p.CreateTime))
// 		actionTable.SetCellSimple(row, 5, p.CreateTime)
// 	}
// }

// func refreshTable(application string, actionTable *tview.Table, wsp *tools.Wingsplatform) {

// 	ps := wings.PipelinePage(wsp.Host, wsp.Project, application, wsp.Branch, 10)
// 	if len(ps) == 0 {
// 		return
// 	}

// 	// whether has a new pipeline
// 	topRowTag := actionTable.GetCell(1, 1).Text
// 	// TODO: 在这么短的刷新时间内, 会出现多个新的pipeline吗?
// 	if topRowTag != ps[0].GetTag() {
// 		actionTable.InsertRow(0)
// 		setActionTableData(actionTable, 0, ps[0])
// 	}

// 	pipelineMap := make(map[string]*wings.Pipeline)
// 	for _, p := range ps {
// 		pipelineMap[p.GetTag()] = p
// 	}

// 	// actionTable.InsertRow()
// 	rowNum := actionTable.GetRowCount()
// 	for i := 1; i < rowNum; i++ {
// 		tag := actionTable.GetCell(i, 1).Text
// 		if p, ok := pipelineMap[tag]; ok {
// 			setActionTableData(actionTable, i, p)
// 		} else {
// 			actionTable.RemoveRow(i)
// 		}
// 	}

// 	zap.S().Infof("action table refreshed : %s \n", application)
// }

// func (pam *PodActionModal) refresh() {
// 	// refreshTable(pam.actionTable, pam.vessel.container)
// 	// pam.podTable.Refresh(-1, false, false, false)
// 	// pam.kusApp.Application.Draw()
// }

// func (pam *PodActionModal) close() bool {
// 	if pam.kusApp.Root.HasPage(actionModalName) {
// 		pam.stop <- struct{}{}
// 		pam.kusApp.Root.RemovePage(actionModalName)
// 		// focus podtable when actionModal disappear
// 		pam.podTable.portal.layout.foucus = pam.podTable.portal.layout.body
// 		pam.podTable.portal.refreshLayout()
// 		return true
// 	}
// 	return false
// }

// func closePodActionModal(kusApp *KusApp) bool {
// 	modal := kusApp.Portal.podTable.actionModal
// 	if modal != nil {
// 		return modal.close()
// 	}
// 	return false
// }
