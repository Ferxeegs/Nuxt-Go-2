// Package pagination provides a generic GORM pagination helper reused by
// every list endpoint (users, media, activity logs, ...).
package pagination

import "gorm.io/gorm"

type Result struct {
	Page       int
	PerPage    int
	TotalData  int64
	TotalPages int
}

const (
	DefaultPage    = 1
	DefaultPerPage = 10
	MaxPerPage     = 100
)

// Normalize clamps page/perPage to sane bounds.
func Normalize(page, perPage int) (int, int) {
	if page < 1 {
		page = DefaultPage
	}
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}
	return page, perPage
}

// Paginate applies offset/limit to the given query and fills dest, returning
// pagination metadata. query should already include WHERE/ORDER BY clauses
// but not Limit/Offset.
func Paginate(query *gorm.DB, page, perPage int, dest any) (Result, error) {
	page, perPage = Normalize(page, perPage)

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return Result{}, err
	}

	offset := (page - 1) * perPage
	if err := query.Session(&gorm.Session{}).Offset(offset).Limit(perPage).Find(dest).Error; err != nil {
		return Result{}, err
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	return Result{Page: page, PerPage: perPage, TotalData: total, TotalPages: totalPages}, nil
}
