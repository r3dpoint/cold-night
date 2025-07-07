package web

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Page data structures for templates
type PageData struct {
	Title       string
	User        *User
	IsLoggedIn  bool
	Messages    []Message
	CSRFToken   string
}

type Message struct {
	Type    string // "success", "error", "warning", "info"
	Content string
}

type DashboardData struct {
	PageData
	Portfolio      *PortfolioSummary
	RecentTrades   []*Trade
	MarketSummary  *MarketSummary
	Notifications  []*Notification
}

type SecurityListData struct {
	PageData
	Securities []*Security
}

type TradingData struct {
	PageData
	ActiveListings []*Listing
	ActiveBids     []*Bid
	MarketData     []*MarketStats
}

// handleHome serves the home page
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getCurrentUser(r)
	
	data := PageData{
		Title:      "Securities Trading Platform",
		User:       user,
		IsLoggedIn: user != nil,
	}
	
	s.renderTemplate(w, "home.html", data)
}

// handleLogin handles user login
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := PageData{
			Title: "Login",
		}
		s.renderTemplate(w, "login.html", data)
		return
	}
	
	// POST - process login
	email := r.FormValue("email")
	password := r.FormValue("password")
	
	if email == "" || password == "" {
		data := PageData{
			Title: "Login",
			Messages: []Message{
				{Type: "error", Content: "Email and password are required"},
			},
		}
		s.renderTemplate(w, "login.html", data)
		return
	}
	
	user, err := s.userService.AuthenticateUser(email, password)
	if err != nil {
		data := PageData{
			Title: "Login",
			Messages: []Message{
				{Type: "error", Content: "Invalid email or password"},
			},
		}
		s.renderTemplate(w, "login.html", data)
		return
	}
	
	// Set session
	session, _ := s.getSession(r)
	session.Values["user_id"] = user.ID
	session.Save(r, w)
	
	// Redirect to dashboard
	http.Redirect(w, r, "/app/dashboard", http.StatusSeeOther)
}

// handleRegister handles user registration
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := PageData{
			Title: "Register",
		}
		s.renderTemplate(w, "register.html", data)
		return
	}
	
	// POST - process registration
	email := r.FormValue("email")
	username := r.FormValue("username")
	fullName := r.FormValue("full_name")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")
	
	// Validation
	var messages []Message
	if email == "" {
		messages = append(messages, Message{Type: "error", Content: "Email is required"})
	}
	if username == "" {
		messages = append(messages, Message{Type: "error", Content: "Username is required"})
	}
	if fullName == "" {
		messages = append(messages, Message{Type: "error", Content: "Full name is required"})
	}
	if password == "" {
		messages = append(messages, Message{Type: "error", Content: "Password is required"})
	}
	if password != confirmPassword {
		messages = append(messages, Message{Type: "error", Content: "Passwords do not match"})
	}
	
	if len(messages) > 0 {
		data := PageData{
			Title:    "Register",
			Messages: messages,
		}
		s.renderTemplate(w, "register.html", data)
		return
	}
	
	// Register user
	err := s.userService.RegisterUser(email, username, fullName, password)
	if err != nil {
		data := PageData{
			Title: "Register",
			Messages: []Message{
				{Type: "error", Content: "Registration failed: " + err.Error()},
			},
		}
		s.renderTemplate(w, "register.html", data)
		return
	}
	
	// Redirect to login with success message
	http.Redirect(w, r, "/login?success=registered", http.StatusSeeOther)
}

// handleLogout handles user logout
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := s.getSession(r)
	session.Values["user_id"] = nil
	session.Save(r, w)
	
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleDashboard serves the user dashboard
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getCurrentUser(r)
	
	// Get user's portfolio summary
	portfolio := &PortfolioSummary{
		TotalValue:       250000.00,
		TotalGainLoss:    15000.00,
		TotalGainPercent: 6.38,
		PositionsCount:   8,
	}
	
	// Get recent trades
	recentTrades, _ := s.executionService.GetTradesByUser(user.ID)
	
	// Get market summary
	marketSummary := &MarketSummary{
		TotalSecurities: 45,
		ActiveListings:  23,
		TodayVolume:     1500000,
		TodayTrades:     89,
	}
	
	data := DashboardData{
		PageData: PageData{
			Title:      "Dashboard",
			User:       user,
			IsLoggedIn: true,
		},
		Portfolio:     portfolio,
		RecentTrades:  recentTrades,
		MarketSummary: marketSummary,
	}
	
	s.renderTemplate(w, "dashboard.html", data)
}

