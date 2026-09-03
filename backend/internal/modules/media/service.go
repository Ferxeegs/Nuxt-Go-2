package media

import (
	"errors"
	"mime/multipart"
	"strings"

	"gorm.io/gorm"

	"go-cp-boilerplate/internal/domain"
	"go-cp-boilerplate/internal/pkg/apperror"
	"go-cp-boilerplate/internal/pkg/pagination"
	"go-cp-boilerplate/internal/pkg/uploader"
)

type Service interface {
	Upload(fh *multipart.FileHeader, module string, uploadedBy string) (*Response, error)
	Delete(id uint) error
	List(page, perPage int, module string) ([]Response, pagination.Result, error)
}

type service struct {
	repo      Repository
	uploader  *uploader.Uploader
	publicURL string
}

// NewService takes the base public URL prefix (e.g. "/uploads") that the
// relative stored path is joined onto to build a client-facing URL.
func NewService(repo Repository, up *uploader.Uploader, publicURL string) Service {
	return &service{repo: repo, uploader: up, publicURL: publicURL}
}

func (s *service) Upload(fh *multipart.FileHeader, module string, uploadedBy string) (*Response, error) {
	saved, err := s.uploader.Save(fh, module)
	if err != nil {
		switch {
		case errors.Is(err, uploader.ErrExtensionNotAllowed), errors.Is(err, uploader.ErrMimeNotAllowed):
			return nil, apperror.BadRequest("Only .jpg, .jpeg, .png, .webp and .pdf files are allowed")
		case errors.Is(err, uploader.ErrFileTooLarge):
			return nil, apperror.BadRequest("File exceeds the maximum allowed upload size")
		default:
			return nil, apperror.Internal("Failed to save uploaded file")
		}
	}

	asset := &domain.MediaAsset{
		FileName:      saved.FileName,
		Path:          saved.RelativePath,
		MimeType:      saved.MimeType,
		Size:          saved.Size,
		UploadedBy:    &uploadedBy,
		RelatedModule: module,
	}
	if err := s.repo.Create(asset); err != nil {
		_ = s.uploader.DeleteObject(saved.RelativePath)
		return nil, apperror.Internal("Failed to record uploaded file")
	}

	url := saved.PublicURL
	if url == "" {
		url = strings.TrimRight(s.publicURL, "/") + "/" + asset.Path
	}

	return &Response{
		ID:            asset.ID,
		FileName:      asset.FileName,
		Path:          asset.Path,
		URL:           url,
		MimeType:      asset.MimeType,
		Size:          asset.Size,
		RelatedModule: asset.RelatedModule,
		CreatedAt:     asset.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *service) Delete(id uint) error {
	asset, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Media asset not found")
		}
		return apperror.Internal("Failed to fetch media asset")
	}
	_ = s.uploader.DeleteObject(asset.Path)
	if err := s.repo.Delete(id); err != nil {
		return apperror.Internal("Failed to delete media asset")
	}
	return nil
}

func (s *service) List(page, perPage int, module string) ([]Response, pagination.Result, error) {
	page, perPage = pagination.Normalize(page, perPage)

	items, total, err := s.repo.List(page, perPage, module)
	if err != nil {
		return nil, pagination.Result{}, apperror.Internal("Failed to list media assets")
	}

	prefix := strings.TrimRight(s.publicURL, "/")
	out := make([]Response, 0, len(items))
	for _, a := range items {
		url := prefix + "/" + a.Path
		out = append(out, Response{
			ID:            a.ID,
			FileName:      a.FileName,
			Path:          a.Path,
			URL:           url,
			MimeType:      a.MimeType,
			Size:          a.Size,
			RelatedModule: a.RelatedModule,
			CreatedAt:     a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}
	return out, pagination.Result{Page: page, PerPage: perPage, TotalData: total, TotalPages: totalPages}, nil
}
