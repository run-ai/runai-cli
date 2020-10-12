package helpers

import (
	v1 "k8s.io/api/core/v1"
)



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
