package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/robintyagi27/k8s-cluster-health-checker/backend/health"
)

func main() {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = filepath.Join(os.Getenv("USERPROFILE"), ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("❌ Error building kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("❌ Error creating Kubernetes clientset: %v", err)
	}

	// Start Prometheus exporter
	health.StartMetricsServer("8081")

	for {
		stats, err := health.CheckClusterHealth(clientset)
		if err != nil {
			log.Printf("⚠️ Health check error: %v", err)
			time.Sleep(15 * time.Second)
			continue
		}

		// Update Prometheus gauges
		health.NodeReadyGauge.Set(float64(stats.ReadyNodes))
		health.PodRunningGauge.Set(float64(stats.RunningPods))
		health.PodFailedGauge.Set(float64(stats.FailedPods))

		log.Printf("[Cluster Health] Nodes Ready: %d/%d | Pods Running: %d/%d | Failed: %d",
			stats.ReadyNodes, stats.NodeCount, stats.RunningPods, stats.TotalPods, stats.FailedPods)

		time.Sleep(30 * time.Second)
	}
}
