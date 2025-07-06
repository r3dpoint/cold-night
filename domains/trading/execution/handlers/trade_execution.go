package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"securities-marketplace/domains/shared/auth"
	"securities-marketplace/domains/shared/web"
	"securities-marketplace/domains/trading/execution"
)

// TradeExecutionHandler handles trade execution HTTP requests
type TradeExecutionHandler struct {
	service  *execution.ExecutionService
	renderer *web.TemplateRenderer
	auth     *auth.Middleware
}

// NewTradeExecutionHandler creates a new trade execution handler
func NewTradeExecutionHandler(service *execution.ExecutionService, renderer *web.TemplateRenderer, auth *auth.Middleware) *TradeExecutionHandler {
	return &TradeExecutionHandler{
		service:  service,
		renderer: renderer,
		auth:     auth,
	}
}

// ExecuteTradeRequest represents the request to execute a trade
type ExecuteTradeRequest struct {
	MatchResult *execution.MatchResult `json:"matchResult"`
}

// RunMatchingRequest represents the request to run order matching
type RunMatchingRequest struct {
	SecurityID string                       `json:"securityId"`
	Algorithm  execution.MatchingAlgorithm  `json:"algorithm"`
}

// ConfirmTradeRequest represents the request to confirm a trade
type ConfirmTradeRequest struct {
	TradeID     string `json:"tradeId"`
	ConfirmedBy string `json:"confirmedBy"`
}

// HandleExecuteTrade executes a single trade from a match result
func (h *TradeExecutionHandler) HandleExecuteTrade(w http.ResponseWriter, r *http.Request) {
	// Require admin or system role for trade execution
	if !h.auth.HasPermission(r, auth.Permission{
		Resource: "trades",
		Action:   "execute",
	}) {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req ExecuteTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.MatchResult == nil {
		http.Error(w, "Match result is required", http.StatusBadRequest)
		return
	}

	// Execute the trade
	trade, err := h.service.ExecuteTradeMatch(req.MatchResult)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute trade: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tradeId": trade.ID,
		"status":  trade.Status,
	})
}

