package auth

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"go-cp-boilerplate/internal/cache"
	"go-cp-boilerplate/internal/config"
	"go-cp-boilerplate/internal/domain"
	"go-cp-boilerplate/internal/pkg/apperror"
	"go-cp-boilerplate/internal/pkg/hash"
	"go-cp-boilerplate/internal/pkg/jwtutil"
	"go-cp-boilerplate/internal/pkg/turnstile"
)

type Service interface {
	Login(ctx context.Context, email, password, turnstileToken, remoteIP string) (*TokenResponse, error)
	Refresh(refreshToken string) (*TokenResponse, error)
	Logout(userID string, tokenJTI string, jwtExpiresAt time.Time) error
	Me(userID string, impersonatorID *string, activeRoleID, activeScopeID uint) (*MeResponse, error)
	UpdateProfile(userID string, impersonatorID *string, activeRoleID, activeScopeID uint, req UpdateProfileRequest) (*MeResponse, error)
	ChangePassword(userID string, impersonatorID *string, req ChangePasswordRequest) error
	DeleteProfile(userID string, impersonatorID *string, req DeleteProfileRequest) error
	Impersonate(actorID string, targetID string, tokenJTI string, jwtExpiresAt time.Time, ipAddress string) (*TokenResponse, error)
	StopImpersonate(targetID string, impersonatorID string, tokenJTI string, jwtExpiresAt time.Time, ipAddress string) (*TokenResponse, error)
	SwitchRole(userID string, impersonatorID *string, roleID, scopeID uint, tokenJTI string, jwtExpiresAt time.Time, ipAddress string) (*TokenResponse, error)
}

type service struct {
	repo        Repository
	jwtManager  *jwtutil.Manager
	redisClient *cache.RedisClient
	cfg         *config.Config
	bcryptCost  int
	turnstile   turnstile.Verifier
}

func NewService(
	repo Repository,
	jwtManager *jwtutil.Manager,
	redisClient *cache.RedisClient,
	cfg *config.Config,
	bcryptCost int,
	turnstileVerifier turnstile.Verifier,
) Service {
	if turnstileVerifier == nil {
		turnstileVerifier = turnstile.New("")
	}
	return &service{
		repo:        repo,
		jwtManager:  jwtManager,
		redisClient: redisClient,
		cfg:         cfg,
		bcryptCost:  bcryptCost,
		turnstile:   turnstileVerifier,
	}
}

type resolvedAssignment struct {
	RoleID    uint
	ScopeID   uint
	RoleName  string
	ScopeCode string
	ScopeName string
}

func (s *service) Login(ctx context.Context, email, password, turnstileToken, remoteIP string) (*TokenResponse, error) {
	if s.turnstile.Enabled() {
		if err := s.turnstile.Verify(ctx, turnstileToken, remoteIP); err != nil {
			return nil, err
		}
	}

	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.Unauthorized("Invalid email or password")
		}
		return nil, apperror.Internal("Failed to look up user")
	}

	if !user.IsActive {
		return nil, apperror.Forbidden("This account has been deactivated")
	}

	if !hash.Verify(password, user.Password) {
		return nil, apperror.Unauthorized("Invalid email or password")
	}

	now := time.Now()
	_ = s.repo.UpdateLastLogin(user.ID, now)

	userID := user.ID
	_ = s.repo.CreateActivityLog(&domain.ActivityLog{
		UserID:      &userID,
		Action:      "login",
		Module:      "auth",
		Description: "User signed in",
		IPAddress:   remoteIP,
	})

	return s.issueTokenPair(user, nil, nil, nil)
}

