package kuboard

import (
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

func GetSelfName() (string, error) {
	c := restyClient().SetRedirectPolicy(resty.NoRedirectPolicy())
	url := "https://" + Host() + "/kuboard-api/selfInfo"
	body := "{\"cluster\":\"GLOBAL\"}"
	r, err := c.R().SetBody(body).Post(url)
	if err != nil {
		return "", err
	}
	resp := &SelfInfoResp{}
	jsoniter.Unmarshal(r.Body(), resp)
	return resp.Username, nil
}

func CheckTokenFailed() bool {
	_, err := GetSelfName()
	return err != nil
}

type SelfInfoResp struct {
	Username string `json:"username"`
	Code     string `json:"code"`
	Reason   string `json:"reason"`
}
