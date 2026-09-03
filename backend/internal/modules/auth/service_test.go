package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"go-cp-boilerplate/internal/config"
	"go-cp-boilerplate/internal/domain"
	"go-cp-boilerplate/internal/pkg/apperror"
	"go-cp-boilerplate/internal/pkg/hash"
	"go-cp-boilerplate/internal/pkg/jwtutil"
	"go-cp-boilerplate/internal/pkg/turnstile"
)

type mockRepo struct {
	users             map[string]*domain.User
	usersByID         map[string]*domain.User
	permissions       map[string][]string
	permissionsByRole map[uint][]string
	assignments       map[string][]domain.RoleAssignment
	refreshTokens     map[string]*domain.RefreshToken
	activityLogs      []*domain.ActivityLog
	nextRTID          uint
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		users:             map[string]*domain.User{},
		usersByID:         map[string]*domain.User{},
		permissions:       map[string][]string{},
		permissionsByRole: map[uint][]string{},
		assignments:       map[string][]domain.RoleAssignment{},
		refreshTokens:     map[string]*domain.RefreshToken{},
	}
}

func (m *mockRepo) FindUserByEmail(email string) (*domain.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return u, nil
}

func (m *mockRepo) FindUserByID(id string) (*domain.User, error) {
	u, ok := m.usersByID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return u, nil
}

func (m *mockRepo) FindPermissionsByRoleID(roleID uint) ([]string, error) {
	if p, ok := m.permissionsByRole[roleID]; ok {
		return p, nil
	}
	return []string{}, nil
}

func (m *mockRepo) FindUnionPermissions(userID string) ([]string, error) {
	return m.permissions[userID], nil
}

func (m *mockRepo) FindAssignmentsByUserID(userID string) ([]domain.RoleAssignment, error) {
	if items, ok := m.assignments[userID]; ok {
		return items, nil
	}
	return []domain.RoleAssignment{}, nil
}

