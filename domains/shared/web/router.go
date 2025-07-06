package web

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

// NewRouter creates and configures the main application router
func NewRouter(db *sql.DB, redis *redis.Client) *mux.Router {
	router := mux.NewRouter()

	// Add middleware
	router.Use(LoggingMiddleware)
	router.Use(CORSMiddleware)
	router.Use(SecurityHeadersMiddleware)

	// Health check endpoint
	router.HandleFunc("/health", HealthCheckHandler).Methods("GET")

	// API routes
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	setupAPIRoutes(apiRouter, db, redis)

	// Web routes (server-rendered HTML)
	webRouter := router.PathPrefix("/").Subrouter()
	setupWebRoutes(webRouter, db, redis)

	// Static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))

	return router
}

// setupAPIRoutes configures API routes
func setupAPIRoutes(router *mux.Router, db *sql.DB, redis *redis.Client) {
	// Authentication routes
	authRouter := router.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/login", LoginHandler(db)).Methods("POST")
	authRouter.HandleFunc("/logout", LogoutHandler).Methods("POST")
	authRouter.HandleFunc("/register", RegisterHandler(db)).Methods("POST")

	// User routes
	userRouter := router.PathPrefix("/users").Subrouter()
	userRouter.Use(AuthenticationMiddleware)
	userRouter.HandleFunc("", GetUsersHandler(db)).Methods("GET")
	userRouter.HandleFunc("/{id}", GetUserHandler(db)).Methods("GET")
	userRouter.HandleFunc("/{id}", UpdateUserHandler(db)).Methods("PUT")

	// Security routes
	securityRouter := router.PathPrefix("/securities").Subrouter()
	securityRouter.Use(AuthenticationMiddleware)
	securityRouter.HandleFunc("", GetSecuritiesHandler(db)).Methods("GET")
	securityRouter.HandleFunc("/{id}", GetSecurityHandler(db)).Methods("GET")
	securityRouter.HandleFunc("", CreateSecurityHandler(db)).Methods("POST")

	// Trading routes
	tradingRouter := router.PathPrefix("/trading").Subrouter()
	tradingRouter.Use(AuthenticationMiddleware)
	tradingRouter.HandleFunc("/listings", GetListingsHandler(db)).Methods("GET")
	tradingRouter.HandleFunc("/listings", CreateListingHandler(db)).Methods("POST")
	tradingRouter.HandleFunc("/bids", GetBidsHandler(db)).Methods("GET")
	tradingRouter.HandleFunc("/bids", CreateBidHandler(db)).Methods("POST")
	tradingRouter.HandleFunc("/trades", GetTradesHandler(db)).Methods("GET")

	// Market data routes
	marketRouter := router.PathPrefix("/market").Subrouter()
	marketRouter.HandleFunc("/data", GetMarketDataHandler(db, redis)).Methods("GET")
	marketRouter.HandleFunc("/prices/{security_id}", GetPriceHistoryHandler(db)).Methods("GET")

	// Admin routes
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(AuthenticationMiddleware)
	adminRouter.Use(AdminAuthorizationMiddleware)
	adminRouter.HandleFunc("/users", AdminGetUsersHandler(db)).Methods("GET")
	adminRouter.HandleFunc("/securities", AdminGetSecuritiesHandler(db)).Methods("GET")
	adminRouter.HandleFunc("/trades", AdminGetTradesHandler(db)).Methods("GET")

	// Compliance routes
	complianceRouter := router.PathPrefix("/compliance").Subrouter()
	complianceRouter.Use(AuthenticationMiddleware)
	complianceRouter.Use(ComplianceAuthorizationMiddleware)
	complianceRouter.HandleFunc("/reports", GetComplianceReportsHandler(db)).Methods("GET")
	complianceRouter.HandleFunc("/activities", GetSuspiciousActivitiesHandler(db)).Methods("GET")
}

// setupWebRoutes configures web routes for server-rendered HTML
func setupWebRoutes(router *mux.Router, db *sql.DB, redis *redis.Client) {
	// Home page
	router.HandleFunc("/", HomeHandler(db)).Methods("GET")

	// Authentication pages
	router.HandleFunc("/login", LoginPageHandler).Methods("GET")
	router.HandleFunc("/register", RegisterPageHandler).Methods("GET")
	router.HandleFunc("/logout", LogoutPageHandler).Methods("GET")

	// User pages
	userRouter := router.PathPrefix("/app").Subrouter()
	userRouter.Use(WebAuthenticationMiddleware)
	userRouter.HandleFunc("/dashboard", UserDashboardHandler(db)).Methods("GET")
	userRouter.HandleFunc("/portfolio", UserPortfolioHandler(db)).Methods("GET")
	userRouter.HandleFunc("/profile", UserProfileHandler(db)).Methods("GET")

	// Trading pages
	userRouter.HandleFunc("/market", MarketHandler(db)).Methods("GET")
	userRouter.HandleFunc("/listings", ListingsHandler(db)).Methods("GET")
	userRouter.HandleFunc("/bids", BidsHandler(db)).Methods("GET")
	userRouter.HandleFunc("/trades", TradesHandler(db)).Methods("GET")

	// Admin pages
	adminRouter := userRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(WebAdminAuthorizationMiddleware)
	adminRouter.HandleFunc("/dashboard", AdminDashboardHandler(db)).Methods("GET")
	adminRouter.HandleFunc("/users", AdminUsersHandler(db)).Methods("GET")
	adminRouter.HandleFunc("/securities", AdminSecuritiesHandler(db)).Methods("GET")
	adminRouter.HandleFunc("/system", AdminSystemHandler(db)).Methods("GET")

	// Compliance pages
	complianceRouter := userRouter.PathPrefix("/compliance").Subrouter()
	complianceRouter.Use(WebComplianceAuthorizationMiddleware)
	complianceRouter.HandleFunc("/dashboard", ComplianceDashboardHandler(db)).Methods("GET")
	complianceRouter.HandleFunc("/monitoring", ComplianceMonitoringHandler(db)).Methods("GET")
	complianceRouter.HandleFunc("/reports", ComplianceReportsHandler(db)).Methods("GET")
	complianceRouter.HandleFunc("/investigations", ComplianceInvestigationsHandler(db)).Methods("GET")

	// Broker pages
	brokerRouter := userRouter.PathPrefix("/broker").Subrouter()
	brokerRouter.Use(WebBrokerAuthorizationMiddleware)
	brokerRouter.HandleFunc("/dashboard", BrokerDashboardHandler(db)).Methods("GET")
	brokerRouter.HandleFunc("/clients", BrokerClientsHandler(db)).Methods("GET")
	brokerRouter.HandleFunc("/orders", BrokerOrdersHandler(db)).Methods("GET")
	brokerRouter.HandleFunc("/settlements", BrokerSettlementsHandler(db)).Methods("GET")
}

