package kuboard

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWsExec(t *testing.T) {
	fmt.Printf("Begin ws exec connecting \n")
	conn, err := WsExec(" ", " ", " ", " ", "")
	if err != nil {
		fmt.Printf("ws exec conn error %s", err.Error())
	}
	defer conn.Close()

	go func() {
		// t := time.NewTicker(time.Second * 8)
		for {
			// cmd := "pwd"
			time.Sleep(time.Second * 3)
			conn.WriteMessage(websocket.TextMessage, []byte(EncodeRawCmd("cd /data0")))
			time.Sleep(time.Second * 1)
			conn.WriteMessage(websocket.TextMessage, []byte(EncodeRawCmd("ls")))
			time.Sleep(time.Second * 1)
			conn.WriteMessage(websocket.TextMessage, []byte(EncodeRawCmd("cd ~")))
			time.Sleep(time.Second * 1)
			conn.WriteMessage(websocket.TextMessage, []byte(EncodeRawCmd("ls")))
			// conn.WriteMessage(websocket.TextMessage, []byte(EncodeRawCmd("yum install -y mysql")))
			time.Sleep(time.Second * 100)
		}
	}()

	for {
		_, bs, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("read message error: %v", err)
			return
		}
		str := string(bs)
		// fmt.Printf("ws exec 1 recv: %s \n", str)
		str = str[1:]
		a, _ := base64.StdEncoding.DecodeString(str)
		output := string(a)
		// fmt.Println("----------")
		fmt.Print(output)
		// fmt.Println("----------")
	}

}

func TestBase64(t *testing.T) {
	b, _ := base64.StdEncoding.DecodeString("DQo=")
	fmt.Printf("[%s]\n", string(b))

	cmd := "pwd"
	input := "1" + base64.StdEncoding.EncodeToString([]byte(cmd)) + "DQo="
	fmt.Printf("input: [%s]\n", input)
}
