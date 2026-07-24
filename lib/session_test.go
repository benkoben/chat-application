package lib_test

import (
	"chat-application/lib"
	"testing"
)

func TestGenerateSessionId(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		want    *lib.Session
		wantErr bool
	}{
		{
			name:    "generate session id",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := lib.GenerateSessionId()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GenerateSessionId() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GenerateSessionId() succeeded unexpectedly")
			}
			if got == nil {
				t.Fatal("GenerateSessionId() returned nil Session")
			}

			if got.Id == nil {
				t.Errorf("GenerateSessionId().Id is all zeros, want non-empty [32]byte")
			}

			if len(got.Id) != 32 {
				t.Errorf("GenerateSessionId().Id has length %d, want 32", len(got.Id))
			}

			// Verify Id is actually non-zero (contains random data)
			allZero := true
			for _, b := range got.Id {
				if b != 0 {
					allZero = false
					break
				}
			}
			if allZero {
				t.Error("GenerateSessionId().Id is all zeros, expected random data")
			}
		})
	}
}
