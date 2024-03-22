package kuboard

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const txtAccept string = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"

func NewToken(username, password string) (string, error) {
	reqId, err := reqId()
	if err != nil {
		return "", err
	}
	token, err := authLdap(reqId, username, password)
	if err != nil {
		return "", err
	}
	zap.S().Infof("Get kuboard token: %s \n", token)
	return token, nil
}

func reqId() (string, error) {
	resp, err := restyClient().R().
		SetHeader("Accept", txtAccept).
		SetQueryString("access_type=offline&client_id=kuboard-sso&redirect_uri=/callback&response_type=code&scope=openid+profile+email+groups&connector_id=ldap").
		Get("https://" + Host() + "/sso/auth")

	if err != nil && !errors.Is(err, resty.ErrAutoRedirectDisabled) {
		zap.S().Errorf("%v \n", err)
		return "", err
	}

	if len(resp.Header()["Location"]) > 0 {
		location := resp.Header()["Location"][0]
		u, err := url.Parse(location)
		if err != nil {
			return "", err
		}
		if len(u.Query()["req"]) > 0 {
			reqId := u.Query()["req"][0]

			restyClient().R().
				Get("https://" + Host() + location)

			return reqId, nil
		}
	}

	return "", errors.New("get kuboard reqId failed")
}

func authLdap(reqId, username, password string) (string, error) {
	resp, err := restyClient().
		SetRedirectPolicy(LogRedirectPolicy()).
		R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetQueryString("req=" + reqId).
		SetFormData(map[string]string{
			"login":    username,
			"password": fmt.Sprintf(`{"password":"%s","passcode":""}`, password),
		}).
		Post("https://" + Host() + "/sso/auth/ldap")
	if err != nil && !errors.Is(err, resty.ErrAutoRedirectDisabled) {
		zap.S().Debugln(err)
	}

	if len(resp.Header()["Set-Cookie"]) > 0 {
		cookie := resp.Header()["Set-Cookie"][0]
		arr := strings.Split(cookie, ";")
		for _, v := range arr {
			if strings.HasPrefix(v, "KuboardToken") {
				return strings.Split(v, "=")[1], nil
			}
		}
	}
	return "", errors.New("get kuboard ldap token failed")
}

func LogRedirectPolicy() resty.RedirectPolicy {
	return resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		// fmt.Printf("\nRedirecting: ")
		// for _, r := range via {
		// 	fmt.Printf("%s ", r.URL)
		// }
		// fmt.Printf(" -> %s \n", req.URL)
		if req.URL.String() == "https://"+Host()+"/" {
			return resty.ErrAutoRedirectDisabled
		}
		return nil
	})
}
