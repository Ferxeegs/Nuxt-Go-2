package user

import (
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"go-cp-boilerplate/internal/domain"
	"go-cp-boilerplate/internal/pkg/apperror"
	"go-cp-boilerplate/internal/pkg/hash"
	"go-cp-boilerplate/internal/pkg/pagination"
)

type Service interface {
	List(q ListQuery) ([]Response, pagination.Result, error)
	Get(id string) (*Response, error)
	Create(req CreateRequest, bcryptCost int, actorRole string) (*Response, error)
	Update(id string, req UpdateRequest, actorRole string) (*Response, error)
	ResetPassword(id string, bcryptCost int, defaultPassword string) error
	Delete(id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(q ListQuery) ([]Response, pagination.Result, error) {
	page, perPage := pagination.Normalize(q.Page, q.PerPage)

	users, total, err := s.repo.List(page, perPage, q.Search)
	if err != nil {
		return nil, pagination.Result{}, apperror.Internal("Failed to list users")
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	out := make([]Response, 0, len(users))
	for _, u := range users {
		resp, err := s.toResponseWithAssignments(u)
		if err != nil {
			return nil, pagination.Result{}, err
		}
		out = append(out, *resp)
	}

	return out, pagination.Result{Page: page, PerPage: perPage, TotalData: total, TotalPages: totalPages}, nil
}

func (s *service) Get(id string) (*Response, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("User not found")
		}
		return nil, apperror.Internal("Failed to fetch user")
	}
	return s.toResponseWithAssignments(*u)
}

func (s *service) Create(req CreateRequest, bcryptCost int, actorRole string) (*Response, error) {
	if len(req.Assignments) == 0 {
		return nil, apperror.BadRequest("At least one role assignment is required")
	}

	if _, err := s.repo.FindByEmail(req.Email); err == nil {
		return nil, apperror.Conflict("Email is already registered")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.Internal("Failed to validate email")
	}

	hashed, err := hash.Hash(req.Password, bcryptCost)
	if err != nil {
		return nil, apperror.Internal("Failed to hash password")
	}

	u := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashed,
		IsActive: true,
	}
	if err := s.repo.Create(u); err != nil {
		return nil, apperror.Internal("Failed to create user")
	}

	if err := s.applyAssignments(u.ID, req.Assignments, actorRole); err != nil {
		_ = s.repo.Delete(u.ID)
		return nil, err
	}

	return s.Get(u.ID)
}

func (s *service) Update(id string, req UpdateRequest, actorRole string) (*Response, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("User not found")
		}
		return nil, apperror.Internal("Failed to fetch user")
	}

	if req.Name != "" {
		u.Name = req.Name
	}
	if req.Email != "" && req.Email != u.Email {
		if _, err := s.repo.FindByEmail(req.Email); err == nil {
			return nil, apperror.Conflict("Email is already registered")
		}
		u.Email = req.Email
	}
	if req.IsActive != nil {
		u.IsActive = *req.IsActive
	}

	if err := s.repo.Update(u); err != nil {
		return nil, apperror.Internal("Failed to update user")
	}

	if req.Assignments != nil {
		if len(*req.Assignments) == 0 {
			return nil, apperror.BadRequest("At least one role assignment is required")
		}
		if err := s.applyAssignments(u.ID, *req.Assignments, actorRole); err != nil {
			return nil, err
		}
	}

	return s.Get(u.ID)
}

func (s *service) ResetPassword(id string, bcryptCost int, defaultPassword string) error {
	if len(strings.TrimSpace(defaultPassword)) < 6 {
		return apperror.Internal("DEFAULT_PASSWORD is not configured")
	}

	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("User not found")
		}
		return apperror.Internal("Failed to fetch user")
	}

	hashed, err := hash.Hash(defaultPassword, bcryptCost)
	if err != nil {
		return apperror.Internal("Failed to hash password")
	}

	if err := s.repo.UpdatePassword(id, hashed); err != nil {
		return apperror.Internal("Failed to reset password")
	}
	_ = s.repo.RevokeAllRefreshTokens(id)
	return nil
}

func (s *service) Delete(id string) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("User not found")
		}
		return apperror.Internal("Failed to fetch user")
	}
	if err := s.repo.Delete(id); err != nil {
		return apperror.Internal("Failed to delete user")
	}
	return nil
}

func (s *service) applyAssignments(userID string, inputs []AssignmentInput, actorRole string) error {
	seen := map[string]struct{}{}
	items := make([]domain.RoleAssignment, 0, len(inputs))

	for _, in := range inputs {
		if in.RoleID == 0 || in.ScopeID == 0 {
			return apperror.BadRequest("Each assignment requires role_id and scope_id")
		}
		dedupeKey := formatAssignmentKey(in.RoleID, in.ScopeID)
		if _, ok := seen[dedupeKey]; ok {
			continue
		}
		seen[dedupeKey] = struct{}{}

		role, err := s.repo.FindRoleByID(in.RoleID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperror.BadRequest("Assignment role not found")
			}
			return apperror.Internal("Failed to validate assignment role")
		}
		if role.Name == domain.RoleSuperAdmin && actorRole != domain.RoleSuperAdmin {
			return apperror.Forbidden("Only SuperAdmin can assign SuperAdmin via role assignments")
		}

		if _, err := s.repo.FindScopeByID(in.ScopeID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperror.BadRequest("Assignment scope not found")
			}
			return apperror.Internal("Failed to validate assignment scope")
		}

		items = append(items, domain.RoleAssignment{
			UserID:  userID,
			RoleID:  in.RoleID,
			ScopeID: in.ScopeID,
		})
	}

	if len(items) == 0 {
		return apperror.BadRequest("At least one role assignment is required")
	}

	if err := s.repo.ReplaceAssignments(userID, items); err != nil {
		return apperror.Internal("Failed to save role assignments")
	}
	return nil
}

func formatAssignmentKey(roleID, scopeID uint) string {
	return strconv.FormatUint(uint64(roleID), 10) + ":" + strconv.FormatUint(uint64(scopeID), 10)
}

func (s *service) toResponseWithAssignments(u domain.User) (*Response, error) {
	resp := toResponse(u)
	assignments, err := s.repo.ListAssignments(u.ID)
	if err != nil {
		return nil, apperror.Internal("Failed to load role assignments")
	}
	resp.Assignments = toAssignmentResponses(assignments)
	resp.Roles = uniqueRoles(assignments)
	return &resp, nil
}

func toResponse(u domain.User) Response {
	resp := Response{
		ID:          u.ID,
		Name:        u.Name,
		Email:       u.Email,
		IsActive:    u.IsActive,
		Roles:       []RoleSummary{},
		Assignments: []AssignmentResponse{},
	}
	if u.Avatar != nil {
		resp.Avatar = *u.Avatar
	}
	return resp
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

func uniqueRoles(items []domain.RoleAssignment) []RoleSummary {
	seen := map[uint]struct{}{}
	out := make([]RoleSummary, 0)
	for _, a := range items {
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

func ParseUserID(raw string) (string, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return "", apperror.BadRequest("Invalid user ID")
	}
	return id.String(), nil
}
