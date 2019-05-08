// Copyright 2017-2019 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package metrics holds prometheus metrics objects and related utility functions. It
// does not abstract away the prometheus client but the caller rarely needs to
// refer to prometheus directly.
package metrics

// Adding a metric
// - Add a metric object of the appropriate type as an exported variable
// - Register the new object in the init function

import (
	"net/http"
	"syscall"

	"github.com/cilium/cilium/api/v1/models"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
)

const (
	// ErrorTimeout is the value used to notify timeout errors.
	ErrorTimeout = "timeout"

	// ErrorProxy is the value used to notify errors on Proxy.
	ErrorProxy = "proxy"

	//L7DNS is the value used to report DNS label on metrics
	L7DNS = "dns"

	// SubsystemBPF is the subsystem to scope metrics related to the bpf syscalls.
	SubsystemBPF = "bpf"

	// SubsystemDatapath is the subsystem to scope metrics related to management of
	// the datapath. It is prepended to metric names and separated with a '_'.
	SubsystemDatapath = "datapath"

	// SubsystemAgent is the subsystem to scope metrics related to the cilium agent itself.
	SubsystemAgent = "agent"

	// SubsystemK8s is the subsystem to scope metrics related to Kubernetes
	SubsystemK8s = "k8s"

	// SubsystemK8sClient is the subsystem to scope metrics related to the kubernetes client.
	SubsystemK8sClient = "k8s_client"

	// SubsystemKVStore is the subsystem to scope metrics related to the kvstore.
	SubsystemKVStore = "kvstore"

	// SubsystemNodes is the subsystem to scope metrics related to the node manager.
	SubsystemNodes = "nodes"

	// SubsystemTriggers is the subsystem to scope metrics related to the trigger package.
	SubsystemTriggers = "triggers"

	// Namespace is used to scope metrics from cilium. It is prepended to metric
	// names and separated with a '_'
	Namespace = "cilium"

	// LabelOutcome indicates whether the outcome of the operation was successful or not
	LabelOutcome = "outcome"

	// LabelAttempts is the number of attempts it took to complete the operation
	LabelAttempts = "attempts"

	// Labels

	// LabelValueOutcomeSuccess is used as a successful outcome of an operation
	LabelValueOutcomeSuccess = "success"

	// LabelValueOutcomeFail is used as an unsuccessful outcome of an operation
	LabelValueOutcomeFail = "fail"

	// LabelEventSourceAPI marks event-related metrics that come from the API
	LabelEventSourceAPI = "api"

	// LabelEventSourceK8s marks event-related metrics that come from k8s
	LabelEventSourceK8s = "k8s"

	// LabelEventSourceFQDN marks event-related metrics that come from pkg/fqdn
	LabelEventSourceFQDN = "fqdn"

	// LabelEventSourceContainerd marks event-related metrics that come from docker
	LabelEventSourceContainerd = "docker"

	// LabelDatapathArea marks which area the metrics are related to (eg, which BPF map)
	LabelDatapathArea = "area"

	// LabelDatapathName marks a unique identifier for this metric.
	// The name should be defined once for a given type of error.
	LabelDatapathName = "name"

	// LabelDatapathFamily marks which protocol family (IPv4, IPV6) the metric is related to.
	LabelDatapathFamily = "family"

	// LabelProtocol marks the L4 protocol (TCP, ANY) for the metric.
	LabelProtocol = "protocol"

	// LabelStatus the label from completed task
	LabelStatus = "status"

	// LabelPolicyEnforcement is the label used to see the enforcement status
	LabelPolicyEnforcement = "enforcement"

	// LabelPolicySource is the label used to see the enforcement status
	LabelPolicySource = "source"

	// LabelScope is the label used to defined multiples scopes in the same
	// metric. For example, one counter may measure a metric over the scope of
	// the entire event (scope=global), or just part of an event
	// (scope=slow_path)
	LabelScope = "scope"

	// LabelProtocolL7 is the label used when working with layer 7 protocols.
	LabelProtocolL7 = "protocol_l7"

	// LabelBuildState is the state a build queue entry is in
	LabelBuildState = "state"

	// LabelBuildQueueName is the name of the build queue
	LabelBuildQueueName = "name"

	// LabelAction is the label used to defined what kind of action was performed in a metric
	LabelAction = "action"

	// LabelSubsystem is the label used to refer to any of the child process
	// started by cilium (Envoy, monitor, etc..)
	LabelSubsystem = "subsystem"

	// LabelKind is the kind a label
	LabelKind = "kind"

	// LabelPath is the label for the API path
	LabelPath = "path"
	// LabelMethod is the label for the HTTP method
	LabelMethod = "method"

	// LabelAPIReturnCode is the HTTP code returned for that API path
	LabelAPIReturnCode = "return_code"

	// LabelOperation is the label for BPF maps operations
	LabelOperation = "operation"

	// LabelMapName is the label for the BPF map name
	LabelMapName = "mapName"
)

