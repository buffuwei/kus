package view

import (
	"buffuwei/kus/kuboard"
	"buffuwei/kus/tools"
	"context"
	"encoding/base64"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/gorilla/websocket"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

type ShellF struct {
	*tview.Flex
	kusApp       *KusApp
	currExec     *Executor                       // current showing executor
	cxecutorsMap map[string]map[string]*Executor // holding all cluster pods executor
}

func (flexApp *KusApp) SetShell() *KusApp {
	shell := &ShellF{
		Flex:         tview.NewFlex().SetDirection(tview.FlexRow),
		kusApp:       flexApp,
		currExec:     nil,
		cxecutorsMap: make(map[string]map[string]*Executor, 10),
	}
	flexApp.Shell = shell
	shell.SetInputCapture(capture(shell))

	return flexApp
}

func capture(shell *ShellF) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			shell.Esc()
		} else if event.Key() == tcell.KeyCtrlQ {
			shell.Esc()
			go func() {
				if shell.currExec != nil {
					shell.currExec.close()
				}
			}()
			return nil
		}
		return event
	}
}

var footip *tview.TextView = tview.NewTextView().SetText("").SetTextColor(LOGO_COLOR)
var footfile *tview.TextView = tview.NewTextView().SetText("").SetTextColor(LOGO_COLOR).SetTextAlign(tview.AlignRight)
var footer *tview.Flex = tview.NewFlex().SetDirection(tview.FlexColumn).
	AddItem(footip, 0, 1, false).AddItem(footfile, 0, 1, false)

func (shell *ShellF) Open(v *Vessel) {
	exec := shell.GetExector(v)
	if shell.currExec != exec {
		shell.currExec = exec
		footfile.SetText(exec.outFilePath)

		shell.Flex.Clear().
			AddItem(exec.out, 0, 1, false).
			AddItem(exec.in, 7, 0, true).
			AddItem(footer, 1, 0, false)
	}

	shell.kusApp.Root.SwitchToPage(PageShell)
}

func (shell *ShellF) Esc() {
	footfile.SetText("")
	shell.kusApp.Root.SwitchToPage(PagePortal)
}

// a pod-executor
// Include: websocket than read and write msg ; I/O showing i/o message
type Executor struct {
	shell       *ShellF
	vessel      *Vessel
	conn        *websocket.Conn
	in          *tview.TextArea
	out         *tview.TextView
	inCh        chan string
	cancelFunc  func()
	history     *History
	outFilePath string
}

type History struct {
	items      []string
	max        int
	pos        int   //位置: 缺省值0, 1表示取倒数第一个
	posChanged int64 //time.Now().Unix()
}

func (shell *ShellF) GetExector(v *Vessel) *Executor {
	if e, ok := shell.PickExecutor(v); ok {
		return e
	}
	return shell.NewExecutor(v)
}

func (shell *ShellF) NewExecutor(v *Vessel) *Executor {
	ctx, cancel := context.WithCancel(context.Background())
	conn, _ := kuboard.WsExec(v.cluster, v.ns, v.pod, v.container)
	exec := &Executor{
		shell:       shell,
		vessel:      v,
		conn:        conn,
		inCh:        make(chan string, 1),
		cancelFunc:  cancel,
		history:     &History{[]string{}, 30, 0, 0},
		outFilePath: getShellOutFilePath(v.cluster, v.ns, v.pod),
	}
	exec.setIn().setOut()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ctx.Done():
				zap.S().Infof("shell exec goroutine exit")
				return
			case <-ticker.C:
				exec.conn.WriteMessage(websocket.TextMessage, []byte("0"))
			case cmd := <-exec.inCh:
				input := "0" + base64.StdEncoding.EncodeToString([]byte(cmd))
				exec.conn.WriteMessage(websocket.TextMessage, []byte(input))
			}
		}
	}()

	go func() {
		file := openFile(exec.outFilePath)
		defer file.Close()
		ansiW := tview.ANSIWriter(exec.out)
		for {
			select {
			case <-ctx.Done():
				zap.S().Infof("shell exec goroutine exit")
				return
			default:
				_, bs, err := conn.ReadMessage()
				if err != nil {
					zap.S().Errorf("read message error: %v", err)
					return
				}

				ba, _ := base64.StdEncoding.DecodeString(string(bs)[1:])
				ansiW.Write(ba)
				exec.out.ScrollToEnd()

				cleanedText := colorRegex.ReplaceAllString(string(ba), "")
				file.Write([]byte(cleanedText))
			}
		}
	}()

	shell.addExecutor(exec)
	return exec
}

var colorRegex = regexp.MustCompile(`\x1B\[[\d;]*m`)

