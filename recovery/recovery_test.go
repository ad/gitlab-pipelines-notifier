package recovery

import (
	"testing"
)

func TestRecovery(t *testing.T) {
	tests := []struct {
		name     string
		testCase string
		result   bool
	}{
		{"success", "a", true},
		{"fail", "c", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			func() {
				defer Recovery()

				if !tt.result {
					panic("panic")
				}
			}()
		})
	}
}
