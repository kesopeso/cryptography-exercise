package server

import (
	"net/http"
)

// GET /api/status
// Retrieve all structures (list of statusIds).
func listStatuses(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// POST /api/status
// Create a new structure and return the statusId.
func createStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// POST /api/status/{statusId}
// Create a new state in the structure and return the index.
func createState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// GET /api/status/{statusId}/{index}
// Return a JWS compact signed message with the status payload.
func getState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// PUT /api/status/{statusId}/{index}
// Set the state at index to true.
func setState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// DELETE /api/status/{statusId}/{index}
// Set the state at index to false.
func deleteState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
