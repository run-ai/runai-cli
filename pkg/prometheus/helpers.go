package prometheus

import (
	"fmt"
	"strconv"
)


func SetFloatFromFirstMetric(num *float64, m MetricResultsByQueryName, key string) error {
	metrics, found := m[key]
	if !found || len(*metrics) == 0 {
		return nil
	}

	n, err := strconv.ParseFloat((*metrics)[0].Value[1].(string), 64)
	if err != nil {
		return err
	}
	*num = n
	return nil
}


func GroupMetrics(groupBy string, metricsByQueryName MetricResultsByQueryName, queryNames ...string) (map[string]map[string]float64, error) {
	result := map[string]map[string]float64{}
	// map the metrics to items by the given 'groupBy' 
	for _, queryName := range queryNames {
		metrics, ok := metricsByQueryName[queryName]
		if !ok {
			continue
		}
		for _, metric := range *metrics {
			key, ok := metric.Metric[groupBy]
			if !ok {
				return nil, fmt.Errorf("[Prometheus] Failed to find key: (%s) on the metric query name: %s",groupBy, queryName )
			}
			item, created := result[key]
			if !created {
				item = map[string]float64{}
				result[key] = item
			}
			n, err := strconv.ParseFloat(metric.Value[1].(string), 64)
			if err != nil {
				return nil, err
			}
			item[queryName] = n
		}
		
	}

	return result, nil
}