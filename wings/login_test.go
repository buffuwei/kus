package wings

import (
	"buffuwei/kus/tools"
	"fmt"
	"testing"
)

func TestMain(m *testing.M) {
	fmt.Printf("Hello, TestMain!\n")
	tools.InitLogger()
	m.Run()
	fmt.Printf("Bye, TestMain!\n")
}

func TestVerifyCookie(tt *testing.T) {
	kuboard := tools.GetConfig().Kuboard
	username, pswd := kuboard.Username, kuboard.Password

	for _, asset := range tools.GetConfig().Assets {
		cookie := renewCookie(username, pswd, asset.Wingsplatform)
		if cookie != "" {
			verifyCookie(cookie, asset.Wingsplatform)

		}
	}
}