func (s *service) issueTokenPair(user *domain.User, impersonatorID *string, preferredRoleID, preferredScopeID *uint) (*TokenResponse, error) {
	assignments, err := s.repo.FindAssignmentsByUserID(user.ID)
	if err != nil {
		return nil, apperror.Internal("Failed to load role assignments")
	}
	if len(assignments) == 0 {
		return nil, apperror.Forbidden("User has no role assignments")
	}

	active, err := resolveAssignment(assignments, preferredRoleID, preferredScopeID)
	if err != nil {
		return nil, err
	}

	permissions, err := s.repo.FindPermissionsByRoleID(active.RoleID)
	if err != nil {
		return nil, apperror.Internal("Failed to load permissions")
	}
	if permissions == nil {
		permissions = []string{}
	}

	accessToken, expiresAt, _, err := s.jwtManager.GenerateAccessToken(jwtutil.AccessTokenInput{
		UserID:         user.ID,
		Role:           active.RoleName,
		Permissions:    permissions,
		ActiveRoleID:   active.RoleID,
		ActiveScopeID:  active.ScopeID,
		ImpersonatorID: impersonatorID,
	})
	if err != nil {
		return nil, apperror.Internal("Failed to generate access token")
	}

	rawRefresh, err := jwtutil.GenerateRefreshToken()
	if err != nil {
		return nil, apperror.Internal("Failed to generate refresh token")
	}

	roleID := active.RoleID
	scopeID := active.ScopeID
	rt := &domain.RefreshToken{
		UserID:         user.ID,
		TokenHash:      hash.HashToken(rawRefresh),
		ImpersonatorID: impersonatorID,
		ActiveRoleID:   &roleID,
		ActiveScopeID:  &scopeID,
		ExpiresAt:      time.Now().Add(s.cfg.JWTRefreshTTL),
	}
	if err := s.repo.CreateRefreshToken(rt); err != nil {
		return nil, apperror.Internal("Failed to persist refresh token")
	}

	return &TokenResponse{
		AccessToken:           accessToken,
		RefreshToken:          rawRefresh,
		TokenType:             "Bearer",
		ExpiresAt:             expiresAt.Format(time.RFC3339),
		RequiresRoleSelection: len(assignments) > 1,
		ActiveRoleID:          active.RoleID,
		ActiveScopeID:         active.ScopeID,
	}, nil
}

func resolveAssignment(assignments []domain.RoleAssignment, preferredRoleID, preferredScopeID *uint) (*resolvedAssignment, error) {
	if preferredRoleID != nil && preferredScopeID != nil && *preferredRoleID > 0 && *preferredScopeID > 0 {
		for i := range assignments {
			a := &assignments[i]
			if a.RoleID == *preferredRoleID && a.ScopeID == *preferredScopeID {
				return assignmentToResolved(a), nil
			}
		}
		// Preferred context no longer valid — fall through to default.
	}
	return pickDefaultAssignment(assignments)
}

func pickDefaultAssignment(assignments []domain.RoleAssignment) (*resolvedAssignment, error) {
	if len(assignments) == 0 {
		return nil, apperror.Forbidden("User has no role assignments")
	}

	var global *domain.RoleAssignment
	var first *domain.RoleAssignment

	for i := range assignments {
		a := &assignments[i]
		if a.Role != nil && a.Role.Name == domain.RoleSuperAdmin {
			return assignmentToResolved(a), nil
		}
		if first == nil {
			first = a
		}
		if a.Scope != nil && a.Scope.Code == domain.ScopeCodeGlobal && global == nil {
			global = a
		}
	}
	if global != nil {
		return assignmentToResolved(global), nil
	}
	return assignmentToResolved(first), nil
}

func assignmentToResolved(a *domain.RoleAssignment) *resolvedAssignment {
	out := &resolvedAssignment{
		RoleID:  a.RoleID,
		ScopeID: a.ScopeID,
	}
	if a.Role != nil {
		out.RoleName = a.Role.Name
	}
	if a.Scope != nil {
		out.ScopeCode = a.Scope.Code
		out.ScopeName = a.Scope.Name
	}
	return out
}

