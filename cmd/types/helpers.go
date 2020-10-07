package types

import (
	"strconv"

	v1 "k8s.io/api/core/v1"
	prom "github.com/run-ai/runai-cli/pkg/prometheus"
)

func setIntPromData(num *int64, m map[string][]prom.MetricValue, key string) error {
	v, found := m[key]
	if !found {
		return nil
	}

	n, err := strconv.Atoi(v[1].(string))
	if err != nil {
		return err
	} 
	*num = int64(n)	
	return nil
}

func setFloatPromData(num *float64, m map[string][]prom.MetricValue, key string) error {
	v, found := m[key]
	if !found {
		return nil
	}
	n, err := strconv.ParseFloat(v[1].(string), 64)
	if err != nil {
		return err
	} 
	*num = n
	return nil
}

func hasError(errors ...error) error{
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}


func kubeQuantityToFloat64(rl v1.ResourceList, key v1.ResourceName) float64 {
	num, ok := rl[key]
	if ok {
		return float64(num.Value())
	}
	return 0
}

func kubeQuantityToMilliFloat64(rl v1.ResourceList, key v1.ResourceName) float64 {
	num, ok := rl[key]
	if ok {
		return float64(num.MilliValue())
	}
	return 0
}