var (
	registry = prometheus.NewPedanticRegistry()

	// APIInteractions is the total time taken to process an API call made
	// to the cilium-agent
	APIInteractions = NoOpObserverVec

	// Endpoint

	// EndpointCount is a function used to collect this metric.
	// It must be thread-safe.
	EndpointCount prometheus.GaugeFunc

	// EndpointCountRegenerating is the number of endpoints currently regenerating
	EndpointCountRegenerating = NoOpGauge

	// EndpointRegenerationCount is a count of the number of times any endpoint
	// has been regenerated and success/fail outcome
	EndpointRegenerationCount = NoOpCounterVec

	// EndpointStateCount is the total count of the endpoints in various states.
	EndpointStateCount = NoOpGaugeVec

	// EndpointRegenerationTimeStats is the total time taken to regenerate
	// endpoints, labeled by span name and status ("success" or "failure")
	EndpointRegenerationTimeStats = NoOpObserverVec

	// Policy

	// PolicyCount is the number of policies loaded into the agent
	PolicyCount = NoOpGauge

	// PolicyRegenerationCount is the total number of successful policy
	// regenerations.
	PolicyRegenerationCount = NoOpCounter

	// PolicyRegenerationTimeStats is the total time taken to generate policies
	PolicyRegenerationTimeStats = NoOpObserverVec

	// PolicyRevision is the current policy revision number for this agent
	PolicyRevision = NoOpGauge

	// PolicyImportErrors is a count of failed policy imports
	PolicyImportErrors = NoOpCounter

	// PolicyEndpointStatus is the number of endpoints with policy labeled by enforcement type
	PolicyEndpointStatus = NoOpGaugeVec

	// PolicyImplementationDelay is a distribution of times taken from adding a
	// policy (and incrementing the policy revision) to seeing it in the datapath
	// per Endpoint. This reflects the actual delay percieved by traffic flowing
	// through the datapath. The longest times will roughly correlate with the
	// time taken to fully deploy an endpoint.
	PolicyImplementationDelay = NoOpObserverVec

	// Events

	// EventTS*is the time in seconds since epoch that we last received an
	// event that we will handle
	// source is one of k8s, docker or apia

	// EventTSK8s is the timestamp of k8s events
	EventTSK8s = NoOpGauge

	// EventTSContainerd is the timestamp of docker events
	EventTSContainerd = NoOpGauge

	// EventTSAPI is the timestamp of docker events
	EventTSAPI = NoOpGauge

	// L7 statistics

	// ProxyRedirects is the number of redirects labelled by protocol
	ProxyRedirects = NoOpGaugeVec

	// ProxyParseErrors is a count of failed parse errors on proxy
	ProxyParseErrors = NoOpCounter

	// ProxyForwarded is a count of all forwarded requests by proxy
	ProxyForwarded = NoOpCounter

	// ProxyDenied is a count of all denied requests by policy by the proxy
	ProxyDenied = NoOpCounter

	// ProxyReceived is a count of all received requests by the proxy
	ProxyReceived = NoOpCounter

	// ProxyUpstreamTime is how long the upstream server took to reply labeled
	// by error, protocol and span time
	ProxyUpstreamTime = NoOpObserverVec

	// L3-L4 statistics

	// DropCount is the total drop requests,
	// tagged by drop reason and direction(ingress/egress)
	DropCount = NoOpCounterVec

	// DropBytes is the total dropped bytes,
	// tagged by drop reason and direction(ingress/egress)
	DropBytes = NoOpCounterVec

	// ForwardCount is the total forwarded packets,
	// tagged by ingress/egress direction
	ForwardCount = NoOpCounterVec

	// ForwardBytes is the total forwarded bytes,
	// tagged by ingress/egress direction
	ForwardBytes = NoOpCounterVec

	// Datapath statistics

	// DatapathErrors is the number of errors managing datapath components
	// such as BPF maps.
	DatapathErrors = NoOpCounterVec

	// ConntrackGCRuns is the number of times that the conntrack GC
	// process was run.
	ConntrackGCRuns = NoOpCounterVec

	// ConntrackGCKeyFallbacks number of times that the conntrack key fallback was invalid.
	ConntrackGCKeyFallbacks = NoOpCounterVec

	// ConntrackGCSize the number of entries in the conntrack table
	ConntrackGCSize = NoOpGaugeVec

	// ConntrackGCDuration the duration of the conntrack GC process in milliseconds.
	ConntrackGCDuration = NoOpObserverVec

	// Services

	// ServicesCount number of services
	ServicesCount = NoOpCounterVec

	// Errors and warnings

	// ErrorsWarnings is the number of errors and warnings in cilium-agent instances
	ErrorsWarnings = NoOpCounterVec

	// ControllerRuns is the number of times that a controller process runs.
	ControllerRuns = NoOpCounterVec

	// ControllerRunsDuration the duration of the controller process in seconds
	ControllerRunsDuration = NoOpObserverVec

	// subprocess, labeled by Subsystem
	SubprocessStart = NoOpCounterVec

	// Kubernetes Events

	// KubernetesEventProcessed is the number of Kubernetes events
	// processed labeled by scope, action and execution result
	KubernetesEventProcessed = NoOpCounterVec

	// KubernetesEventReceived is the number of Kubernetes events received
	// labeled by scope, action, valid data and equalness.
	KubernetesEventReceived = NoOpCounterVec

	// Kubernetes interactions

	// KubernetesAPIInteractions is the total time taken to process an API call made
	// to the kube-apiserver
	KubernetesAPIInteractions = NoOpObserverVec

	// KubernetesAPICalls is the counter for all API calls made to
	// kube-apiserver.
	KubernetesAPICalls = NoOpCounterVec

	// KubernetesCNPStatusCompletion is the number of seconds it takes to
	// complete a CNP status update
	KubernetesCNPStatusCompletion = NoOpObserverVec

	// IPAM events

	// IpamEvent is the number of IPAM events received labeled by action and
	// datapath family type
	IpamEvent = NoOpCounterVec

	// KVstore events

	// KVStoreOperationsDuration records the duration of kvstore operations
	KVStoreOperationsDuration = NoOpObserverVec

	// KVStoreEventsQueueDuration records the duration in seconds of time
	// received event was blocked before it could be queued
	KVStoreEventsQueueDuration = NoOpObserverVec

	// FQDNGarbageCollectorCleanedTotal is the number of domains cleaned by the
	// GC job.
	FQDNGarbageCollectorCleanedTotal = NoOpCounter

	// BPFSyscallDuration is the metric for bpf syscalls duration.
	BPFSyscallDuration = NoOpObserverVec

	// BPFMapOps is the metric to measure the number of operations done to a
	// bpf map.
	BPFMapOps = NoOpCounterVec
)

