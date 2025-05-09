package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"log"
	"context"

	"github.com/redhat-et/TKDK/cargohold/pkg/accelerator"
	"github.com/redhat-et/TKDK/cargohold/pkg/preflightcheck"
	"github.com/redhat-et/TKDK/cargohold/pkg/config"
)

func main() {
	// Create a new context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal catching
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Starting tkm-agent...")

	// Start the agent functionality
	go func() {
		if err := startAgent(ctx); err != nil {
			log.Fatalf("Failed to start agent: %v", err)
		}
	}()

	// Wait for a termination signal
	sig := <-sigChan
	log.Printf("Received signal %v, shutting down...", sig)
	cancel()
}

func startAgent(ctx context.Context) error {
	log.Println("Initializing accelerator registry...")
	registry := accelerator.GetRegistry()

	accs := registry.GetAccelerators()
	if len(accs) == 0 {
		log.Println("No accelerators found on this node.")
		return nil
	}

	for accType, acc := range accs {
		log.Printf("Found accelerator: %s - Running: %v", accType, acc.IsRunning())
	}

	// Update node status with accelerator info
	updateNodeStatus(accs)

	// Start monitoring CRD updates
	go monitorCacheCRD(accs)

	log.Println("Agent started successfully.")
	<-ctx.Done()
	log.Println("tkm-agent stopped.")
	return nil
}

func updateNodeStatus(accs map[string]accelerator.Accelerator) {
	log.Println("Updating node status with accelerator info...")
	for accType, acc := range accs {
		log.Printf("Node status update - Accelerator: %s, Running: %v", accType, acc.IsRunning())
	}
}

func monitorCacheCRD(accs map[string]accelerator.Accelerator) {
	log.Println("Monitoring cache CRD updates...")
	for {
		// Simulate watching for CRD update
		crdVerified := checkCRDVerified()
		if crdVerified {
			log.Println("CRD marked as Verified. Running preflight checks...")
			if err := runPreflightChecks(accs); err != nil {
				log.Printf("Preflight check failed: %v", err)
			} else {
				log.Println("Preflight check passed.")
			}
		} else {
			log.Println("Waiting for CRD verification...")
		}
		// Sleep to avoid busy-waiting
		select {
		case <-time.After(10 * time.Second):
		case <-context.Background().Done():
			return
		}
	}
}

func checkCRDVerified() bool {
	// Placeholder logic to check CRD status
	return true // Simulate CRD being marked as Verified
}

func runPreflightChecks(accs map[string]accelerator.Accelerator) error {
	log.Println("Performing preflight checks...")
	for accType, acc := range accs {
		if err := preflightcheck.CompareTritonCacheImageToGPU(nil, acc); err != nil {
			return fmt.Errorf("accelerator %s is not compatible: %v", accType, err)
		}
	}
	return nil
}
