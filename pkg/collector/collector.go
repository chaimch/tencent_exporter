package collector

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tencentyun/tencentcloud-exporter/pkg/config"
	"github.com/tencentyun/tencentcloud-exporter/pkg/metric"
	"sync"
	"time"
)

const exporterNamespace = "tcm"

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(exporterNamespace, "scrape", "collector_duration_seconds"),
		"qcloud_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(exporterNamespace, "scrape", "collector_success"),
		"qcloud_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
)

const (
	defaultHandlerEnabled = true
)

var (
	collectorState = make(map[string]int)
)

type TcMonitorCollector struct {
	Collectors map[string]*TcProductCollector
	config     *config.TencentConfig
	logger     log.Logger
	lock       sync.Mutex
}

func (n *TcMonitorCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

func (n *TcMonitorCollector) Collect(ch chan<- prometheus.Metric) {
	n.lock.Lock()
	defer n.lock.Unlock()

	wg := sync.WaitGroup{}
	wg.Add(len(n.Collectors))
	for name, c := range n.Collectors {
		go func(name string, c *TcProductCollector) {
			defer wg.Done()
			collect(name, c, ch, n.logger)
		}(name, c)
	}
	wg.Wait()
}

func collect(name string, c *TcProductCollector, ch chan<- prometheus.Metric, logger log.Logger) {
	begin := time.Now()
	level.Info(logger).Log("msg", "Start collect......", "name", name)

	err := c.Collect(ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		level.Error(logger).Log("msg", "Collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		success = 0
	} else {
		level.Info(logger).Log("msg", "Collect done", "name", name, "duration_seconds", duration.Seconds())
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}

func NewTcMonitorCollector(conf *config.TencentConfig, logger log.Logger) (*TcMonitorCollector, error) {
	collectors := make(map[string]*TcProductCollector)

	metricRepo, err := metric.NewTcmMetricRepository(conf, logger)
	if err != nil {
		return nil, err
	}
	// 使用meta缓存
	metricRepoCache := metric.NewTcmMetricCache(metricRepo, logger)

	for _, namespace := range conf.GetNamespaces() {
		state, exists := collectorState[namespace]
		if exists && state == 1 {
			continue
		}

		collector, err := NewTcProductCollector(namespace, metricRepoCache, conf, logger)
		if err != nil {
			panic(fmt.Sprintf("Create product collecter fail, err=%s, Namespace=%s", err, namespace))
		}
		collectors[namespace] = collector
		collectorState[namespace] = 1
		level.Info(logger).Log("msg", "Create product collecter ok", "Namespace", namespace)
	}

	level.Info(logger).Log("msg", "Create all product collecter ok", "num", len(collectors))
	return &TcMonitorCollector{
		Collectors: collectors,
		config:     conf,
		logger:     logger,
	}, nil
}
