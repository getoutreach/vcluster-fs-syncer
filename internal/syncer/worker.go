// Copyright 2023 Outreach Corporation. All Rights Reserved.

// Description: Implements the syncer worker logic for client-go.

package syncer

// processNextWorkItem processes a work item handling
// retry logic, queueing and rate limiting.
func (s *Syncer) processNextWorkItem() bool {
	einf, quit := s.queue.Get()
	if quit {
		return false
	}
	defer s.queue.Done(einf)

	e, ok := einf.(*event)
	if !ok {
		s.queue.Forget(einf)
		return true
	}

	// Invoke the method containing the business logic
	err := s.reconcile(e)
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		s.queue.Forget(einf)
		return true
	}

	if s.queue.NumRequeues(einf) < 10 {
		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		s.queue.AddRateLimited(einf)
		return true
	}

	// Retries exceeded. Forgetting for this reconciliation loop
	s.queue.Forget(einf)
	return true
}

func (s *Syncer) runWorker() {
	for s.processNextWorkItem() {
	}
}
