package view

import (
	"buffuwei/kus/kuboard"
	"buffuwei/kus/tools"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/gorilla/websocket"
	"github.com/rivo/tview"
	"go.uber.org/zap"
	"golang.design/x/clipboard"
)

// Logger is a reader from container ws connection and writer to log file
type Logger struct {
	vessel      *Vessel
	conn        *websocket.Conn
	cancel      context.CancelFunc
	logFile     *os.File
	logFilePath string
}

// LoggerF is a TextView for specified Logger
type LoggerF struct {
	*tview.Flex
	kusApp      *KusApp
	textView    *tview.TextView
	filter      *tview.InputField
	filterReg   *regexp.Regexp
	filterAfter int
	stream      chan []byte
	running     bool
	logger      *Logger
	loggersMap  map[string]map[string]*Logger // holding all logger; k: cluster-ns ; v: k:pod v: Logger
}

func (kusApp *KusApp) SetLogger() *KusApp {
	lf := &LoggerF{
		Flex:        tview.NewFlex(),
		kusApp:      kusApp,
		textView:    tview.NewTextView(),
		filter:      tview.NewInputField(),
		filterReg:   nil,
		filterAfter: 0,
		stream:      make(chan []byte, 10240*10),
		running:     false,
		logger:      nil,
		loggersMap:  make(map[string]map[string]*Logger, 10),
	}
	kusApp.Logger = lf

	lf.filter.SetFieldStyle(tcell.StyleDefault).
		SetLabel(" 🐶 Filter: ").
		SetLabelColor(LOGO_COLOR).
		SetFieldTextColor(tcell.ColorGreen).
		SetBorder(true).
		SetBorderColor(CYAN_COLOR)

	lf.filter.SetDoneFunc(func(key tcell.Key) {
		s := strings.Trim(lf.filter.GetText(), " ")
		if s == "" {
			lf.filterReg = nil
			lf.filterAfter = 0
		} else {
			title := fmt.Sprintf(" Logger %s ", lf.logger.logFilePath)

			arr := strings.Split(s, "+")
			s = arr[0]
			if reg, err := regexp.Compile("(?i)" + s); err == nil {
				lf.filterReg = reg
				title = title + "  " + reg.String()
				if len(arr) > 1 {
					lf.filterAfter, _ = strconv.Atoi(arr[1])
					title = title + " +" + arr[1]
				}
			} else {
				title += "  filter regexp error "
			}

			lf.textView.SetTitle(title)
		}
		lf.textView.ScrollToEnd()
		zap.S().Infof("Logger filter: [%s] +%dlines \n", lf.filterReg, lf.filterAfter)
	})

	lf.SetDirection(tview.FlexRow).
		AddItem(lf.filter, 3, 0, false).
		AddItem(lf.textView, 0, 2, true)

	lf.textView.SetScrollable(true).
		SetMaxLines(5000).
		SetTextColor(tcell.GetColor("#A1ABAB")).
		SetBorder(true).
		SetTitle(" Logger ").
		SetTitleColor(LOGO_COLOR).
		SetBorderPadding(0, 0, 1, 1).
		SetBorderColor(tcell.ColorYellowGreen).
		SetBorderAttributes(tcell.AttrDim).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				kusApp.Logger.Hide()
				return nil
			}
			return event
		})

	go lf.logging()

	return kusApp
}

const bufCap = 1024 * 5
const bufMax = 1024 * 3

func (lf *LoggerF) logging() {
	var after int
	ticker := time.NewTicker(time.Second * 5)
	var buf []byte
	for {
		select {
		case <-ticker.C:
			if len(buf) > 0 {
				lf.kusApp.QueueUpdateDraw(func() {
					lf.textView.Write(buf)
				})
				buf = make([]byte, 0, bufCap)
			}
		case b := <-lf.stream:
			if lf.filterReg != nil {
				if lf.filterReg.MatchString(string(b)) {
					after = lf.filterAfter
				} else {
					if after > 0 {
						after -= 1
					} else {
						b = nil
					}
				}
			}

			if b != nil {
				buf = append(buf, b...)
				if len(buf) > bufMax {
					lf.kusApp.QueueUpdateDraw(func() {
						lf.textView.Write(buf)
					})
					buf = make([]byte, 0, bufCap)
				}
			}

		}
	}
}

func (lf *LoggerF) Show() {
	lf.kusApp.Root.SwitchToPage("logger")
}

func (lf *LoggerF) Hide() {
	zap.S().Infof("Stop log view \n")
	lf.logger = nil
	lf.running = false
	lf.kusApp.Root.SwitchToPage("portal")

	zap.S().Infof("exit hide log view \n")
}

func (lv *LoggerF) OpenLogView(vessel *Vessel) {
	for len(lv.stream) > 0 {
		<-lv.stream
	}

	lv.Show()

	logger := lv.GetLogger(vessel)
	if logger == nil {
		lv.stream <- []byte("Logger not found 😭")
		return
	}

	lv.filter.SetText("")
	lv.filterReg = nil
	lv.filterAfter = 0
	lv.logger = logger
	lv.textView.SetTitle(fmt.Sprintf(" Logger %s ", logger.logFilePath))
	lv.textView.ScrollToEnd()
	zap.S().Infof("Finished view log %v \n", vessel)
}

