package web

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// Server represents the web server
type Server struct {
	router      *mux.Router
	templates   *template.Template
	sessionStore sessions.Store
	port        string
	
	// Services
	userService     UserService
	securityService SecurityService
	listingService  ListingService
	biddingService  BiddingService
	executionService ExecutionService
}

// UserService interface for user operations
type UserService interface {
	GetUser(userID string) (*User, error)
	AuthenticateUser(email, password string) (*User, error)
	RegisterUser(email, username, fullName, password string) error
}

// SecurityService interface for security operations
type SecurityService interface {
	GetSecurity(securityID string) (*Security, error)
	GetSecurities() ([]*Security, error)
	GetUserSecurities(userID string) ([]*Security, error)
}

// ListingService interface for listing operations
type ListingService interface {
	GetListing(listingID string) (*Listing, error)
	GetActiveListings() ([]*Listing, error)
	GetListingsBySecurity(securityID string) ([]*Listing, error)
	GetListingsByUser(userID string) ([]*Listing, error)
}

// BiddingService interface for bidding operations
type BiddingService interface {
	GetBid(bidID string) (*Bid, error)
	GetActiveBids() ([]*Bid, error)
	GetBidsBySecurity(securityID string) ([]*Bid, error)
	GetBidsByUser(userID string) ([]*Bid, error)
}

// ExecutionService interface for execution operations
type ExecutionService interface {
	GetTrade(tradeID string) (*Trade, error)
	GetTradesByUser(userID string) ([]*Trade, error)
	GetTradesBySecurity(securityID string) ([]*Trade, error)
	GetMarketStatistics(securityID string, period time.Duration) (*MarketStats, error)
}

// NewServer creates a new web server
func NewServer(port string, sessionSecret string) *Server {
	server := &Server{
		router:       mux.NewRouter(),
		sessionStore: sessions.NewCookieStore([]byte(sessionSecret)),
		port:         port,
	}
	
	server.loadTemplates()
	server.setupRoutes()
	server.setupMiddleware()
	
	return server
}

// SetServices sets the domain services
func (s *Server) SetServices(
	userService UserService,
	securityService SecurityService,
	listingService ListingService,
	biddingService BiddingService,
	executionService ExecutionService,
) {
	s.userService = userService
	s.securityService = securityService
	s.listingService = listingService
	s.biddingService = biddingService
	s.executionService = executionService
}

// loadTemplates loads HTML templates
func (s *Server) loadTemplates() {
	templatePattern := "templates/**/*.html"
	templates, err := template.ParseGlob(templatePattern)
	if err != nil {
		log.Printf("Warning: Failed to load templates: %v", err)
		// Create empty template to prevent panics
		s.templates = template.New("empty")
		return
	}
	
	s.templates = templates
}

// setupMiddleware sets up common middleware
func (s *Server) setupMiddleware() {
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.securityHeadersMiddleware)
	s.router.Use(s.sessionMiddleware)
}

// setupRoutes sets up HTTP routes
func (s *Server) setupRoutes() {
	// Static files
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))
	
	// Public routes
	s.router.HandleFunc("/", s.handleHome).Methods("GET")
	s.router.HandleFunc("/login", s.handleLogin).Methods("GET", "POST")
	s.router.HandleFunc("/register", s.handleRegister).Methods("GET", "POST")
	s.router.HandleFunc("/logout", s.handleLogout).Methods("POST")
	
	// Protected routes (require authentication)
	protected := s.router.PathPrefix("/app").Subrouter()
	protected.Use(s.authMiddleware)
	
	// Dashboard
	protected.HandleFunc("/dashboard", s.handleDashboard).Methods("GET")
	
	// Securities
	protected.HandleFunc("/securities", s.handleSecurities).Methods("GET")
	protected.HandleFunc("/securities/{id}", s.handleSecurityDetail).Methods("GET")
	
	// Trading
	protected.HandleFunc("/trading", s.handleTrading).Methods("GET")
	protected.HandleFunc("/trading/listings", s.handleListings).Methods("GET")
	protected.HandleFunc("/trading/listings/{id}", s.handleListingDetail).Methods("GET")
	protected.HandleFunc("/trading/bids", s.handleBids).Methods("GET")
	protected.HandleFunc("/trading/bids/{id}", s.handleBidDetail).Methods("GET")
	
	// Portfolio
	protected.HandleFunc("/portfolio", s.handlePortfolio).Methods("GET")
	protected.HandleFunc("/portfolio/trades", s.handleTrades).Methods("GET")
	protected.HandleFunc("/portfolio/trades/{id}", s.handleTradeDetail).Methods("GET")
	
	// Admin routes (require admin role)
	admin := protected.PathPrefix("/admin").Subrouter()
	admin.Use(s.adminMiddleware)
	admin.HandleFunc("/", s.handleAdminDashboard).Methods("GET")
	admin.HandleFunc("/users", s.handleAdminUsers).Methods("GET")
	admin.HandleFunc("/compliance", s.handleAdminCompliance).Methods("GET")
	admin.HandleFunc("/audit", s.handleAdminAudit).Methods("GET")
	
	// API routes for AJAX/JSON
	api := s.router.PathPrefix("/api").Subrouter()
	api.Use(s.authMiddleware)
	api.HandleFunc("/market-data/{securityId}", s.handleAPIMarketData).Methods("GET")
	api.HandleFunc("/portfolio/summary", s.handleAPIPortfolioSummary).Methods("GET")
}

// Start starts the web server
func (s *Server) Start() error {
	log.Printf("Starting web server on port %s", s.port)
	return http.ListenAndServe(":"+s.port, s.router)
}

// StartWithContext starts the web server with context for graceful shutdown
func (s *Server) StartWithContext(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":" + s.port,
		Handler: s.router,
	}
	
	// Start server in goroutine
	go func() {
		log.Printf("Starting web server on port %s", s.port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()
	
	// Wait for context cancellation
	<-ctx.Done()
	
	// Graceful shutdown
	log.Println("Shutting down web server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	return server.Shutdown(shutdownCtx)
}

// renderTemplate renders an HTML template with data
func (s *Server) renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	err := s.templates.ExecuteTemplate(w, templateName, data)
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// renderJSON renders JSON response
func (s *Server) renderJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	
	// Simple JSON encoding - in production use proper JSON library
	fmt.Fprintf(w, `{"data": %v}`, data)
}

// getSession gets the user session
func (s *Server) getSession(r *http.Request) (*sessions.Session, error) {
	return s.sessionStore.Get(r, "session")
}

// getCurrentUser gets the current authenticated user from session
func (s *Server) getCurrentUser(r *http.Request) (*User, error) {
	session, err := s.getSession(r)
	if err != nil {
		return nil, err
	}
	
	userID, ok := session.Values["user_id"].(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("no authenticated user")
	}
	
	return s.userService.GetUser(userID)
}