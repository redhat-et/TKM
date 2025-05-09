package agent

import (
	"context"
	"log"

	"github.com/redhat-et/TKDK/cargohold/pkg/accelerator"
	"github.com/redhat-et/TKM/internal/tkm-agent/crd"
	"github.com/redhat-et/TKM/internal/tkm-agent/node"
	"github.com/redhat-et/TKM/internal/tkm-agent/preflight-check"
)

func Start(ctx context.Context) error {
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
	node.UpdateNodeStatus(accs)

	// Start monitoring CRD updates
	go crd.MonitorCacheCRD(accs)

	log.Println("Agent started successfully.")
	<-ctx.Done()
	log.Println("tkm-agent stopped.")
	return nil
}
