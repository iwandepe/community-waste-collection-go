package worker

import (
	"context"
	"log"
	"time"

	"github.com/iwandp/community-waste-collection-go/internal/domain"
)

// OrganicCancelWorker cancels organic pickups that have been pending for more than 3 days.
// It runs on a ticker and shuts down cleanly when ctx is cancelled.
type OrganicCancelWorker struct {
	repo     domain.PickupRepository
	interval time.Duration
}

func NewOrganicCancelWorker(repo domain.PickupRepository, interval time.Duration) *OrganicCancelWorker {
	return &OrganicCancelWorker{repo: repo, interval: interval}
}

func (w *OrganicCancelWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Println("[worker] organic cancel worker started")
	for {
		select {
		case <-ctx.Done():
			log.Println("[worker] organic cancel worker stopped")
			return
		case <-ticker.C:
			w.run()
		}
	}
}

func (w *OrganicCancelWorker) run() {
	cutoff := time.Now().Add(-3 * 24 * time.Hour)
	n, err := w.repo.CancelOrganicExpired(cutoff)
	if err != nil {
		log.Printf("[worker] organic cancel error: %v", err)
		return
	}
	if n > 0 {
		log.Printf("[worker] canceled %d expired organic pickup(s)", n)
	}
}
