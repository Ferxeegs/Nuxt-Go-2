package activitylogger

import (
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"go-cp-boilerplate/internal/domain"
)

// Logger writes activity_logs rows. Failures are logged and ignored so
// audit recording never blocks the primary request path.
type Logger struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Logger {
	return &Logger{db: db}
}

type Entry struct {
	UserID      *string
	Action      string
	Module      string
	EntityID    *string
	Description string
	IPAddress   string
}

func (l *Logger) Record(entry Entry) {
	if l == nil || l.db == nil {
		return
	}
	row := &domain.ActivityLog{
		UserID:      entry.UserID,
		Action:      entry.Action,
		Module:      entry.Module,
		EntityID:    entry.EntityID,
		Description: entry.Description,
		IPAddress:   entry.IPAddress,
	}
	if err := l.db.Create(row).Error; err != nil {
		log.Warn().Err(err).Str("action", entry.Action).Str("module", entry.Module).Msg("failed to write activity log")
	}
}
