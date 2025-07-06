package auth

import (
	"errors"
	"strings"
)

// Role represents a user role in the system
type Role string

const (
	RoleClient     Role = "client"
	RoleBroker     Role = "broker"
	RoleAdmin      Role = "admin"
	RoleCompliance Role = "compliance"
	RoleSystem     Role = "system"
)

// Permission represents a specific permission
type Permission string

const (
	// User permissions
	PermissionUserRead   Permission = "user:read"
	PermissionUserWrite  Permission = "user:write"
	PermissionUserDelete Permission = "user:delete"

	// Trading permissions
	PermissionTradeRead    Permission = "trade:read"
	PermissionTradeWrite   Permission = "trade:write"
	PermissionTradeExecute Permission = "trade:execute"
	PermissionTradeCancel  Permission = "trade:cancel"

	// Security permissions
	PermissionSecurityRead  Permission = "security:read"
	PermissionSecurityWrite Permission = "security:write"
	PermissionSecurityList  Permission = "security:list"

	// Compliance permissions
	PermissionComplianceRead   Permission = "compliance:read"
	PermissionComplianceWrite  Permission = "compliance:write"
	PermissionComplianceVerify Permission = "compliance:verify"
	PermissionComplianceRevoke Permission = "compliance:revoke"

	// Admin permissions
	PermissionAdminRead  Permission = "admin:read"
	PermissionAdminWrite Permission = "admin:write"
	PermissionAdminUsers Permission = "admin:users"

	// Market data permissions
	PermissionMarketDataRead Permission = "market_data:read"
	PermissionMarketDataWrite Permission = "market_data:write"

	// Reporting permissions
	PermissionReportRead  Permission = "report:read"
	PermissionReportWrite Permission = "report:write"
)

// RBAC manages role-based access control
type RBAC struct {
	rolePermissions map[Role][]Permission
}

// NewRBAC creates a new RBAC instance with default role permissions
func NewRBAC() *RBAC {
	rbac := &RBAC{
		rolePermissions: make(map[Role][]Permission),
	}

	// Define default role permissions
	rbac.setDefaultPermissions()
	return rbac
}

// setDefaultPermissions sets up the default permission mappings for each role
func (r *RBAC) setDefaultPermissions() {
	// Client permissions - can read own data, place trades
	r.rolePermissions[RoleClient] = []Permission{
		PermissionUserRead,
		PermissionTradeRead,
		PermissionTradeWrite,
		PermissionSecurityRead,
		PermissionSecurityList,
		PermissionMarketDataRead,
	}

	// Broker permissions - can manage clients and execute trades
	r.rolePermissions[RoleBroker] = []Permission{
		PermissionUserRead,
		PermissionUserWrite,
		PermissionTradeRead,
		PermissionTradeWrite,
		PermissionTradeExecute,
		PermissionTradeCancel,
		PermissionSecurityRead,
		PermissionSecurityWrite,
		PermissionSecurityList,
		PermissionMarketDataRead,
		PermissionMarketDataWrite,
		PermissionReportRead,
	}

	// Compliance permissions - can verify users and monitor activities
	r.rolePermissions[RoleCompliance] = []Permission{
		PermissionUserRead,
		PermissionUserWrite,
		PermissionTradeRead,
		PermissionSecurityRead,
		PermissionSecurityList,
		PermissionComplianceRead,
		PermissionComplianceWrite,
		PermissionComplianceVerify,
		PermissionComplianceRevoke,
		PermissionMarketDataRead,
		PermissionReportRead,
		PermissionReportWrite,
	}

	// Admin permissions - full system access
	r.rolePermissions[RoleAdmin] = []Permission{
		PermissionUserRead,
		PermissionUserWrite,
		PermissionUserDelete,
		PermissionTradeRead,
		PermissionTradeWrite,
		PermissionTradeExecute,
		PermissionTradeCancel,
		PermissionSecurityRead,
		PermissionSecurityWrite,
		PermissionSecurityList,
		PermissionComplianceRead,
		PermissionComplianceWrite,
		PermissionComplianceVerify,
		PermissionComplianceRevoke,
		PermissionAdminRead,
		PermissionAdminWrite,
		PermissionAdminUsers,
		PermissionMarketDataRead,
		PermissionMarketDataWrite,
		PermissionReportRead,
		PermissionReportWrite,
	}

	// System permissions - all permissions for internal operations
	r.rolePermissions[RoleSystem] = []Permission{
		PermissionUserRead,
		PermissionUserWrite,
		PermissionUserDelete,
		PermissionTradeRead,
		PermissionTradeWrite,
		PermissionTradeExecute,
		PermissionTradeCancel,
		PermissionSecurityRead,
		PermissionSecurityWrite,
		PermissionSecurityList,
		PermissionComplianceRead,
		PermissionComplianceWrite,
		PermissionComplianceVerify,
		PermissionComplianceRevoke,
		PermissionAdminRead,
		PermissionAdminWrite,
		PermissionAdminUsers,
		PermissionMarketDataRead,
		PermissionMarketDataWrite,
		PermissionReportRead,
		PermissionReportWrite,
	}
}

// HasPermission checks if a user with given roles has a specific permission
func (r *RBAC) HasPermission(roles []string, permission Permission) bool {
	for _, roleStr := range roles {
		role := Role(roleStr)
		permissions, exists := r.rolePermissions[role]
		if !exists {
			continue
		}

		for _, p := range permissions {
			if p == permission {
				return true
			}
		}
	}
	return false
}

