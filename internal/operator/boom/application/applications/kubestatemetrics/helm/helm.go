package helm

import "github.com/caos/orbos/internal/operator/boom/templator/helm/chart"

func GetChartInfo() *chart.Chart {
	return &chart.Chart{
		Name:    "kube-state-metrics",
		Version: "2.9.7",
		Index: &chart.Index{
			Name: "kube-state-metrics",
			URL:  "kubernetes.github.io/kube-state-metrics",
		},
	}
}

func GetImageTags() map[string]string {
	return map[string]string{
		"quay.io/coreos/kube-state-metrics": "v1.9.7",
	}
}