func (s *service) Refresh(refreshToken string) (*TokenResponse, error) {
	hashed := hash.HashToken(refreshToken)
	rt, err := s.repo.FindRefreshTokenByHash(hashed)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.Unauthorized("Invalid refresh token")
		}
		return nil, apperror.Internal("Failed to look up refresh token")
	}

	if rt.Revoked || time.Now().After(rt.ExpiresAt) {
		return nil, apperror.Unauthorized("Refresh token expired or revoked")
	}

	user, err := s.repo.FindUserByID(rt.UserID)
	if err != nil {
		return nil, apperror.Unauthorized("Invalid refresh token")
	}
	if !user.IsActive {
		return nil, apperror.Forbidden("This account has been deactivated")
	}

	if err := s.repo.RevokeRefreshToken(rt.ID); err != nil {
		return nil, apperror.Internal("Failed to rotate refresh token")
	}

	return s.issueTokenPair(user, rt.ImpersonatorID, rt.ActiveRoleID, rt.ActiveScopeID)
}

func (s *service) Logout(userID string, tokenJTI string, jwtExpiresAt time.Time) error {
	if err := s.repo.RevokeAllUserRefreshTokens(userID); err != nil {
		return apperror.Internal("Failed to revoke refresh tokens")
	}

	s.blacklistAccessToken(tokenJTI, jwtExpiresAt)
	return nil
}

func (s *service) SwitchRole(userID string, impersonatorID *string, roleID, scopeID uint, tokenJTI string, jwtExpiresAt time.Time, ipAddress string) (*TokenResponse, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("User not found")
		}
		return nil, apperror.Internal("Failed to look up user")
	}
	if !user.IsActive {
		return nil, apperror.Forbidden("This account has been deactivated")
	}

	assignment, err := s.repo.FindAssignment(userID, roleID, scopeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.Forbidden("Selected role assignment is not available for this user")
		}
		return nil, apperror.Internal("Failed to validate role assignment")
	}

	s.blacklistAccessToken(tokenJTI, jwtExpiresAt)
	_ = s.repo.RevokeAllUserRefreshTokens(userID)

	preferredRole := assignment.RoleID
	preferredScope := assignment.ScopeID
	tokens, err := s.issueTokenPair(user, impersonatorID, &preferredRole, &preferredScope)
	if err != nil {
		return nil, err
	}
	// After an explicit switch, do not force the choose-role screen again.
	tokens.RequiresRoleSelection = false

	_ = s.repo.CreateActivityLog(&domain.ActivityLog{
		UserID:      &userID,
		Action:      "role.switch",
		Module:      "auth",
		Description: "Switched active role assignment",
		IPAddress:   ipAddress,
	})

	return tokens, nil
}

func (s *service) Me(userID string, impersonatorID *string, activeRoleID, activeScopeID uint) (*MeResponse, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, apperror.NotFound("User not found")
	}
	assignments, err := s.repo.FindAssignmentsByUserID(user.ID)
	if err != nil {
		return nil, apperror.Internal("Failed to load role assignments")
	}

	var preferredRole, preferredScope *uint
	if activeRoleID > 0 && activeScopeID > 0 {
		preferredRole = &activeRoleID
		preferredScope = &activeScopeID
	}
	active, err := resolveAssignment(assignments, preferredRole, preferredScope)
	if err != nil {
		return nil, err
	}

	permissions, err := s.repo.FindPermissionsByRoleID(active.RoleID)
	if err != nil {
		return nil, apperror.Internal("Failed to load permissions")
	}
	return toMeResponse(user, impersonatorID, permissions, assignments, active), nil
}

