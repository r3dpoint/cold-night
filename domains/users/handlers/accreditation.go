package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"securities-marketplace/domains/users"
	"securities-marketplace/domains/shared/events"
)

// AccreditationHandler handles accreditation-related requests
type AccreditationHandler struct {
	userService *users.UserService
	eventBus    events.EventBus
}

// NewAccreditationHandler creates a new accreditation handler
func NewAccreditationHandler(userService *users.UserService, eventBus events.EventBus) *AccreditationHandler {
	return &AccreditationHandler{
		userService: userService,
		eventBus:    eventBus,
	}
}

// RegisterRoutes registers the handler routes
func (h *AccreditationHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/users/{userId}/accreditation", h.ShowAccreditationForm).Methods("GET")
	router.HandleFunc("/users/{userId}/accreditation", h.SubmitAccreditation).Methods("POST")
	router.HandleFunc("/api/users/{userId}/accreditation/submit", h.HandleAPISubmitAccreditation).Methods("POST")
	router.HandleFunc("/api/users/{userId}/accreditation/verify", h.HandleAPIVerifyAccreditation).Methods("POST")
	router.HandleFunc("/api/users/{userId}/accreditation/revoke", h.HandleAPIRevokeAccreditation).Methods("POST")
}

// ShowAccreditationForm displays the accreditation submission form
func (h *AccreditationHandler) ShowAccreditationForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	_, err := h.userService.GetUser(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// TODO: Implement template rendering with user data
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<h1>Accreditation Form</h1>"))
}

// SubmitAccreditation handles form-based accreditation submission
func (h *AccreditationHandler) SubmitAccreditation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Process uploaded documents
	documents := []users.DocumentInfo{}
	if files := r.MultipartForm.File["documents"]; files != nil {
		for _, fileHeader := range files {
			// In a real implementation, you would:
			// 1. Validate file type and size
			// 2. Scan for malware
			// 3. Store in secure storage (S3, etc.)
			// 4. Generate document hash for integrity
			
			doc := users.DocumentInfo{
				Name: fileHeader.Filename,
				Type: r.FormValue("documentType"),
				Size: fileHeader.Size,
				Hash: "placeholder-hash", // Calculate actual hash
				UploadedAt: time.Now(),
			}
			documents = append(documents, doc)
		}
	}

	cmd := &users.SubmitAccreditationCommand{
		UserID:            userID,
		AccreditationType: r.FormValue("accreditationType"),
		Documents:         documents,
		SubmissionDetails: map[string]string{
			"notes":           r.FormValue("notes"),
			"income_range":    r.FormValue("incomeRange"),
			"net_worth_range": r.FormValue("netWorthRange"),
		},
	}

	if err := h.processSubmitAccreditation(cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Redirect to profile page
	http.Redirect(w, r, "/users/"+userID+"/profile?submitted=true", http.StatusSeeOther)
}

// HandleAPISubmitAccreditation handles JSON API accreditation submission
func (h *AccreditationHandler) HandleAPISubmitAccreditation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var cmd users.SubmitAccreditationCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd.UserID = userID

	if err := h.processSubmitAccreditation(&cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Accreditation submitted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleAPIVerifyAccreditation handles accreditation verification (admin only)
func (h *AccreditationHandler) HandleAPIVerifyAccreditation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var cmd users.VerifyAccreditationCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd.UserID = userID

	if err := h.processVerifyAccreditation(&cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Accreditation verified successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleAPIRevokeAccreditation handles accreditation revocation (admin only)
func (h *AccreditationHandler) HandleAPIRevokeAccreditation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var cmd users.RevokeAccreditationCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd.UserID = userID

	if err := h.processRevokeAccreditation(&cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Accreditation revoked successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// processSubmitAccreditation handles the core accreditation submission logic
func (h *AccreditationHandler) processSubmitAccreditation(cmd *users.SubmitAccreditationCommand) error {
	return h.userService.SubmitAccreditation(cmd)
}

// processVerifyAccreditation handles the core accreditation verification logic
func (h *AccreditationHandler) processVerifyAccreditation(cmd *users.VerifyAccreditationCommand) error {
	return h.userService.VerifyAccreditation(cmd)
}

// processRevokeAccreditation handles the core accreditation revocation logic
func (h *AccreditationHandler) processRevokeAccreditation(cmd *users.RevokeAccreditationCommand) error {
	return h.userService.RevokeAccreditation(cmd)
}