package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"securities-marketplace/domains/users"
	"securities-marketplace/domains/shared/events"
	"securities-marketplace/domains/shared/web"
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

	user, err := h.userService.GetByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"Title":        "Accreditation Submission",
		"User":         user,
		"AccreditationTypes": []string{
			"individual_accredited",
			"institutional_accredited", 
			"qualified_institutional_buyer",
			"family_office",
		},
		"DocumentTypes": []string{
			"tax_return",
			"bank_statement",
			"investment_statement",
			"cpa_letter",
			"employer_verification",
		},
	}

	web.RenderTemplate(w, "users/accreditation.html", data)
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
		web.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd.UserID = userID

	if err := h.processSubmitAccreditation(&cmd); err != nil {
		web.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Accreditation submitted successfully",
	}

	web.WriteJSON(w, response, http.StatusOK)
}

// HandleAPIVerifyAccreditation handles accreditation verification (admin only)
func (h *AccreditationHandler) HandleAPIVerifyAccreditation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var cmd users.VerifyAccreditationCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		web.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd.UserID = userID

	if err := h.processVerifyAccreditation(&cmd); err != nil {
		web.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Accreditation verified successfully",
	}

	web.WriteJSON(w, response, http.StatusOK)
}

// HandleAPIRevokeAccreditation handles accreditation revocation (admin only)
func (h *AccreditationHandler) HandleAPIRevokeAccreditation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var cmd users.RevokeAccreditationCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		web.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd.UserID = userID

	if err := h.processRevokeAccreditation(&cmd); err != nil {
		web.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Accreditation revoked successfully",
	}

	web.WriteJSON(w, response, http.StatusOK)
}

// processSubmitAccreditation handles the core accreditation submission logic
func (h *AccreditationHandler) processSubmitAccreditation(cmd *users.SubmitAccreditationCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	user, err := h.userService.GetByID(cmd.UserID)
	if err != nil {
		return err
	}

	if err := user.SubmitAccreditation(cmd.AccreditationType, cmd.Documents, cmd.SubmissionDetails); err != nil {
		return err
	}

	if err := h.userService.Save(user); err != nil {
		return err
	}

	// Publish events
	for _, event := range user.GetUncommittedEvents() {
		h.eventBus.Publish(event)
	}

	user.MarkEventsAsCommitted()
	return nil
}

// processVerifyAccreditation handles the core accreditation verification logic
func (h *AccreditationHandler) processVerifyAccreditation(cmd *users.VerifyAccreditationCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	user, err := h.userService.GetByID(cmd.UserID)
	if err != nil {
		return err
	}

	if err := user.VerifyAccreditation(cmd.AccreditationType, cmd.ValidUntil, cmd.VerifiedBy, cmd.VerificationNotes); err != nil {
		return err
	}

	if err := h.userService.Save(user); err != nil {
		return err
	}

	// Publish events
	for _, event := range user.GetUncommittedEvents() {
		h.eventBus.Publish(event)
	}

	user.MarkEventsAsCommitted()
	return nil
}

// processRevokeAccreditation handles the core accreditation revocation logic
func (h *AccreditationHandler) processRevokeAccreditation(cmd *users.RevokeAccreditationCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	user, err := h.userService.GetByID(cmd.UserID)
	if err != nil {
		return err
	}

	if err := user.RevokeAccreditation(cmd.Reason, cmd.RevokedBy); err != nil {
		return err
	}

	if err := h.userService.Save(user); err != nil {
		return err
	}

	// Publish events
	for _, event := range user.GetUncommittedEvents() {
		h.eventBus.Publish(event)
	}

	user.MarkEventsAsCommitted()
	return nil
}