func (m *mockRepo) FindAssignment(userID string, roleID, scopeID uint) (*domain.RoleAssignment, error) {
	for _, a := range m.assignments[userID] {
		if a.RoleID == roleID && a.ScopeID == scopeID {
			cp := a
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockRepo) UpdateLastLogin(userID string, at time.Time) error { return nil }

func (m *mockRepo) UpdateUserFields(userID string, fields map[string]interface{}) error {
	u, ok := m.usersByID[userID]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if name, ok := fields["name"].(string); ok {
		u.Name = name
	}
	if email, ok := fields["email"].(string); ok {
		delete(m.users, u.Email)
		u.Email = email
		m.users[email] = u
	}
	if avatar, ok := fields["avatar"].(string); ok {
		u.Avatar = &avatar
	}
	if password, ok := fields["password"].(string); ok {
		u.Password = password
	}
	if active, ok := fields["is_active"].(bool); ok {
		u.IsActive = active
	}
	m.usersByID[userID] = u
	return nil
}

func (m *mockRepo) EmailTakenByOther(email string, userID string) (bool, error) {
	for id, u := range m.usersByID {
		if u.Email == email && id != userID {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockRepo) CreateRefreshToken(rt *domain.RefreshToken) error {
	m.nextRTID++
	rt.ID = m.nextRTID
	m.refreshTokens[rt.TokenHash] = rt
	return nil
}

func (m *mockRepo) FindRefreshTokenByHash(hashVal string) (*domain.RefreshToken, error) {
	rt, ok := m.refreshTokens[hashVal]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return rt, nil
}

func (m *mockRepo) RevokeRefreshToken(id uint) error {
	for _, rt := range m.refreshTokens {
		if rt.ID == id {
			rt.Revoked = true
		}
	}
	return nil
}

func (m *mockRepo) RevokeAllUserRefreshTokens(userID string) error {
	for _, rt := range m.refreshTokens {
		if rt.UserID == userID {
			rt.Revoked = true
		}
	}
	return nil
}

func (m *mockRepo) CreateActivityLog(log *domain.ActivityLog) error {
	m.activityLogs = append(m.activityLogs, log)
	return nil
}

func newTestService(repo *mockRepo) Service {
	cfg := &config.Config{JWTRefreshTTL: time.Hour}
	jwtManager := jwtutil.NewManager("test-secret", time.Minute*15)
	return NewService(repo, jwtManager, nil, cfg, 4, turnstile.New(""))
}

func seedUser(repo *mockRepo, id string, email, password string, active bool) {
	seedUserWithRole(repo, id, email, password, active, 1, domain.RoleAdmin, []string{"setting.view"})
}

func seedUserWithRole(repo *mockRepo, id string, email, password string, active bool, roleID uint, roleName string, perms []string) {
	hashed, _ := hash.Hash(password, 4)
	u := &domain.User{
		ID:       id,
		Name:     email,
		Email:    email,
		Password: hashed,
		IsActive: active,
	}
	repo.users[email] = u
	repo.usersByID[id] = u
	repo.permissions[id] = perms
	repo.permissionsByRole[roleID] = perms
	repo.assignments[id] = []domain.RoleAssignment{{
		UserID:  id,
		RoleID:  roleID,
		ScopeID: 1,
		Role:    &domain.Role{ID: roleID, Name: roleName},
		Scope:   &domain.Scope{ID: 1, Code: domain.ScopeCodeGlobal, Name: "Global"},
	}}
}

func TestLogin_Success(t *testing.T) {
	repo := newMockRepo()
	seedUser(repo, "11111111-1111-1111-1111-111111111111", "admin@example.com", "correct-password", true)
	svc := newTestService(repo)

	tokens, err := svc.Login(context.Background(), "admin@example.com", "correct-password", "", "")
	if err != nil {
		t.Fatalf("expected successful login, got error: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("expected non-empty access and refresh tokens")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockRepo()
	seedUser(repo, "11111111-1111-1111-1111-111111111111", "admin@example.com", "correct-password", true)
	svc := newTestService(repo)

	_, err := svc.Login(context.Background(), "admin@example.com", "wrong-password", "", "")
	assertAppErrorStatus(t, err, 401)
}

func TestLogin_UnknownEmail(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)

	_, err := svc.Login(context.Background(), "nobody@example.com", "whatever", "", "")
	assertAppErrorStatus(t, err, 401)
}

func TestLogin_InactiveUser(t *testing.T) {
	repo := newMockRepo()
	seedUser(repo, "11111111-1111-1111-1111-111111111111", "admin@example.com", "correct-password", false)
	svc := newTestService(repo)

	_, err := svc.Login(context.Background(), "admin@example.com", "correct-password", "", "")
	assertAppErrorStatus(t, err, 403)
}

func TestRefresh_RotatesToken(t *testing.T) {
	repo := newMockRepo()
	seedUser(repo, "11111111-1111-1111-1111-111111111111", "admin@example.com", "correct-password", true)
	svc := newTestService(repo)

	tokens, err := svc.Login(context.Background(), "admin@example.com", "correct-password", "", "")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	refreshed, err := svc.Refresh(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("expected successful refresh, got error: %v", err)
	}
	if refreshed.RefreshToken == tokens.RefreshToken {
		t.Error("expected refresh to rotate the refresh token")
	}

	// Old refresh token must now be rejected (replay prevention).
	if _, err := svc.Refresh(tokens.RefreshToken); err == nil {
		t.Error("expected the rotated-out refresh token to be rejected")
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)

	_, err := svc.Refresh("does-not-exist")
	assertAppErrorStatus(t, err, 401)
}

func TestLogout_RevokesRefreshTokens(t *testing.T) {
	repo := newMockRepo()
	seedUser(repo, "11111111-1111-1111-1111-111111111111", "admin@example.com", "correct-password", true)
	svc := newTestService(repo)

	tokens, err := svc.Login(context.Background(), "admin@example.com", "correct-password", "", "")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if err := svc.Logout("11111111-1111-1111-1111-111111111111", "some-jti", time.Now().Add(time.Minute)); err != nil {
		t.Fatalf("expected successful logout, got error: %v", err)
	}

	if _, err := svc.Refresh(tokens.RefreshToken); err == nil {
		t.Error("expected refresh token to be revoked after logout")
	}
}

func TestMe_Success(t *testing.T) {
	repo := newMockRepo()
	seedUser(repo, "11111111-1111-1111-1111-111111111111", "admin@example.com", "correct-password", true)
	svc := newTestService(repo)

	me, err := svc.Me("11111111-1111-1111-1111-111111111111", nil, 0, 0)
	if err != nil {
		t.Fatalf("expected Me to succeed, got error: %v", err)
	}
	if me.Email != "admin@example.com" {
		t.Errorf("expected email admin@example.com, got %s", me.Email)
	}
	if me.Role != domain.RoleAdmin {
		t.Errorf("expected role %s, got %s", domain.RoleAdmin, me.Role)
	}
	if me.ImpersonatorID != nil {
		t.Error("expected nil ImpersonatorID")
	}
}

func TestMe_WithImpersonator(t *testing.T) {
	repo := newMockRepo()
	seedUser(repo, "22222222-2222-2222-2222-222222222222", "editor@example.com", "password", true)
	svc := newTestService(repo)
	actor := "11111111-1111-1111-1111-111111111111"

	me, err := svc.Me("22222222-2222-2222-2222-222222222222", &actor, 0, 0)
	if err != nil {
		t.Fatalf("expected Me to succeed, got error: %v", err)
	}
	if me.ImpersonatorID == nil || *me.ImpersonatorID != actor {
		t.Fatalf("expected ImpersonatorID actor, got %v", me.ImpersonatorID)
	}
}

func TestMe_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)

	_, err := svc.Me("99999999-9999-9999-9999-999999999999", nil, 0, 0)
	assertAppErrorStatus(t, err, 404)
}

func TestImpersonate_Success(t *testing.T) {
	repo := newMockRepo()
	seedUserWithRole(repo, "11111111-1111-1111-1111-111111111111", "super@example.com", "password", true, 1, domain.RoleSuperAdmin, []string{"user.impersonate"})
	seedUserWithRole(repo, "22222222-2222-2222-2222-222222222222", "editor@example.com", "password", true, 2, domain.RoleEditor, []string{"setting.view"})
	svc := newTestService(repo)
	jwtManager := jwtutil.NewManager("test-secret", time.Minute*15)

	tokens, err := svc.Impersonate("11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222", "actor-jti", time.Now().Add(time.Minute), "127.0.0.1")
	if err != nil {
		t.Fatalf("expected impersonate to succeed, got %v", err)
	}

	claims, err := jwtManager.ParseAccessToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("parse access token: %v", err)
	}
	if claims.UserID != "22222222-2222-2222-2222-222222222222" {
		t.Errorf("expected effective user target, got %s", claims.UserID)
	}
	if claims.ImpersonatorID == nil || *claims.ImpersonatorID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("expected impersonator 1, got %v", claims.ImpersonatorID)
	}
	if claims.ActiveRoleID == 0 || claims.ActiveScopeID == 0 {
		t.Fatalf("expected active assignment on impersonation token, got role=%d scope=%d", claims.ActiveRoleID, claims.ActiveScopeID)
	}
	if len(repo.activityLogs) != 1 || repo.activityLogs[0].Action != "impersonate.start" {
		t.Errorf("expected impersonate.start activity log, got %+v", repo.activityLogs)
	}

	hashed := hash.HashToken(tokens.RefreshToken)
	rt, err := repo.FindRefreshTokenByHash(hashed)
	if err != nil {
		t.Fatalf("refresh token missing: %v", err)
	}
	if rt.UserID != "22222222-2222-2222-2222-222222222222" || rt.ImpersonatorID == nil || *rt.ImpersonatorID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("refresh token actor not preserved: %+v", rt)
	}
}