type Configuration struct {
	APIInteractionsEnabled                  bool
	EndpointCountRegeneratingEnabled        bool
	EndpointRegenerationCountEnabled        bool
	EndpointStateCountEnabled               bool
	EndpointRegenerationTimeStatsEnabled    bool
	PolicyCountEnabled                      bool
	PolicyRegenerationCountEnabled          bool
	PolicyRegenerationTimeStatsEnabled      bool
	PolicyRevisionEnabled                   bool
	PolicyImportErrorsEnabled               bool
	PolicyEndpointStatusEnabled             bool
	PolicyImplementationDelayEnabled        bool
	EventTSK8sEnabled                       bool
	EventTSContainerdEnabled                bool
	EventTSAPIEnabled                       bool
	ProxyRedirectsEnabled                   bool
	ProxyParseErrorsEnabled                 bool
	ProxyForwardedEnabled                   bool
	ProxyDeniedEnabled                      bool
	ProxyReceivedEnabled                    bool
	NoOpObserverVecEnabled                  bool
	DropCountEnabled                        bool
	DropBytesEnabled                        bool
	NoOpCounterVecEnabled                   bool
	ForwardBytesEnabled                     bool
	DatapathErrorsEnabled                   bool
	ConntrackGCRunsEnabled                  bool
	ConntrackGCKeyFallbacksEnabled          bool
	ConntrackGCSizeEnabled                  bool
	ConntrackGCDurationEnabled              bool
	ServicesCountEnabled                    bool
	ErrorsWarningsEnabled                   bool
	ControllerRunsEnabled                   bool
	ControllerRunsDurationEnabled           bool
	SubprocessStartEnabled                  bool
	KubernetesEventProcessedEnabled         bool
	KubernetesEventReceivedEnabled          bool
	KubernetesAPIInteractionsEnabled        bool
	KubernetesAPICallsEnabled               bool
	KubernetesCNPStatusCompletionEnabled    bool
	IpamEventEnabled                        bool
	KVStoreOperationsDurationEnabled        bool
	KVStoreEventsQueueDurationEnabled       bool
	FQDNGarbageCollectorCleanedTotalEnabled bool
	BPFSyscallDurationEnabled               bool
	BPFMapOps                               bool
}

