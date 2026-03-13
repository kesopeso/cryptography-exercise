package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kesopeso/cryptography-exercise/internal/service"
	"github.com/kesopeso/cryptography-exercise/pkg/cryptography"
)

// statusHandlers holds dependencies for the status API route handlers.
type statusHandlers struct {
	statusService  service.StatusService
	signJWSKeyPath string
}

// newStatusHandlers creates a statusHandlers with the given StatusService and signing key path.
func newStatusHandlers(statusService service.StatusService, signJWSKeyPath string) *statusHandlers {
	return &statusHandlers{statusService: statusService, signJWSKeyPath: signJWSKeyPath}
}

// POST /api/status
// Create a new structure and return the statusId.
// Expects JSON body: { "status": [true, false, ...] }
func (h *statusHandlers) createStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status []bool `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	id, err := h.statusService.CreateStatus(r.Context(), body.Status)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create status, error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"statusId": id.String()})
}

// GET /api/status
// Retrieve all structures (list of statusIds).
func (h *statusHandlers) getStatusIds(w http.ResponseWriter, r *http.Request) {
	ids, err := h.statusService.GetStatusIds(r.Context())
	if err != nil {
		http.Error(w, "failed to list statuses", http.StatusInternalServerError)
		return
	}

	statusIds := make([]string, len(ids))
	for i, id := range ids {
		statusIds[i] = id.String()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"statusIds": statusIds})
}

// POST /api/status/{statusId}
// Create a new state in the structure and return the index.
// Expects JSON body: { "value": true/false }
func (h *statusHandlers) createStatusValue(w http.ResponseWriter, r *http.Request) {
	statusId := chi.URLParam(r, "statusId")

	var body struct {
		Value bool `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	valueIndex, err := h.statusService.CreateStatusValue(r.Context(), statusId, body.Value)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create status value, %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"valueIndex": valueIndex})
}

// GET /api/status/{statusId}/{index}
// Return a JWS compact signed message with the status payload.
func (h *statusHandlers) getStatusValue(w http.ResponseWriter, r *http.Request) {
	statusId := chi.URLParam(r, "statusId")

	index, err := strconv.Atoi(chi.URLParam(r, "index"))
	if err != nil || index < 0 {
		http.Error(w, "index must be a non-negative integer", http.StatusBadRequest)
		return
	}

	encodedList, err := h.statusService.GetEncodedStatus(r.Context(), statusId)
	if err != nil {
		http.Error(w, "failed to get status", http.StatusNotFound)
		return
	}

	issuer := "http://" + r.Host + "/api/status/" + statusId
	token, err := cryptography.SignStatusJWS(h.signJWSKeyPath, issuer, encodedList, index)
	if err != nil {
		http.Error(w, "failed to sign token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"jws": token})
}

// PUT /api/status/{statusId}/{index}
// Set the state at index to true.
func (h *statusHandlers) updateStatusValueToTrue(w http.ResponseWriter, r *http.Request) {
	statusId := chi.URLParam(r, "statusId")

	index, err := strconv.Atoi(chi.URLParam(r, "index"))
	if err != nil || index < 0 {
		http.Error(w, "index must be a non-negative integer", http.StatusBadRequest)
		return
	}

	if err := h.statusService.UpdateStatusValue(r.Context(), statusId, index, true); err != nil {
		http.Error(w, "failed to update status value", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/status/{statusId}/{index}
// Set the state at index to false.
func (h *statusHandlers) updateStatusValueToFalse(w http.ResponseWriter, r *http.Request) {
	statusId := chi.URLParam(r, "statusId")

	index, err := strconv.Atoi(chi.URLParam(r, "index"))
	if err != nil || index < 0 {
		http.Error(w, "index must be a non-negative integer", http.StatusBadRequest)
		return
	}

	if err := h.statusService.UpdateStatusValue(r.Context(), statusId, index, false); err != nil {
		http.Error(w, "failed to update status value", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
