package websetting

import (
	"gorm.io/gorm"

	"go-cp-boilerplate/internal/domain"
)

type Repository interface {
	All() ([]domain.WebSetting, error)
	Upsert(key, value, typ string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) All() ([]domain.WebSetting, error) {
	var items []domain.WebSetting
	err := r.db.Order("`key` ASC").Find(&items).Error
	return items, err
}

func (r *repository) Upsert(key, value, typ string) error {
	var existing domain.WebSetting
	err := r.db.Where("`key` = ?", key).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		if typ == "" {
			typ = "string"
		}
		return r.db.Create(&domain.WebSetting{Key: key, Value: value, Type: typ}).Error
	}
	if err != nil {
		return err
	}

	existing.Value = value
	if typ != "" {
		existing.Type = typ
	}
	return r.db.Save(&existing).Error
}