func DefaultMetrics() map[string]struct{} {
	return map[string]struct{}{
		Namespace + "_" + SubsystemAgent + "_api_process_time_seconds":            {},
		Namespace + "_endpoint_regenerating":                                      {},
		Namespace + "_endpoint_regenerations":                                     {},
		Namespace + "_endpoint_state":                                             {},
		Namespace + "_endpoint_regeneration_time_stats_seconds":                   {},
		Namespace + "_policy_count":                                               {},
		Namespace + "_policy_regeneration_total":                                  {},
		Namespace + "_policy_regeneration_time_stats_seconds":                     {},
		Namespace + "_policy_max_revision":                                        {},
		Namespace + "_policy_import_errors":                                       {},
		Namespace + "_policy_endpoint_enforcement_status":                         {},
		Namespace + "_policy_implementation_delay":                                {},
		Namespace + "_event_ts":                                                   {},
		Namespace + "_proxy_redirects":                                            {},
		Namespace + "_policy_l7_parse_errors_total":                               {},
		Namespace + "_policy_l7_forwarded_total":                                  {},
		Namespace + "_policy_l7_denied_total":                                     {},
		Namespace + "_policy_l7_received_total":                                   {},
		Namespace + "_proxy_upstream_reply_seconds":                               {},
		Namespace + "_drop_count_total":                                           {},
		Namespace + "_drop_bytes_total":                                           {},
		Namespace + "_forward_count_total":                                        {},
		Namespace + "_forward_bytes_total":                                        {},
		Namespace + "_" + SubsystemDatapath + "_errors_total":                     {},
		Namespace + "_" + SubsystemDatapath + "_conntrack_gc_runs_total":          {},
		Namespace + "_" + SubsystemDatapath + "_conntrack_gc_key_fallbacks_total": {},
		Namespace + "_" + SubsystemDatapath + "_conntrack_gc_entries":             {},
		Namespace + "_" + SubsystemDatapath + "_conntrack_gc_duration_seconds":    {},
		Namespace + "_services_events_total":                                      {},
		Namespace + "_errors_warnings_total":                                      {},
		Namespace + "_controllers_runs_total":                                     {},
		Namespace + "_controllers_runs_duration_seconds":                          {},
		Namespace + "_subprocess_start_total":                                     {},
		Namespace + "_kubernetes_events_total":                                    {},
		Namespace + "_kubernetes_events_received_total":                           {},
		Namespace + "_" + SubsystemK8sClient + "_api_latency_time_seconds":        {},
		Namespace + "_" + SubsystemK8sClient + "_api_calls_counter":               {},
		Namespace + "_" + SubsystemK8s + "_cnp_status_completion_seconds":         {},
		Namespace + "_ipam_events_total":                                          {},
		Namespace + "_" + SubsystemKVStore + "_operations_duration_seconds":       {},
		Namespace + "_" + SubsystemKVStore + "_events_queue_seconds":              {},
		Namespace + "_fqdn_gc_deletions_total":                                    {},
		Namespace + "_" + SubsystemBPF + "_map_ops_total":                         {},
	}
}

