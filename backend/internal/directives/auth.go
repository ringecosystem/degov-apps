package directives

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/ringecosystem/degov-apps/graph/model"
	"github.com/ringecosystem/degov-apps/internal/middleware"
)

// AuthDirective handles @auth directive
func AuthDirective(ctx context.Context, obj interface{}, next graphql.Resolver, required *bool) (interface{}, error) {
	// Default to required=true if not specified
	isRequired := true
	if required != nil {
		isRequired = *required
	}

	if isRequired {
		// Authentication is required
		_, err := middleware.RequireAuth(ctx)
		if err != nil {
			return nil, fmt.Errorf("authentication required: %v", err)
		}
	} else {
		// Authentication is optional - just extract user info if available
		// This allows resolvers to provide personalized data when user is authenticated
		// but still work for unauthenticated users
		claims, ok := middleware.GetUserFromContext(ctx)
		if ok && claims != nil {
			// User is authenticated, add user info to context for resolver to use
			ctx = context.WithValue(ctx, "authenticated_user", claims.Address)
		}
	}

	return next(ctx)
}

// AuthorizeDirective handles @authorize directive
func AuthorizeDirective(ctx context.Context, obj interface{}, next graphql.Resolver, rule model.AuthRule) (interface{}, error) {
	// Get authenticated user
	claims, err := middleware.RequireAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required for authorization: %v", err)
	}

	switch rule {
	case model.AuthRuleOwnerOnly:
		// For OWNER_ONLY, we need to extract the resource owner from the arguments
		// This is a simplified implementation - in practice, you might need more sophisticated logic
		return authorizeOwnerOnly(ctx, claims, obj, next)

	case model.AuthRuleAdminOnly:
		// Check if user is admin (you'll need to implement this based on your user model)
		if !isAdmin(claims.Address) {
			return nil, fmt.Errorf("admin privileges required")
		}

	case model.AuthRulePublic:
		// Public access - no additional authorization needed
		break

	default:
		return nil, fmt.Errorf("unknown authorization rule: %v", rule)
	}

	return next(ctx)
}

// authorizeOwnerOnly checks if the authenticated user owns the resource
func authorizeOwnerOnly(ctx context.Context, claims *middleware.AuthClaims, obj interface{}, next graphql.Resolver) (interface{}, error) {
	// Extract field arguments to find the resource owner
	fieldCtx := graphql.GetFieldContext(ctx)
	if fieldCtx == nil {
		return nil, fmt.Errorf("failed to get field context")
	}

	// Check different argument patterns to find the user/owner ID
	args := fieldCtx.Args

	// Pattern 1: Direct userId argument
	if userID, ok := args["userId"].(string); ok {
		if claims.Address != userID {
			return nil, fmt.Errorf("permission denied: can only access your own resources")
		}
		return next(ctx)
	}

	// Pattern 2: Input object with userId field
	if input, ok := args["input"]; ok {
		if inputMap, ok := input.(map[string]interface{}); ok {
			if userID, ok := inputMap["userId"].(string); ok {
				if claims.Address != userID {
					return nil, fmt.Errorf("permission denied: can only access your own resources")
				}
				return next(ctx)
			}

			// Pattern 3: Input object with address field (for createUser)
			if address, ok := inputMap["address"].(string); ok {
				if claims.Address != address {
					return nil, fmt.Errorf("permission denied: can only create user for your own address")
				}
				return next(ctx)
			}
		}
	}

	// Pattern 4: For "me" query - always allow since user is accessing their own data
	if fieldCtx.Field.Name == "me" {
		return next(ctx)
	}

	// Pattern 5: Check if this is a user update operation without explicit user ID
	// In this case, we assume the user is updating their own profile
	if fieldCtx.Field.Name == "updateUser" {
		return next(ctx)
	}

	// If we can't determine ownership, deny access
	return nil, fmt.Errorf("permission denied: unable to verify resource ownership")
}

// isAdmin checks if the user has admin privileges
// You should implement this based on your user model and admin logic
func isAdmin(address string) bool {
	// TODO: Implement admin check logic
	// This could check a database, smart contract, or configuration
	// For now, return false (no admins)
	return false
}
