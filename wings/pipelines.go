package wings

import (
	"buffuwei/kus/tools"
	"fmt"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

var restyClient *resty.Client

func init() {
	restyClient = resty.New().SetTimeout(time.Second*3).
		SetRedirectPolicy(resty.NoRedirectPolicy()).
		SetHeader("Connection", "keep-alive").
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "*/*").
		SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36 Edg/128.0.0.0").
		SetContentLength(true)
}

func AppPipelines(app string, wsp *tools.Wingsplatform, pageSize int) []*Pipeline {
	return PipelinePage(wsp.Host, wsp.Project, app, wsp.Branch, pageSize)
}

func AssetPipelines(wsp *tools.Wingsplatform, pageSize int) []*Pipeline {
	return PipelinePage(wsp.Host, wsp.Project, "", wsp.Branch, pageSize)
}

func PipelinePage(host, project, app, branch string, pageSize int) (pipelines []*Pipeline) {
	start := time.Now().Nanosecond()
	url := host + "/api/v1/ci/projects/" + project + "/source-types/GitEvent/pipelines"
	queryStr := fmt.Sprintf("?current=1&pageSize=%d&application=%s&branch=%s&project=%s", pageSize, app, branch, app)
	resp, err1 := restyClient.R().Get(url + queryStr)
	if err1 != nil {
		zap.S().Errorf("Failed get wings pipelies: %s\n", err1.Error())
		return nil
	}

	pipelineResp := &PipelineResp{}
	err1 = jsoniter.Unmarshal(resp.Body(), pipelineResp)
	if err1 != nil {
		zap.S().Errorf("Failed unmarshal wings pipelies: %s\n", err1.Error())
		return nil
	}
	end := time.Now().Nanosecond()
	zap.S().Infof("Get wings pipelines: %d ns\n", end-start)
	return pipelineResp.Data.Pipelines
}

type PipelineResp struct {
	ServerTime int64  `json:"serverTime"`
	Ret        int    `json:"ret"`
	Msg        string `json:"msg"`
	Data       struct {
		Total     int         `json:"total"`
		Pipelines []*Pipeline `json:"pipelines"`
	} `json:"data"`
}

type Pipeline struct {
	ApplicationName string `json:"applicationName"`
	CreateTime      string `json:"createTime"`
	Commits         struct {
		Branch   string `json:"branch"`
		CommitId string `json:"commitId"`
		Message  string `json:"messageeee"`
		UserName string `json:"userName"`
	} `json:"commits"`
	Runners []Runner `json:"runners"`
}

func (p Pipeline) GetTag() string {
	return p.Commits.Branch + "-" + p.Commits.CommitId
}

type Runner struct {
	JobUrl string `json:"jobUrl"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Type   string `json:"type"`
}

func Deploy(app string, wsp *tools.Wingsplatform, targetImagTag string) bool {
	zap.S().Infof("begin pipeline deploy: %s  %+v\n", app, wsp)

	// TODO: parameter registry
	registry := "aliheyuan-registry-vpc.cn-heyuan.cr.aliyuncs.com"
	imageName := fmt.Sprintf("%s/%s/%s", registry, wsp.Project, app)

	deployParam := &DeployParam{
		ProjectName:  wsp.Project,
		EnvName:      wsp.Env,
		CellName:     wsp.DeployCell,
		ResourceName: app,
		MultiImages: []MultiImage{
			{
				Type:            "containers",
				Name:            app,
				ImageName:       imageName,
				HistoryImageTag: "", // not a must
				ImageTag:        targetImagTag,
				Image:           imageName + ":" + targetImagTag,
				ID:              1,
				Key:             1,
				HasSet:          true,
			},
		},
	}

	url := wsp.Host + "/api/v1/projects/%s/regions/%s/releases/%s/set-multi-images"
	url = fmt.Sprintf(url, wsp.Project, wsp.Regin, wsp.DeployCell)
	resp, err := restyClient.EnableTrace().R().
		SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36 Edg/128.0.0.0").
		SetHeader("Content-Type", "application/json").
		SetHeader("Cookie", _cookie[wsp.Host]).
		SetBody(deployParam).
		Post(url)
	zap.S().Debugf("Cookie: %s\n", _cookie[wsp.Host])
	zap.S().Debugf("Param: %+v\n", deployParam)
	if err != nil {
		zap.S().Errorf("failed deploy %s, %s, %s\n", app, url, err.Error())
		return false
	}
	if resp.StatusCode() != 200 {
		zap.S().Errorf("failed deploy %s, %s, %+v\n", app, url, resp)
		return false
	} else {
		zap.S().Infof("succeed deploy %s, %s, %s\n", app, url, string(resp.Body()))
		return true
	}
}

type DeployParam struct {
	ProjectName  string       `json:"projectName"`
	EnvName      string       `json:"envName"`
	CellName     string       `json:"cellName"`
	ResourceName string       `json:"resourceName"`
	MultiImages  []MultiImage `json:"multiImages"`
}

// MultiImage 代表 multiImages 数组中的单个对象
type MultiImage struct {
	Type            string `json:"type"`
	Name            string `json:"name"`
	ImageName       string `json:"imageName"`
	HistoryImageTag string `json:"historyImageTag"`
	ImageTag        string `json:"imageTag"`
	ID              int    `json:"id"`
	Key             int    `json:"key"`
	HasSet          bool   `json:"hasSet"`
	Image           string `json:"image"`
}
