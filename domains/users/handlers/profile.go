package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"securities-marketplace/domains/users"
	"securities-marketplace/domains/shared/events"
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

	_, err := h.userService.GetUser(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// TODO: Implement template rendering with user data
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<h1>User Profile</h1>"))
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

	user, err := h.userService.GetUser(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profileData)
}

// HandleAPIUpdateProfile handles JSON API profile updates
func (h *ProfileHandler) HandleAPIUpdateProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var cmd users.UpdateUserProfileCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd.UserID = userID
	cmd.UpdatedBy = getCurrentUserID(r)

	if err := h.processUpdateProfile(&cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Profile updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleAPISuspendUser handles user suspension (admin only)
func (h *ProfileHandler) HandleAPISuspendUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var cmd users.SuspendUserCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd.UserID = userID
	cmd.SuspendedBy = getCurrentUserID(r)

	if err := h.processSuspendUser(&cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "User suspended successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleAPIReinstateUser handles user reinstatement (admin only)
func (h *ProfileHandler) HandleAPIReinstateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	var cmd users.ReinstateUserCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd.UserID = userID
	cmd.ReinstatedBy = getCurrentUserID(r)

	if err := h.processReinstateUser(&cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "User reinstated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// processUpdateProfile handles the core profile update logic
func (h *ProfileHandler) processUpdateProfile(cmd *users.UpdateUserProfileCommand) error {
	return h.userService.UpdateUserProfile(cmd)
}

// processSuspendUser handles the core user suspension logic
func (h *ProfileHandler) processSuspendUser(cmd *users.SuspendUserCommand) error {
	return h.userService.SuspendUser(cmd)
}

// processReinstateUser handles the core user reinstatement logic
func (h *ProfileHandler) processReinstateUser(cmd *users.ReinstateUserCommand) error {
	return h.userService.ReinstateUser(cmd)
}

// getCurrentUserID extracts the current user ID from the request context
// This would typically come from JWT token or session
func getCurrentUserID(r *http.Request) string {
	// TODO: Implement actual user ID extraction from JWT/session
	return "system" // Placeholder
}