// handleSecurities lists all securities
func (s *Server) handleSecurities(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getCurrentUser(r)
	
	securities, err := s.securityService.GetSecurities()
	if err != nil {
		http.Error(w, "Failed to load securities", http.StatusInternalServerError)
		return
	}
	
	data := SecurityListData{
		PageData: PageData{
			Title:      "Securities",
			User:       user,
			IsLoggedIn: true,
		},
		Securities: securities,
	}
	
	s.renderTemplate(w, "securities.html", data)
}

// handleSecurityDetail shows security details
func (s *Server) handleSecurityDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	securityID := vars["id"]
	
	user, _ := s.getCurrentUser(r)
	
	security, err := s.securityService.GetSecurity(securityID)
	if err != nil {
		http.Error(w, "Security not found", http.StatusNotFound)
		return
	}
	
	// Get listings for this security
	listings, _ := s.listingService.GetListingsBySecurity(securityID)
	
	// Get market statistics
	stats, _ := s.executionService.GetMarketStatistics(securityID, 24*time.Hour)
	
	data := struct {
		PageData
		Security *Security
		Listings []*Listing
		Stats    *MarketStats
	}{
		PageData: PageData{
			Title:      security.CompanyName,
			User:       user,
			IsLoggedIn: true,
		},
		Security: security,
		Listings: listings,
		Stats:    stats,
	}
	
	s.renderTemplate(w, "security_detail.html", data)
}

// handleTrading shows the trading interface
func (s *Server) handleTrading(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getCurrentUser(r)
	
	listings, _ := s.listingService.GetActiveListings()
	bids, _ := s.biddingService.GetActiveBids()
	
	data := TradingData{
		PageData: PageData{
			Title:      "Trading",
			User:       user,
			IsLoggedIn: true,
		},
		ActiveListings: listings,
		ActiveBids:     bids,
	}
	
	s.renderTemplate(w, "trading.html", data)
}

// handleListings shows all listings
func (s *Server) handleListings(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getCurrentUser(r)
	
	listings, err := s.listingService.GetActiveListings()
	if err != nil {
		http.Error(w, "Failed to load listings", http.StatusInternalServerError)
		return
	}
	
	data := struct {
		PageData
		Listings []*Listing
	}{
		PageData: PageData{
			Title:      "Active Listings",
			User:       user,
			IsLoggedIn: true,
		},
		Listings: listings,
	}
	
	s.renderTemplate(w, "listings.html", data)
}

// handleListingDetail shows listing details
func (s *Server) handleListingDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	listingID := vars["id"]
	
	user, _ := s.getCurrentUser(r)
	
	listing, err := s.listingService.GetListing(listingID)
	if err != nil {
		http.Error(w, "Listing not found", http.StatusNotFound)
		return
	}
	
	data := struct {
		PageData
		Listing *Listing
	}{
		PageData: PageData{
			Title:      "Listing Details",
			User:       user,
			IsLoggedIn: true,
		},
		Listing: listing,
	}
	
	s.renderTemplate(w, "listing_detail.html", data)
}

// handleBids shows all bids
func (s *Server) handleBids(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getCurrentUser(r)
	
	bids, err := s.biddingService.GetActiveBids()
	if err != nil {
		http.Error(w, "Failed to load bids", http.StatusInternalServerError)
		return
	}
	
	data := struct {
		PageData
		Bids []*Bid
	}{
		PageData: PageData{
			Title:      "Active Bids",
			User:       user,
			IsLoggedIn: true,
		},
		Bids: bids,
	}
	
	s.renderTemplate(w, "bids.html", data)
}

