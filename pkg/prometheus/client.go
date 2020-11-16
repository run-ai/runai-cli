package prometheus

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	prometheusSchema                    = "http"
	namespace                           = "runai"
	promLabel                           = "prometheus-operator-prometheus"
	SuccessStatus    MetricStatusResult = "success"
)

type (
	MetricStatusResult string
	MetricType         string
	QueryNameToQuery   = map[string]string
	// MetricResultsByItems is a map of itemId => item[key] => MetricValue
	MetricResultsByItems     = map[string]MetricResultsByQueryName
	MetricResultsByQueryName = map[string]*[]MetricResult

	Metric struct {
		Status MetricStatusResult `json:"status,inline"`
		Data   MetricData         `json:"data,omitempty"`
	}

	MetricData struct {
		Result     []MetricResult `json:"result"`
		ResultType string         `json:"resultType"`
	}

	MetricResult struct {
		Metric map[string]string `json:"metric"`
		Value  []MetricValue     `json:"value"`
	}

	queryResult struct {
		name   string
		metric *MetricData
		err    error
	}

	MetricValue interface{}

	Client struct {
		client  kubernetes.Interface
		service v1.Service
	}
)

func BuildPrometheusClient(c kubernetes.Interface) (*Client, error) {
	ps := &Client{
		client: c,
	}
	service, err := ps.GetPrometheusService()
	if err != nil {
		return nil, err
	}
	ps.service = *service

	return ps, nil
}

func (ps *Client) GetPrometheusService() (service *v1.Service, err error) {

	list, err := ps.client.CoreV1().Services(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", promLabel),
	})

	if err != nil {
		return
	} else if len(list.Items) > 0 {
		service = &list.Items[0]
		return
	}

	return nil, fmt.Errorf("no available services of promethues")
}

func (ps *Client) Query(query string) (*MetricData, error) {
	queryResponse := ps.client.CoreV1().Services(ps.service.Namespace).ProxyGet(prometheusSchema, ps.service.Name, "9090", "api/v1/query", map[string]string{
		"query": query,
		"time":  strconv.FormatInt(time.Now().Unix(), 10),
	})

	log.Debugf("Query prometheus for by %s in ns %s", query, ps.service.Namespace)
	rawMetric, err := queryResponse.DoRaw()
	if err != nil {
		log.Debugf("Query prometheus failed due to err %v", err)
		log.Debugf("Query prometheus failed due to result %s", string(rawMetric))
		return nil, err
	}
	metricResponse := &Metric{}
	err = json.Unmarshal(rawMetric, metricResponse)
	log.Debugf("Prometheus metric:%v", metricResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall response: %v", err)
	}
	if metricResponse.Status != SuccessStatus {
		return nil, fmt.Errorf("failed to query prometheus, status: %s", metricResponse.Status)
	}
	if len(metricResponse.Data.Result) == 0 {
		log.Debugf("The metric is not exist in prometheus for query  %s", query)
	}
	return &metricResponse.Data, nil
}

// GroupMultiQueriesToItems map multiple queries to items by given itemId
func (ps *Client) GroupMultiQueriesToItems(queryMap QueryNameToQuery, labelId string) (MetricResultsByItems, error) {
	metricResults := MetricResultsByItems{}
	queryResultsByNames, err := ps.queryAndGetResponse(queryMap)
	if err != nil {
		return metricResults, err
	}

	for queryName, queryResult := range queryResultsByNames {
		for _, metricResult := range queryResult.Result {
			labelIdValue, ok := metricResult.Metric[labelId]
			if !ok {
				return nil, fmt.Errorf("[Prometheus] Failed to find key: (%s) on the metric query: %s => %s", labelId, queryName, queryMap[queryName])
			}
			resultsMap, created := metricResults[labelIdValue]
			if !created {
				resultsMap = map[string]*[]MetricResult{}
				metricResults[labelIdValue] = resultsMap
			}
			queryResult, ok := resultsMap[queryName]
			if !ok {
				queryResult = &[]MetricResult{}
				resultsMap[queryName] = queryResult
			}

			*queryResult = append(*queryResult, metricResult)
		}

	}
	return metricResults, nil
}

func (ps *Client) queryAndGetResponse(queryMap QueryNameToQuery) (map[string]MetricData, error) {
	queryResults := map[string]MetricData{}
	var prometheusResultChanel = make(chan queryResult)
	for queryName, query := range queryMap {
		go (func(query, name string) {
			metric, err := ps.Query(query)
			prometheusResultChanel <- queryResult{name, metric, err}
		})(query, queryName)
	}
	for i := 0; i < len(queryMap); i++ {
		queryResult := <-prometheusResultChanel
		if queryResult.err != nil {
			return nil, queryResult.err
		}
		queryResults[queryResult.name] = *queryResult.metric
	}
	return queryResults, nil
}
