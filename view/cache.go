package view

import (
	"buffuwei/kus/kuboard"
	"buffuwei/kus/tools"
	"buffuwei/kus/wings"
	"errors"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

type GlobalCacher struct {
	*cache.Cache
	kusApp            *KusApp
	podsReCached      chan PodsCached
	clustersReCached  chan struct{}
	pipelinesReCached chan struct{}
}

func (kusApp *KusApp) SetCacher() *KusApp {
	cache := &GlobalCacher{
		Cache:             cache.New(60*time.Minute, 60*time.Minute),
		kusApp:            kusApp,
		podsReCached:      make(chan PodsCached, 2),
		clustersReCached:  make(chan struct{}, 2),
		pipelinesReCached: make(chan struct{}, 2),
	}
	kusApp.Cacher = cache

	go func() {
		ticker := time.NewTicker(time.Second * 3)
		for range ticker.C {
			cache.CacheKuPods("", "", false)
		}
	}()

	go func() {
		clusters, _ := cache.CacheClusters()
		for _, cluster := range clusters {
			go cache.CacheNs(cluster)
		}

		ticker := time.NewTicker(time.Minute * 10)
		for range ticker.C {
			cs, _ := cache.CacheClusters()
			for _, cluster := range cs {
				go cache.CacheNs(cluster)
			}
		}
	}()

	go func() {
		zap.S().Debugln("Begin to cache pipelines firstly")
		cache.CachePipelines()
		zap.S().Debugln("Begin new ticker")
		ticker := time.NewTicker(time.Second * 3)
		for range ticker.C {
			cache.CachePipelines()
		}
	}()

	return kusApp
}

const KEY_CLUSTER = "key-all-clusters"

func (cacher *GlobalCacher) CacheClusters() ([]string, error) {
	cs, err := kuboard.Clusters()
	if err != nil {
		return []string{}, err
	}
	zap.S().Infof("All clusters: %-v \n", cs)
	cacher.Set(KEY_CLUSTER, cs, time.Hour)
	cacher.clustersReCached <- struct{}{}
	return cs, nil
}

func (cacher *GlobalCacher) GetClusters() []string {
	if x, found := cacher.Get(KEY_CLUSTER); found {
		return x.([]string)
	}
	go cacher.CacheClusters()
	return []string{}
}

const KEY_NS = "key-%s-all-ns" // +cluster

func (cacher *GlobalCacher) CacheNs(cluster string) []string {
	ns, err := kuboard.Ns(cluster)
	if err != nil {
		zap.S().Infof("Get %s ns error: %s \n", cluster, err)
		return []string{}
	}
	key := fmt.Sprintf(KEY_NS, cluster)
	cacher.Set(key, ns, time.Hour)
	return ns
}

func (cacher *GlobalCacher) GetNs(cluster string) []string {
	key := fmt.Sprintf(KEY_NS, cluster)
	if x, found := cacher.Get(key); found {
		return x.([]string)
	}
	return cacher.CacheNs(cluster)
}

// CacheKuPods refresh pod cache
// changed means cluster/ns whether changed
func (cacher *GlobalCacher) CacheKuPods(cluster, ns string, ctxChanged bool) {
	if cluster == "" || ns == "" {
		if cacher.kusApp.Portal == nil {
			return
		}
		cluster = cacher.kusApp.Portal.cluster
		ns = cacher.kusApp.Portal.namespace
	}
	if ns == "" || cluster == "" {
		// zap.S().Infof("Skip recache pod \n")
		return
	}

	kuPods, err := kuboard.KuPods(cluster, ns)
	if err != nil {
		return
	}

	key := fmt.Sprintf(KEY_PODS, cluster, ns)
	cacher.Set(key, &UpdatedPods{kuPods, tools.CurrDateTime()}, time.Hour)
	cacher.podsReCached <- PodsCached{cluster, ns, ctxChanged}
}

const KEY_PODS = "key-%s-%s-pods" // cluster ns

type PodsCached struct {
	Cluster string
	Ns      string
	Changed bool
}

func (cacher *GlobalCacher) GetKuPods(cluster, ns string) ([]*kuboard.KuPod, string, error) {
	key := fmt.Sprintf(KEY_PODS, cluster, ns)
	if x, found := cacher.Get(key); found {
		updatedPods := x.(*UpdatedPods)
		// zap.S().Infof("Get %d cached pods : %s %s \n", len(updatedPods.KuPods), cluster, ns)
		return updatedPods.KuPods, updatedPods.time, nil
	} else {
		zap.S().Errorf("Failed get cached pods: %s %s \n", cluster, ns)
		return nil, "", errors.New("waiting pod cache")
	}
}

type UpdatedPods struct {
	KuPods []*kuboard.KuPod
	time   string
}

const KEY_PIPELINE = "key-%s-%s-pipeline" // project(not a git project)-branch

func (cacher *GlobalCacher) CachePipelines() {
	// TODO: 缓存什么级别的pipeline？
	// TODO: 如何定义独立的Pipeline Table 的展示？
	// assets := tools.GetConfig().Assets
	// for _, asset := range assets {
	// 	ps := wings.AssetPipelines(&asset.Wingsplatform, 50)
	// 	if ps != nil {
	// 		key := fmt.Sprintf(KEY_PIPELINE, asset., app)
	// 		// zap.S().Debugf("Begin to cache pipelines %s \n", key)
	// 		cacher.Set(key, ps, time.Hour)
	// 		cacher.pipelinesReCached <- struct{}{}
	// 		// zap.S().Debugf("Finished to cache pipelines %s \n", key)
	// 	}
	// }
}

func (cacher *GlobalCacher) GetPipelines() []*wings.Pipeline {

	return []*wings.Pipeline{}
}
