package kuboard

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// wss://kuboard.xxxxx.net
// /k8s-ws/CLUSTER/api/v1/namespaces/NS/pods/content-server-686c8b447c-fnzwh/exec?
// stdin=true&stdout=true&stderr=true&tty=true&command=%2Fbin%2Fbash&container=content-server
func WsExec(cluster, ns, pod, container string) (*websocket.Conn, error) {
	path := fmt.Sprintf("/k8s-ws/%s/api/v1/namespaces/%s/pods/%s/exec", cluster, ns, pod)
	query := fmt.Sprintf("stdin=true&stdout=true&stderr=true&tty=true&command=/bin/sh&container=%s", container)
	u := url.URL{
		Scheme:   "wss",
		Host:     Host(),
		Path:     path,
		RawQuery: query,
	}
	zap.S().Infof("ws exec connecting : %s \n", u.String())

	header := make(map[string][]string)
	header["Cookie"] = []string{Cookie()}
	header["Sec-Websocket-Protocol"] = []string{"base64.channel.k8s.io"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return nil, err
	}
	afterConnected(c)
	return c, nil
}

func afterConnected(conn *websocket.Conn) {
	if conn != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("4"+base64.StdEncoding.EncodeToString([]byte(`{"Width":150,"Height":15}`))))
		time.AfterFunc(time.Millisecond*100, func() {
			WriteBytes(conn, []byte("\r"))
		})
		time.AfterFunc(time.Millisecond*300, func() {
			WriteBytes(conn, []byte("\r"))
		})
		time.AfterFunc(time.Millisecond*500, func() {
			WriteBytes(conn, []byte("\r"))
		})
	}
}

func WriteBytes(conn *websocket.Conn, bytes []byte) {
	input := "0" + base64.StdEncoding.EncodeToString(bytes)
	conn.WriteMessage(websocket.TextMessage, []byte(input))
}

func EncodeRawCmd(cmd string) string {
	input := "0" + base64.StdEncoding.EncodeToString([]byte(cmd+"\n"))
	return input
}