func TestImpersonate_RejectSelf(t *testing.T) {
	repo := newMockRepo()
	seedUserWithRole(repo, "11111111-1111-1111-1111-111111111111", "super@example.com", "password", true, 1, domain.RoleSuperAdmin, []string{"user.impersonate"})
	svc := newTestService(repo)

	_, err := svc.Impersonate("11111111-1111-1111-1111-111111111111", "11111111-1111-1111-1111-111111111111", "jti", time.Now().Add(time.Minute), "")
	assertAppErrorStatus(t, err, 400)
}

func TestImpersonate_RejectSuperAdmin(t *testing.T) {
	repo := newMockRepo()
	seedUserWithRole(repo, "11111111-1111-1111-1111-111111111111", "super@example.com", "password", true, 1, domain.RoleSuperAdmin, []string{"user.impersonate"})
	seedUserWithRole(repo, "22222222-2222-2222-2222-222222222222", "other-super@example.com", "password", true, 1, domain.RoleSuperAdmin, []string{"user.impersonate"})
	svc := newTestService(repo)

	_, err := svc.Impersonate("11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222", "jti", time.Now().Add(time.Minute), "")
	assertAppErrorStatus(t, err, 403)
}

func TestImpersonate_RejectInactive(t *testing.T) {
	repo := newMockRepo()
	seedUserWithRole(repo, "11111111-1111-1111-1111-111111111111", "super@example.com", "password", true, 1, domain.RoleSuperAdmin, []string{"user.impersonate"})
	seedUserWithRole(repo, "22222222-2222-2222-2222-222222222222", "editor@example.com", "password", false, 2, domain.RoleEditor, []string{"setting.view"})
	svc := newTestService(repo)

	_, err := svc.Impersonate("11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222", "jti", time.Now().Add(time.Minute), "")
	assertAppErrorStatus(t, err, 403)
}

