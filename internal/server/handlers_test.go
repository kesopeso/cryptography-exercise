package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kesopeso/cryptography-exercise/internal/bitset"
)

type mockStatusStore struct {
	createStatusFn func(ctx context.Context, status *bitset.Bitset) (uuid.UUID, error)
}

func (m *mockStatusStore) CreateStatus(ctx context.Context, status *bitset.Bitset) (uuid.UUID, error) {
	return m.createStatusFn(ctx, status)
}

func TestCreateStatus(t *testing.T) {
	fixedID := uuid.MustParse("01961234-5678-7abc-8def-0123456789ab")

	tests := []struct {
		name           string
		body           string
		store          *mockStatusStore
		wantStatusCode int
		wantStatusId   string
	}{
		{
			name: "valid request with statuses",
			body: `{"status": [true, false, true]}`,
			store: &mockStatusStore{
				createStatusFn: func(_ context.Context, _ *bitset.Bitset) (uuid.UUID, error) {
					return fixedID, nil
				},
			},
			wantStatusCode: http.StatusCreated,
			wantStatusId:   fixedID.String(),
		},
		{
			name: "valid request with empty statuses",
			body: `{"status": []}`,
			store: &mockStatusStore{
				createStatusFn: func(_ context.Context, _ *bitset.Bitset) (uuid.UUID, error) {
					return fixedID, nil
				},
			},
			wantStatusCode: http.StatusCreated,
			wantStatusId:   fixedID.String(),
		},
		{
			name:           "invalid json body",
			body:           `not json`,
			store:          &mockStatusStore{},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "store error",
			body: `{"status": [true]}`,
			store: &mockStatusStore{
				createStatusFn: func(_ context.Context, _ *bitset.Bitset) (uuid.UUID, error) {
					return uuid.UUID{}, errors.New("db error")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newStatusHandlers(tt.store)

			req := httptest.NewRequest(http.MethodPost, "/api/status", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			h.createStatus(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusId != "" {
				var resp map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp["statusId"] != tt.wantStatusId {
					t.Errorf("got statusId %q, want %q", resp["statusId"], tt.wantStatusId)
				}
			}
		})
	}
}
