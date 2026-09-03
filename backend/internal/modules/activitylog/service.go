package activitylog

import (
	"errors"

	"gorm.io/gorm"

	"go-cp-boilerplate/internal/domain"
	"go-cp-boilerplate/internal/pkg/apperror"
	"go-cp-boilerplate/internal/pkg/pagination"
)

type Service interface {
	List(q ListQuery) ([]Response, pagination.Result, error)
	Get(id uint) (*Response, error)
	Delete(id uint) error
	ListModules() ([]string, error)
	ListActions() ([]string, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(q ListQuery) ([]Response, pagination.Result, error) {
	page, perPage := pagination.Normalize(q.Page, q.PerPage)
	items, total, err := s.repo.List(page, perPage, Filter{
		Module: q.Module,
		Action: q.Action,
		UserID: q.UserID,
		Q:      q.Q,
	})
	if err != nil {
		return nil, pagination.Result{}, apperror.Internal("Failed to list activity logs")
	}
	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}
	return toResponses(items), pagination.Result{
		Page: page, PerPage: perPage, TotalData: total, TotalPages: totalPages,
	}, nil
}

func (s *service) Get(id uint) (*Response, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Activity log not found")
		}
		return nil, apperror.Internal("Failed to fetch activity log")
	}
	resp := toResponse(*item)
	return &resp, nil
}

func (s *service) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Activity log not found")
		}
		return apperror.Internal("Failed to fetch activity log")
	}
	if err := s.repo.Delete(id); err != nil {
		return apperror.Internal("Failed to delete activity log")
	}
	return nil
}

func (s *service) ListModules() ([]string, error) {
	modules, err := s.repo.ListModules()
	if err != nil {
		return nil, apperror.Internal("Failed to list activity modules")
	}
	return modules, nil
}

func (s *service) ListActions() ([]string, error) {
	actions, err := s.repo.ListActions()
	if err != nil {
		return nil, apperror.Internal("Failed to list activity actions")
	}
	return actions, nil
}

func toResponse(item domain.ActivityLog) Response {
	resp := Response{
		ID:          item.ID,
		UserID:      item.UserID,
		Action:      item.Action,
		Module:      item.Module,
		EntityID:    item.EntityID,
		Description: item.Description,
		IPAddress:   item.IPAddress,
		CreatedAt:   item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if item.User != nil {
		resp.UserName = item.User.Name
		resp.UserEmail = item.User.Email
	}
	return resp
}

func toResponses(items []domain.ActivityLog) []Response {
	out := make([]Response, 0, len(items))
	for _, item := range items {
		out = append(out, toResponse(item))
	}
	return out
}
