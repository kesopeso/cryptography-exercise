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
	createStatusFn  func(ctx context.Context, status *bitset.Bitset) (uuid.UUID, error)
	getStatusIdsFn  func(ctx context.Context) ([]uuid.UUID, error)
}

func (m *mockStatusStore) CreateStatus(ctx context.Context, status *bitset.Bitset) (uuid.UUID, error) {
	return m.createStatusFn(ctx, status)
}

func (m *mockStatusStore) GetStatusIds(ctx context.Context) ([]uuid.UUID, error) {
	return m.getStatusIdsFn(ctx)
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

func TestListStatuses(t *testing.T) {
	id1 := uuid.MustParse("01961234-5678-7abc-8def-0123456789ab")
	id2 := uuid.MustParse("01961234-5678-7abc-8def-0123456789ac")

	tests := []struct {
		name           string
		store          *mockStatusStore
		wantStatusCode int
		wantStatusIds  []string
	}{
		{
			name: "returns multiple status ids",
			store: &mockStatusStore{
				getStatusIdsFn: func(_ context.Context) ([]uuid.UUID, error) {
					return []uuid.UUID{id1, id2}, nil
				},
			},
			wantStatusCode: http.StatusOK,
			wantStatusIds:  []string{id1.String(), id2.String()},
		},
		{
			name: "returns empty list",
			store: &mockStatusStore{
				getStatusIdsFn: func(_ context.Context) ([]uuid.UUID, error) {
					return []uuid.UUID{}, nil
				},
			},
			wantStatusCode: http.StatusOK,
			wantStatusIds:  []string{},
		},
		{
			name: "store error",
			store: &mockStatusStore{
				getStatusIdsFn: func(_ context.Context) ([]uuid.UUID, error) {
					return nil, errors.New("db error")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newStatusHandlers(tt.store)

			req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
			rec := httptest.NewRecorder()

			h.listStatuses(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusIds != nil {
				var resp map[string][]string
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if len(resp["statusIds"]) != len(tt.wantStatusIds) {
					t.Fatalf("got %d statusIds, want %d", len(resp["statusIds"]), len(tt.wantStatusIds))
				}
				for i, id := range resp["statusIds"] {
					if id != tt.wantStatusIds[i] {
						t.Errorf("statusIds[%d] = %q, want %q", i, id, tt.wantStatusIds[i])
					}
				}
			}
		})
	}
}