// CreateConfiguration returns a Configuration with all metrics that are
// considered enabled from the given slice of metricsEnabled as well as a slice
// of prometheus.Collectors that must be registered in the prometheus default
// register.
func CreateConfiguration(metricsEnabled []string) (Configuration, []prometheus.Collector) {
	var collectors []prometheus.Collector
	c := Configuration{}

	for _, metricName := range metricsEnabled {
		switch metricName {
		case Namespace + "_" + SubsystemAgent + "_api_process_time_seconds":
			APIInteractions = prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: Namespace,
				Subsystem: SubsystemAgent,
				Name:      "api_process_time_seconds",
				Help:      "Duration of processed API calls labeled by path, method and return code.",
			}, []string{LabelPath, LabelMethod, LabelAPIReturnCode})

			collectors = append(collectors, APIInteractions)
			c.APIInteractionsEnabled = true

		case Namespace + "_endpoint_regenerating":
			EndpointCountRegenerating = prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace: Namespace,
				Name:      "endpoint_regenerating",
				Help:      "Number of endpoints currently regenerating. Deprecated. Use endpoint_state with proper labels instead",
			})

			collectors = append(collectors, EndpointCountRegenerating)
			c.EndpointCountRegeneratingEnabled = true

		case Namespace + "_endpoint_regenerations":
			EndpointRegenerationCount = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "endpoint_regenerations",
				Help:      "Count of all endpoint regenerations that have completed, tagged by outcome",
			}, []string{"outcome"})

			collectors = append(collectors, EndpointRegenerationCount)
			c.EndpointRegenerationCountEnabled = true

		case Namespace + "_endpoint_state":
			EndpointStateCount = prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: Namespace,
					Name:      "endpoint_state",
					Help:      "Count of all endpoints, tagged by different endpoint states",
				},
				[]string{"endpoint_state"},
			)

			collectors = append(collectors, EndpointStateCount)
			c.EndpointStateCountEnabled = true

		case Namespace + "_endpoint_regeneration_time_stats_seconds":
			EndpointRegenerationTimeStats = prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: Namespace,
				Name:      "endpoint_regeneration_time_stats_seconds",
				Help:      "Endpoint regeneration time stats labeled by the scope",
			}, []string{LabelScope, LabelStatus})

			collectors = append(collectors, EndpointRegenerationTimeStats)
			c.EndpointRegenerationTimeStatsEnabled = true

		case Namespace + "_policy_count":
			PolicyCount = prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace: Namespace,
				Name:      "policy_count",
				Help:      "Number of policies currently loaded",
			})

			collectors = append(collectors, PolicyCount)
			c.PolicyCountEnabled = true

		case Namespace + "_policy_regeneration_total":
			PolicyRegenerationCount = prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "policy_regeneration_total",
				Help:      "Total number of successful policy regenerations",
			})

			collectors = append(collectors, PolicyRegenerationCount)
			c.PolicyRegenerationCountEnabled = true

		case Namespace + "_policy_regeneration_time_stats_seconds":
			PolicyRegenerationTimeStats = prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: Namespace,
				Name:      "policy_regeneration_time_stats_seconds",
				Help:      "Policy regeneration time stats labeled by the scope",
			}, []string{LabelScope, LabelStatus})

			collectors = append(collectors, PolicyRegenerationTimeStats)
			c.PolicyRegenerationTimeStatsEnabled = true

		case Namespace + "_policy_max_revision":
			PolicyRevision = prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace: Namespace,
				Name:      "policy_max_revision",
				Help:      "Highest policy revision number in the agent",
			})

			collectors = append(collectors, PolicyRevision)
			c.PolicyRegenerationTimeStatsEnabled = true

		case Namespace + "_policy_import_errors":
			PolicyImportErrors = prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "policy_import_errors",
				Help:      "Number of times a policy import has failed",
			})

			collectors = append(collectors, PolicyImportErrors)
			c.PolicyImportErrorsEnabled = true

		case Namespace + "_policy_endpoint_enforcement_status":
			PolicyEndpointStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: Namespace,
				Name:      "policy_endpoint_enforcement_status",
				Help:      "Number of endpoints labeled by policy enforcement status",
			}, []string{LabelPolicyEnforcement})

			collectors = append(collectors, PolicyEndpointStatus)
			c.PolicyEndpointStatusEnabled = true

		case Namespace + "_policy_implementation_delay":
			PolicyImplementationDelay = prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: Namespace,
				Name:      "policy_implementation_delay",
				Help:      "Time between a policy change and it being fully deployed into the datapath",
			}, []string{LabelPolicySource})

			collectors = append(collectors, PolicyImplementationDelay)
			c.PolicyImplementationDelayEnabled = true

		case Namespace + "_event_ts":
			EventTSK8s = prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace:   Namespace,
				Name:        "event_ts",
				Help:        "Last timestamp when we received an event",
				ConstLabels: prometheus.Labels{"source": LabelEventSourceK8s},
			})

			collectors = append(collectors, EventTSK8s)
			c.EventTSK8sEnabled = true

			EventTSContainerd = prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace:   Namespace,
				Name:        "event_ts",
				Help:        "Last timestamp when we received an event",
				ConstLabels: prometheus.Labels{"source": LabelEventSourceContainerd},
			})

			collectors = append(collectors, EventTSContainerd)
			c.EventTSContainerdEnabled = true

			EventTSAPI = prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace:   Namespace,
				Name:        "event_ts",
				Help:        "Last timestamp when we received an event",
				ConstLabels: prometheus.Labels{"source": LabelEventSourceAPI},
			})

			collectors = append(collectors, EventTSAPI)
			c.EventTSAPIEnabled = true

		case Namespace + "_proxy_redirects":
			ProxyRedirects = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: Namespace,
				Name:      "proxy_redirects",
				Help:      "Number of redirects installed for endpoints, labeled by protocol",
			}, []string{LabelProtocolL7})

			collectors = append(collectors, ProxyRedirects)
			c.ProxyRedirectsEnabled = true

		case Namespace + "_policy_l7_parse_errors_total":
			ProxyParseErrors = prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "policy_l7_parse_errors_total",
				Help:      "Number of total L7 parse errors",
			})

			collectors = append(collectors, ProxyParseErrors)
			c.ProxyParseErrorsEnabled = true

		case Namespace + "_policy_l7_forwarded_total":
			ProxyForwarded = prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "policy_l7_forwarded_total",
				Help:      "Number of total L7 forwarded requests/responses",
			})

			collectors = append(collectors, ProxyForwarded)
			c.ProxyForwardedEnabled = true

		case Namespace + "_policy_l7_denied_total":
			ProxyDenied = prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "policy_l7_denied_total",
				Help:      "Number of total L7 denied requests/responses due to policy",
			})

			collectors = append(collectors, ProxyDenied)
			c.ProxyDeniedEnabled = true

		case Namespace + "_policy_l7_received_total":
			ProxyReceived = prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "policy_l7_received_total",
				Help:      "Number of total L7 received requests/responses",
			})

			collectors = append(collectors, ProxyReceived)
			c.ProxyReceivedEnabled = true

		case Namespace + "_proxy_upstream_reply_seconds":
			ProxyUpstreamTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: Namespace,
				Name:      "proxy_upstream_reply_seconds",
				Help:      "Seconds waited to get a reply from a upstream server",
			}, []string{"error", LabelProtocolL7, LabelScope})

			collectors = append(collectors, ProxyUpstreamTime)
			c.NoOpObserverVecEnabled = true

		case Namespace + "_drop_count_total":
			DropCount = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "drop_count_total",
				Help:      "Total dropped packets, tagged by drop reason and ingress/egress direction",
			},
				[]string{"reason", "direction"})

			collectors = append(collectors, DropCount)
			c.DropCountEnabled = true

		case Namespace + "_drop_bytes_total":
			DropBytes = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "drop_bytes_total",
				Help:      "Total dropped bytes, tagged by drop reason and ingress/egress direction",
			},
				[]string{"reason", "direction"})

			collectors = append(collectors, DropBytes)
			c.DropBytesEnabled = true

		case Namespace + "_forward_count_total":
			ForwardCount = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "forward_count_total",
				Help:      "Total forwarded packets, tagged by ingress/egress direction",
			},
				[]string{"direction"})

			collectors = append(collectors, ForwardCount)
			c.NoOpCounterVecEnabled = true

		case Namespace + "_forward_bytes_total":
			ForwardBytes = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "forward_bytes_total",
				Help:      "Total forwarded bytes, tagged by ingress/egress direction",
			},
				[]string{"direction"})

			collectors = append(collectors, ForwardBytes)
			c.ForwardBytesEnabled = true

		case Namespace + "_" + SubsystemDatapath + "_errors_total":
			DatapathErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: SubsystemDatapath,
				Name:      "errors_total",
				Help:      "Number of errors that occurred in the datapath or datapath management",
			}, []string{LabelDatapathArea, LabelDatapathName, LabelDatapathFamily})

			collectors = append(collectors, DatapathErrors)
			c.DatapathErrorsEnabled = true

		case Namespace + "_" + SubsystemDatapath + "_conntrack_gc_runs_total":
			ConntrackGCRuns = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: SubsystemDatapath,
				Name:      "conntrack_gc_runs_total",
				Help: "Number of times that the conntrack garbage collector process was run " +
					"labeled by completion status",
			}, []string{LabelDatapathFamily, LabelProtocol, LabelStatus})

			collectors = append(collectors, ConntrackGCRuns)
			c.ConntrackGCRunsEnabled = true

		case Namespace + "_" + SubsystemDatapath + "_conntrack_gc_key_fallbacks_total":
			ConntrackGCKeyFallbacks = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: SubsystemDatapath,
				Name:      "conntrack_gc_key_fallbacks_total",
				Help:      "Number of times a key fallback was needed when iterating over the BPF map",
			}, []string{LabelDatapathFamily, LabelProtocol})

			collectors = append(collectors, ConntrackGCKeyFallbacks)
			c.ConntrackGCKeyFallbacksEnabled = true

		case Namespace + "_" + SubsystemDatapath + "_conntrack_gc_entries":
			ConntrackGCSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: Namespace,
				Subsystem: SubsystemDatapath,
				Name:      "conntrack_gc_entries",
				Help: "The number of alive and deleted conntrack entries at the end " +
					"of a garbage collector run labeled by datapath family.",
			}, []string{LabelDatapathFamily, LabelProtocol, LabelStatus})

			collectors = append(collectors, ConntrackGCSize)
			c.ConntrackGCSizeEnabled = true

		case Namespace + "_" + SubsystemDatapath + "_conntrack_gc_duration_seconds":
			ConntrackGCDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: Namespace,
				Subsystem: SubsystemDatapath,
				Name:      "conntrack_gc_duration_seconds",
				Help: "Duration in seconds of the garbage collector process " +
					"labeled by datapath family and completion status",
			}, []string{LabelDatapathFamily, LabelProtocol, LabelStatus})

			collectors = append(collectors, ConntrackGCDuration)
			c.ConntrackGCDurationEnabled = true

		case Namespace + "_services_events_total":
			ServicesCount = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "services_events_total",
				Help:      "Number of services events labeled by action type",
			}, []string{LabelAction})

			collectors = append(collectors, ServicesCount)
			c.ServicesCountEnabled = true

		case Namespace + "_errors_warnings_total":
			ErrorsWarnings = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "errors_warnings_total",
				Help:      "Number of total errors in cilium-agent instances",
			}, []string{"level", "subsystem"})

			collectors = append(collectors, ErrorsWarnings)
			c.ErrorsWarningsEnabled = true

		case Namespace + "_controllers_runs_total":
			ControllerRuns = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "controllers_runs_total",
				Help:      "Number of times that a controller process was run labeled by completion status",
			}, []string{LabelStatus})

			collectors = append(collectors, ControllerRuns)
			c.ControllerRunsEnabled = true

		case Namespace + "_controllers_runs_duration_seconds":
			ControllerRunsDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: Namespace,
				Name:      "controllers_runs_duration_seconds",
				Help:      "Duration in seconds of the controller process labeled by completion status",
			}, []string{LabelStatus})

			collectors = append(collectors, ControllerRunsDuration)
			c.ControllerRunsDurationEnabled = true

		case Namespace + "_subprocess_start_total":
			SubprocessStart = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "subprocess_start_total",
				Help:      "Number of times that Cilium has started a subprocess, labeled by subsystem",
			}, []string{LabelSubsystem})

			collectors = append(collectors, SubprocessStart)
			c.SubprocessStartEnabled = true

		case Namespace + "_kubernetes_events_total":
			KubernetesEventProcessed = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "kubernetes_events_total",
				Help:      "Number of Kubernetes events processed labeled by scope, action and execution result",
			}, []string{LabelScope, LabelAction, LabelStatus})

			collectors = append(collectors, KubernetesEventProcessed)
			c.KubernetesEventProcessedEnabled = true

		case Namespace + "_kubernetes_events_received_total":
			KubernetesEventReceived = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "kubernetes_events_received_total",
				Help:      "Number of Kubernetes events processed labeled by scope, action and execution result",
			}, []string{LabelScope, LabelAction, "valid", "equal"})

			collectors = append(collectors, KubernetesEventReceived)
			c.KubernetesEventReceivedEnabled = true

		case Namespace + "_" + SubsystemK8sClient + "_api_latency_time_seconds":
			KubernetesAPIInteractions = prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: Namespace,
				Subsystem: SubsystemK8sClient,
				Name:      "api_latency_time_seconds",
				Help:      "Duration of processed API calls labeled by path and method.",
			}, []string{LabelPath, LabelMethod})

			collectors = append(collectors, KubernetesAPIInteractions)
			c.KubernetesAPIInteractionsEnabled = true

		case Namespace + "_" + SubsystemK8sClient + "_api_calls_counter":
			KubernetesAPICalls = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: SubsystemK8sClient,
				Name:      "api_calls_counter",
				Help:      "Number of API calls made to kube-apiserver labeled by host, method and return code.",
			}, []string{"host", LabelMethod, LabelAPIReturnCode})

			collectors = append(collectors, KubernetesAPICalls)
			c.KubernetesAPICallsEnabled = true

		case Namespace + "_" + SubsystemK8s + "_cnp_status_completion_seconds":
			KubernetesCNPStatusCompletion = prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: Namespace,
				Subsystem: SubsystemK8s,
				Name:      "cnp_status_completion_seconds",
				Help:      "Duration in seconds in how long it took to complete a CNP status update",
			}, []string{LabelAttempts, LabelOutcome})

			collectors = append(collectors, KubernetesCNPStatusCompletion)
			c.KubernetesCNPStatusCompletionEnabled = true

		case Namespace + "_ipam_events_total":
			IpamEvent = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "ipam_events_total",
				Help:      "Number of IPAM events received labeled by action and datapath family type",
			}, []string{LabelAction, LabelDatapathFamily})

			collectors = append(collectors, IpamEvent)
			c.IpamEventEnabled = true

		case Namespace + "_" + SubsystemKVStore + "_operations_duration_seconds":
			KVStoreOperationsDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: Namespace,
				Subsystem: SubsystemKVStore,
				Name:      "operations_duration_seconds",
				Help:      "Duration in seconds of kvstore operations",
			}, []string{LabelScope, LabelKind, LabelAction, LabelOutcome})

			collectors = append(collectors, KVStoreOperationsDuration)
			c.KVStoreOperationsDurationEnabled = true

		case Namespace + "_" + SubsystemKVStore + "_events_queue_seconds":
			KVStoreEventsQueueDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: Namespace,
				Subsystem: SubsystemKVStore,
				Name:      "events_queue_seconds",
				Help:      "Duration in seconds of time received event was blocked before it could be queued",
				Buckets:   []float64{.002, .005, .01, .015, .025, .05, .1, .25, .5, .75, 1},
			}, []string{LabelScope, LabelAction})

			collectors = append(collectors, KVStoreEventsQueueDuration)
			c.KVStoreEventsQueueDurationEnabled = true

		case Namespace + "_fqdn_gc_deletions_total":
			FQDNGarbageCollectorCleanedTotal = prometheus.NewCounter(prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "fqdn_gc_deletions_total",
				Help:      "Number of FQDNs that have been cleaned on FQDN Garbage collector job",
			})

			collectors = append(collectors, FQDNGarbageCollectorCleanedTotal)
			c.FQDNGarbageCollectorCleanedTotalEnabled = true

		case Namespace + "_" + SubsystemBPF + "_syscall_duration_seconds":
			BPFSyscallDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: Namespace,
				Subsystem: SubsystemBPF,
				Name:      "syscall_duration_seconds",
				Help:      "Duration of BPF system calls",
			}, []string{LabelOperation, LabelOutcome})

			collectors = append(collectors, BPFSyscallDuration)
			c.BPFSyscallDurationEnabled = true

		case Namespace + "_" + SubsystemBPF + "_map_ops_total":
			BPFMapOps = prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: SubsystemBPF,
				Name:      "map_ops_total",
				Help:      "Total operations on map, tagged by map name",
			}, []string{LabelMapName, LabelOperation, LabelOutcome})

			collectors = append(collectors, BPFMapOps)
			c.BPFMapOps = true
		}
	}

	return c, collectors
}

