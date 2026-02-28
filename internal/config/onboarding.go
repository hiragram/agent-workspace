package config

import (
	"fmt"
	"os"
)

// EnsureOnboardingState creates the onboarding state file with {} if it
// doesn't exist or is empty.
func (s *DefaultSyncer) EnsureOnboardingState(path string) error {
	info, err := os.Stat(path)
	if err == nil && info.Size() > 0 {
		// File exists and is non-empty, leave it alone
		return nil
	}

	if err := os.WriteFile(path, []byte("{}\n"), 0644); err != nil {
		return fmt.Errorf("creating onboarding state: %w", err)
	}
	return nil
}
