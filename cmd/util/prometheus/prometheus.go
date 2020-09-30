package prometheus

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/run-ai/runai-cli/cmd/util"
	"github.com/run-ai/runai-cli/pkg/client"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PROMETHEUS_SCHEME = "http"
)

var (
	cacheServiceGetFunc = func () (interface{}, error) { 
		data, err := getPrometheusService()
		return interface{}(data), err
	}
	PrometheusServiceCache = util.NewCache(cacheServiceGetFunc)
)

func intoInterface(d interface{}) interface{} {
	return d
}

type (
	// MultiQueries is a simple map for queryName => query
	MultiQueries = map[string]string
	// ItemsMap is a map of itemId => item[key] => MetricValue
	ItemsMap = map[string]map[string]MetricValue

	Metric struct {
		Status string     `json:"status,inline"`
		Data   MetricData `json:"data,omitempty"`
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
)


func GetPrometheusService() ( *v1.Service,  error){
	data, err := PrometheusServiceCache.Get()
	switch t := data.(type) {
	case *v1.Service:
		return t, err;
	default:
		return nil, err;
	}
}


func getPrometheusService() (service *v1.Service, err error) {
	

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
	} else {
		return nil, fmt.Errorf("Not found a server for prometheus")
	}

	return
}

func Query(query string) (data MetricData,  err error) {
	var rst *Metric
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
	rst = &Metric{}
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
	}
	return rst.Data, nil
 }


 // MultipuleQueriesToItemsMap map multipule queries to items by given itemId
func MultipuleQueriesToItemsMap(q MultiQueries, itemID string) ( ItemsMap, error) {
	queryResults := map[string]MetricData{}
	rst := ItemsMap{}
	funcs := []func() error{}
	var mux sync.Mutex
	// init the promethus server before the parrall
	// it is not the best way to solve that but it ok for now
	GetPrometheusService()
	for queryName, query := range q {
		funcs = append(funcs, func() error {
			rst, err := Query(query)
			mux.Lock()
			queryResults[queryName] = rst
			mux.Unlock()
			return err
		})
	}
	err := util.Parallel(funcs...)
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
				item = map[string]MetricValue{}
				rst[key] = item
			}
			item[queryName] = val
		}
		
	}
	return rst, nil
}

