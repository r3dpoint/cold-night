package web

import (
	"log"
	"net/http"
	"time"
)

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Call the next handler
		next.ServeHTTP(w, r)
		
		// Log the request
		log.Printf(
			"%s %s %s %v",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			time.Since(start),
		)
	})
}

// securityHeadersMiddleware adds security headers
func (s *Server) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// sessionMiddleware manages user sessions
func (s *Server) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get or create session
		session, err := s.getSession(r)
		if err != nil {
			log.Printf("Session error: %v", err)
		} else {
			// Set session options
			session.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   86400 * 7, // 7 days
				HttpOnly: true,
				Secure:   false, // Set to true in production with HTTPS
				SameSite: http.SameSiteStrictMode,
			}
		}
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// authMiddleware requires user authentication
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.getSession(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		
		userID, ok := session.Values["user_id"].(string)
		if !ok || userID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		
		// Verify user still exists and is active
		user, err := s.userService.GetUser(userID)
		if err != nil || !user.IsActive {
			// Clear invalid session
			session.Values["user_id"] = nil
			session.Save(r, w)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// adminMiddleware requires admin role
func (s *Server) adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.getCurrentUser(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		if user.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// complianceMiddleware requires compliance role
func (s *Server) complianceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.getCurrentUser(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		if user.Role != "admin" && user.Role != "compliance" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// brokerMiddleware requires broker role
func (s *Server) brokerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.getCurrentUser(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		if user.Role != "admin" && user.Role != "broker" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// tradingMiddleware requires trading permissions
func (s *Server) tradingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.getCurrentUser(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		if !user.CanTrade {
			http.Error(w, "Trading not permitted", http.StatusForbidden)
			return
		}
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}