func (s *service) UpdateProfile(userID string, impersonatorID *string, activeRoleID, activeScopeID uint, req UpdateProfileRequest) (*MeResponse, error) {
	if impersonatorID != nil {
		return nil, apperror.Forbidden("Profile cannot be edited while impersonating")
	}

	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("User not found")
		}
		return nil, apperror.Internal("Failed to fetch user")
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" && req.Email != user.Email {
		taken, err := s.repo.EmailTakenByOther(req.Email, userID)
		if err != nil {
			return nil, apperror.Internal("Failed to validate email")
		}
		if taken {
			return nil, apperror.Conflict("Email is already registered")
		}
		user.Email = req.Email
	}
	if req.Avatar != "" {
		user.Avatar = &req.Avatar
	}

	fields := map[string]interface{}{
		"name":  user.Name,
		"email": user.Email,
	}
	if user.Avatar != nil {
		fields["avatar"] = *user.Avatar
	}

	if err := s.repo.UpdateUserFields(userID, fields); err != nil {
		return nil, apperror.Internal("Failed to update profile")
	}

	return s.Me(userID, impersonatorID, activeRoleID, activeScopeID)
}

func (s *service) ChangePassword(userID string, impersonatorID *string, req ChangePasswordRequest) error {
	if impersonatorID != nil {
		return apperror.Forbidden("Password cannot be changed while impersonating")
	}

	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("User not found")
		}
		return apperror.Internal("Failed to fetch user")
	}

	if !hash.Verify(req.CurrentPassword, user.Password) {
		return apperror.BadRequest("Current password is incorrect")
	}

	hashed, err := hash.Hash(req.NewPassword, s.bcryptCost)
	if err != nil {
		return apperror.Internal("Failed to hash password")
	}

	if err := s.repo.UpdateUserFields(userID, map[string]interface{}{"password": hashed}); err != nil {
		return apperror.Internal("Failed to update password")
	}
	_ = s.repo.RevokeAllUserRefreshTokens(userID)
	return nil
}

func (s *service) DeleteProfile(userID string, impersonatorID *string, req DeleteProfileRequest) error {
	if impersonatorID != nil {
		return apperror.Forbidden("Account cannot be deleted while impersonating")
	}

	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("User not found")
		}
		return apperror.Internal("Failed to fetch user")
	}

	if !hash.Verify(req.Password, user.Password) {
		return apperror.BadRequest("Password is incorrect")
	}

	if err := s.repo.UpdateUserFields(userID, map[string]interface{}{"is_active": false}); err != nil {
		return apperror.Internal("Failed to deactivate account")
	}
	_ = s.repo.RevokeAllUserRefreshTokens(userID)
	return nil
}

func toMeResponse(user *domain.User, impersonatorID *string, permissions []string, assignments []domain.RoleAssignment, active *resolvedAssignment) *MeResponse {
	avatar := ""
	if user.Avatar != nil {
		avatar = *user.Avatar
	}

	if permissions == nil {
		permissions = []string{}
	}

	resp := &MeResponse{
		ID:                    user.ID,
		Name:                  user.Name,
		Email:                 user.Email,
		Role:                  active.RoleName,
		RoleID:                active.RoleID,
		Roles:                 uniqueRoleSummaries(assignments),
		Avatar:                avatar,
		IsActive:              user.IsActive,
		ImpersonatorID:        impersonatorID,
		Permissions:           permissions,
		Assignments:           toAssignmentResponses(assignments),
		ActiveRoleID:          active.RoleID,
		ActiveScopeID:         active.ScopeID,
		ActiveRoleName:        active.RoleName,
		ActiveScopeCode:       active.ScopeCode,
		ActiveScopeName:       active.ScopeName,
		RequiresRoleSelection: false,
		CanSwitchRole:         len(assignments) > 1,
	}
	if !user.CreatedAt.IsZero() {
		resp.CreatedAt = user.CreatedAt.Format(time.RFC3339)
	}
	if user.LastLoginAt != nil && !user.LastLoginAt.IsZero() {
		resp.LastLoginAt = user.LastLoginAt.Format(time.RFC3339)
	}
	return resp
}

