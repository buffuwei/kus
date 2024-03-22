package kuboard

import (
	"buffuwei/kus/tools"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

const UserAgent string = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

var (
	client *resty.Client
	once   sync.Once
)

func Cookie() string {
	if tools.GetConfig().Kuboard.Token == "" {
		return ""
	}
	token := tools.GetConfig().Kuboard.Token
	return "KuboardToken=" + token
}

func Host() string {
	return tools.GetConfig().Kuboard.Host
}

func restyClient() *resty.Client {
	once.Do(func() {
		client = resty.New().SetTimeout(time.Second*3).
			SetRedirectPolicy(resty.NoRedirectPolicy()).
			SetHeader("Connection", "keep-alive").
			SetHeader("Content-Type", "application/json").
			SetHeader("Cookie", Cookie()).
			SetHeader("Accept", "*/*").
			SetHeader("User-Agent", UserAgent).
			SetHeader("Host", Host()).
			SetContentLength(true)
	})
	return client
}
