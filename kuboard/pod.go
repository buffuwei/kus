package kuboard

import (
	"buffuwei/kus/tools"
	"fmt"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

type KuPod struct {
	Name       string
	IsReady    bool
	Restarts   int
	ReadyRatio string
	Phase      string
	Ip         string
	Port       string
	Node       string
	ImageTag   string
	StartTime  string
	Age        string
	AgeSeconds int64
	Container  string
}

func KuPods(cluster, ns string) ([]*KuPod, error) {
	resp, err := Pods(cluster, ns)
	if err != nil {
		return nil, err
	}
	kuPods := []*KuPod{}
	for _, podItem := range resp.Items {
		isReady, readyRatio := podItem.GetReady()
		startTime, age, ageSeconds := podItem.GetStartTime()
		pod := &KuPod{
			Name:       podItem.Metadata.Name,
			IsReady:    isReady,
			Restarts:   podItem.Status.GetRestartCount(),
			ReadyRatio: readyRatio,
			Phase:      podItem.Status.Phase,
			Ip:         podItem.Status.PodIP,
			Port:       podItem.GetPort(),
			Node:       podItem.Spec.NodeName,
			ImageTag:   podItem.GetImageTag(),
			StartTime:  startTime,
			Age:        age,
			AgeSeconds: ageSeconds,
			Container:  podItem.GetContainerName(),
		}
		kuPods = append(kuPods, pod)
	}
	return kuPods, nil
}

func Pods(cluster, namespace string) (*PodsResp, error) {
	url := "https://" + Host() + "/k8s-api/" + cluster + "/api/v1/namespaces/" + namespace + "/pods"
	r, err := restyClient().R().Get(url)
	if err != nil {
		return nil, err
	}

	podsResp := &PodsResp{}
	jsoniter.Unmarshal(r.Body(), podsResp)

	// sort.Slice(podsResp.Items, func(i, j int) bool {
	// 	return podsResp.Items[i].Metadata.Name < podsResp.Items[j].Metadata.Name
	// })

	return podsResp, nil
}

type PodsResp struct {
	Items []*PodItem `json:"items"`
}

type PodItem struct {
	Status   *ItemStatus   `json:"status"`
	Metadata *ItemMetadata `json:"metadata"`
	Spec     *ItemSpec     `json:"spec"`
}

type ItemSpec struct {
	NodeName   string `json:"nodeName"`
	Containers []struct {
		Name  string `json:"name"`
		Image string `json:"image"`
		Ports []struct {
			ContainerPort int32  `json:"containerPort"`
			Protocol      string `json:"protocol"`
		} `json:"ports"`
		Resources struct {
			Limits struct {
				CPU    string `json:"cpu"`
				Memory string `json:"memory"`
			} `json:"limits"`
			Requests struct {
				CPU    string `json:"cpu"`
				Memory string `json:"memory"`
			} `json:"requests"`
		} `json:"resources"`
	} `json:"containers"`
}

func (podItem *PodItem) GetPort() string {
	for _, c := range podItem.Spec.Containers {
		if c.Name != "logagent" {
			var ports string
			for _, p := range c.Ports {
				ports += fmt.Sprintf("%d/%s ", p.ContainerPort, p.Protocol)
			}
			return ports
		}
	}
	return ""
}

func (podItem *PodItem) GetContainerName() string {
	for _, c := range podItem.Spec.Containers {
		if c.Name != "logagent" {
			return c.Name
		}
	}
	return ""
}

func (podItem *PodItem) GetReady() (bool, string) {
	isReady := false
	readyCnt := 0
	for _, v := range podItem.Status.ContainerStatuses {
		if v.Ready {
			readyCnt++
		}
	}
	if readyCnt != 0 && readyCnt == len(podItem.Status.ContainerStatuses) {
		isReady = true
	}
	readyRatio := fmt.Sprintf("%d/%d", readyCnt, len(podItem.Status.ContainerStatuses))
	return isReady, readyRatio
}

func (podItem *PodItem) GetImageTag() string {
	for _, c := range podItem.Spec.Containers {
		if c.Name != "logagent" && strings.Contains(c.Image, ":") {
			return strings.Split(c.Image, ":")[1]
		}
	}
	return ""
}

func (podItem *PodItem) GetStartTime() (string, string, int64) {
	t, _ := time.Parse(time.RFC3339, podItem.Status.StartTime)
	t = t.In(time.FixedZone("CST", 8*60*60))
	elapsed := time.Now().Unix() - t.Unix()
	return t.Format(time.DateTime),
		tools.FormatDuration(time.Now().Unix() - t.Unix()),
		elapsed
}

type ItemStatus struct {
	Phase             string `json:"phase"`
	HostIP            string `json:"hostIP"`
	PodIP             string `json:"podIP"`
	StartTime         string `json:"startTime"`
	ContainerStatuses []*struct {
		Name         string `json:"name"`
		Ready        bool   `json:"ready"`
		RestartCount int    `json:"restartCount"`
	} `json:"containerStatuses"`
}

func (itemStatus *ItemStatus) GetRestartCount() int {
	cnt := 0
	for _, v := range itemStatus.ContainerStatuses {
		cnt += v.RestartCount
	}
	return cnt
}

type ItemMetadata struct {
	Name              string `json:"name"`
	CreationTimestamp string `json:"creationTimestamp"`
}

func KillPod(cluster, namespace, pod string) {
	zap.S().Infof("Del pod: %s %s %s \n", cluster, namespace, pod)
	url := "https://%s/k8s-api/%s/api/v1/namespaces/%s/pods/%s?gracePeriodSeconds=60"
	url = fmt.Sprintf(url, Host(), cluster, namespace, pod)
	restyClient().R().Delete(url)
}
