package helm

import (
	"github.com/caos/orbos/pkg/kubernetes/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func DefaultValues(imageTags map[string]string, image string) *Values {
	adminAnnotations := map[string]string{"app.kubernetes.io/use": "admin-service"}

	return &Values{
		Env: map[string]string{
			"POLL_EVERY_SECS": "0",
		},
		NodeSelector:     map[string]string{},
		FullnameOverride: "ambassador",
		AdminService: &AdminService{
			Annotations: adminAnnotations,
			Create:      true,
			Port:        8877,
			Type:        "ClusterIP",
		},
		AuthService: &AuthService{
			Create: true,
		},
		Autoscaling: &Autoscaling{
			Enabled: false,
		},
		Crds: &Crds{
			Create:  true,
			Enabled: true,
			Keep:    true,
		},
		CreateDevPortalMapping: false,
		DaemonSet:              false,
		DeploymentStrategy: &DeploymentStrategy{
			Type: "RollingUpdate",
		},
		DNSPolicy:   "ClusterFirst",
		HostNetwork: false,
		Image: &Image{
			PullPolicy: "IfNotPresent",
			Repository: image,
			Tag:        imageTags[image],
		},
		LicenseKey: &LicenseKey{
			CreateSecret: true,
		},
		LivenessProbe: &LivenessProbe{
			FailureThreshold:    3,
			InitialDelaySeconds: 30,
			PeriodSeconds:       3,
		},
		PrometheusExporter: &PrometheusExporter{
			Enabled:    false,
			PullPolicy: "IfNotPresent",
			Repository: "prom/statsd-exporter",
			Tag:        imageTags["prom/statsd-exporter"],
			Resources: &k8s.Resources{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("50m"),
					corev1.ResourceMemory: resource.MustParse("100Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10m"),
					corev1.ResourceMemory: resource.MustParse("20Mi"),
				},
			},
		},
		RateLimit: &RateLimit{
			Create: true,
		},
		Rbac: &Rbac{
			Create: true,
		},
		ReadinessProbe: &ReadinessProbe{
			FailureThreshold:    3,
			InitialDelaySeconds: 30,
			PeriodSeconds:       3,
		},

		Redis: &Redis{
			Create: true,
			Resources: &k8s.Resources{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("360Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("50m"),
					corev1.ResourceMemory: resource.MustParse("180Mi"),
				},
			},
			Annotations: &RedisAnnotations{
				Deployment: map[string]string{},
				Service:    map[string]string{},
			},
			NodeSelector: map[string]string{},
		},
		ReplicaCount: 3,
		Scope: &Scope{
			SingleNamespace: false,
		},
		Security: &Security{
			PodSecurityContext: &PodSecurityContext{
				RunAsUser: 8888,
			},
			ContainerSecurityContext: &ContainerSecurityContext{
				AllowPrivilegeEscalation: false,
			},
		},
		Service: &Service{
			Type: "NodePort",
			Ports: []*Port{{
				Name:       "http",
				Port:       80,
				TargetPort: 8080,
				NodePort:   30080,
			}, {
				Name:       "https",
				Port:       443,
				TargetPort: 8443,
				NodePort:   30443,
			},
			},
			Annotations: &ModuleAnnotation{
				Module: &AmbassadorModuleAnnotation{
					ApiVersion: "ambassador/v2",
					Kind:       "Module",
					Name:       "ambassador",
					Config: &AmbassadorModuleConfig{
						UseProxyProto: true,
						EnableGRPCWeb: false,
						Diagnostics: &AmbassadorDiagnosticst{
							Enabled: false,
						},
					},
				},
			},
		},
		ServiceAccount: &ServiceAccount{
			Create: true,
		},
		Resources: &k8s.Resources{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("500Mi"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("250m"),
				corev1.ResourceMemory: resource.MustParse("250Mi"),
			},
		},
	}
}

func defaultServiceAnnotations() map[string]string {
	return map[string]string{
		"getambassador.io/config": `---
apiVersion: ambassador/v1
kind: Module
name: tls
config:
  server:
    enabled: True
    # secret: MY_TLS_SECRET_NAME
    redirect_cleartext_from: 8080`,
	}
}

