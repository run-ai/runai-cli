package prometheus

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/run-ai/runai-cli/pkg/authentication/kubeconfig"
	"github.com/run-ai/runai-cli/pkg/client"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	prometheusSchema                                = "http"
	thanosSchema                                    = "https"
	namespace                                       = "monitoring"
	openshiftMonitoringNamespace                    = "openshift-monitoring"
	thanosRouteName                                 = "thanos-querier"
	promLabel                                       = "kube-prometheus-stack-prometheus"
	SuccessStatus                MetricStatusResult = "success"
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
		client             kubernetes.Interface
		dynamicClient      dynamic.Interface
		isOpenshift        bool
		prometheusService  v1.Service
		thanosRouteService *ThanosRouteService
	}

	ThanosRouteService struct {
		url                string
		authorizationToken string
	}
)

func BuildPrometheusClient(c *client.Client) (*Client, error) {
	ps := &Client{
		client:        c.GetClientset(),
		dynamicClient: c.GetDynamicClient(),
	}
	service, err := ps.getPrometheusService()
	if err != nil {
		return nil, err
	}
	if service != nil {
		ps.prometheusService = *service
		return ps, nil
	}

	thanos, err := ps.getThanosRouteService()
	if err != nil {
		return nil, err
	}
	if thanos != nil {
		ps.isOpenshift = true
		ps.thanosRouteService = thanos
		return ps, nil
	}
	return nil, nil
}

func (ps *Client) getPrometheusService() (service *v1.Service, err error) {
	list, err := ps.client.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", promLabel),
	})

	if err != nil {
		return
	} else if len(list.Items) > 0 {
		service = &list.Items[0]
		return
	}

	return nil, nil
}

func (ps *Client) getThanosRouteService() (*ThanosRouteService, error) {
	openshiftRouteSchema := schema.GroupVersionResource{
		Group:    "route.openshift.io",
		Version:  "v1",
		Resource: "routes",
	}

	thanosRoute, err := ps.dynamicClient.Resource(openshiftRouteSchema).Namespace(openshiftMonitoringNamespace).Get(context.TODO(), thanosRouteName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	thanosRouteSpec := thanosRoute.Object["spec"].(map[string]interface{})
	thanosRouteSpecHost := thanosRouteSpec["host"]

	thanosUrl := fmt.Sprintf("%s://%s/", thanosSchema, thanosRouteSpecHost.(string))

	userOcToken, err := kubeconfig.GetOpenshiftToken()
	if err != nil {
		return nil, err
	}

	return &ThanosRouteService{url: thanosUrl, authorizationToken: fmt.Sprintf("bearer %s", userOcToken)}, nil
}

func (ps *Client) queryPrometheus(query string) (*MetricData, error) {
	queryResponse := ps.client.CoreV1().Services(ps.prometheusService.Namespace).ProxyGet(prometheusSchema, ps.prometheusService.Name, "9090", "api/v1/query", map[string]string{
		"query": query,
		"time":  strconv.FormatInt(time.Now().Unix(), 10),
	})

	log.Debugf("Query prometheus for by %s in ns %s", query, ps.prometheusService.Namespace)
	rawMetrics, err := queryResponse.DoRaw(context.TODO())
	if err != nil {
		log.Debugf("Query prometheus failed due to err %v", err)
		log.Debugf("Query prometheus failed due to result %s", string(rawMetrics))
		return nil, err
	}
	return handleQueryResponse(rawMetrics, query)
}

func (ps *Client) queryThanos(query string) (*MetricData, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := http.Client{}

	requestUrl := fmt.Sprintf("%s%s", ps.thanosRouteService.url, "api/v1/query")
	request, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Add("query", query)
	q.Add("time", strconv.FormatInt(time.Now().Unix(), 10))
	request.URL.RawQuery = q.Encode()

	request.Header.Set("Authorization", ps.thanosRouteService.authorizationToken)

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	log.Debugf("Query thanos for by %s in ns %s", query, ps.prometheusService.Namespace)

	rawMetrics, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Debugf("Query thanos failed due to err %v", err)
		log.Debugf("Query thanos failed due to result %s", string(rawMetrics))
		return nil, err
	}
	return handleQueryResponse(rawMetrics, query)
}

func handleQueryResponse(rawMetric []byte, query string) (*MetricData, error) {
	var err error
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
		log.Debugf("The metric is not exist in prometheus for query %s", query)
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
			var err error
			var metric *MetricData
			if ps.isOpenshift {
				metric, err = ps.queryThanos(query)
			} else {
				metric, err = ps.queryPrometheus(query)
			}
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
