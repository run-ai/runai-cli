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


type (
	MetricStatusResult string
	MetricType string
	QueryNameToQuery = map[string]string
	// MetricResultsAsItems is a map of itemId => item[key] => MetricValue
	MetricResultsAsItems = map[string]map[string][]MetricValue

	Metric struct {
		Status MetricStatusResult     `json:"status,inline"`
		Data   MetricData 		`json:"data,omitempty"`
	}

	MetricData struct {
		Result     []MetricResult `json:"result"`
		ResultType string         `json:"resultType"`
	}

	MetricResult struct {
		Metric map[string]string    `json:"metric"`
		Value  []MetricValue 		`json:"value"`
	}

	MetricValue interface{}

	Client struct{
		client kubernetes.Interface
		service v1.Service
	}
)

const (

	SuccessStatus MetricStatusResult = "success"
	ErrorStatus MetricStatusResult = "error"

	MatrixResult MetricType = "matrix" 
	VectorResult MetricType = "vector" 
	ScalarResult MetricType = "scalar" 
	StringResult MetricType = "string"

	prometheusSchema = "http"
	// todo: the namespace can be different from runai
	namespace = "runai"
	promLabel = "prometheus-operator-prometheus"

)

func BuildPrometheusClient(c kubernetes.Interface) (*Client, error) {

	ps := &Client {
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

	list, err := ps.client.CoreV1().Services(namespace).List( metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", promLabel),
	})

	if err != nil {
		return 
	} else if len(list.Items) > 0 {
		service = &list.Items[0]
		return
	} 
	
	return nil, fmt.Errorf("No available services of promethues")

}

func (ps *Client)  Query( query string) (data MetricData,  err error) {
	var rst *Metric

	req := ps.client.CoreV1().Services(ps.service.Namespace).ProxyGet(prometheusSchema, ps.service.Name , "9090", "api/v1/query", map[string]string{
		"query": query,
		"time": strconv.FormatInt(time.Now().Unix(), 10),
	})

	log.Debugf("Query prometheus for by %s in ns %s", query, ps.service.Namespace)
	metric, err := req.DoRaw()
	if err != nil {
		log.Debugf("Query prometheus failed due to err %v", err)
		log.Debugf("Query prometheus failed due to result %s", string(metric))
		return
	}
	rst = &Metric{}
	err = json.Unmarshal(metric, rst)
	log.Debugf("Prometheus metric:%v", rst)
	if err != nil {
		err = fmt.Errorf("failed to unmarshall response: %v", err)
		return
	}
	if rst.Status != SuccessStatus {
		err = fmt.Errorf("failed to query prometheus, status: %s", rst.Status)
		return 
	}
	if len(rst.Data.Result) == 0 {
		log.Debugf("The metric is not exist in prometheus for query  %s", query)
	}
	data = rst.Data
	return
 }

type queryResult struct {
	name string
	metric MetricData
	err error
}

 // GroupMultiQueriesToItems map multipule queries to items by given itemId
func (ps *Client) GroupMultiQueriesToItems(q QueryNameToQuery, itemID string) ( MetricResultsAsItems, error) {
	queryResults := map[string]MetricData{}
	results := MetricResultsAsItems{}
	var prometheusResultChanel = make(chan queryResult)
	for queryName, query := range q {
		go (func(query, name string) {
			metric, err := ps.Query(query)
			prometheusResultChanel <- queryResult{name, metric, err}
		})(query, queryName)
	}
	for i := 0; i< len(q); i++ {
		queryResult := <-prometheusResultChanel
		if queryResult.err != nil {
			return nil, queryResult.err
		}
		queryResults[queryResult.name] = queryResult.metric
	}

	// map the result to items by the given 'itemId' 
	for queryName, queryResult := range queryResults {
		// todo: now we are handling only result type = "vector", consider handling more result type in the future 
		for _, metricResult := range queryResult.Result {
			key, ok := metricResult.Metric[itemID]
			if !ok {
				return nil, fmt.Errorf("[Prometheus] Failed to find key: (%s) on the metric query: %s => %s",itemID, queryName, q[queryName] )
			}
			val := metricResult.Value
			item, created := results[key]
			if !created {
				item = map[string][]MetricValue{}
				results[key] = item
			}
			item[queryName] = val
		}
		
	}
	return results, nil
}

