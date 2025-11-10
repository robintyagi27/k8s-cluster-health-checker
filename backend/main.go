package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// ClusterHealth stores basic cluster info
type ClusterHealth struct {
	NodeCount   int
	ReadyNodes  int
	TotalPods   int
	RunningPods int
	FailedPods  int
}

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

	log.Println("✅ Connected to Kubernetes cluster successfully")

	for {
		stats, err := checkClusterHealth(clientset)
		if err != nil {
			log.Printf("⚠️ Error checking cluster: %v", err)
		} else {
			log.Printf("[Cluster Health] Nodes Ready: %d/%d | Pods Running: %d/%d | Failed: %d",
				stats.ReadyNodes, stats.NodeCount, stats.RunningPods, stats.TotalPods, stats.FailedPods)
		}
		time.Sleep(30 * time.Second)
	}
}

func checkClusterHealth(clientset *kubernetes.Clientset) (*ClusterHealth, error) {
	ctx := context.Background()

	// ✅ Use correct context + ListOptions syntax
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	stats := &ClusterHealth{}
	stats.NodeCount = len(nodes.Items)

	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				stats.ReadyNodes++
			}
		}
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
