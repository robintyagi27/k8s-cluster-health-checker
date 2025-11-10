package health

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"
)

var (
	// Simulated CPU utilization percentage
	CPUUtilizationGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "k8s_cpu_utilization_percent",
		Help: "Simulated cluster CPU utilization percentage (0-100%)",
	})

	// Desired replica count (as if HPA decided)
	DesiredReplicasGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "k8s_desired_replicas_total",
		Help: "Desired number of replicas simulated by autoscaler",
	})
)

// Register metrics once
func init() {
	prometheus.MustRegister(CPUUtilizationGauge)
	prometheus.MustRegister(DesiredReplicasGauge)
}

var (
	currentReplicas = 3 // start baseline
	mu              sync.Mutex
)

// SimulateNodeScaling mimics HorizontalPodAutoscaler decisions
func SimulateNodeScaling(clientset *kubernetes.Clientset) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		mu.Lock()

		// Simulate random CPU load 20-100%
		cpuLoad := rand.Float64()*80 + 20
		CPUUtilizationGauge.Set(cpuLoad)

		// Fake scaling logic
		if cpuLoad > 75 && currentReplicas < 10 {
			currentReplicas++
			log.Printf("⚡ HPA triggered scale-up → %d replicas (CPU %.1f%%)", currentReplicas, cpuLoad)
		} else if cpuLoad < 35 && currentReplicas > 2 {
			currentReplicas--
			log.Printf("🌀 HPA triggered scale-down → %d replicas (CPU %.1f%%)", currentReplicas, cpuLoad)
		}

		DesiredReplicasGauge.Set(float64(currentReplicas))
		mu.Unlock()
	}
}
