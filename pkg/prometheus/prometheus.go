package prometheus

import (
	"encoding/json"
	"fmt"
	//"strconv"
	"sync"
	//"time"

	"github.com/run-ai/runai-cli/pkg/util"
	"k8s.io/client-go/kubernetes"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


type (
	// MetricStatus the possible result status
	MetricStatus string
	// MetricType the possible metice type
	MetricType string
	// MultiQueries is a simple map for queryName => query
	MultiQueries = map[string]string
	// ItemsMap is a map of itemId => item[key] => MetricValue
	ItemsMap = map[string]map[string][]MetricValue

	Metric struct {
		Status MetricStatus     `json:"status,inline"`
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

	SuccessStatus MetricStatus = "success"
	ErrorStatus MetricStatus = "error"

	MatrixResult MetricType = "matrix" 
	VectorResult MetricType = "vector" 
	ScalarResult MetricType = "scalar" 
	StringResult MetricType = "string"

	prometheusSchema = "http"
	// todo: the namespace can be different from runai
	namespace = "runai"
	promLabel = "prometheus-operator-prometheus"

)

func BuildPromethuseClient(c kubernetes.Interface) (*Client, error) {

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
	} else {
		return nil, fmt.Errorf("Not found a server for prometheus")
	}

	return
}

func (ps *Client)  Query( query string) (data MetricData,  err error) {
	var rst *Metric

	req := ps.client.CoreV1().Services(ps.service.Namespace).ProxyGet(prometheusSchema, ps.service.Name , "9090", "api/v1/query", map[string]string{
		"query": query,
		"time":  "1601565109", // strconv.FormatInt(time.Now().Unix(), 10),
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
		err = fmt.Errorf("failed to unmarshall heapster response: %v", err)
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


 // MultipuleQueriesToItemsMap map multipule queries to items by given itemId
func (ps *Client) MultipuleQueriesToItemsMap(q MultiQueries, itemID string) ( ItemsMap, error) {
	queryResults := map[string]MetricData{}
	rst := ItemsMap{}
	funcs := []func() error{}
	var mux sync.Mutex

	var err error;

	for queryName, query := range q {
		getFunc := func() error {
			rst, err := ps.Query(query)
			mux.Lock()
			queryResults[queryName] = rst
			fmt.Print(queryName,"\n\n",rst, "\n\n")
			mux.Unlock()
			return err
		}
		funcs = append(funcs, getFunc)
	}
	err = util.Parallel(funcs...)
	if err != nil {
		return nil, err
	}

	// map the result to items by the given 'itemId' 
	for queryName, qr := range queryResults {
		// todo: check the metric type
		for _, metricResult := range qr.Result {
			// search the itemId in metric labels
			key, ok := metricResult.Metric[itemID]
			if !ok {
				return nil, fmt.Errorf("[Prometheos] Not found an 'itemID' (%s) on the metric query: %s => %s",itemID, queryName, q[queryName] )
			}
			val := metricResult.Value
			item, created := rst[key]
			if !created {
				item = map[string][]MetricValue{}
				rst[key] = item
			}
			item[queryName] = val
		}
		
	}
	return rst, nil
}

