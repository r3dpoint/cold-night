package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"securities-marketplace/domains/users"
	"securities-marketplace/domains/shared/events"
	"securities-marketplace/domains/shared/web"
)

// ProfileHandler handles user profile requests
type ProfileHandler struct {
	userService *users.UserService
	eventBus    events.EventBus
}

// NewProfileHandler creates a new profile handler
func NewProfileHandler(userService *users.UserService, eventBus events.EventBus) *ProfileHandler {
	return &ProfileHandler{
		userService: userService,
		eventBus:    eventBus,
	}
}

// RegisterRoutes registers the handler routes
func (h *ProfileHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/users/{userId}/profile", h.ShowProfile).Methods("GET")
	router.HandleFunc("/users/{userId}/profile", h.UpdateProfile).Methods("POST")
	router.HandleFunc("/api/users/{userId}/profile", h.HandleAPIGetProfile).Methods("GET")
	router.HandleFunc("/api/users/{userId}/profile", h.HandleAPIUpdateProfile).Methods("PUT")
	router.HandleFunc("/api/users/{userId}/suspend", h.HandleAPISuspendUser).Methods("POST")
	router.HandleFunc("/api/users/{userId}/reinstate", h.HandleAPIReinstateUser).Methods("POST")
}

// ShowProfile displays the user profile page
func (h *ProfileHandler) ShowProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	user, err := h.userService.GetByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"Title": "User Profile",
		"User":  user,
	}

	web.RenderTemplate(w, "users/profile.html", data)
}

// UpdateProfile handles form-based profile updates
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	updatedFields := make(map[string]interface{})
	
	if firstName := r.FormValue("firstName"); firstName != "" {
		updatedFields["firstName"] = firstName
	}
	if lastName := r.FormValue("lastName"); lastName != "" {
		updatedFields["lastName"] = lastName
	}
	if email := r.FormValue("email"); email != "" {
		updatedFields["email"] = email
	}

	cmd := &users.UpdateUserProfileCommand{
		UserID:        userID,
		UpdatedFields: updatedFields,
		UpdatedBy:     getCurrentUserID(r), // Get from session/JWT
	}

	if err := h.processUpdateProfile(cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Redirect back to profile
	http.Redirect(w, r, "/users/"+userID+"/profile?updated=true", http.StatusSeeOther)
}

// HandleAPIGetProfile handles JSON API profile retrieval
func (h *ProfileHandler) HandleAPIGetProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	user, err := h.userService.GetByID(userID)
	if err != nil {
		web.WriteJSONError(w, "User not found", http.StatusNotFound)
		return
	}

	// Sanitize sensitive information
	profileData := map[string]interface{}{
		"userId":    user.ID,
		"email":     user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"status":    user.Status,
		"createdAt": user.CreatedAt,
		"accreditation": map[string]interface{}{
			"status":      user.Accreditation.Status,
			"type":        user.Accreditation.Type,
			"validUntil":  user.Accreditation.ValidUntil,
			"verifiedAt":  user.Accreditation.VerifiedAt,
		},
		"compliance": map[string]interface{}{
			"overallStatus":   user.Compliance.OverallStatus,
			"kycStatus":       user.Compliance.KYCStatus,
			"amlStatus":       user.Compliance.AMLStatus,
			"sanctionsStatus": user.Compliance.SanctionsStatus,
			"riskScore":       user.Compliance.RiskScore,
			"lastCheck":       user.Compliance.LastCheck,
			"nextReview":      user.Compliance.NextReview,
		},
		"capabilities": map[string]interface{}{
			"canTrade":      user.CanTrade(),
			"isAccredited":  user.IsAccredited(),
			"isCompliant":   user.IsCompliant(),
		},
	}

	web.WriteJSON(w, profileData, http.StatusOK)
}

// HandleAPIUpdateProfile handles JSON API profile updates
func (h *ProfileHandler) HandleAPIUpdateProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var cmd users.UpdateUserProfileCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		web.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd.UserID = userID
	cmd.UpdatedBy = getCurrentUserID(r)

	if err := h.processUpdateProfile(&cmd); err != nil {
		web.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Profile updated successfully",
	}

	web.WriteJSON(w, response, http.StatusOK)
}

// HandleAPISuspendUser handles user suspension (admin only)
func (h *ProfileHandler) HandleAPISuspendUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var cmd users.SuspendUserCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		web.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd.UserID = userID
	cmd.SuspendedBy = getCurrentUserID(r)

	if err := h.processSuspendUser(&cmd); err != nil {
		web.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "User suspended successfully",
	}

	web.WriteJSON(w, response, http.StatusOK)
}

// HandleAPIReinstateUser handles user reinstatement (admin only)
func (h *ProfileHandler) HandleAPIReinstateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var cmd users.ReinstateUserCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		web.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd.UserID = userID
	cmd.ReinstatedBy = getCurrentUserID(r)

	if err := h.processReinstateUser(&cmd); err != nil {
		web.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "User reinstated successfully",
	}

	web.WriteJSON(w, response, http.StatusOK)
}

// processUpdateProfile handles the core profile update logic
func (h *ProfileHandler) processUpdateProfile(cmd *users.UpdateUserProfileCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	user, err := h.userService.GetByID(cmd.UserID)
	if err != nil {
		return err
	}

	if err := user.UpdateProfile(cmd.UpdatedFields, cmd.UpdatedBy); err != nil {
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

// processSuspendUser handles the core user suspension logic
func (h *ProfileHandler) processSuspendUser(cmd *users.SuspendUserCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	user, err := h.userService.GetByID(cmd.UserID)
	if err != nil {
		return err
	}

	if err := user.Suspend(cmd.Reason, cmd.SuspendedBy, cmd.Duration); err != nil {
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

// processReinstateUser handles the core user reinstatement logic
func (h *ProfileHandler) processReinstateUser(cmd *users.ReinstateUserCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	user, err := h.userService.GetByID(cmd.UserID)
	if err != nil {
		return err
	}

	if err := user.Reinstate(cmd.ReinstatedBy, cmd.Reason); err != nil {
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

// getCurrentUserID extracts the current user ID from the request context
// This would typically come from JWT token or session
func getCurrentUserID(r *http.Request) string {
	// TODO: Implement actual user ID extraction from JWT/session
	return "system" // Placeholder
}