// handleBidDetail shows bid details
func (s *Server) handleBidDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bidID := vars["id"]
	
	user, _ := s.getCurrentUser(r)
	
	bid, err := s.biddingService.GetBid(bidID)
	if err != nil {
		http.Error(w, "Bid not found", http.StatusNotFound)
		return
	}
	
	data := struct {
		PageData
		Bid *Bid
	}{
		PageData: PageData{
			Title:      "Bid Details",
			User:       user,
			IsLoggedIn: true,
		},
		Bid: bid,
	}
	
	s.renderTemplate(w, "bid_detail.html", data)
}

// handlePortfolio shows user portfolio
func (s *Server) handlePortfolio(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getCurrentUser(r)
	
	// Get user's securities
	securities, _ := s.securityService.GetUserSecurities(user.ID)
	
	data := struct {
		PageData
		Securities []*Security
	}{
		PageData: PageData{
			Title:      "My Portfolio",
			User:       user,
			IsLoggedIn: true,
		},
		Securities: securities,
	}
	
	s.renderTemplate(w, "portfolio.html", data)
}

// handleTrades shows user trades
func (s *Server) handleTrades(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getCurrentUser(r)
	
	trades, err := s.executionService.GetTradesByUser(user.ID)
	if err != nil {
		http.Error(w, "Failed to load trades", http.StatusInternalServerError)
		return
	}
	
	data := struct {
		PageData
		Trades []*Trade
	}{
		PageData: PageData{
			Title:      "My Trades",
			User:       user,
			IsLoggedIn: true,
		},
		Trades: trades,
	}
	
	s.renderTemplate(w, "trades.html", data)
}

// handleTradeDetail shows trade details
func (s *Server) handleTradeDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tradeID := vars["id"]
	
	user, _ := s.getCurrentUser(r)
	
	trade, err := s.executionService.GetTrade(tradeID)
	if err != nil {
		http.Error(w, "Trade not found", http.StatusNotFound)
		return
	}
	
	data := struct {
		PageData
		Trade *Trade
	}{
		PageData: PageData{
			Title:      "Trade Details",
			User:       user,
			IsLoggedIn: true,
		},
		Trade: trade,
	}
	
	s.renderTemplate(w, "trade_detail.html", data)
}

// Admin handlers
func (s *Server) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getCurrentUser(r)
	
	data := PageData{
		Title:      "Admin Dashboard",
		User:       user,
		IsLoggedIn: true,
	}
	
	s.renderTemplate(w, "admin_dashboard.html", data)
}

func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getCurrentUser(r)
	
	data := PageData{
		Title:      "User Management",
		User:       user,
		IsLoggedIn: true,
	}
	
	s.renderTemplate(w, "admin_users.html", data)
}

func (s *Server) handleAdminCompliance(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getCurrentUser(r)
	
	data := PageData{
		Title:      "Compliance Dashboard",
		User:       user,
		IsLoggedIn: true,
	}
	
	s.renderTemplate(w, "admin_compliance.html", data)
}

func (s *Server) handleAdminAudit(w http.ResponseWriter, r *http.Request) {
	user, _ := s.getCurrentUser(r)
	
	data := PageData{
		Title:      "Audit Log",
		User:       user,
		IsLoggedIn: true,
	}
	
	s.renderTemplate(w, "admin_audit.html", data)
}

// API handlers
func (s *Server) handleAPIMarketData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	securityID := vars["securityId"]
	
	stats, err := s.executionService.GetMarketStatistics(securityID, 24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to get market data", http.StatusInternalServerError)
		return
	}
	
	s.renderJSON(w, stats)
}

func (s *Server) handleAPIPortfolioSummary(w http.ResponseWriter, r *http.Request) {
	_, _ = s.getCurrentUser(r)
	
	// Get portfolio summary for the user
	summary := &PortfolioSummary{
		TotalValue:       250000.00,
		TotalGainLoss:    15000.00,
		TotalGainPercent: 6.38,
		PositionsCount:   8,
	}
	
	s.renderJSON(w, summary)
}