package view

import (
	"buffuwei/kus/tools"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

type ClusterF struct {
	*tview.Flex
	kusApp *KusApp
	// tree   *tview.TreeView
	filter *tview.InputField
	table  *tview.Table
}

func (kusApp *KusApp) SetCluster() *KusApp {
	cf := &ClusterF{
		Flex:   tview.NewFlex(),
		kusApp: kusApp,
		filter: tview.NewInputField(),
		table:  tview.NewTable(),
	}
	kusApp.Cluster = cf

	cf.SetDirection(tview.FlexRow).
		AddItem(cf.filter, 3, 0, false).
		AddItem(cf.table, 0, 2, true)

	setClusterFilter(cf)
	setClusterTable(cf)

	// root := tview.NewTreeNode("_")
	// cf.tree.SetRoot(root).
	// 	SetCurrentNode(root).
	// 	SetBorder(true).
	// 	SetTitle(" switch cluster namespace ").
	// 	SetTitleColor(LOGO_COLOR).
	// 	SetBorderColor(CYAN_COLOR).
	// 	SetBorderPadding(0, 0, 2, 2)
	//
	// cf.SetDirection(tview.FlexRow).
	// 	AddItem(cf.filter, 3, 0, false).
	// 	AddItem(cf.tree, 0, 2, true)
	//
	// cf.tree.SetSelectedFunc(func(node *tview.TreeNode) {
	// 	if node.GetLevel() == 1 {
	// 		if node.IsExpanded() {
	// 			node.CollapseAll()
	// 		} else {
	// 			node.ExpandAll()
	// 		}
	// 	} else if node.GetLevel() == 2 {
	// 		ns := node.GetText()
	// 		clu := node.GetReference().(string)
	//
	// 		if clu == kusApp.Portal.cluster && ns == kusApp.Portal.namespace {
	// 			kusApp.Root.SwitchToPage("portal")
	// 			return
	// 		} else {
	// 			kusApp.Portal.cluster = clu
	// 			kusApp.Portal.namespace = ns
	// 			go kusApp.Cacher.CacheKuPods(clu, ns, true)
	// 			go tools.GetConfig().UpdateSelectedCtx(clu, ns)
	// 			kusApp.Portal.podTable.Refresh(1, true, true, true)
	// 			kusApp.Root.SwitchToPage("portal")
	// 		}
	// 	}
	// })

	// cf.tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
	// 	if event.Key() == tcell.KeyEsc {
	// 		kusApp.Root.SwitchToPage("portal")
	// 	}
	// 	return event
	// })

	// cf.filter.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
	// 	if event.Key() == tcell.KeyEnter {
	// 		zap.S().Infof("cluster filter entered \n")
	// 		refreshClusterNs(cf)
	// 		cf.kusApp.SetFocus(cf.tree)
	// 		return nil
	// 	}
	// 	return event
	// })

	go func() {
		for {
			select {
			case <-kusApp.Cacher.clustersReCached:
				refreshClusterNs(cf)
			}
		}
	}()

	return kusApp
}

func setClusterFilter(cf *ClusterF) {
	cf.filter.SetFieldStyle(tcell.StyleDefault).
		SetText("").
		SetLabel(" 🐶 Filter: ").
		SetLabelColor(LOGO_COLOR).
		SetFieldTextColor(tcell.ColorGreen).
		SetBorder(true).
		SetBorderColor(CYAN_COLOR)
}

func setClusterTable(cf *ClusterF) {
	cf.table.SetSelectable(true, false).
		SetFixed(1, 0).
		SetSelectedStyle(tcell.Style{}).
		SetEvaluateAllRows(true).
		SetTitle(" Assets ").
		// SetTitleColor(LOGO_COLOR).
		SetBorder(true).
		SetBorderPadding(0, 0, 2, 0).
		SetBorderStyle(tcell.StyleDefault.Foreground(CYAN_COLOR))

	setHeader(cf.table)
	setData(cf.table, cf.kusApp.Cacher)

	cf.table.SetSelectedFunc(func(row, column int) {
		clu := cf.table.GetCell(row, 1).Text
		ns := cf.table.GetCell(row, 2).Text
		zap.S().Infof("select %s %s \n", clu, ns)

		// TODO:
		cf.kusApp.Portal.cluster = clu
		cf.kusApp.Portal.namespace = ns
		go cf.kusApp.Cacher.CacheKuPods(clu, ns, true)
		go tools.GetConfig().UpdateSelectedCtx(clu, ns)
		cf.kusApp.Portal.podTable.Refresh(1, true, true, true)
		cf.kusApp.Root.SwitchToPage("portal")
	})

}

func setHeader(table *tview.Table) {
	table.SetCell(0, 0, &tview.TableCell{Text: "No.", Align: tview.AlignCenter, NotSelectable: true})
	table.SetCell(0, 1, &tview.TableCell{Text: "Cluster", Align: tview.AlignCenter, NotSelectable: true})
	table.SetCell(0, 2, &tview.TableCell{Text: "Namespace", Align: tview.AlignCenter, NotSelectable: true})
	table.SetCell(0, 3, &tview.TableCell{Text: "Okay", Align: tview.AlignCenter, NotSelectable: true})
	table.SetCell(0, 4, &tview.TableCell{Text: "Wings", Align: tview.AlignCenter, NotSelectable: true})
	table.SetCell(0, 5, &tview.TableCell{Text: "Remark", Align: tview.AlignCenter, NotSelectable: true})
}

func setData(table *tview.Table, cacher *GlobalCacher) {
	assets := tools.GetConfig().Assets
	for i, asset := range assets {
		idx := i + 1
		wings := ""
		if asset.Wingsplatform != nil {
			wings = asset.Wingsplatform.Env + " " + asset.Wingsplatform.Regin + " " + asset.Wingsplatform.Branch
		}
		table.SetCellSimple(idx, 0, fmt.Sprintf("%d", i))
		table.SetCellSimple(idx, 1, asset.Cluster)
		table.SetCellSimple(idx, 2, asset.Namespace)
		table.SetCellSimple(idx, 3, "-")
		table.SetCellSimple(idx, 4, wings)
		table.SetCellSimple(idx, 5, "")
	}
}

func refreshClusterNs(cf *ClusterF) {
	// cacher := cf.kusApp.Cacher

	// root := cf.tree.GetRoot()
	// root.ClearChildren()

	// for _, asset := range tools.GetConfig().Assets {
	// 	cluster := asset.Cluster
	// 	nss := cacher.GetNs(cluster)
	// 	scopedNs := lo.Filter(nss, func(ns string, _ int) bool {
	// 		return lo.Contains(asset.Namespaces, ns)
	// 	})

	// 	clusterNode := tview.NewTreeNode(cluster).SetColor(CYAN_COLOR)
	// 	root.AddChild(clusterNode)
	// 	for _, ns := range scopedNs {
	// 		nsNode := tview.NewTreeNode(ns).SetColor(CYAN_COLOR).SetReference(cluster)
	// 		clusterNode.AddChild(nsNode)
	// 	}
	// }

	// clusters := cacher.GetClusters()
	// for _, cluster := range clusters {
	// 	filteredNs := []string{}
	// 	for _, ns := range cacher.GetNs(cluster) {
	// 		if strings.Contains(ns, filter) {
	// 			filteredNs = append(filteredNs, ns)
	// 		}
	// 	}
	//
	// 	if len(filteredNs) > 0 {
	// 		clusterNode := tview.NewTreeNode(cluster).SetColor(CYAN_COLOR)
	// 		root.AddChild(clusterNode)
	// 		for _, ns := range filteredNs {
	// 			nsNode := tview.NewTreeNode(ns).SetColor(CYAN_COLOR).SetReference(cluster)
	// 			clusterNode.AddChild(nsNode)
	// 		}
	// 	}
	// }
}