func uniqueRoleSummaries(assignments []domain.RoleAssignment) []RoleSummary {
	seen := map[uint]struct{}{}
	out := make([]RoleSummary, 0)
	for _, a := range assignments {
		if _, ok := seen[a.RoleID]; ok {
			continue
		}
		seen[a.RoleID] = struct{}{}
		name := ""
		if a.Role != nil {
			name = a.Role.Name
		}
		out = append(out, RoleSummary{ID: a.RoleID, Name: name})
	}
	return out
}

func toAssignmentResponses(items []domain.RoleAssignment) []AssignmentResponse {
	out := make([]AssignmentResponse, 0, len(items))
	for _, a := range items {
		item := AssignmentResponse{
			RoleID:  a.RoleID,
			ScopeID: a.ScopeID,
		}
		if a.Role != nil {
			item.RoleName = a.Role.Name
		}
		if a.Scope != nil {
			item.ScopeCode = a.Scope.Code
			item.ScopeName = a.Scope.Name
		}
		out = append(out, item)
	}
	return out
}

func (s *service) Impersonate(actorID string, targetID string, tokenJTI string, jwtExpiresAt time.Time, ipAddress string) (*TokenResponse, error) {
	if actorID == targetID {
		return nil, apperror.BadRequest("Cannot impersonate yourself")
	}

	target, err := s.repo.FindUserByID(targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("User not found")
		}
		return nil, apperror.Internal("Failed to look up target user")
	}

	if !target.IsActive {
		return nil, apperror.Forbidden("Cannot impersonate an inactive user")
	}

	targetAssignments, err := s.repo.FindAssignmentsByUserID(target.ID)
	if err != nil {
		return nil, apperror.Internal("Failed to load target role assignments")
	}
	active, err := pickDefaultAssignment(targetAssignments)
	if err != nil {
		return nil, err
	}
	if active.RoleName == domain.RoleSuperAdmin {
		return nil, apperror.Forbidden("Cannot impersonate a SuperAdmin")
	}

	s.blacklistAccessToken(tokenJTI, jwtExpiresAt)

	tokens, err := s.issueTokenPair(target, &actorID, nil, nil)
	if err != nil {
		return nil, err
	}

	entityID := targetID
	_ = s.repo.CreateActivityLog(&domain.ActivityLog{
		UserID:      &actorID,
		Action:      "impersonate.start",
		Module:      "auth",
		EntityID:    &entityID,
		Description: "Started impersonation session",
		IPAddress:   ipAddress,
	})

	return tokens, nil
}

func (s *service) StopImpersonate(targetID string, impersonatorID string, tokenJTI string, jwtExpiresAt time.Time, ipAddress string) (*TokenResponse, error) {
	actor, err := s.repo.FindUserByID(impersonatorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.Unauthorized("Impersonator account not found")
		}
		return nil, apperror.Internal("Failed to look up impersonator")
	}

	if !actor.IsActive {
		return nil, apperror.Forbidden("Impersonator account has been deactivated")
	}

	s.blacklistAccessToken(tokenJTI, jwtExpiresAt)
	_ = s.repo.RevokeAllUserRefreshTokens(targetID)

	tokens, err := s.issueTokenPair(actor, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	entityID := targetID
	_ = s.repo.CreateActivityLog(&domain.ActivityLog{
		UserID:      &impersonatorID,
		Action:      "impersonate.stop",
		Module:      "auth",
		EntityID:    &entityID,
		Description: "Stopped impersonation session",
		IPAddress:   ipAddress,
	})

	return tokens, nil
}

func (s *service) blacklistAccessToken(tokenJTI string, jwtExpiresAt time.Time) {
	if s.redisClient == nil || tokenJTI == "" {
		return
	}
	ttl := time.Until(jwtExpiresAt)
	if ttl > 0 {
		_ = s.redisClient.BlacklistToken(context.Background(), tokenJTI, ttl)
	}
}