// HandleRunMatching executes order matching for a security
func (h *TradeExecutionHandler) HandleRunMatching(w http.ResponseWriter, r *http.Request) {
	// Require admin or system role for running matching
	if !h.auth.HasPermission(r, auth.Permission{
		Resource: "matching",
		Action:   "execute",
	}) {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req RunMatchingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SecurityID == "" {
		http.Error(w, "Security ID is required", http.StatusBadRequest)
		return
	}

	if req.Algorithm == "" {
		req.Algorithm = execution.PriceTimePriority // Default algorithm
	}

	// Run matching
	trades, err := h.service.RunMatching(req.SecurityID, req.Algorithm)
	if err != nil {
		http.Error(w, fmt.Sprintf("Matching failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare response
	tradeResults := make([]map[string]interface{}, len(trades))
	for i, trade := range trades {
		tradeResults[i] = map[string]interface{}{
			"tradeId":      trade.ID,
			"buyerId":      trade.BuyerID,
			"sellerId":     trade.SellerID,
			"securityId":   trade.SecurityID,
			"sharesTraded": trade.SharesTraded,
			"tradePrice":   trade.TradePrice,
			"totalAmount":  trade.TotalAmount,
			"status":       trade.Status,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"tradesCreated": len(trades),
		"trades":     tradeResults,
		"algorithm":  req.Algorithm,
	})
}

// HandleConfirmTrade confirms a trade
func (h *TradeExecutionHandler) HandleConfirmTrade(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tradeID := vars["tradeId"]

	if tradeID == "" {
		http.Error(w, "Trade ID is required", http.StatusBadRequest)
		return
	}

	// Get user from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Confirm the trade
	err := h.service.ConfirmTrade(tradeID, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to confirm trade: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Trade confirmed successfully",
	})
}

// HandleGetTrade retrieves a specific trade
func (h *TradeExecutionHandler) HandleGetTrade(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tradeID := vars["tradeId"]

	if tradeID == "" {
		http.Error(w, "Trade ID is required", http.StatusBadRequest)
		return
	}

	trade, err := h.service.GetTrade(tradeID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get trade: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trade)
}

// HandleGetTradesByUser retrieves all trades for a user
func (h *TradeExecutionHandler) HandleGetTradesByUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Check permission - users can only see their own trades unless they're admin
	currentUserID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if userID != currentUserID && !h.auth.HasPermission(r, auth.Permission{
		Resource: "trades",
		Action:   "read_all",
	}) {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	trades, err := h.service.GetTradesByUser(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get trades: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trades)
}

// HandleGetMarketStatistics retrieves market statistics for a security
func (h *TradeExecutionHandler) HandleGetMarketStatistics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	securityID := vars["securityId"]

	if securityID == "" {
		http.Error(w, "Security ID is required", http.StatusBadRequest)
		return
	}

	// Parse period parameter (default to 24 hours)
	periodStr := r.URL.Query().Get("period")
	var period time.Duration = 24 * time.Hour

	if periodStr != "" {
		var err error
		period, err = time.ParseDuration(periodStr)
		if err != nil {
			http.Error(w, "Invalid period format", http.StatusBadRequest)
			return
		}
	}

	stats, err := h.service.GetMarketStatistics(securityID, period)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get market statistics: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleTradesPage renders the trades management page
func (h *TradeExecutionHandler) HandleTradesPage(w http.ResponseWriter, r *http.Request) {
	// Require admin role for trades management page
	if !h.auth.HasPermission(r, auth.Permission{
		Resource: "trades",
		Action:   "read_all",
	}) {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	// Parse query parameters
	status := r.URL.Query().Get("status")
	securityID := r.URL.Query().Get("securityId")
	pageStr := r.URL.Query().Get("page")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Get trades based on filters
	var trades []*execution.TradeAggregate
	var err error

	if status != "" {
		if tradeStatus, ok := parseTradeStatus(status); ok {
			trades, err = h.service.GetTradesByStatus(tradeStatus)
		} else {
			http.Error(w, "Invalid status", http.StatusBadRequest)
			return
		}
	} else if securityID != "" {
		trades, err = h.service.GetTradesBySecurity(securityID)
	} else {
		// For demo purposes, get pending settlements
		trades, err = h.service.GetPendingSettlements()
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get trades: %v", err), http.StatusInternalServerError)
		return
	}

	// Render the page
	data := map[string]interface{}{
		"PageTitle": "Trade Management",
		"Trades":    trades,
		"Filters": map[string]string{
			"status":     status,
			"securityId": securityID,
		},
		"Pagination": map[string]interface{}{
			"CurrentPage": page,
			"TotalPages":  1, // Simplified for now
		},
	}

	err = h.renderer.RenderTemplate(w, "execution/trades", data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
	}
}

// HandleMatchingPage renders the order matching page
func (h *TradeExecutionHandler) HandleMatchingPage(w http.ResponseWriter, r *http.Request) {
	// Require admin role for matching page
	if !h.auth.HasPermission(r, auth.Permission{
		Resource: "matching",
		Action:   "execute",
	}) {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	data := map[string]interface{}{
		"PageTitle": "Order Matching",
		"Algorithms": []map[string]string{
			{"Value": string(execution.PriceTimePriority), "Label": "Price-Time Priority"},
			{"Value": string(execution.UniformPriceAuction), "Label": "Uniform Price Auction"},
			{"Value": string(execution.NegotiatedTrading), "Label": "Negotiated Trading"},
		},
	}

	err := h.renderer.RenderTemplate(w, "execution/matching", data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
	}
}

// HandleSettlementPage renders the settlement management page
func (h *TradeExecutionHandler) HandleSettlementPage(w http.ResponseWriter, r *http.Request) {
	// Require admin role for settlement page
	if !h.auth.HasPermission(r, auth.Permission{
		Resource: "settlements",
		Action:   "manage",
	}) {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	// Get pending settlements
	pendingTrades, err := h.service.GetPendingSettlements()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get pending settlements: %v", err), http.StatusInternalServerError)
		return
	}

	// Get overdue trades
	overdueTrades, err := h.service.GetOverdueTrades()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get overdue trades: %v", err), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"PageTitle":      "Settlement Management",
		"PendingTrades":  pendingTrades,
		"OverdueTrades":  overdueTrades,
	}

	err = h.renderer.RenderTemplate(w, "execution/settlement", data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
	}
}

// Helper functions

func parseTradeStatus(status string) (execution.TradeStatus, bool) {
	switch strings.ToLower(status) {
	case "matched":
		return execution.TradeStatusMatched, true
	case "confirmed":
		return execution.TradeStatusConfirmed, true
	case "settlement_initiated":
		return execution.TradeStatusSettlementInitiated, true
	case "settled":
		return execution.TradeStatusSettled, true
	case "failed":
		return execution.TradeStatusFailed, true
	case "cancelled":
		return execution.TradeStatusCancelled, true
	default:
		return "", false
	}
}

// RegisterRoutes registers all trade execution routes
func (h *TradeExecutionHandler) RegisterRoutes(r *mux.Router) {
	// API routes
	api := r.PathPrefix("/api/trades").Subrouter()
	api.Use(h.auth.RequireAuthentication)

	api.HandleFunc("/execute", h.HandleExecuteTrade).Methods("POST")
	api.HandleFunc("/matching/run", h.HandleRunMatching).Methods("POST")
	api.HandleFunc("/{tradeId}/confirm", h.HandleConfirmTrade).Methods("POST")
	api.HandleFunc("/{tradeId}", h.HandleGetTrade).Methods("GET")
	api.HandleFunc("/user/{userId}", h.HandleGetTradesByUser).Methods("GET")
	api.HandleFunc("/statistics/{securityId}", h.HandleGetMarketStatistics).Methods("GET")

	// Page routes
	pages := r.PathPrefix("/trades").Subrouter()
	pages.Use(h.auth.RequireAuthentication)

	pages.HandleFunc("", h.HandleTradesPage).Methods("GET")
	pages.HandleFunc("/matching", h.HandleMatchingPage).Methods("GET")
	pages.HandleFunc("/settlement", h.HandleSettlementPage).Methods("GET")
}