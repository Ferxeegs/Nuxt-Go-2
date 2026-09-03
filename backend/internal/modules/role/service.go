package role

import (
	"errors"

	"gorm.io/gorm"

	"go-cp-boilerplate/internal/domain"
	"go-cp-boilerplate/internal/pkg/apperror"
)

type Service interface {
	List() ([]Response, error)
	Get(id uint) (*Response, error)
	Create(req CreateRequest) (*Response, error)
	Update(id uint, req UpdateRequest) (*Response, error)
	Delete(id uint) error
	ListPermissions() ([]PermissionResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List() ([]Response, error) {
	roles, err := s.repo.List()
	if err != nil {
		return nil, apperror.Internal("Failed to list roles")
	}
	return toResponses(roles), nil
}

func (s *service) Get(id uint) (*Response, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Role not found")
		}
		return nil, apperror.Internal("Failed to fetch role")
	}
	resp := toResponse(*r)
	return &resp, nil
}

func (s *service) Create(req CreateRequest) (*Response, error) {
	r := &domain.Role{Name: req.Name, Description: req.Description}
	if err := s.repo.Create(r); err != nil {
		return nil, apperror.Internal("Failed to create role")
	}
	if len(req.PermissionIDs) > 0 {
		if err := s.repo.ReplacePermissions(r.ID, req.PermissionIDs); err != nil {
			return nil, apperror.Internal("Failed to assign permissions")
		}
	}
	created, _ := s.repo.FindByID(r.ID)
	resp := toResponse(*created)
	return &resp, nil
}

func (s *service) Update(id uint, req UpdateRequest) (*Response, error) {
	r, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Role not found")
		}
		return nil, apperror.Internal("Failed to fetch role")
	}

	if req.Name != "" {
		r.Name = req.Name
	}
	if req.Description != "" {
		r.Description = req.Description
	}
	if err := s.repo.Update(r); err != nil {
		return nil, apperror.Internal("Failed to update role")
	}

	if req.PermissionIDs != nil {
		if err := s.repo.ReplacePermissions(r.ID, *req.PermissionIDs); err != nil {
			return nil, apperror.Internal("Failed to update permissions")
		}
	}

	updated, _ := s.repo.FindByID(r.ID)
	resp := toResponse(*updated)
	return &resp, nil
}

func (s *service) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Role not found")
		}
		return apperror.Internal("Failed to fetch role")
	}
	if err := s.repo.Delete(id); err != nil {
		return apperror.Internal("Failed to delete role")
	}
	return nil
}

func (s *service) ListPermissions() ([]PermissionResponse, error) {
	perms, err := s.repo.AllPermissions()
	if err != nil {
		return nil, apperror.Internal("Failed to list permissions")
	}
	out := make([]PermissionResponse, 0, len(perms))
	for _, p := range perms {
		out = append(out, PermissionResponse{ID: p.ID, Name: p.Name, Description: p.Description})
	}
	return out, nil
}

func toResponse(r domain.Role) Response {
	perms := make([]string, 0, len(r.Permissions))
	for _, p := range r.Permissions {
		perms = append(perms, p.Name)
	}
	return Response{ID: r.ID, Name: r.Name, Description: r.Description, Permissions: perms}
}

func toResponses(roles []domain.Role) []Response {
	out := make([]Response, 0, len(roles))
	for _, r := range roles {
		out = append(out, toResponse(r))
	}
	return out
}