// HasAnyPermission checks if a user has any of the specified permissions
func (r *RBAC) HasAnyPermission(roles []string, permissions []Permission) bool {
	for _, permission := range permissions {
		if r.HasPermission(roles, permission) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if a user has all of the specified permissions
func (r *RBAC) HasAllPermissions(roles []string, permissions []Permission) bool {
	for _, permission := range permissions {
		if !r.HasPermission(roles, permission) {
			return false
		}
	}
	return true
}

// GetPermissions returns all permissions for given roles
func (r *RBAC) GetPermissions(roles []string) []Permission {
	permissionSet := make(map[Permission]bool)
	
	for _, roleStr := range roles {
		role := Role(roleStr)
		permissions, exists := r.rolePermissions[role]
		if !exists {
			continue
		}

		for _, permission := range permissions {
			permissionSet[permission] = true
		}
	}

	var result []Permission
	for permission := range permissionSet {
		result = append(result, permission)
	}
	
	return result
}

// AddRolePermission adds a permission to a role
func (r *RBAC) AddRolePermission(role Role, permission Permission) {
	if r.rolePermissions[role] == nil {
		r.rolePermissions[role] = []Permission{}
	}
	r.rolePermissions[role] = append(r.rolePermissions[role], permission)
}

// RemoveRolePermission removes a permission from a role
func (r *RBAC) RemoveRolePermission(role Role, permission Permission) {
	permissions := r.rolePermissions[role]
	for i, p := range permissions {
		if p == permission {
			r.rolePermissions[role] = append(permissions[:i], permissions[i+1:]...)
			break
		}
	}
}

// IsValidRole checks if a role is valid
func (r *RBAC) IsValidRole(role string) bool {
	validRoles := []Role{RoleClient, RoleBroker, RoleAdmin, RoleCompliance, RoleSystem}
	for _, validRole := range validRoles {
		if Role(role) == validRole {
			return true
		}
	}
	return false
}

// GetRoleHierarchy returns the role hierarchy (higher roles inherit lower role permissions)
func (r *RBAC) GetRoleHierarchy() map[Role][]Role {
	return map[Role][]Role{
		RoleSystem:     {RoleAdmin, RoleCompliance, RoleBroker, RoleClient},
		RoleAdmin:      {RoleCompliance, RoleBroker, RoleClient},
		RoleCompliance: {RoleClient},
		RoleBroker:     {RoleClient},
		RoleClient:     {},
	}
}

// ExpandRoles expands roles based on hierarchy
func (r *RBAC) ExpandRoles(roles []string) []string {
	hierarchy := r.GetRoleHierarchy()
	expandedRoles := make(map[string]bool)
	
	// Add original roles
	for _, role := range roles {
		expandedRoles[role] = true
	}
	
	// Add inherited roles
	for _, role := range roles {
		inheritedRoles, exists := hierarchy[Role(role)]
		if exists {
			for _, inheritedRole := range inheritedRoles {
				expandedRoles[string(inheritedRole)] = true
			}
		}
	}
	
	var result []string
	for role := range expandedRoles {
		result = append(result, role)
	}
	
	return result
}

// Resource-based permission checking

// ResourcePermission represents a permission on a specific resource
type ResourcePermission struct {
	Action   string `json:"action"`
	Resource string `json:"resource"`
	UserID   string `json:"userId,omitempty"`
}

// CanAccessResource checks if a user can access a specific resource
func (r *RBAC) CanAccessResource(userID string, roles []string, resource, action string) error {
	// Build permission string
	permission := Permission(resource + ":" + action)
	
	// Check if user has the required permission
	if !r.HasPermission(roles, permission) {
		return errors.New("insufficient permissions")
	}
	
	return nil
}

// CanAccessUserResource checks if a user can access their own or others' resources
func (r *RBAC) CanAccessUserResource(currentUserID string, targetUserID string, roles []string, resource, action string) error {
	// Users can always access their own resources (with basic permissions)
	if currentUserID == targetUserID {
		basicPermission := Permission(resource + ":read")
		if action == "read" && r.HasPermission(roles, basicPermission) {
			return nil
		}
		// For write operations on own data, check write permission
		if strings.Contains(action, "write") || strings.Contains(action, "update") {
			writePermission := Permission(resource + ":write")
			if r.HasPermission(roles, writePermission) {
				return nil
			}
		}
	}
	
	// For accessing other users' resources, use standard permission check
	return r.CanAccessResource(currentUserID, roles, resource, action)
}

// GetUserRoles determines user roles based on user type and status
func GetUserRoles(userType, complianceStatus, accreditationStatus string) []string {
	roles := []string{string(RoleClient)} // Default role
	
	switch strings.ToLower(userType) {
	case "broker":
		roles = []string{string(RoleBroker), string(RoleClient)}
	case "admin":
		roles = []string{string(RoleAdmin), string(RoleCompliance), string(RoleBroker), string(RoleClient)}
	case "compliance":
		roles = []string{string(RoleCompliance), string(RoleClient)}
	}
	
	// Restrict permissions based on compliance status
	if complianceStatus != "clear" {
		// Remove trading permissions for non-compliant users
		roles = filterRoles(roles, []string{string(RoleBroker)})
	}
	
	return roles
}

// filterRoles removes specified roles from a role list
func filterRoles(roles []string, toRemove []string) []string {
	removeMap := make(map[string]bool)
	for _, role := range toRemove {
		removeMap[role] = true
	}
	
	var filtered []string
	for _, role := range roles {
		if !removeMap[role] {
			filtered = append(filtered, role)
		}
	}
	
	return filtered
}