func defaultExporterConfig() string {
	return `---
defaults:
  timer_type: histogram
mappings:
###### Envoy global

### Downstream RQ
- match: envoy.http.*.downstream_rq_total
  name: envoy_http_downstream_rq_total
  labels: 
    cluster: "$1"
- match: envoy.http.*.rq_total
  name: envoy_http_rq_total
  labels: 
    cluster: "$1"
- match: envoy.http.*.downstream_cx_total
  name: envoy_http_downstream_cx_total
  labels: 
    cluster: "$1"
- match: envoy.http.*.downstream_cx_ssl_total
  name: envoy_http_downstream_cx_ssl_total
  labels: 
    cluster: "$1"
- match: envoy\.http\.(.*)\.downstream_rq_(.*)
  match_type: regex
  name: envoy_http_downstream_rq_xxx
  labels: 
    cluster: "$1"
    response_code_class: "$2"
- match: envoy.http.*.downstream_cx_active
  name: envoy_http_downstream_cx_active
  labels: 
    cluster: "$1"
- match: envoy.http.*.downstream_cx_ssl_active
  name: envoy_http_downstream_cx_ssl_active
  labels: 
    cluster: "$1"
- match: envoy.http.*.downstream_rq_active
  name: envoy_http_downstream_rq_active
  labels: 
    cluster: "$1"
- match: envoy.http.*.downstream_cx_length_ms
  name: envoy_http_downstream_cx_length_ms
  labels: 
    cluster: "$1"
- match: envoy.http.*.downstream_cx_rx_bytes_total
  name: envoy_http_downstream_cx_rx_bytes_total
  labels: 
    cluster: "$1"
- match: envoy.http.*.downstream_cx_tx_bytes_total
  name: envoy_http_downstream_cx_tx_bytes_total
  labels: 
    cluster: "$1"

### Upstream CX
- match: envoy.cluster.*.upstream_cx_total
  name: envoy_cluster_upstream_cx_total
  labels:
    cluster: "$1"
- match: envoy.cluster.*.upstream_cx_active
  name: envoy_cluster_upstream_cx_active
  labels:
    cluster: "$1"
- match: envoy.cluster.*.upstream_connect_fail
  name: envoy_cluster_upstream_connect_fail
  labels:
    cluster: "$1"    
- match: envoy.cluster.*.upstream_cx_connect_timeout
  name: envoy_cluster_upstream_cx_connect_timeout
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.upstream_cx_destroy_local_with_active_rq
  name: envoy_cluster_upstream_cx_destroy_local_with_active_rq
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.upstream_cx_destroy_remote_active_rq
  name: envoy_cluster_upstream_cx_destroy_remote_active_rq
  labels: 
    cluster: "$1"

### Upstream RQ
- match: envoy\.cluster\.(.*)\.upstream_rq_(.*)
  match_type: regex
  name: envoy_cluster_upstream_rq_xxx
  labels: 
    cluster: "$1"
    response_code_class: "$2"
- match: envoy.cluster.*.upstream_rq_completed
  name: envoy_cluster_upstream_rq_completed
  labels: 
    cluster: "$1"
    response_code_class: "$2"

- match: envoy.cluster.*.upstream_rq_timeout
  name: envoy_cluster_upstream_rq_timeout
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.upstream_rq_per_try_timeout
  name: envoy_cluster_upstream_rq_per_try_timeout
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.upstream_rq_pending_overflow
  name: envoy_cluster_upstream_rq_pending_overflow
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.upstream_rq_pending_failure_eject
  name: envoy_cluster_upstream_rq_pending_failure_eject
  labels: 
    cluster: "$1"

- match: envoy.cluster.*.upstream_rq_retry
  name: envoy_cluster_upstream_rq_retry
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.upstream_rq_retry_success
  name: envoy_cluster_upstream_rq_retry_success
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.upstream_rq_retry_overflow
  name: envoy_cluster_upstream_rq_retry_overflow
  labels: 
    cluster: "$1"

### Outlier
- match: envoy.cluster.*.outlier_detection_ejections_active
  name: envoy_cluster_outlier_detection_ejections_active
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.outlier_detection_ejections_enforced_total
  name: envoy_cluster_outlier_detection_ejections_enforced_total
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.outlier_detection_ejections_overflow
  name: envoy_cluster_outlier_detection_ejections_overflow
  labels: 
    cluster: "$1"

### Healtcheck
- match: envoy.cluster.*.health_check.attempt
  name: envoy_cluster_health_check_attempt
  labels:
    cluster: "$1"
- match: envoy.cluster.*.health_check.success
  name: envoy_cluster_health_check_success
  labels:
    cluster: "$1"
- match: envoy.cluster.*.health_check.failure
  name: envoy_cluster_health_check_failure
  labels:
    cluster: "$1"

### Envoy Service
- match: envoy.cluster.*.upstream_rq_pending_active
  name: envoy_cluster_upstream_rq_pending_active
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.upstream_rq_active
  name: envoy_cluster_upstream_rq_active
  labels: 
    cluster: "$1"
- match: envoy\.cluster\.(.*)\.downstream_rq_(.*)
  match_type: regex
  name: envoy_cluster_downstream_rq_xxx
  labels: 
    cluster: "$1"
    response_code_class: "$2"

- match: envoy.http.*.downstream_cx_destroy_remote_active_rq
  name: envoy_http_downstream_cx_destroy_remote_active_rq
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.upstream_rq_maintenance_mode
  name: envoy_cluster_upstream_rq_maintenance_mode
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.upstream_rq_retry
  name: envoy_cluster_upstream_rq_retry
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.upstream_rq_rx_reset
  name: envoy_cluster_upstream_rq_rx_reset
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.upstream_rq_retry_success
  name: envoy_cluster_upstream_rq_retry_success
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.upstream_rq_retry_overflow
  name: envoy_cluster_upstream_rq_retry_overflow
  labels: 
    cluster: "$1"

# Upstream Flow Control
- match: envoy.cluster.*.upstream_flow_control_paused_reading_total
  name: envoy_cluster_upstream_flow_control_paused_reading_total
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.upstream_flow_control_resumed_reading_total
  name: envoy_cluster_upstream_flow_control_resumed_reading_total
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.upstream_flow_control_backed_up_total
  name: envoy_cluster_upstream_flow_control_backed_up_total
  labels: 
    cluster: "$1"
- match: envoy.cluster.*.upstream_flow_control_drained_total
  name: envoy_cluster_upstream_flow_control_drained_total
  labels: 
    cluster: "$1"

### Upstream time
- match: envoy.cluster.*.upstream_rq_time
  name: envoy_cluster_upstream_rq_time
  labels:
    cluster: "$1"
- match: envoy.cluster.*.upstream_rq_time_count
  name: envoy_cluster_upstream_rq_time_count
  labels:
    cluster: "$1"
- match: envoy.cluster.*.upstream_rq_time_sum
  name: envoy_cluster_upstream_rq_time_sum
  labels:
    cluster: "$1"
- match: envoy.cluster.*.upstream_rq_time_bucket
  name: envoy_cluster_upstream_rq_time_bucket
  labels:
    cluster: "$1"

### Downstream time
- match: envoy.http.*.downstream_rq_time
  name: envoy_http_downstream_rq_time
  labels:
    cluster: "$1"
- match: envoy.http.*.downstream_rq_time_count
  name: envoy_http_downstream_rq_time_count
  labels:
    cluster: "$1"
- match: envoy.http.*.downstream_rq_time_sum
  name: envoy_http_downstream_rq_time_sum
  labels:
    cluster: "$1"
- match: envoy.http.*.downstream_rq_time_bucket
  name: envoy_http_downstream_rq_time
  labels:
    cluster: "$1"

### BEGIN General
- match: envoy.cluster.*.membership_healthy
  name: envoy_cluster_membership_healthy
  labels:
    cluster: "$1"
- match: envoy.cluster.*.membership_change
  name: envoy_cluster_membership_change
  labels:
    cluster: "$1"
- match: envoy.cluster.*.membership_total
  name: envoy_cluster_membership_total
  labels:
    cluster: "$1" `
}