func TestRefresh_PreservesImpersonator(t *testing.T) {
	repo := newMockRepo()
	seedUserWithRole(repo, "11111111-1111-1111-1111-111111111111", "super@example.com", "password", true, 1, domain.RoleSuperAdmin, []string{"user.impersonate"})
	seedUserWithRole(repo, "22222222-2222-2222-2222-222222222222", "editor@example.com", "password", true, 2, domain.RoleEditor, []string{"setting.view"})
	svc := newTestService(repo)
	jwtManager := jwtutil.NewManager("test-secret", time.Minute*15)

	tokens, err := svc.Impersonate("11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222", "jti", time.Now().Add(time.Minute), "")
	if err != nil {
		t.Fatalf("impersonate failed: %v", err)
	}

	refreshed, err := svc.Refresh(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	claims, err := jwtManager.ParseAccessToken(refreshed.AccessToken)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != "22222222-2222-2222-2222-222222222222" {
		t.Errorf("expected user target, got %s", claims.UserID)
	}
	if claims.ImpersonatorID == nil || *claims.ImpersonatorID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("expected impersonator preserved, got %v", claims.ImpersonatorID)
	}
}

func TestStopImpersonate_Success(t *testing.T) {
	repo := newMockRepo()
	seedUserWithRole(repo, "11111111-1111-1111-1111-111111111111", "super@example.com", "password", true, 1, domain.RoleSuperAdmin, []string{"user.impersonate"})
	seedUserWithRole(repo, "22222222-2222-2222-2222-222222222222", "editor@example.com", "password", true, 2, domain.RoleEditor, []string{"setting.view"})
	svc := newTestService(repo)
	jwtManager := jwtutil.NewManager("test-secret", time.Minute*15)

	impTokens, err := svc.Impersonate("11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222", "jti", time.Now().Add(time.Minute), "")
	if err != nil {
		t.Fatalf("impersonate failed: %v", err)
	}

	restored, err := svc.StopImpersonate("22222222-2222-2222-2222-222222222222", "11111111-1111-1111-1111-111111111111", "imp-jti", time.Now().Add(time.Minute), "127.0.0.1")
	if err != nil {
		t.Fatalf("stop failed: %v", err)
	}

	claims, err := jwtManager.ParseAccessToken(restored.AccessToken)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("expected restored user actor, got %s", claims.UserID)
	}
	if claims.ImpersonatorID != nil {
		t.Fatalf("expected no impersonator after stop, got %v", claims.ImpersonatorID)
	}

	if _, err := svc.Refresh(impTokens.RefreshToken); err == nil {
		t.Error("expected impersonation refresh token revoked after stop")
	}

	foundStop := false
	for _, log := range repo.activityLogs {
		if log.Action == "impersonate.stop" {
			foundStop = true
		}
	}
	if !foundStop {
		t.Error("expected impersonate.stop activity log")
	}
}

