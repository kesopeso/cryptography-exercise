package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kesopeso/cryptography-exercise/pkg/cryptography"
)

type mockStatusService struct {
	createStatusFn      func(ctx context.Context, status []bool) (uuid.UUID, error)
	getStatusIdsFn      func(ctx context.Context) ([]uuid.UUID, error)
	getEncodedStatusFn  func(ctx context.Context, statusId string) (string, error)
	createStatusValueFn func(ctx context.Context, statusId string, value bool) (int, error)
	updateStatusValueFn func(ctx context.Context, statusId string, index int, value bool) error
	verifyValueExistsFn func() (bool, error)
}

func (m *mockStatusService) CreateStatus(ctx context.Context, status []bool) (uuid.UUID, error) {
	return m.createStatusFn(ctx, status)
}

func (m *mockStatusService) GetStatusIds(ctx context.Context) ([]uuid.UUID, error) {
	return m.getStatusIdsFn(ctx)
}

func (m *mockStatusService) GetEncodedStatus(ctx context.Context, statusId string) (string, error) {
	return m.getEncodedStatusFn(ctx, statusId)
}

func (m *mockStatusService) CreateStatusValue(ctx context.Context, statusId string, value bool) (int, error) {
	return m.createStatusValueFn(ctx, statusId, value)
}

func (m *mockStatusService) UpdateStatusValue(ctx context.Context, statusId string, index int, value bool) error {
	return m.updateStatusValueFn(ctx, statusId, index, value)
}

func (m *mockStatusService) VerifyValueExists(index int, encodedStatus string) (bool, error) {
	return m.verifyValueExistsFn()
}

