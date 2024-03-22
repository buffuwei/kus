package kuboard

import (
	"fmt"
	"sort"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

func Clusters() ([]string, error) {
	url := "https://%s/kuboard-api/kind/KubernetesCluster"
	url = fmt.Sprintf(url, Host())
	resp, err := restyClient().R().Get(url)

	if err != nil {
		zap.S().Errorf("get cluster error: %s \n", err)
		return nil, err
	}

	clusterResp := &ClusterResp{}
	jsoniter.Unmarshal(resp.Body(), clusterResp)

	result := []string{}
	for _, item := range clusterResp.Items {
		result = append(result, item.Metadata.Name)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result, nil
}

type ClusterResp struct {
	Items []ClusterItem `json:"items"`
	// ...
}

type ClusterItem struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
}

func Ns(cluster string) ([]string, error) {
	url := "https://%s/k8s-api/%s/api/v1/namespaces?labelSelector=!k8s.kuboard.cn/hide"
	url = fmt.Sprintf(url, Host(), cluster)
	resp, err := restyClient().R().Get(url)

	if err != nil {
		zap.S().Errorf("get cluster error: %s", err)
		return nil, err
	}

	nsrResp := &NsResp{}
	jsoniter.Unmarshal(resp.Body(), nsrResp)

	result := []string{}
	for _, item := range nsrResp.Items {
		result = append(result, item.Metadata.Name)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result, nil
}

type NsResp struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
	} `json:"items"`
}