func (lf *LoggerF) OpenLogBackground(vessel *Vessel) {
	logger := lf.GetLogger(vessel)
	if logger == nil {
		clipboard.Write(clipboard.FmtText, []byte("Logger not found"))
		lf.kusApp.ShowErr("Logger not found 😭")
		return
	}
	clipboard.Write(clipboard.FmtText, []byte(logger.logFilePath))
}

// new a logger or get existed logger
func (lv *LoggerF) GetLogger(vessel *Vessel) *Logger {
	if l := lv.PickLogger(vessel); l != nil {
		zap.S().Infof("Get existed logger %v \n", vessel)
		return l
	}

	logger, _ := lv.newLogger(vessel)

	zap.S().Infof("Get new logger : %v \n", vessel)
	return logger
}

func getTmux(kusApp *KusApp) string {
	bin, _ := exec.LookPath("tmux")

	if bin == "" {
		kusApp.ShowErr(`Can not find tmux, try "brew install tmux"`)
	}

	return bin
}

func (lf *LoggerF) LoggingInTmux(vessel *Vessel) {
	l := lf.GetLogger(vessel)

	bin := getTmux(lf.kusApp)
	if bin == "" {
		return
	}

	sessionName := fmt.Sprintf("kus-%d-LOG", GetPort())

	lf.kusApp.Suspend(func() {
		zap.S().Infof("tmux logging %s \n", l.logFilePath)
		defer func() {
			exec.Command(bin, "kill-session", "-t", sessionName).Run()
		}()

		cmd := exec.Command(bin, "new-session", "-t", sessionName)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

		time.AfterFunc(time.Millisecond*500, func() {
			exec.Command(bin, "set", "-t", sessionName, "status-right-length", "80").Run()
			s := fmt.Sprintf("#[fg=black]%s | %s #[green]", sessionName, l.vessel.pod)
			exec.Command(bin, "set", "-t", sessionName, "status-right", s).Run()
			exec.Command(bin, "send-keys", "-t", sessionName, "tail -f "+l.logFilePath, "Enter").Run()
		})

		err := cmd.Run()
		zap.S().Infof("tmux logging done: %v \n", err)
	})
}

func (lv *LoggerF) newLogger(v *Vessel) (*Logger, error) {
	conn, err := kuboard.WsLog(v.cluster, v.ns, v.pod, v.container)
	if err != nil {
		return nil, err
	}

	logFilePath := getLogFilePath(v)
	ctx, cancel := context.WithCancel(context.Background())
	logger := &Logger{v, conn, cancel, nil, logFilePath}
	lv.addLogger(logger)

	go logger.read(ctx, lv)

	go logger.ping(ctx)

	return logger, nil
}

func (lv *LoggerF) PickLogger(vessel *Vessel) *Logger {
	if lv.loggersMap[vessel.cluster+vessel.ns] == nil {
		return nil
	}
	return lv.loggersMap[vessel.cluster+vessel.ns][vessel.pod]
}

func (lv *LoggerF) addLogger(l *Logger) {
	diff := l.vessel.cluster + l.vessel.ns
	if lv.loggersMap[diff] == nil {
		lv.loggersMap[diff] = make(map[string]*Logger, 100)
	}
	lv.loggersMap[diff][l.vessel.pod] = l
	lv.kusApp.Portal.ReTopInfo()
}

func (lv *LoggerF) CloseLogger(vessel *Vessel) {
	logger := lv.PickLogger(vessel)
	if logger != nil {
		zap.S().Infof("Close logger : %v \n", vessel)
		logger.cancel()
		delete(lv.loggersMap[vessel.cluster+vessel.ns], vessel.pod)
	}
	lv.kusApp.Portal.ReTopInfo()
}

func (lv *LoggerF) percentLogger(cluster, ns string) string {
	return tools.Percent(lv.loggersMap, cluster+ns)
}

func (logger *Logger) read(ctx context.Context, lv *LoggerF) {
	file := openFile(logger.logFilePath)
	defer file.Close()

	for {
		select {
		case <-ctx.Done():
			zap.S().Infof("Close logger %s \n", file.Name())
			logger.conn.Close()
			return
		default:
			_, bs, err := logger.conn.ReadMessage()
			if err != nil {
				zap.S().Errorf("Stop ws read : %s \n", err.Error())
				return
			}
			decodedBs, err := base64.StdEncoding.DecodeString(string(bs))
			if err != nil {
				zap.S().Errorln(err)
				continue
			}
			if lv.logger == logger {
				lv.stream <- decodedBs
			}
			file.Write(decodedBs)
		}
	}
}

func (logger *Logger) ping(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 6)

	for {
		select {
		case <-ctx.Done():
			zap.S().Infof("Close ping ws \n")
			ticker.Stop()
			return
		case <-ticker.C:
			err := logger.conn.WriteMessage(websocket.TextMessage, []byte("0"))
			if err != nil {
				zap.S().Errorln("Ping ws error ", err)
				return
			}
		}
	}
}

func openFile(filePath string) *os.File {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		zap.S().Errorln("Open log file error", err)
	}
	return f
}

func getLogFilePath(v *Vessel) string {
	dir := tools.HomeDir() + "/" + v.cluster
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	filePath := dir + "/" + v.pod + ".log"
	return filePath
}