func TestCreateStatus(t *testing.T) {
	fixedID := uuid.MustParse("01961234-5678-7abc-8def-0123456789ab")

	tests := []struct {
		name           string
		body           string
		service        *mockStatusService
		wantStatusCode int
		wantStatusId   string
	}{
		{
			name: "valid request with statuses",
			body: `{"status": [true, false, true]}`,
			service: &mockStatusService{
				createStatusFn: func(_ context.Context, _ []bool) (uuid.UUID, error) {
					return fixedID, nil
				},
			},
			wantStatusCode: http.StatusCreated,
			wantStatusId:   fixedID.String(),
		},
		{
			name: "valid request with empty statuses",
			body: `{"status": []}`,
			service: &mockStatusService{
				createStatusFn: func(_ context.Context, _ []bool) (uuid.UUID, error) {
					return fixedID, nil
				},
			},
			wantStatusCode: http.StatusCreated,
			wantStatusId:   fixedID.String(),
		},
		{
			name:           "invalid json body",
			body:           `not json`,
			service:        &mockStatusService{},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: `{"status": [true]}`,
			service: &mockStatusService{
				createStatusFn: func(_ context.Context, _ []bool) (uuid.UUID, error) {
					return uuid.UUID{}, errors.New("db error")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newStatusHandlers(tt.service, "")

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
		service        *mockStatusService
		wantStatusCode int
		wantStatusIds  []string
	}{
		{
			name: "returns multiple status ids",
			service: &mockStatusService{
				getStatusIdsFn: func(_ context.Context) ([]uuid.UUID, error) {
					return []uuid.UUID{id1, id2}, nil
				},
			},
			wantStatusCode: http.StatusOK,
			wantStatusIds:  []string{id1.String(), id2.String()},
		},
		{
			name: "returns empty list",
			service: &mockStatusService{
				getStatusIdsFn: func(_ context.Context) ([]uuid.UUID, error) {
					return []uuid.UUID{}, nil
				},
			},
			wantStatusCode: http.StatusOK,
			wantStatusIds:  []string{},
		},
		{
			name: "service error",
			service: &mockStatusService{
				getStatusIdsFn: func(_ context.Context) ([]uuid.UUID, error) {
					return nil, errors.New("db error")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newStatusHandlers(tt.service, "")

			req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
			rec := httptest.NewRecorder()

			h.getStatusIds(rec, req)

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

// withChiURLParams adds chi URL parameters to the request context.
// Accepts alternating key-value pairs.
func withChiURLParams(r *http.Request, kvs ...string) *http.Request {
	rctx := chi.NewRouteContext()
	for i := 0; i < len(kvs)-1; i += 2 {
		rctx.URLParams.Add(kvs[i], kvs[i+1])
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestCreateStatusValue(t *testing.T) {
	tests := []struct {
		name           string
		statusId       string
		body           string
		service        *mockStatusService
		wantStatusCode int
		wantIndex      int
	}{
		{
			name:     "valid request with true value",
			statusId: "01961234-5678-7abc-8def-0123456789ab",
			body:     `{"value": true}`,
			service: &mockStatusService{
				createStatusValueFn: func(_ context.Context, statusId string, value bool) (int, error) {
					if statusId != "01961234-5678-7abc-8def-0123456789ab" {
						t.Errorf("got statusId %q, want %q", statusId, "01961234-5678-7abc-8def-0123456789ab")
					}
					if !value {
						t.Error("got value false, want true")
					}
					return 5, nil
				},
			},
			wantStatusCode: http.StatusCreated,
			wantIndex:      5,
		},
		{
			name:     "valid request with false value",
			statusId: "01961234-5678-7abc-8def-0123456789ab",
			body:     `{"value": false}`,
			service: &mockStatusService{
				createStatusValueFn: func(_ context.Context, _ string, value bool) (int, error) {
					if value {
						t.Error("got value true, want false")
					}
					return 0, nil
				},
			},
			wantStatusCode: http.StatusCreated,
			wantIndex:      0,
		},
		{
			name:           "invalid json body",
			statusId:       "01961234-5678-7abc-8def-0123456789ab",
			body:           `not json`,
			service:        &mockStatusService{},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:     "service error",
			statusId: "01961234-5678-7abc-8def-0123456789ab",
			body:     `{"value": true}`,
			service: &mockStatusService{
				createStatusValueFn: func(_ context.Context, _ string, _ bool) (int, error) {
					return -1, errors.New("db error")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newStatusHandlers(tt.service, "")

			req := httptest.NewRequest(http.MethodPost, "/api/status/"+tt.statusId, strings.NewReader(tt.body))
			req = withChiURLParams(req, "statusId", tt.statusId)
			rec := httptest.NewRecorder()

			h.createStatusValue(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusCreated {
				var resp map[string]int
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp["valueIndex"] != tt.wantIndex {
					t.Errorf("got valueIndex %d, want %d", resp["valueIndex"], tt.wantIndex)
				}
			}
		})
	}
}

func TestUpdateStatusValueToFalse(t *testing.T) {
	statusId := "01961234-5678-7abc-8def-0123456789ab"

	tests := []struct {
		name           string
		statusId       string
		index          string
		service        *mockStatusService
		wantStatusCode int
	}{
		{
			name:     "valid request",
			statusId: statusId,
			index:    "3",
			service: &mockStatusService{
				updateStatusValueFn: func(_ context.Context, gotStatusId string, gotIndex int, gotValue bool) error {
					if gotStatusId != statusId {
						t.Errorf("got statusId %q, want %q", gotStatusId, statusId)
					}
					if gotIndex != 3 {
						t.Errorf("got index %d, want 3", gotIndex)
					}
					if gotValue {
						t.Error("got value true, want false")
					}
					return nil
				},
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name:     "index zero is valid",
			statusId: statusId,
			index:    "0",
			service: &mockStatusService{
				updateStatusValueFn: func(_ context.Context, _ string, gotIndex int, _ bool) error {
					if gotIndex != 0 {
						t.Errorf("got index %d, want 0", gotIndex)
					}
					return nil
				},
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name:           "negative index",
			statusId:       statusId,
			index:          "-1",
			service:        &mockStatusService{},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "non-numeric index",
			statusId:       statusId,
			index:          "abc",
			service:        &mockStatusService{},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:     "service error",
			statusId: statusId,
			index:    "0",
			service: &mockStatusService{
				updateStatusValueFn: func(_ context.Context, _ string, _ int, _ bool) error {
					return errors.New("db error")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newStatusHandlers(tt.service, "")

			req := httptest.NewRequest(http.MethodDelete, "/api/status/"+tt.statusId+"/"+tt.index, nil)
			req = withChiURLParams(req, "statusId", tt.statusId, "index", tt.index)
			rec := httptest.NewRecorder()

			h.updateStatusValueToFalse(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestUpdateStatusValueToTrue(t *testing.T) {
	statusId := "01961234-5678-7abc-8def-0123456789ab"

	tests := []struct {
		name           string
		statusId       string
		index          string
		service        *mockStatusService
		wantStatusCode int
	}{
		{
			name:     "valid request",
			statusId: statusId,
			index:    "3",
			service: &mockStatusService{
				updateStatusValueFn: func(_ context.Context, gotStatusId string, gotIndex int, gotValue bool) error {
					if gotStatusId != statusId {
						t.Errorf("got statusId %q, want %q", gotStatusId, statusId)
					}
					if gotIndex != 3 {
						t.Errorf("got index %d, want 3", gotIndex)
					}
					if !gotValue {
						t.Error("got value false, want true")
					}
					return nil
				},
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name:     "index zero is valid",
			statusId: statusId,
			index:    "0",
			service: &mockStatusService{
				updateStatusValueFn: func(_ context.Context, _ string, _ int, _ bool) error {
					return nil
				},
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name:           "negative index",
			statusId:       statusId,
			index:          "-1",
			service:        &mockStatusService{},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "non-numeric index",
			statusId:       statusId,
			index:          "abc",
			service:        &mockStatusService{},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:     "service error",
			statusId: statusId,
			index:    "0",
			service: &mockStatusService{
				updateStatusValueFn: func(_ context.Context, _ string, _ int, _ bool) error {
					return errors.New("db error")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newStatusHandlers(tt.service, "")

			req := httptest.NewRequest(http.MethodPut, "/api/status/"+tt.statusId+"/"+tt.index, nil)
			req = withChiURLParams(req, "statusId", tt.statusId, "index", tt.index)
			rec := httptest.NewRecorder()

			h.updateStatusValueToTrue(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestGetStatusValue(t *testing.T) {
	statusId := "01961234-5678-7abc-8def-0123456789ab"

	validKeyPath := func(t *testing.T) string {
		t.Helper()
		path := t.TempDir() + "/key.pem"
		if err := cryptography.GenerateAndSaveECDSAKey(path); err != nil {
			t.Fatalf("failed to generate key: %v", err)
		}
		return path
	}

	tests := []struct {
		name           string
		statusId       string
		index          string
		service        *mockStatusService
		keyPath        string
		wantStatusCode int
		wantJWS        bool
	}{
		{
			name:     "valid request",
			statusId: statusId,
			index:    "2",
			service: &mockStatusService{
				getEncodedStatusFn: func(_ context.Context, gotStatusId string) (string, error) {
					if gotStatusId != statusId {
						t.Errorf("got statusId %q, want %q", gotStatusId, statusId)
					}
					return "encodedListValue", nil
				},
				verifyValueExistsFn: func() (bool, error) {
					return true, nil
				},
			},
			wantStatusCode: http.StatusOK,
			wantJWS:        true,
		},
		{
			name:           "non-numeric index",
			statusId:       statusId,
			index:          "abc",
			service:        &mockStatusService{},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:     "status not found",
			statusId: statusId,
			index:    "0",
			service: &mockStatusService{
				getEncodedStatusFn: func(_ context.Context, _ string) (string, error) {
					return "", errors.New("db error")
				},
			},
			keyPath:        "",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:     "value check error",
			statusId: statusId,
			index:    "0",
			service: &mockStatusService{
				getEncodedStatusFn: func(_ context.Context, _ string) (string, error) {
					return "encodedListValue", nil
				},
				verifyValueExistsFn: func() (bool, error) {
					return false, errors.New("can't determine if value exists")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:     "no such value",
			statusId: statusId,
			index:    "3",
			service: &mockStatusService{
				getEncodedStatusFn: func(_ context.Context, _ string) (string, error) {
					return "encodedListValue", nil
				},
				verifyValueExistsFn: func() (bool, error) {
					return false, nil
				},
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:     "sign error",
			statusId: statusId,
			index:    "0",
			service: &mockStatusService{
				getEncodedStatusFn: func(_ context.Context, _ string) (string, error) {
					return "encodedListValue", nil
				},
				verifyValueExistsFn: func() (bool, error) {
					return true, nil
				},
			},
			keyPath:        "/nonexistent/key.pem",
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyPath := tt.keyPath
			if tt.wantJWS {
				keyPath = validKeyPath(t)
			}

			h := newStatusHandlers(tt.service, keyPath)

			req := httptest.NewRequest(http.MethodGet, "/api/status/"+tt.statusId+"/"+tt.index, nil)
			req = withChiURLParams(req, "statusId", tt.statusId, "index", tt.index)
			rec := httptest.NewRecorder()

			h.getStatusValue(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatusCode)
			}

			if tt.wantJWS {
				var resp map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp["jws"] == "" {
					t.Error("expected non-empty jws in response")
				}
			}
		})
	}
}