// Placeholder handlers - these will be implemented in respective domain packages
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"securities-marketplace"}`))
}

// Placeholder functions for handlers that will be implemented later
func LoginHandler(db *sql.DB) http.HandlerFunc                      { return notImplemented }
func LogoutHandler(w http.ResponseWriter, r *http.Request)           { notImplemented(w, r) }
func RegisterHandler(db *sql.DB) http.HandlerFunc                    { return notImplemented }
func GetUsersHandler(db *sql.DB) http.HandlerFunc                    { return notImplemented }
func GetUserHandler(db *sql.DB) http.HandlerFunc                     { return notImplemented }
func UpdateUserHandler(db *sql.DB) http.HandlerFunc                  { return notImplemented }
func GetSecuritiesHandler(db *sql.DB) http.HandlerFunc               { return notImplemented }
func GetSecurityHandler(db *sql.DB) http.HandlerFunc                 { return notImplemented }
func CreateSecurityHandler(db *sql.DB) http.HandlerFunc              { return notImplemented }
func GetListingsHandler(db *sql.DB) http.HandlerFunc                 { return notImplemented }
func CreateListingHandler(db *sql.DB) http.HandlerFunc               { return notImplemented }
func GetBidsHandler(db *sql.DB) http.HandlerFunc                     { return notImplemented }
func CreateBidHandler(db *sql.DB) http.HandlerFunc                   { return notImplemented }
func GetTradesHandler(db *sql.DB) http.HandlerFunc                   { return notImplemented }
func GetMarketDataHandler(db *sql.DB, redis *redis.Client) http.HandlerFunc { return notImplemented }
func GetPriceHistoryHandler(db *sql.DB) http.HandlerFunc             { return notImplemented }
func AdminGetUsersHandler(db *sql.DB) http.HandlerFunc               { return notImplemented }
func AdminGetSecuritiesHandler(db *sql.DB) http.HandlerFunc          { return notImplemented }
func AdminGetTradesHandler(db *sql.DB) http.HandlerFunc              { return notImplemented }
func GetComplianceReportsHandler(db *sql.DB) http.HandlerFunc        { return notImplemented }
func GetSuspiciousActivitiesHandler(db *sql.DB) http.HandlerFunc     { return notImplemented }

// Web page handlers
func HomeHandler(db *sql.DB) http.HandlerFunc                        { return notImplemented }
func LoginPageHandler(w http.ResponseWriter, r *http.Request)         { notImplemented(w, r) }
func RegisterPageHandler(w http.ResponseWriter, r *http.Request)      { notImplemented(w, r) }
func LogoutPageHandler(w http.ResponseWriter, r *http.Request)        { notImplemented(w, r) }
func UserDashboardHandler(db *sql.DB) http.HandlerFunc               { return notImplemented }
func UserPortfolioHandler(db *sql.DB) http.HandlerFunc               { return notImplemented }
func UserProfileHandler(db *sql.DB) http.HandlerFunc                 { return notImplemented }
func MarketHandler(db *sql.DB) http.HandlerFunc                      { return notImplemented }
func ListingsHandler(db *sql.DB) http.HandlerFunc                    { return notImplemented }
func BidsHandler(db *sql.DB) http.HandlerFunc                        { return notImplemented }
func TradesHandler(db *sql.DB) http.HandlerFunc                      { return notImplemented }
func AdminDashboardHandler(db *sql.DB) http.HandlerFunc              { return notImplemented }
func AdminUsersHandler(db *sql.DB) http.HandlerFunc                  { return notImplemented }
func AdminSecuritiesHandler(db *sql.DB) http.HandlerFunc             { return notImplemented }
func AdminSystemHandler(db *sql.DB) http.HandlerFunc                 { return notImplemented }
func ComplianceDashboardHandler(db *sql.DB) http.HandlerFunc         { return notImplemented }
func ComplianceMonitoringHandler(db *sql.DB) http.HandlerFunc        { return notImplemented }
func ComplianceReportsHandler(db *sql.DB) http.HandlerFunc           { return notImplemented }
func ComplianceInvestigationsHandler(db *sql.DB) http.HandlerFunc    { return notImplemented }
func BrokerDashboardHandler(db *sql.DB) http.HandlerFunc             { return notImplemented }
func BrokerClientsHandler(db *sql.DB) http.HandlerFunc               { return notImplemented }
func BrokerOrdersHandler(db *sql.DB) http.HandlerFunc                { return notImplemented }
func BrokerSettlementsHandler(db *sql.DB) http.HandlerFunc           { return notImplemented }

func notImplemented(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"Not implemented yet"}`))
}