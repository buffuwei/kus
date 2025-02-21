package wings

import (
	"buffuwei/kus/tools"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"log"
	"runtime/debug"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var _cookie map[string]string = make(map[string]string)

func verifyCookie(cookie string, wsp *tools.Wingsplatform) {
	resp, err := restyClient.R().
		SetHeader("Cookie", cookie).
		Get(wsp.Host + "/api/v1/user/currentUser")
	if err != nil {
		zap.S().Errorf("failed to get user info: %v", err)
	}
	zap.S().Infof("current/user response: %s \n", string(resp.Body()))

	currUser := currentUser{}
	json.Unmarshal(resp.Body(), &currUser)
	if currUser.Data.Name == "" {
		zap.S().Errorf("cookie is invalid \n")
	} else {
		zap.S().Infof("cookid is valid \n")
	}
}

func init() {
	zap.S().Infof("init wings components\n")
	go func() {
		renewAllCookie()
		ticker := time.NewTicker(time.Minute * 5)
		for {
			<-ticker.C
			renewAllCookie()
		}
	}()
}

type currentUser struct {
	Ret  int `json:"ret"`
	Data struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"data"`
}

func renewAllCookie() {
	defer func() {
		if err := recover(); err != nil {
			stack := string(debug.Stack())
			zap.S().Errorf("renew cookie failed: %s, %s", err, stack)
		}
	}()

	kuboard := tools.GetConfig().Kuboard
	username, pswd := kuboard.Username, kuboard.Password
	assets := tools.GetConfig().Assets
	for _, asset := range assets {
		wsp := asset.Wingsplatform
		if wsp != nil {
			renewCookie(username, pswd, wsp)
		}
	}
}

func renewCookie(username, pswd string, wsp *tools.Wingsplatform) string {
	if wsp.Host == "" || wsp.Login == "" {
		return ""
	}
	host, loginUrl := wsp.Host, wsp.Login

	state := uuid.NewString()
	code := login(host, loginUrl, username, pswd, state)
	if code == "" {
		zap.S().Errorf("login failed")
		return ""
	}
	setCookie := loginCallback(host, code, state)
	_cookie[host] = setCookie
	zap.S().Infof("renew %s cookie : %s\n", host, setCookie)
	return setCookie
}

func loginCallback(host, code, state string) string {
	url := host + "/api/dex/callback"
	resp, err := resty.New().
		SetRedirectPolicy(resty.NoRedirectPolicy()).
		R().
		SetQueryParam("code", code).
		SetQueryParam("state", state).
		Get(url)
	if err != nil && !errors.Is(err, resty.ErrAutoRedirectDisabled) {
		zap.S().Errorf("login callback failed %v \n", err)
	}
	setCookie := resp.Header().Get("Set-Cookie")
	zap.S().Infof("get set-cookie: %s\n", setCookie)
	return setCookie
}

func login(host, loginUrl, username, passwrod, state string) string {
	body := map[string]string{
		"application":  "application_ou2gon",
		"username":     username,
		"password":     "",
		"password_v2":  encryptPassword(passwrod),
		"signinMethod": "Password",
		"type":         "code",
	}
	resp, err := restyClient.R().
		SetBody(body).
		SetQueryParam("clientId", "058b9b1cc3").
		SetQueryParam("responseType", "code").
		SetQueryParam("redirectUri", host+"/api/dex/callback").
		SetQueryParam("type", "code").
		SetQueryParam("scope", "openid rofile email").
		SetQueryParam("state", state).
		SetQueryParam("withKey", "true").
		Post(loginUrl)
	if err != nil {
		zap.S().Errorf("Login failed: %v\n", err)
		return ""
	}
	loginResp := loginResp{}
	err = json.Unmarshal(resp.Body(), &loginResp)
	if err != nil {
		zap.S().Errorf("unmarshal response error: %s, %v", resp, err)
		return ""
	}

	code := loginResp.Data
	zap.S().Infof("get login code: %s\n", code)
	return code
}

type loginResp struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Data   string `json:"data"`
}

func encryptPassword(password string) string {
	bs, _ := hex.DecodeString(s)
	bs, _ = base64.StdEncoding.DecodeString(string(bs))
	block, _ := pem.Decode(bs)
	pub, _ := x509.ParsePKIXPublicKey(block.Bytes)
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		log.Fatalf("not an public key")
	}
	encryptedBytes, _ := rsa.EncryptPKCS1v15(rand.Reader, rsaPub, []byte(password))
	passwordV2 := base64.StdEncoding.EncodeToString(encryptedBytes)
	return passwordV2
}

var s = "4c5330744c5331435255644a5469425156554a4d53554d675330565a4c5330744c53304b54556c4a51306c71515535435a32747861477470527a6c334d454a4255555647515546505130466e4f45464e53556c445132644c5130466e52554533654463786545784753444236636c68575248467357575a456367704d5745746c5533524959546454656d4533634552574e6e524d526a564e6348684c593068686558677a5647387264555a3062475674526b56795255786156476f35535568465233566a5646704e65556f355a793876436c644a57565635526a46785957394d54576f314f58463063484652526c464d4d573958513059784e446856574373305932315962575a6d6569744b5957597251305a536554685554585978566d6b34536e68534f46514b5a6d356165466850656a453363315652546c4a584d6e42514d7a6c356244685553576c42564552425753393257486477654731794e464978534449334e6c637952306f7a5554513354465a7a53575249546e6b7a4e677069566a4a575233643363574a574f456471576e4d766432394f596d68314e7a6c695a7974754d4374695a3363794f5539496130524e4f5646316144566b5333526e55545630546c706f646a457964336c3255455177436a5654554867795557356d65546431566b5232656d526953476858616c644353453136593074564b3070765248466e624668564e7a466d626b6f795345566a5430557755455a56646b46614f4852444e6c497a55336b4b626c52564e58644c6330705257546461567a597a54485a4663305675655445305a3046595544566d59326731636a6b355158684d4d7974485346646e6447387a566e4a734e53737257586b334d465a5a576d5a4855676f34567a684552546b7759544e6e5a693952636e5676566939364e457456643164795245777a4d6c5679546e5669654446586257354a5a33637854304a7364323170566b643663473557556b786b636a524952324a50436a6c775769747263566c5861314e6b625752784d6b74515454453155574a355157524663556458626a464e6257354f64476c4757555576556c5a76516b785764445676613156504d3277304b307854615551305a30514b65566c7464477056613264485254687864586471636c4a764e307842566b4a354f55316f5432356e595777315a4852515656425264584d72623074794d6c6843637a59354c7a5a4662693872625842695a544134635170776545646e6156426153546376635663314d546c30616c7052556e4e786147746153544e346448687159337045566e56594f574e584e4768756132745752446457575456356457685a54307033516b746d52544673436a52344f43396b5343395a5957355061456c746232353055334e4c59316446513046335255464255543039436930744c5330745255354549464256516b784a5179424c52566b744c5330744c513d3d"
