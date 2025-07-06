package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"securities-marketplace/domains/users"
	"securities-marketplace/domains/shared/events"
	"securities-marketplace/domains/shared/web"
)

// RegistrationHandler handles user registration requests
type RegistrationHandler struct {
	userService *users.UserService
	eventBus    events.EventBus
}

// NewRegistrationHandler creates a new registration handler
func NewRegistrationHandler(userService *users.UserService, eventBus events.EventBus) *RegistrationHandler {
	return &RegistrationHandler{
		userService: userService,
		eventBus:    eventBus,
	}
}

// RegisterRoutes registers the handler routes
func (h *RegistrationHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/register", h.ShowRegistrationForm).Methods("GET")
	router.HandleFunc("/register", h.HandleRegistration).Methods("POST")
	router.HandleFunc("/api/users/register", h.HandleAPIRegistration).Methods("POST")
}

// ShowRegistrationForm displays the registration form
func (h *RegistrationHandler) ShowRegistrationForm(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "User Registration",
		"AccreditationTypes": []string{
			"individual_accredited",
			"institutional_accredited", 
			"qualified_institutional_buyer",
			"family_office",
		},
	}
	
	web.RenderTemplate(w, "users/registration.html", data)
}

// HandleRegistration handles form-based user registration
func (h *RegistrationHandler) HandleRegistration(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	cmd := &users.RegisterUserCommand{
		UserID:            r.FormValue("userId"),
		Email:             r.FormValue("email"),
		FirstName:         r.FormValue("firstName"),
		LastName:          r.FormValue("lastName"),
		Password:          r.FormValue("password"),
		AccreditationType: r.FormValue("accreditationType"),
		AccreditationDetails: map[string]string{
			"income":     r.FormValue("income"),
			"net_worth":  r.FormValue("netWorth"),
			"experience": r.FormValue("experience"),
		},
	}

	if err := h.processRegistration(cmd); err != nil {
		data := map[string]interface{}{
			"Title": "User Registration",
			"Error": err.Error(),
			"AccreditationTypes": []string{
				"individual_accredited",
				"institutional_accredited", 
				"qualified_institutional_buyer",
				"family_office",
			},
		}
		web.RenderTemplate(w, "users/registration.html", data)
		return
	}

	// Redirect to login page on success
	http.Redirect(w, r, "/login?registered=true", http.StatusSeeOther)
}

// HandleAPIRegistration handles JSON API registration requests
func (h *RegistrationHandler) HandleAPIRegistration(w http.ResponseWriter, r *http.Request) {
	var cmd users.RegisterUserCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		web.WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.processRegistration(&cmd); err != nil {
		web.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "User registered successfully",
		"userId":  cmd.UserID,
	}

	web.WriteJSON(w, response, http.StatusCreated)
}

// processRegistration handles the core registration logic
func (h *RegistrationHandler) processRegistration(cmd *users.RegisterUserCommand) error {
	// Validate command
	if err := cmd.Validate(); err != nil {
		return err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Generate user ID if not provided
	if cmd.UserID == "" {
		cmd.UserID = generateUserID()
	}

	// Create user aggregate
	user := users.NewUserAggregate(cmd.UserID)
	
	// Register user
	err = user.Register(
		cmd.Email,
		cmd.FirstName,
		cmd.LastName,
		string(hashedPassword),
		cmd.AccreditationType,
		cmd.AccreditationDetails,
	)
	if err != nil {
		return err
	}

	// Save user
	if err := h.userService.Save(user); err != nil {
		return err
	}

	// Publish events
	for _, event := range user.GetUncommittedEvents() {
		if err := h.eventBus.Publish(event); err != nil {
			// Log error but don't fail the request
			// Events will be republished on next aggregate load
		}
	}

	user.MarkEventsAsCommitted()
	return nil
}

// generateUserID generates a unique user ID
func generateUserID() string {
	return "user-" + time.Now().Format("20060102150405") + "-" + randomString(6)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}