func (exec *Executor) setIn() *Executor {
	in := tview.NewTextArea().
		SetWrap(true).SetWordWrap(true).SetLabel(">").
		SetTextStyle(tcell.StyleDefault.Foreground(tcell.GetColor("#2CAD00")))

	in.SetBorder(true).
		SetBorderColor(tcell.ColorYellowGreen).
		SetTitle(" " + exec.vessel.pod + " ").
		SetTitleColor(LOGO_COLOR).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyCtrlC {
				exec.clear()
				exec.sendETX()
				return nil
			}
			if event.Key() == tcell.KeyUp {
				exec.in.SetText(exec.history.next(true), true)
				return nil
			}
			if event.Key() == tcell.KeyDown {
				exec.in.SetText(exec.history.next(false), true)
				return nil
			}
			if event.Key() == tcell.KeyEnter {
				exec.send()
				return nil
			}
			return event
		})
	exec.in = in
	return exec
}

func (exec *Executor) setOut() *Executor {
	exec.out = tview.NewTextView()

	exec.out.SetDynamicColors(true).
		SetChangedFunc(func() {
			exec.shell.kusApp.Draw()
		}).
		SetWrap(true).SetWordWrap(true).
		ScrollToEnd().
		SetTextColor(tcell.GetColor("#2CAD00")).
		// SetTitle(" Out " + exec.outFilePath + " ").
		// SetTitleColor(LOGO_COLOR).
		// SetBorder(true).
		SetBorderColor(tcell.ColorYellowGreen)

	return exec
}

func (ex *Executor) close() {
	ex.cancelFunc()
	delete(ex.shell.cxecutorsMap[ex.vessel.cluster+ex.vessel.ns], ex.vessel.pod)
	ex.shell.kusApp.Portal.ReTopInfo()
}

func (shell *ShellF) addExecutor(e *Executor) {
	if shell.cxecutorsMap[e.vessel.cluster+e.vessel.ns] == nil {
		shell.cxecutorsMap[e.vessel.cluster+e.vessel.ns] = make(map[string]*Executor, 10)
	}
	shell.cxecutorsMap[e.vessel.cluster+e.vessel.ns][e.vessel.pod] = e
	shell.kusApp.Portal.ReTopInfo()
}

func (shell *ShellF) PickExecutor(v *Vessel) (*Executor, bool) {
	if shell.cxecutorsMap[v.cluster+v.ns] == nil {
		return nil, false
	}
	ex := shell.cxecutorsMap[v.cluster+v.ns][v.pod]
	return ex, ex != nil
}

func (shell *ShellF) percentShell(cluster, ns string) string {
	return tools.Percent(shell.cxecutorsMap, cluster+ns)
}

func (exec *Executor) send() {
	cmd := tidyInput(exec.in.GetText())
	exec.inCh <- cmd + CR
	exec.clear().addHistory(cmd)
}

var CR string = string('\u000D')  // Carriage Return
var ETX string = string('\u0003') // \u0003 End of Text  https://symbl.cc/en/0003/

func (exec *Executor) sendETX() {
	exec.inCh <- ETX
}

func (exec *Executor) clear() *Executor {
	exec.in.SetText("", false)
	exec.history.reset()
	return exec
}

func (exec *Executor) addHistory(cmd string) {
	if len(tools.RemoveUnprintable(cmd)) == 0 {
		return
	}
	if cmd == exec.history.last() {
		return
	}
	if len(exec.history.items) >= exec.history.max {
		exec.history.items = exec.history.items[1:]
	}
	exec.history.items = append(exec.history.items, cmd)
}

var BL string = string([]rune{'\u005c', '\u000a'})

func tidyInput(s string) string {
	s = strings.ReplaceAll(s, BL, " ")

	rs := []rune(s)
	for i, v := range rs {
		if !unicode.IsPrint(v) {
			rs[i] = ' '
		}
	}
	return string(rs)
}

func getShellOutFilePath(cluster, namespace, pod string) string {
	dir := tools.HomeDir() + "/" + cluster
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	filePath := dir + "/shell-" + pod + ".log"
	return filePath
}

// true -> up  ;  false -> down
func (his *History) next(up bool) string {
	if time.Now().Unix()-his.posChanged > 5 {
		his.pos = 0
	}
	his.posChanged = time.Now().Unix()
	if up {
		his.pos++
	} else {
		his.pos--
	}
	if his.pos < 1 {
		his.pos = 1
	}
	if his.pos > len(his.items) {
		his.pos = len(his.items)
	}
	if his.pos == 0 {
		return ""
	} else {
		return his.items[len(his.items)-his.pos]
	}
}

func (his *History) last() string {
	if len(his.items) == 0 {
		return ""
	} else {
		return his.items[len(his.items)-1]
	}
}

func (his *History) reset() {
	his.pos = 0
	his.posChanged = 0
}
