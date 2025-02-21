package view

import (
	"buffuwei/kus/frontend"
	"buffuwei/kus/kuboard"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var port = 9900

func GetPort() int {
	return port
}

func openTerm(vessel *Pea) {
	url := "http://localhost:%d/sh?cluster=%s&ns=%s&pod=%s&container=%s"
	url = fmt.Sprintf(url, port, vessel.cluster, vessel.ns, vessel.pod, vessel.container)
	browseURL(url)
}

func (kusApp *KusApp) serv() {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	router := gin.Default()
	dispatch(router)

	for i := 9900; i < 100000; i++ {
		port = i
		zap.S().Infof("Run gin port %d\n ", port)
		err := router.Run(fmt.Sprintf(":%d", port))
		if err != nil {
			zap.S().Errorf("Run gin failed: [%s]\n", err)
		}
	}
}

func dispatch(router *gin.Engine) {
	// router.Static("/dist", "./frontend/dist")

	router.Use(static.Serve("/", static.EmbedFolder(frontend.Dist, ".")))

	LoadHTMLFromEmbedFS(router, frontend.Dist, "*html")
	// router.LoadHTMLFiles("./frontend/dist/index.html")

	router.GET("/sh", handlePage)

	router.GET("/ws", handleWs)
}

func handlePage(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			zap.S().Errorf("Recover handlePage panic: %v \n", err)
		}
	}()

	clu := c.Query("cluster")
	ns := c.Query("ns")
	pod := c.Query("pod")
	container := c.Query("container")
	zap.S().Infof("Sh page : [%s][%s][%s][%s] \n", clu, ns, pod, container)

	c.HTML(http.StatusOK, "dist/index.html",
		gin.H{
			"cluster":     clu,
			"ns":          ns,
			"pod":         pod,
			"container":   container,
			"description": clu + " " + ns + " " + pod,
		})
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWs(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			zap.S().Errorf("Recover handleWs panic: %v \n", err)
		}
	}()

	cluster, ns, pod, container := c.Query("cluster"), c.Query("ns"), c.Query("pod"), c.Query("container")
	zap.S().Infof("Term [%s] [%s] [%s] [%s] \n", cluster, ns, pod, container)

	termConn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil || termConn == nil {
		zap.S().Infof("Ws upgrade error %s \n", err)
		return
	} else {
		defer termConn.Close()
	}

	podConn, _ := kuboard.WsExec(cluster, ns, pod, container, "bash")
	if podConn == nil {
		m := fmt.Sprintf("executor not found [%s][%s][%s][%s]", cluster, ns, pod, container)
		termConn.WriteMessage(websocket.TextMessage, []byte(m))
		return
	} else {
		defer podConn.Close()
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				zap.S().Errorf("Recover pod-read panic: %v \n", err)
			}
		}()
		for {
			_, bs, err := podConn.ReadMessage()
			if err != nil {
				zap.S().Errorf("read message error: %v", err)
				return
			}
			ba, _ := base64.StdEncoding.DecodeString(string(bs)[1:])
			zap.S().Infof("Term output: [%s]\n", string(ba))
			termConn.WriteMessage(websocket.TextMessage, ba)
		}
	}()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				zap.S().Errorf("Recover term-read panic: %v \n", err)
			}
		}()
		for {
			_, message, err := termConn.ReadMessage()
			if err != nil {
				continue
			}
			zap.S().Debugf("Term input: [%s]\n", string(message))
			podConn.WriteMessage(websocket.TextMessage, []byte(message))
			// kuboard.WriteBytes(podConn, message)
		}
	}()

	ping(podConn, termConn)
}

func ping(podConn, termConn *websocket.Conn) {
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		podConn.WriteMessage(websocket.TextMessage, []byte("0"))
		termConn.WriteMessage(websocket.TextMessage, []byte("0DQ=="))
	}
}

func browseURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		// Add support for other operating systems as needed
		zap.S().Warnf("unsupported platform: %s\n", runtime.GOOS)
	}

	err := cmd.Start()
	if err != nil {
		zap.S().Errorf("Cmd start failed: %s\n", err.Error())
	}
}

// ========================================================================================//
// Thanks to https://github.com/j1mmyson/LoadHTMLFromEmbedFS-example/blob/main/main.go     //
// ========================================================================================//

func LoadHTMLFromEmbedFS(engine *gin.Engine, embedFS embed.FS, pattern string) {
	root := template.New("")
	tmpl := template.Must(root, LoadAndAddToRoot(engine.FuncMap, root, embedFS, pattern))
	engine.SetHTMLTemplate(tmpl)
}

func LoadAndAddToRoot(funcMap template.FuncMap, rootTemplate *template.Template, embedFS embed.FS, pattern string) error {
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "*", ".*")

	err := fs.WalkDir(embedFS, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		zap.S().Debugf("Walk to [%s] \n", path)
		if matched, _ := regexp.MatchString(pattern, path); !d.IsDir() && matched {
			data, readErr := embedFS.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			t := rootTemplate.New(path).Funcs(funcMap)
			if _, parseErr := t.Parse(string(data)); parseErr != nil {
				return parseErr
			}
		}
		return nil
	})
	return err
}