func TestSwitchRole_Success(t *testing.T) {
	repo := newMockRepo()
	id := "11111111-1111-1111-1111-111111111111"
	seedUserWithRole(repo, id, "multi@example.com", "password", true, 2, domain.RoleAdmin, []string{"setting.view", "user.view"})
	repo.permissionsByRole[3] = []string{"setting.view"}
	repo.assignments[id] = append(repo.assignments[id], domain.RoleAssignment{
		UserID:  id,
		RoleID:  3,
		ScopeID: 2,
		Role:    &domain.Role{ID: 3, Name: domain.RoleEditor},
		Scope:   &domain.Scope{ID: 2, Code: "campus-a", Name: "Campus A"},
	})
	svc := newTestService(repo)
	jwtManager := jwtutil.NewManager("test-secret", time.Minute*15)

	loginTokens, err := svc.Login(context.Background(), "multi@example.com", "password", "", "")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if !loginTokens.RequiresRoleSelection {
		t.Fatal("expected requires_role_selection for multi-assignment user")
	}

	switched, err := svc.SwitchRole(id, nil, 3, 2, "old-jti", time.Now().Add(time.Minute), "127.0.0.1")
	if err != nil {
		t.Fatalf("switch failed: %v", err)
	}
	if switched.RequiresRoleSelection {
		t.Error("expected requires_role_selection false after explicit switch")
	}
	if switched.ActiveRoleID != 3 || switched.ActiveScopeID != 2 {
		t.Fatalf("unexpected active assignment in response: %+v", switched)
	}

	claims, err := jwtManager.ParseAccessToken(switched.AccessToken)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.Role != domain.RoleEditor {
		t.Errorf("expected Editor role claim, got %s", claims.Role)
	}
	if claims.ActiveRoleID != 3 || claims.ActiveScopeID != 2 {
		t.Fatalf("unexpected claims assignment: %+v", claims)
	}
	if len(claims.Permissions) != 1 || claims.Permissions[0] != "setting.view" {
		t.Fatalf("expected permissions of selected role only, got %v", claims.Permissions)
	}
}

func TestSwitchRole_RejectUnknownAssignment(t *testing.T) {
	repo := newMockRepo()
	seedUser(repo, "11111111-1111-1111-1111-111111111111", "admin@example.com", "password", true)
	svc := newTestService(repo)

	_, err := svc.SwitchRole("11111111-1111-1111-1111-111111111111", nil, 99, 99, "jti", time.Now().Add(time.Minute), "")
	assertAppErrorStatus(t, err, 403)
}

func TestRefresh_PreservesActiveAssignment(t *testing.T) {
	repo := newMockRepo()
	id := "11111111-1111-1111-1111-111111111111"
	seedUserWithRole(repo, id, "multi@example.com", "password", true, 2, domain.RoleAdmin, []string{"setting.view", "user.view"})
	repo.permissionsByRole[3] = []string{"setting.view"}
	repo.assignments[id] = append(repo.assignments[id], domain.RoleAssignment{
		UserID:  id,
		RoleID:  3,
		ScopeID: 2,
		Role:    &domain.Role{ID: 3, Name: domain.RoleEditor},
		Scope:   &domain.Scope{ID: 2, Code: "campus-a", Name: "Campus A"},
	})
	svc := newTestService(repo)
	jwtManager := jwtutil.NewManager("test-secret", time.Minute*15)

	switched, err := svc.SwitchRole(id, nil, 3, 2, "jti", time.Now().Add(time.Minute), "")
	if err != nil {
		t.Fatalf("switch failed: %v", err)
	}

	refreshed, err := svc.Refresh(switched.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	claims, err := jwtManager.ParseAccessToken(refreshed.AccessToken)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.ActiveRoleID != 3 || claims.ActiveScopeID != 2 {
		t.Fatalf("expected active assignment preserved on refresh, got role=%d scope=%d", claims.ActiveRoleID, claims.ActiveScopeID)
	}
}

func assertAppErrorStatus(t *testing.T, err error, wantStatus int) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var ae *apperror.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *apperror.AppError, got %T", err)
	}
	if ae.HTTPStatus != wantStatus {
		t.Errorf("expected HTTP status %d, got %d", wantStatus, ae.HTTPStatus)
	}
}
