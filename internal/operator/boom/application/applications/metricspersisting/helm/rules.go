package helm

import (
	"gopkg.in/yaml.v3"
)

const (
	cockroachStatefulsetName = "cockroachdb"
	cockroachServiceName     = "cockroachdb-public"
	cockroachNamespace       = "caos-zitadel"
	caosToolingNamespace     = "caos-system"
	nodeExporterService      = "node-exporter"
	kubeStateMetricsService  = "kube-state-metrics"
)

func GetDefaultRules(labels map[string]string) (*AdditionalPrometheusRules, error) {
	rulesStr := `name: node-exporter.rules
groups:
- name: node-exporter.rules
  rules:
  - expr: |-
      count without (cpu) (
        count without (mode) (
          node_cpu_seconds_total{job="` + nodeExporterService + `"}
        )
      )
    record: instance:node_num_cpu:sum
  - expr: |-
      1 - avg without (cpu, mode) (
        rate(node_cpu_seconds_total{job="` + nodeExporterService + `", mode="idle"}[1m])
      )
    record: instance:node_cpu_utilisation:rate1m
  - expr: |-
      (
        node_load1{job="` + nodeExporterService + `"}
      /
        instance:node_num_cpu:sum{job="` + nodeExporterService + `"}
      )
    record: instance:node_load1_per_cpu:ratio
  - expr: |-
      1 - (
        node_memory_MemAvailable_bytes{job="` + nodeExporterService + `"}
      /
        node_memory_MemTotal_bytes{job="` + nodeExporterService + `"}
      )
    record: instance:node_memory_utilisation:ratio
  - expr: rate(node_vmstat_pgmajfault{job="` + nodeExporterService + `"}[1m])
    record: instance:node_vmstat_pgmajfault:rate1m
  - expr: rate(node_disk_io_time_seconds_total{job="` + nodeExporterService + `", device=~"nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+"}[1m])
    record: instance_device:node_disk_io_time_seconds:rate1m
  - expr: rate(node_disk_io_time_weighted_seconds_total{job="` + nodeExporterService + `", device=~"nvme.+|rbd.+|sd.+|vd.+|xvd.+|dm-.+"}[1m])
    record: instance_device:node_disk_io_time_weighted_seconds:rate1m
  - expr: |-
      sum without (device) (
        rate(node_network_receive_bytes_total{job="` + nodeExporterService + `", device!="lo"}[1m])
      )
    record: instance:node_network_receive_bytes_excluding_lo:rate1m
  - expr: |-
      sum without (device) (
        rate(node_network_transmit_bytes_total{job="` + nodeExporterService + `", device!="lo"}[1m])
      )
    record: instance:node_network_transmit_bytes_excluding_lo:rate1m
  - expr: |-
      sum without (device) (
        rate(node_network_receive_drop_total{job="` + nodeExporterService + `", device!="lo"}[1m])
      )
    record: instance:node_network_receive_drop_excluding_lo:rate1m
  - expr: |-
      sum without (device) (
        rate(node_network_transmit_drop_total{job="` + nodeExporterService + `", device!="lo"}[1m])
      )
    record: instance:node_network_transmit_drop_excluding_lo:rate1m
- name: node.rules
  rules:
  - expr: sum(min(kube_pod_info) by (node))
    record: ':kube_pod_info_node_count:'
  - expr: max(label_replace(kube_pod_info{job="` + kubeStateMetricsService + `"}, "pod", "$1", "pod", "(.*)")) by (node, namespace, pod)
    record: 'node_namespace_pod:kube_pod_info:'
  - expr: |-
      count by (node) (sum by (node, cpu) (
        node_cpu_seconds_total{job="` + nodeExporterService + `"}
      * on (namespace, pod) group_left(node)
        node_namespace_pod:kube_pod_info:
      ))
    record: node:node_num_cpu:sum
  - expr: |-
      sum(
        node_memory_MemAvailable_bytes{job="` + nodeExporterService + `"} or
        (
          node_memory_Buffers_bytes{job="` + nodeExporterService + `"} +
          node_memory_Cached_bytes{job="` + nodeExporterService + `"} +
          node_memory_MemFree_bytes{job="` + nodeExporterService + `"} +
          node_memory_Slab_bytes{job="` + nodeExporterService + `"}
        )
      )
    record: :node_memory_MemAvailable_bytes:sum
- name: caos.rules
  rules:
   - expr: dist_node_boot_time_seconds
     record: caos_node_boot_time_seconds
   - expr: floor(avg_over_time(dist_systemd_unit_active[5m])+0.2)
     record: caos_systemd_unit_active
   - expr: min(min_over_time(caos_systemd_unit_active[5m])) by (instance)
     record: caos_systemd_ryg
   - expr: avg(max_over_time(caos_probe{type="Upstream",name!="httpingress"}[1m])) by (name)
     record: caos_upstream_probe_ryg
   - expr: max_over_time(caos_probe{type="VIP"}[1m])
     record: caos_vip_probe_ryg
   - expr: sum(1 - avg(rate(dist_node_cpu_seconds_total[5m])))
     record: caos_cluster_cpu_utilisation_5m
   - expr: 100 - (avg by (instance) (irate(dist_node_cpu_seconds_total[5m])) * 100)
     record: caos_node_cpu_utilisation_5m
   - expr: (clamp_max(clamp_min(100-caos_node_cpu_utilisation_5m, 10),20)-10)/10
     record: caos_node_cpu_ryg
   - expr: |-
       sum by (instance) (100 -
       (
         dist_node_memory_MemAvailable_bytes
       /
         dist_node_memory_MemTotal_bytes
       * 100
       ))
     record: caos_node_memory_utilisation
   - expr: (clamp_max(clamp_min(100-caos_node_memory_utilisation, 10),20)-10)/10
     record: caos_node_memory_ryg
   - expr: |-
      100 - (
       min by (instance) (dist_node_filesystem_avail_bytes)
       / min by (instance) (dist_node_filesystem_size_bytes)
       * 100)
     record: caos_node_disk_utilisation
   - expr: dist_kube_node_status_condition
     record: caos_node_ready
   - expr: min_over_time(caos_node_ready[5m])
     record: caos_k8s_node_ryg
   - expr: dist_etcd_server_has_leader or on(instance) up{job="caos_remote_etcd"}
     record: caos_etcd_server_has_leader_and_is_up
   - expr: min_over_time(caos_etcd_server_has_leader_and_is_up[5m])
     record: caos_etcd_ryg
   - expr: |-
       clamp_max(
         clamp_min(
           (
             max_over_time(dist_kube_deployment_status_replicas_available{namespace=~"(kube-system|` + caosToolingNamespace + `|` + cockroachNamespace + `)"}[5m]) -
             dist_kube_deployment_spec_replicas{namespace=~"(kube-system|` + caosToolingNamespace + `|` + cockroachNamespace + `)"} or
             max_over_time(dist_kube_statefulset_status_replicas_ready{namespace=~"(kube-system|` + caosToolingNamespace + `|` + cockroachNamespace + `)"}[5m]) -
             dist_kube_statefulset_replicas{namespace=~"(kube-system|` + caosToolingNamespace + `|` + cockroachNamespace + `)"} or
             max_over_time(dist_kube_daemonset_status_number_available{namespace=~"(kube-system|` + caosToolingNamespace + `|` + cockroachNamespace + `)"}[5m]) -
             dist_kube_daemonset_status_desired_number_scheduled{namespace=~"(kube-system|` + caosToolingNamespace + `|` + cockroachNamespace + `)"}
           ) + 
           1,
           0
         ),
         1
       )
     record: caos_ready_pods_ryg
   - expr: |-
       clamp_max(
         clamp_min(
           (
             max_over_time(dist_kube_deployment_status_replicas{namespace=~"(kube-system|` + caosToolingNamespace + `|` + cockroachNamespace + `)"}[5m]) -
             dist_kube_deployment_spec_replicas{namespace=~"(kube-system|` + caosToolingNamespace + `|` + cockroachNamespace + `)"} or
             max_over_time(dist_kube_statefulset_status_replicas_current{namespace=~"(kube-system|` + caosToolingNamespace + `|` + cockroachNamespace + `)"}[5m]) -
             dist_kube_statefulset_replicas{namespace=~"(kube-system|` + caosToolingNamespace + `|` + cockroachNamespace + `)"} or
             max_over_time(dist_kube_daemonset_status_current_number_scheduled{namespace=~"(kube-system|` + caosToolingNamespace + `|` + cockroachNamespace + `)"}[5m]) -
             dist_kube_daemonset_status_desired_number_scheduled{namespace=~"(kube-system|` + caosToolingNamespace + `|` + cockroachNamespace + `)"}
           ) +
           1,
           0
         ),
         1
       )          
     record: caos_scheduled_pods_ryg
   - expr: |-
       sum(dist_kube_deployment_spec_replicas) + sum(dist_kube_statefulset_replicas) + sum(dist_kube_daemonset_status_desired_number_scheduled) 
     record: caos_desired_pods
   - expr: |-
       sum(dist_kube_deployment_status_replicas) + sum(dist_kube_statefulset_status_replicas_current) + sum(dist_kube_daemonset_status_current_number_scheduled)
     record: caos_scheduled_pods
   - expr: |-
       sum(dist_kube_deployment_status_replicas_available) + sum(dist_kube_statefulset_status_replicas_ready) + sum(dist_kube_daemonset_status_number_available)
     record: caos_ready_pods

# Cockroachdb aggregation rules 
   - record: node:capacity
     expr: sum without(store) (capacity{job="` + cockroachServiceName + `"})
   - record: cluster:capacity
     expr: sum without(instance) (node:capacity{job="` + cockroachServiceName + `"})
   - record: node:capacity_available
     expr: sum without(store) (capacity_available{job="` + cockroachServiceName + `"})
   - record: cluster:capacity_available
     expr: sum without(instance) (node:capacity_available{job="` + cockroachServiceName + `"})
   - record: capacity_available:ratio
     expr: capacity_available{job="` + cockroachServiceName + `"} / capacity{job="` + cockroachServiceName + `"}
   - record: node:capacity_available:ratio
     expr: node:capacity_available{job="` + cockroachServiceName + `"} / node:capacity{job="` + cockroachServiceName + `"}
   - record: cluster:capacity_available:ratio
     expr: cluster:capacity_available{job="` + cockroachServiceName + `"} / cluster:capacity{job="` + cockroachServiceName + `"}
   # Histogram rules: these are fairly expensive to compute live, so we precompute a few percetiles.
   - record: txn_durations_bucket:rate1m
     expr: rate(txn_durations_bucket{job="` + cockroachServiceName + `"}[1m])
   - record: txn_durations:rate1m:quantile_50
     expr: histogram_quantile(0.5, txn_durations_bucket:rate1m)
   - record: txn_durations:rate1m:quantile_75
     expr: histogram_quantile(0.75, txn_durations_bucket:rate1m)
   - record: txn_durations:rate1m:quantile_90
     expr: histogram_quantile(0.9, txn_durations_bucket:rate1m)
   - record: txn_durations:rate1m:quantile_95
     expr: histogram_quantile(0.95, txn_durations_bucket:rate1m)
   - record: txn_durations:rate1m:quantile_99
     expr: histogram_quantile(0.99, txn_durations_bucket:rate1m)
   - record: exec_latency_bucket:rate1m
     expr: rate(exec_latency_bucket{job="` + cockroachServiceName + `"}[1m])
   - record: exec_latency:rate1m:quantile_50
     expr: histogram_quantile(0.5, exec_latency_bucket:rate1m)
   - record: exec_latency:rate1m:quantile_75
     expr: histogram_quantile(0.75, exec_latency_bucket:rate1m)
   - record: exec_latency:rate1m:quantile_90
     expr: histogram_quantile(0.9, exec_latency_bucket:rate1m)
   - record: exec_latency:rate1m:quantile_95
     expr: histogram_quantile(0.95, exec_latency_bucket:rate1m)
   - record: exec_latency:rate1m:quantile_99
     expr: histogram_quantile(0.99, exec_latency_bucket:rate1m)
   - record: round_trip_latency_bucket:rate1m
     expr: rate(round_trip_latency_bucket{job="` + cockroachServiceName + `"}[1m])
   - record: round_trip_latency:rate1m:quantile_50
     expr: histogram_quantile(0.5, round_trip_latency_bucket:rate1m)
   - record: round_trip_latency:rate1m:quantile_75
     expr: histogram_quantile(0.75, round_trip_latency_bucket:rate1m)
   - record: round_trip_latency:rate1m:quantile_90
     expr: histogram_quantile(0.9, round_trip_latency_bucket:rate1m)
   - record: round_trip_latency:rate1m:quantile_95
     expr: histogram_quantile(0.95, round_trip_latency_bucket:rate1m)
   - record: round_trip_latency:rate1m:quantile_99
     expr: histogram_quantile(0.99, round_trip_latency_bucket:rate1m)
   - record: sql_exec_latency_bucket:rate1m
     expr: rate(sql_exec_latency_bucket{job="` + cockroachServiceName + `"}[1m])
   - record: sql_exec_latency:rate1m:quantile_50
     expr: histogram_quantile(0.5, sql_exec_latency_bucket:rate1m)
   - record: sql_exec_latency:rate1m:quantile_75
     expr: histogram_quantile(0.75, sql_exec_latency_bucket:rate1m)
   - record: sql_exec_latency:rate1m:quantile_90
     expr: histogram_quantile(0.9, sql_exec_latency_bucket:rate1m)
   - record: sql_exec_latency:rate1m:quantile_95
     expr: histogram_quantile(0.95, sql_exec_latency_bucket:rate1m)
   - record: sql_exec_latency:rate1m:quantile_99
     expr: histogram_quantile(0.99, sql_exec_latency_bucket:rate1m)
   - record: raft_process_logcommit_latency_bucket:rate1m
     expr: rate(raft_process_logcommit_latency_bucket{job="` + cockroachServiceName + `"}[1m])
   - record: raft_process_logcommit_latency:rate1m:quantile_50
     expr: histogram_quantile(0.5, raft_process_logcommit_latency_bucket:rate1m)
   - record: raft_process_logcommit_latency:rate1m:quantile_75
     expr: histogram_quantile(0.75, raft_process_logcommit_latency_bucket:rate1m)
   - record: raft_process_logcommit_latency:rate1m:quantile_90
     expr: histogram_quantile(0.9, raft_process_logcommit_latency_bucket:rate1m)
   - record: raft_process_logcommit_latency:rate1m:quantile_95
     expr: histogram_quantile(0.95, raft_process_logcommit_latency_bucket:rate1m)
   - record: raft_process_logcommit_latency:rate1m:quantile_99
     expr: histogram_quantile(0.99, raft_process_logcommit_latency_bucket:rate1m)
   - record: raft_process_commandcommit_latency_bucket:rate1m
     expr: rate(raft_process_commandcommit_latency_bucket{job="cockroachdb"}[1m])
   - record: raft_process_commandcommit_latency:rate1m:quantile_50
     expr: histogram_quantile(0.5, raft_process_commandcommit_latency_bucket:rate1m)
   - record: raft_process_commandcommit_latency:rate1m:quantile_75
     expr: histogram_quantile(0.75, raft_process_commandcommit_latency_bucket:rate1m)
   - record: raft_process_commandcommit_latency:rate1m:quantile_90
     expr: histogram_quantile(0.9, raft_process_commandcommit_latency_bucket:rate1m)
   - record: raft_process_commandcommit_latency:rate1m:quantile_95
     expr: histogram_quantile(0.95, raft_process_commandcommit_latency_bucket:rate1m)
   - record: raft_process_commandcommit_latency:rate1m:quantile_99
     expr: histogram_quantile(0.99, raft_process_commandcommit_latency_bucket:rate1m)

# ZITADEL CockroachDB Runtime
   - record: cr_runtime_pod_flapping
     expr: resets(sys_uptime{job="` + cockroachServiceName + `"}[10m])
   - record: cr_runtime_pod_flapping_ryg
     expr: clamp_min(1 - cr_runtime_pod_flapping, 0)
   - record: cr_runtime_high_open_fd_count
     expr: sys_fd_open{job="` + cockroachServiceName + `"} / sys_fd_softlimit{job="` + cockroachServiceName + `"}
   - record: cr_runtime_high_open_fd_count_ryg
     expr: (1 - (clamp_max(clamp_min(floor(cr_runtime_high_open_fd_count * 100), 20), 30) - 20) / 10)
   - record: caos_cr_runtime_pods_ryg
     expr: |-
       cr_runtime_pod_flapping_ryg
       * cr_runtime_high_open_fd_count_ryg

   - record: cr_runtime_version_mismatches
     expr: count by(cluster) (count_values by(tag, cluster) ("version", build_timestamp{job="` + cockroachServiceName + `"}))
   - record: cr_runtime_version_mismatches_ryg
     expr: clamp_min(2 - cr_runtime_version_mismatches, 0)
   - record: caos_cr_runtime_cluster_ryg
     expr: |-
       min(caos_ready_pods_ryg{controller="` + cockroachStatefulsetName + `",namespace="` + cockroachNamespace + `"})
       * min(caos_scheduled_pods_ryg{controller="` + cockroachStatefulsetName + `",namespace="` + cockroachNamespace + `"})
       * cr_runtime_version_mismatches
   - record: caos_security_certificate_expiration_ca
     expr: (security_certificate_expiration_ca{job="` + cockroachServiceName + `"} - time()) / 86400
   - record: caos_security_certificate_expiration_node
     expr: (security_certificate_expiration_node{job="` + cockroachServiceName + `"} - time()) / 86400
       

# ZITADEL CockroachDB Capacity
   - record: node:capacity
     expr: sum without(store) (capacity{job="` + cockroachServiceName + `"})
   - record: cluster:capacity
     expr: sum without(instance) (node:capacity{job="` + cockroachServiceName + `"})
   - record: node:capacity_available
     expr: sum without(store) (capacity_available{job="` + cockroachServiceName + `"})
   - record: cluster:capacity_available
     expr: sum without(instance) (node:capacity_available{job="` + cockroachServiceName + `"})
   - record: capacity_available:ratio
     expr: capacity_available{job="` + cockroachServiceName + `"} / capacity{job="` + cockroachServiceName + `"}
   - record: node:capacity_available:ratio
     expr: node:capacity_available{job="` + cockroachServiceName + `"} / node:capacity{job="` + cockroachServiceName + `"}
   - record: cluster:capacity_available:ratio
     expr: cluster:capacity_available{job="` + cockroachServiceName + `"} / cluster:capacity{job="` + cockroachServiceName + `"}

   - record: caos_cr_capacity_store_ryg
     expr: (clamp_max(clamp_min(100 * capacity_available:ratio{job="` + cockroachServiceName + `"}, 15), 30)-15) /15
   - record: caos_cr_capacity_cluster_ryg
     expr: (clamp_max(clamp_min(100 * cluster:capacity_available:ratio{job="` + cockroachServiceName + `"}, 20), 30) - 20) /10

# ZITADEL CockroachDB Replicas
   - record: cr_replicas_clock_offset_near_max_ryg
     expr: 1 - (clamp_max(clamp_min(abs(clock_offset_meannanos), 200 * 1000 * 1000), 300 * 1000 * 1000) - 200 * 1000 * 1000) / (100 * 1000 * 1000)
   - record: cr_ca_certificate_expires_soon_ryg
     expr: (clamp_max(clamp_min(security_certificate_expiration_ca{job="` + cockroachServiceName + `"}- time(), 86400 * 180), 86400 * 366) - 86400 * 180) / (86400 * 186)
#   - record: cr_client_ca_certificate_expires_soon_ryg
#     expr: (clamp_max(clamp_min(security_certificate_expiration_client_ca{job="` + cockroachServiceName + `"}- time(), 86400 * 180), 86400 * 366) - 86400 * 180) / (86400 * 186)
   - record: cr_node_certificate_expires_soon_ryg
     expr: (clamp_max(clamp_min(security_certificate_expiration_node{job="` + cockroachServiceName + `"}- time(), 86400 * 180), 86400 * 366) - 86400 * 180) / (86400 * 186)
#   - record: cr_node_client_certificate_expires_soon_ryg
#     expr: (clamp_max(clamp_min(cr_node_client_certificate_expires_soon{job="` + cockroachServiceName + `"}- time(), 86400 * 180), 86400 * 366) - 86400 * 180) / (86400 * 186)
   - record: caos_cr_replicas_nodes_ryg
     expr: |-
       cr_replicas_clock_offset_near_max_ryg
       * cr_ca_certificate_expires_soon_ryg
       * cr_node_certificate_expires_soon_ryg
#       * cr_client_ca_certificate_expires_soon_ryg
#       * cr_node_client_certificate_expires_soon_ryg
   - record: caos_cr_replicas_stores_ryg
     expr: |-
       clamp_min(1 - requests_slow_latch{job="` + cockroachServiceName + `"}, 0)
       * clamp_min(1 - requests_slow_lease{job="` + cockroachServiceName + `"}, 0)
       * clamp_min(1 - requests_slow_raft{job="` + cockroachServiceName + `"}, 0)

# Cluster RYG
   - record: caos_orb_ryg
     expr: |-
       # Platform Monitoring
       min(caos_node_cpu_ryg)
       * min(caos_systemd_ryg)
       * min(caos_vip_probe_ryg)
       * min(caos_upstream_probe_ryg)
       * min(caos_node_memory_ryg)
       * min(caos_k8s_node_ryg)
       * avg(caos_etcd_ryg)
       * min(caos_ready_pods_ryg{namespace=~"(kube-system|` + caosToolingNamespace + `|` + cockroachNamespace + `)"})
       * min(caos_scheduled_pods_ryg{namespace=~"(kube-system|` + caosToolingNamespace + `|` + cockroachNamespace + `)"})
       # ZITADEL Monitoring
       * min(caos_cr_runtime_cluster_ryg) or on() 1 - avg(up{endpoint="http",job="` + cockroachServiceName + `",namespace="` + cockroachNamespace + `",pod="` + cockroachStatefulsetName + `-0",service="` + cockroachServiceName + `"} or on() vector(0)) 
       * min(caos_cr_runtime_pods_ryg) or on() 1 - (up{endpoint="http",job="` + cockroachServiceName + `",namespace="` + cockroachNamespace + `",pod="` + cockroachStatefulsetName + `-0",service="` + cockroachServiceName + `"} or on() vector(0))
       * min(caos_cr_replicas_nodes_ryg) or on() 1 - (up{endpoint="http",job="` + cockroachServiceName + `",namespace="` + cockroachNamespace + `",pod="` + cockroachStatefulsetName + `-0",service="` + cockroachServiceName + `"} or on() vector(0))
       * min(caos_cr_replicas_stores_ryg) or on() 1 - (up{endpoint="http",job="` + cockroachServiceName + `",namespace="` + cockroachNamespace + `",pod="` + cockroachStatefulsetName + `-0",service="` + cockroachServiceName + `"} or on() vector(0))
       * min(caos_cr_capacity_cluster_ryg) or on() 1 - (up{endpoint="http",job="` + cockroachServiceName + `",namespace="` + cockroachNamespace + `",pod="` + cockroachStatefulsetName + `-0",service="` + cockroachServiceName + `"} or on() vector(0))
       * min(caos_cr_capacity_store_ryg) or on() 1 - (up{endpoint="http",job="` + cockroachServiceName + `",namespace="` + cockroachNamespace + `",pod="` + cockroachStatefulsetName + `-0",service="` + cockroachServiceName + `"} or on() vector(0))
`

	struc := &AdditionalPrometheusRules{
		AdditionalLabels: labels,
	}
	rulesByte := []byte(rulesStr)
	if err := yaml.Unmarshal(rulesByte, struc); err != nil {
		return nil, err
	}
	return struc, nil
}
