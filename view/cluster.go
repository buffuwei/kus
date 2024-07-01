package view

import (
	"buffuwei/kus/tools"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

type ClusterF struct {
	*tview.Flex
	kusApp *KusApp
	tree   *tview.TreeView
	filter *tview.InputField
}

func (kusApp *KusApp) SetCluster() *KusApp {
	cf := &ClusterF{
		Flex:   tview.NewFlex(),
		kusApp: kusApp,
		tree:   tview.NewTreeView(),
		filter: tview.NewInputField(),
	}
	kusApp.Cluster = cf

	cf.filter.SetFieldStyle(tcell.StyleDefault).
		SetText("comm").
		SetLabel(" 🐶 Filter: ").
		SetLabelColor(LOGO_COLOR).
		SetFieldTextColor(tcell.ColorGreen).
		SetBorder(true).
		SetBorderColor(CYAN_COLOR)

	root := tview.NewTreeNode("_")
	cf.tree.SetRoot(root).
		SetCurrentNode(root).
		SetBorder(true).
		SetTitle(" switch cluster namespace ").
		SetTitleColor(LOGO_COLOR).
		SetBorderColor(CYAN_COLOR).
		SetBorderPadding(0, 0, 2, 2)

	cf.SetDirection(tview.FlexRow).
		AddItem(cf.filter, 3, 0, false).
		AddItem(cf.tree, 0, 2, true)

	cf.tree.SetSelectedFunc(func(node *tview.TreeNode) {
		if node.GetLevel() == 1 {
			if node.IsExpanded() {
				node.CollapseAll()
			} else {
				node.ExpandAll()
			}
		} else if node.GetLevel() == 2 {
			ns := node.GetText()
			clu := node.GetReference().(string)

			if clu == kusApp.Portal.cluster && ns == kusApp.Portal.namespace {
				kusApp.Root.SwitchToPage("portal")
				return
			} else {
				kusApp.Portal.cluster = clu
				kusApp.Portal.namespace = ns
				go kusApp.Cacher.CacheKuPods(clu, ns, true)
				go tools.GetConfig().UpdateSelectedCtx(clu, ns)
				kusApp.Portal.podTable.Refresh(1, true, true, true)
				kusApp.Root.SwitchToPage("portal")
			}
		}
	})

	cf.tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			kusApp.Root.SwitchToPage("portal")
		}
		return event
	})

	cf.filter.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			zap.S().Infof("cluster filter entered \n")
			refreshClusterNs(cf)
			cf.kusApp.SetFocus(cf.tree)
			return nil
		}
		return event
	})

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

func refreshClusterNs(cf *ClusterF) {
	cacher := cf.kusApp.Cacher
	filter := cf.filter.GetText()

	clusters := tools.GetConfig().Kuboard.Clusters
	if len(clusters) == 0 {
		clusters = cacher.GetClusters()
	}

	root := cf.tree.GetRoot()
	root.ClearChildren()
	for _, cluster := range clusters {
		filteredNs := []string{}
		for _, ns := range cacher.GetNs(cluster) {
			if strings.Contains(ns, filter) {
				filteredNs = append(filteredNs, ns)
			}
		}

		if len(filteredNs) > 0 {
			clusterNode := tview.NewTreeNode(cluster).SetColor(CYAN_COLOR)
			root.AddChild(clusterNode)
			for _, ns := range filteredNs {
				nsNode := tview.NewTreeNode(ns).SetColor(CYAN_COLOR).SetReference(cluster)
				clusterNode.AddChild(nsNode)
			}
		}
	}
}
