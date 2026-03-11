package server

import (
	"encoding/json"
	"net/http"

	"github.com/kesopeso/cryptography-exercise/internal/bitset"
	"github.com/kesopeso/cryptography-exercise/internal/store"
)

// statusHandlers holds dependencies for the status API route handlers.
type statusHandlers struct {
	statusStore store.StatusStore
}

// newStatusHandlers creates a statusHandlers with the given StatusStore.
func newStatusHandlers(statusStore store.StatusStore) *statusHandlers {
	return &statusHandlers{statusStore: statusStore}
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

	bs := bitset.NewBitset()
	for _, v := range body.Status {
		bs.Add(v)
	}

	id, err := h.statusStore.CreateStatus(r.Context(), bs)
	if err != nil {
		http.Error(w, "failed to create status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"statusId": id.String()})
}

// GET /api/status
// Retrieve all structures (list of statusIds).
func (h *statusHandlers) listStatuses(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// POST /api/status/{statusId}
// Create a new state in the structure and return the index.
func (h *statusHandlers) createState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// GET /api/status/{statusId}/{index}
// Return a JWS compact signed message with the status payload.
func (h *statusHandlers) getState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// PUT /api/status/{statusId}/{index}
// Set the state at index to true.
func (h *statusHandlers) setState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// DELETE /api/status/{statusId}/{index}
// Set the state at index to false.
func (h *statusHandlers) deleteState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