func init() {
	MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{Namespace: Namespace}))
	// TODO: Figure out how to put this into a Namespace
	// MustRegister(prometheus.NewGoCollector())
}

// MustRegister adds the collector to the registry, exposing this metric to
// prometheus scrapes.
// It will panic on error.
func MustRegister(c ...prometheus.Collector) {
	registry.MustRegister(c...)
}

// Register registers a collector
func Register(c prometheus.Collector) error {
	return registry.Register(c)
}

// RegisterList registers a list of collectors. If registration of one
// collector fails, no collector is registered.
func RegisterList(list []prometheus.Collector) error {
	registered := []prometheus.Collector{}

	for _, c := range list {
		if err := Register(c); err != nil {
			for _, c := range registered {
				Unregister(c)
			}
			return err
		}

		registered = append(registered, c)
	}

	return nil
}

// Unregister unregisters a collector
func Unregister(c prometheus.Collector) bool {
	return registry.Unregister(c)
}

// Enable begins serving prometheus metrics on the address passed in. Addresses
// of the form ":8080" will bind the port on all interfaces.
func Enable(addr string) <-chan error {
	errs := make(chan error, 1)

	go func() {
		// The Handler function provides a default handler to expose metrics
		// via an HTTP server. "/metrics" is the usual endpoint for that.
		http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
		errs <- http.ListenAndServe(addr, nil)
	}()

	return errs
}

