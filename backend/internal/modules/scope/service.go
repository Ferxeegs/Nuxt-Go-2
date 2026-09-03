package scope

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"go-cp-boilerplate/internal/domain"
	"go-cp-boilerplate/internal/pkg/apperror"
	"go-cp-boilerplate/internal/pkg/pagination"
)

type Service interface {
	List(q ListQuery) ([]Response, pagination.Result, error)
	Get(id uint) (*Response, error)
	Create(req CreateRequest) (*Response, error)
	Update(id uint, req UpdateRequest) (*Response, error)
	Delete(id uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(q ListQuery) ([]Response, pagination.Result, error) {
	page, perPage := pagination.Normalize(q.Page, q.PerPage)
	items, total, err := s.repo.List(page, perPage, q.Search)
	if err != nil {
		return nil, pagination.Result{}, apperror.Internal("Failed to list scopes")
	}
	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}
	return toResponses(items), pagination.Result{Page: page, PerPage: perPage, TotalData: total, TotalPages: totalPages}, nil
}

func (s *service) Get(id uint) (*Response, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Scope not found")
		}
		return nil, apperror.Internal("Failed to fetch scope")
	}
	resp := toResponse(*m)
	return &resp, nil
}

func (s *service) Create(req CreateRequest) (*Response, error) {
	code := strings.ToLower(strings.TrimSpace(req.Code))
	if _, err := s.repo.FindByCode(code); err == nil {
		return nil, apperror.Conflict("Scope code already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.Internal("Failed to validate scope code")
	}

	if req.ParentID != nil {
		if _, err := s.repo.FindByID(*req.ParentID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperror.BadRequest("Parent scope not found")
			}
			return nil, apperror.Internal("Failed to validate parent scope")
		}
	}

	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}

	m := &domain.Scope{
		Code:        code,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		ParentID:    req.ParentID,
		IsActive:    active,
	}
	if err := s.repo.Create(m); err != nil {
		return nil, apperror.Internal("Failed to create scope")
	}
	created, _ := s.repo.FindByID(m.ID)
	resp := toResponse(*created)
	return &resp, nil
}

func (s *service) Update(id uint, req UpdateRequest) (*Response, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Scope not found")
		}
		return nil, apperror.Internal("Failed to fetch scope")
	}

	if req.Code != "" {
		code := strings.ToLower(strings.TrimSpace(req.Code))
		if code != m.Code {
			if existing, err := s.repo.FindByCode(code); err == nil && existing.ID != m.ID {
				return nil, apperror.Conflict("Scope code already exists")
			} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperror.Internal("Failed to validate scope code")
			}
			m.Code = code
		}
	}
	if req.Name != "" {
		m.Name = strings.TrimSpace(req.Name)
	}
	if req.Description != "" {
		m.Description = strings.TrimSpace(req.Description)
	}
	if req.ClearParent {
		m.ParentID = nil
	} else if req.ParentID != nil {
		if *req.ParentID == m.ID {
			return nil, apperror.BadRequest("Scope cannot be its own parent")
		}
		if _, err := s.repo.FindByID(*req.ParentID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperror.BadRequest("Parent scope not found")
			}
			return nil, apperror.Internal("Failed to validate parent scope")
		}
		m.ParentID = req.ParentID
	}
	if req.IsActive != nil {
		m.IsActive = *req.IsActive
	}

	if err := s.repo.Update(m); err != nil {
		return nil, apperror.Internal("Failed to update scope")
	}
	updated, _ := s.repo.FindByID(m.ID)
	resp := toResponse(*updated)
	return &resp, nil
}

func (s *service) Delete(id uint) error {
	m, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Scope not found")
		}
		return apperror.Internal("Failed to fetch scope")
	}
	if m.Code == domain.ScopeCodeGlobal {
		return apperror.Forbidden("Cannot delete the global scope")
	}

	children, err := s.repo.CountChildren(id)
	if err != nil {
		return apperror.Internal("Failed to check child scopes")
	}
	if children > 0 {
		return apperror.Conflict("Scope has child scopes; reassign or delete them first")
	}

	assignments, err := s.repo.CountAssignments(id)
	if err != nil {
		return apperror.Internal("Failed to check role assignments")
	}
	if assignments > 0 {
		return apperror.Conflict("Scope has role assignments; remove them first")
	}

	if err := s.repo.Delete(id); err != nil {
		return apperror.Internal("Failed to delete scope")
	}
	return nil
}

func toResponse(m domain.Scope) Response {
	resp := Response{
		ID:          m.ID,
		Code:        m.Code,
		Name:        m.Name,
		Description: m.Description,
		ParentID:    m.ParentID,
		IsActive:    m.IsActive,
	}
	if m.Parent != nil {
		resp.ParentCode = m.Parent.Code
		resp.ParentName = m.Parent.Name
	}
	if !m.CreatedAt.IsZero() {
		resp.CreatedAt = m.CreatedAt.Format(time.RFC3339)
	}
	return resp
}

func toResponses(items []domain.Scope) []Response {
	out := make([]Response, 0, len(items))
	for _, m := range items {
		out = append(out, toResponse(m))
	}
	return out
}
