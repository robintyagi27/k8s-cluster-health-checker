package health

import (
	"context"
	"log"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus metrics
var (
	NodeReadyGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "k8s_node_ready_total",
		Help: "Number of ready Kubernetes nodes",
	})

	PodRunningGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "k8s_pod_running_total",
		Help: "Number of running pods across all namespaces",
	})

	PodFailedGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "k8s_pod_failed_total",
		Help: "Number of failed pods across all namespaces",
	})
)

// Register metrics
func init() {
	prometheus.MustRegister(NodeReadyGauge)
	prometheus.MustRegister(PodRunningGauge)
	prometheus.MustRegister(PodFailedGauge)
}

// Struct to return cluster stats
type ClusterHealth struct {
	NodeCount   int
	ReadyNodes  int
	TotalPods   int
	RunningPods int
	FailedPods  int
}

// Start Prometheus metrics server
func StartMetricsServer(port string) {
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Printf("Starting metrics server on :%s/metrics ...", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()
}

// CheckClusterHealth checks the cluster’s pod & node status
func CheckClusterHealth(clientset *kubernetes.Clientset) (*ClusterHealth, error) {
	ctx := context.Background()
	stats := &ClusterHealth{}

	// --- Node status ---
	nodes, err := clientset.CoreV1().Nodes().List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	stats.NodeCount = len(nodes.Items)
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				stats.ReadyNodes++
			}
		}
	}

	// --- Pod status ---
	pods, err := clientset.CoreV1().Pods("").List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	stats.TotalPods = len(pods.Items)
	for _, pod := range pods.Items {
		switch pod.Status.Phase {
		case "Running":
			stats.RunningPods++
		case "Failed":
			stats.FailedPods++
		}
	}

	return stats, nil
}
