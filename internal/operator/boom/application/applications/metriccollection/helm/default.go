package helm

import (
	"github.com/caos/orbos/v5/pkg/kubernetes/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func DefaultValues(imageTags map[string]string, image string) *Values {
	return &Values{
		FullnameOverride: "prometheus",
		DefaultRules: &DefaultRules{
			Create: false,
			Rules: &Rules{
				Alertmanager:                false,
				Etcd:                        false,
				General:                     false,
				K8S:                         false,
				KubeApiserver:               false,
				KubePrometheusNodeAlerting:  false,
				KubePrometheusNodeRecording: false,
				KubernetesAbsent:            false,
				KubernetesApps:              false,
				KubernetesResources:         false,
				KubernetesStorage:           false,
				KubernetesSystem:            false,
				KubeScheduler:               false,
				Network:                     false,
				Node:                        false,
				Prometheus:                  false,
				PrometheusOperator:          false,
				Time:                        false,
			},
		},
		Global: &Global{
			Rbac: &Rbac{
				Create:     true,
				PspEnabled: true,
			},
		},
		Alertmanager: &DisabledToolServicePerReplica{
			Enabled:           false,
			ServicePerReplica: &DisabledTool{Enabled: false},
			IngressPerReplica: &DisabledTool{Enabled: false},
		},
		Grafana: &DisabledTool{
			Enabled: false,
		},
		KubeAPIServer: &DisabledTool{
			Enabled: false,
		},
		Kubelet: &DisabledTool{
			Enabled: false,
		},
		KubeControllerManager: &DisabledTool{
			Enabled: false,
		},
		CoreDNS: &DisabledTool{
			Enabled: false,
		},
		KubeDNS: &DisabledTool{
			Enabled: false,
		},
		KubeEtcd: &DisabledTool{
			Enabled: false,
		},
		KubeScheduler: &DisabledTool{
			Enabled: false,
		},
		KubeProxy: &DisabledTool{
			Enabled: false,
		},
		KubeStateMetricsScrap: &DisabledTool{
			Enabled: false,
		},
		KubeStateMetrics: &DisabledTool{
			Enabled: false,
		},
		NodeExporter: &DisabledTool{
			Enabled: false,
		},
		PrometheusNodeExporter: &DisabledTool{
			Enabled: false,
		},
		PrometheusOperator: &PrometheusOperatorValues{
			Enabled: true,
			TLSProxy: &TLSProxy{
				Enabled: false,
				Image: &Image{
					Repository: "squareup/ghostunnel",
					Tag:        imageTags["squareup/ghostunnel"],
					PullPolicy: "IfNotPresent",
				},
			},
			AdmissionWebhooks: &AdmissionWebhooks{
				FailurePolicy: "Fail",
				Enabled:       false,
				Patch: &Patch{
					Enabled: false,
					Image: &Image{
						Repository: "jettech/kube-webhook-certgen",
						Tag:        imageTags["jettech/kube-webhook-certgen"],
						PullPolicy: "IfNotPresent",
					},
					PriorityClassName: "",
				},
			},
			DenyNamespaces: []string{},
			ServiceAccount: &ServiceAccount{
				Create: true,
			},
			Service: &Service{
				NodePort:    30080,
				NodePortTLS: 30443,
				Type:        "ClusterIP",
			},
			CreateCustomResource:  true,
			CrdAPIGroup:           "monitoring.coreos.com",
			CleanupCustomResource: false,
			KubeletService: &KubeletService{
				Enabled:   false,
				Namespace: "kube-system",
			},
			ServiceMonitor: &ServiceMonitor{
				Interval:    "",
				SelfMonitor: false,
			},
			SecurityContext: &SecurityContext{
				RunAsNonRoot: true,
				RunAsUser:    65534,
			},
			Image: &Image{
				Repository: image,
				Tag:        imageTags[image],
				PullPolicy: "IfNotPresent",
			},
			ConfigmapReloadImage: &Image{
				Repository: "quay.io/coreos/configmap-reload",
				Tag:        imageTags["quay.io/coreos/configmap-reload"],
				PullPolicy: "IfNotPresent",
			},
			PrometheusConfigReloaderImage: &Image{
				Repository: "quay.io/coreos/prometheus-config-reloader",
				Tag:        imageTags["quay.io/coreos/prometheus-config-reloader"],
				PullPolicy: "IfNotPresent",
			},
			ConfigReloaderCPU:    "100m",
			ConfigReloaderMemory: "25Mi",

			HyperkubeImage: &Image{
				Repository: "k8s.gcr.io/hyperkube",
				Tag:        imageTags["k8s.gcr.io/hyperkube"],
				PullPolicy: "IfNotPresent",
			},
			NodeSelector: map[string]string{},
			Resources: &k8s.Resources{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("20m"),
					corev1.ResourceMemory: resource.MustParse("200Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("100Mi"),
				},
			},
		},
		Prometheus: &DisabledToolServicePerReplica{
			Enabled:           false,
			ServicePerReplica: &DisabledTool{Enabled: false},
			IngressPerReplica: &DisabledTool{Enabled: false},
		},
	}
}
