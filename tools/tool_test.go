package tools

import (
	"encoding/base64"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"testing"
	"time"
	"unicode"

	"github.com/emirpasic/gods/queues/circularbuffer"
)

func TestFormatDuration(t *testing.T) {
	fmt.Println(FormatDuration(23))
	fmt.Println(FormatDuration(167))
	fmt.Println(FormatDuration(167 + 60*60*3))
	fmt.Println(FormatDuration(167 + 60*60*3 + 60*60*24*3))
}

func TestF(t *testing.T) {
	b, _ := base64.StdEncoding.DecodeString("DQ==")
	fmt.Printf("%d , [%d] |\n", len(string(b)), string(b)[0])

	c, _ := base64.StdEncoding.DecodeString("DQo=")
	fmt.Printf("%d , [%d][%d] |\n", len(string(c)), string(c)[0], string(c)[1])
}

func TestG(t *testing.T) {
	res := base64ToUnicode("DQ==")
	fmt.Printf("|%v| \n", res)
}

func base64ToUnicode(base64str string) []string {
	b, _ := base64.StdEncoding.DecodeString(base64str)
	runes := []rune(string(b))
	var codeHexPoints []string
	for i := 0; i < len(runes); i++ {
		// fmt.Printf("[%d] \n", runes[i])
		codeHexPoints = append(codeHexPoints, fmt.Sprintf("%X", runes[i]))
	}

	return codeHexPoints
}

func TestBashColorReg(t *testing.T) {
	// s := " root root 4.0K Sep  7 16:34 [01;34m..[0m"
	s := `[0m[01;34mapm[0m  [01;34mapollo[0m  [01;34marthas[0m  [01;34miast[0m  [01;34mlogs[0m  [34;42msa[0m  [01;34mwebservice[0m`
	// var bashcolorreg = regexp.MustCompile(`[[0-9;]*m`)
	var bashcolorreg = regexp.MustCompile("\u001B[[0-9;]*m")

	// intrr := bashcolorreg.FindAllIndex([]byte(s), -1)
	// for _, v := range intrr {
	// 	fmt.Printf("Matches from %d to %d \n", v[0], v[1])
	// }

	s = bashcolorreg.ReplaceAllString(s, "")
	fmt.Printf("[%s]", s)
}

func TestHH(t *testing.T) {
	s := "total 189484"
	var bashcolorreg = regexp.MustCompile("\u001B[[0-9;]*m")
	s = bashcolorreg.ReplaceAllString(s, "")
	fmt.Printf("[%s]", s)
}

func TestJJKD(testing *testing.T) {
	t, _ := time.Parse(time.RFC3339, "2024-02-22T12:22:49Z")
	t = t.In(time.FixedZone("CST", 8*60*60))
	fmt.Println(t.Unix())

	fmt.Println(time.Now().Unix() - t.Unix())
	fmt.Println(int64(time.Minute * 10))
	recently := time.Now().Unix()-t.Unix() < 10*60
	fmt.Println(recently)
}

func TestKLK(tt *testing.T) {
	cq := circularbuffer.New(10)
	for i := 0; i < 20; i++ {
		cq.Enqueue(i)
	}
	fmt.Println(cq.Size())
	for i := 0; i < 20; i++ {
		fmt.Println(cq.Dequeue())
	}

	cq.Values()
}

func TestLKLK(tt *testing.T) {
	b64 := "Aw=="
	barr, _ := base64.StdEncoding.DecodeString(b64)
	runes := []rune(string(barr))
	fmt.Printf("reunes-len: [%d] , [%d]  \n", len(runes), runes[0])
	fmt.Printf("Printable %v   \n", unicode.IsPrint(runes[0]))
}

// 回车输入 DQ== 一个Unicode  U+000D(Carriage Return)
//  返回  DQo=  两个Unicode [13] [10] U+000D    U+000A(New Lin)
// Ctrl-C. Aw==

func openURL(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		// Add support for other operating systems as needed
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

func TestOpenUrlUsingDefaultBroser(tt *testing.T) {
	url := "https://www.baidu1.com"
	err := openURL(url)
	if err != nil {
		fmt.Println("Error opening URL:", err)
	} else {
		fmt.Println("Successfully opened", url)
	}
}

func TestBuf(tt *testing.T) {
	reg := regexp.MustCompile("(?i)\\[error\\]")
	matched := reg.MatchString(" 4add5aed69d0dd39] [ERROR] [XNIO-1 task-8] - c")
	tt.Logf("[%v] \n", matched)
	matched = reg.MatchString(" 4add5aed69d0dd39] ERROR [XNIO-1 task-8] - c")
	tt.Logf("[%v] \n", matched)
}