// GetCounterValue returns the current value
// stored for the counter
func GetCounterValue(m prometheus.Counter) float64 {
	var pm dto.Metric
	err := m.Write(&pm)
	if err == nil {
		return *pm.Counter.Value
	}
	return 0
}

// DumpMetrics gets the current Cilium metrics and dumps all into a
// models.Metrics structure.If metrics cannot be retrieved, returns an error
func DumpMetrics() ([]*models.Metric, error) {
	result := []*models.Metric{}
	currentMetrics, err := registry.Gather()
	if err != nil {
		return result, err
	}

	for _, val := range currentMetrics {

		metricName := val.GetName()
		metricType := val.GetType()

		for _, metricLabel := range val.Metric {
			labels := map[string]string{}
			for _, label := range metricLabel.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}

			var value float64
			switch metricType {
			case dto.MetricType_COUNTER:
				value = metricLabel.Counter.GetValue()
			case dto.MetricType_GAUGE:
				value = metricLabel.GetGauge().GetValue()
			case dto.MetricType_UNTYPED:
				value = metricLabel.GetUntyped().GetValue()
			case dto.MetricType_SUMMARY:
				value = metricLabel.GetSummary().GetSampleSum()
			case dto.MetricType_HISTOGRAM:
				value = metricLabel.GetHistogram().GetSampleSum()
			default:
				continue
			}

			metric := &models.Metric{
				Name:   metricName,
				Labels: labels,
				Value:  value,
			}
			result = append(result, metric)
		}
	}
	return result, nil
}

// Error2Outcome converts an error to LabelOutcome
func Error2Outcome(err error) string {
	if err != nil {
		return LabelValueOutcomeFail
	}

	return LabelValueOutcomeSuccess
}

// Errno2Outcome converts a syscall.Errno to LabelOutcome
func Errno2Outcome(errno syscall.Errno) string {
	if errno != 0 {
		return LabelValueOutcomeFail
	}

	return LabelValueOutcomeSuccess
}
