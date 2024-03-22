package kuboard

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WsLogger struct {
	Cluster string
	Ns      string
	Pod     string
	Conn    *websocket.Conn
	// Stop    chan int
	LogFile   *os.File
	Cancel    context.CancelFunc
	LogCh     chan string
	LogSwitch chan bool
}

func WsLog(cluster, namespace, pod, container string) (*websocket.Conn, error) {
	path := fmt.Sprintf("/k8s-ws/%s/api/v1/namespaces/%s/pods/%s/log", cluster, namespace, pod)
	query := fmt.Sprintf("stdout=true&stdin=true&tty=true&follow=true&tailLines=100&container=%s", container)
	u := url.URL{
		Scheme:   "wss",
		Host:     Host(),
		Path:     path,
		RawQuery: query,
	}
	zap.S().Infof("ws log connecting : %s \n", u.String())

	header := make(map[string][]string)
	header["Cookie"] = []string{Cookie()}
	header["Sec-Websocket-Protocol"] = []string{"base64.binary.k8s.io"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return nil, err
	}
	return c, nil
}
