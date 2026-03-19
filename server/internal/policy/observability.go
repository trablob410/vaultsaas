package policy

import "sync"

type ResolutionObservation struct {
	Source       string
	Status       string
	WarningCount int
	WarningCodes []string
}

var (
	obsMu       sync.RWMutex
	lastResolve ResolutionObservation
	counters    = map[string]int64{}
)

// ObserveResolution records policy resolution fields for logs/metrics hooks.
func ObserveResolution(source, status string, warningCodes []string) {
	obsMu.Lock()
	defer obsMu.Unlock()
	counters["policy_resolution_total|source="+source+"|status="+status]++
	copyCodes := make([]string, len(warningCodes))
	copy(copyCodes, warningCodes)
	lastResolve = ResolutionObservation{
		Source:       source,
		Status:       status,
		WarningCount: len(copyCodes),
		WarningCodes: copyCodes,
	}
}

// ObserveValidationFailure increments validation failure counter by stage.
func ObserveValidationFailure(stage string) {
	obsMu.Lock()
	defer obsMu.Unlock()
	counters["policy_validation_fail_total|stage="+stage]++
}

// LastResolutionObservation exposes latest structured fields for callers/tests.
func LastResolutionObservation() ResolutionObservation {
	obsMu.RLock()
	defer obsMu.RUnlock()
	copyCodes := make([]string, len(lastResolve.WarningCodes))
	copy(copyCodes, lastResolve.WarningCodes)
	res := lastResolve
	res.WarningCodes = copyCodes
	return res
}

// SnapshotCounters returns a safe copy for tests/integration hooks.
func SnapshotCounters() map[string]int64 {
	obsMu.RLock()
	defer obsMu.RUnlock()
	res := make(map[string]int64, len(counters))
	for key, val := range counters {
		res[key] = val
	}
	return res
}
