package util

import (
	"fmt"
	"strconv"
	"time"
	"encoding/json"
	"github.com/run-ai/runai-cli/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/api/core/v1"
	log "github.com/sirupsen/logrus"
)

const (
	PROMETHEUS_SCHEME = "http"
)

var (
	runaiPrometheusServer *v1.Service = nil
)

type PrometheusMetric struct {
	Status string               `json:"status,inline"`
	Data   PrometheusMetricData `json:"data,omitempty"`
}

type PrometheusMetricData struct {
	Result     []PrometheusMetricResult `json:"result"`
	ResultType string                   `json:"resultType"`
}

type PrometheusMetricResult struct {
	Metric map[string]string       `json:"metric"`
	Value  []PrometheusMetricValue `json:"value"`
}

type PrometheusMetricValue interface{}

func GetPrometheusService() (service *v1.Service, err error) {
	if runaiPrometheusServer != nil {
		service = runaiPrometheusServer
		return
	}

	c, err := client.GetClient()

	if err != nil {
		return 
	}

	// todo: the namespace can be different from runai
	namespace := "runai"
	promLabel := "prometheus-operator-prometheus"

	list, err := c.GetClientset().CoreV1().Services(namespace).List( metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", promLabel),
	})

	if err != nil {
		return 
	} else if len(list.Items) > 0 {
		service = &list.Items[0]
	}
	runaiPrometheusServer = service
	return
}

func PrometheusQuery(query string) (rst *PrometheusMetric,  err error) {
	service, err := GetPrometheusService()
	if err != nil {
		return
	}

	c, err := client.GetClient()

	if err != nil {
		return 
	}
	
	req := c.GetClientset().CoreV1().Services(service.Namespace).ProxyGet(PROMETHEUS_SCHEME, service.Name , "9090", "api/v1/query", map[string]string{
		"query": query,
		"time":  strconv.FormatInt(time.Now().Unix(), 10),
	})

	log.Debugf("Query prometheus for by %s in ns %s", query, service.Namespace)
	metric, err := req.DoRaw()
	if err != nil {
		log.Debugf("Query prometheus failed due to err %v", err)
		log.Debugf("Query prometheus failed due to result %s", string(metric))
		return
	}
	rst = &PrometheusMetric{}
	err = json.Unmarshal(metric, rst)
	log.Debugf("Prometheus metric:%v", rst)
	if err != nil {
		err = fmt.Errorf("failed to unmarshall heapster response: %v", err)
		return
	}
	if rst.Status != "success" {
		err = fmt.Errorf("failed to query prometheus, status: %s", rst.Status)
		return 
	}
	if len(rst.Data.Result) == 0 {
		log.Debugf("The metric is not exist in prometheus for query  %s", query)
		return nil, nil
	}
	return
 }
