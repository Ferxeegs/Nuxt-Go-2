package websetting

import (
	"encoding/json"
	"strings"

	"go-cp-boilerplate/internal/domain"
	"go-cp-boilerplate/internal/pkg/apperror"
)

type Service interface {
	All() ([]Response, error)
	Upsert(req UpsertRequest) ([]Response, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) All() ([]Response, error) {
	items, err := s.repo.All()
	if err != nil {
		return nil, apperror.Internal("Failed to fetch settings")
	}
	return toResponses(items), nil
}

func (s *service) Upsert(req UpsertRequest) ([]Response, error) {
	for _, item := range req.Settings {
		typ := inferSettingType(item.Key, item.Value, item.Type)
		if err := s.repo.Upsert(item.Key, item.Value, typ); err != nil {
			return nil, apperror.Internal("Failed to update setting: " + item.Key)
		}
	}
	return s.All()
}

func inferSettingType(key, value, explicit string) string {
	k := strings.ToLower(strings.TrimSpace(key))
	known := map[string]string{
		"site_name":        "string",
		"logo":             "image",
		"contact_email":    "string",
		"contact_phone":    "string",
		"social_facebook":  "string",
		"footer_text":      "string",
	}
	if t, ok := known[k]; ok {
		return t
	}
	switch {
	case matchSettingToken(k, "logo", "image", "avatar", "banner", "icon", "thumbnail", "cover"):
		return "image"
	case matchSettingToken(k, "json", "config", "meta") || looksLikeJSON(value):
		return "json"
	case matchSettingToken(k, "text", "body", "content", "description", "html", "bio") || strings.Contains(value, "\n"):
		return "text"
	case explicit == "string" || explicit == "text" || explicit == "json" || explicit == "image":
		// Fallback for unknown keys only when client still sends a valid type.
		return explicit
	default:
		return "string"
	}
}

func matchSettingToken(key string, tokens ...string) bool {
	for _, token := range tokens {
		if key == token || strings.HasPrefix(key, token+"_") || strings.HasSuffix(key, "_"+token) || strings.Contains(key, "_"+token+"_") {
			return true
		}
	}
	return false
}

func looksLikeJSON(value string) bool {
	v := strings.TrimSpace(value)
	if v == "" || (v[0] != '{' && v[0] != '[') {
		return false
	}
	var raw any
	return json.Unmarshal([]byte(v), &raw) == nil
}

func toResponses(items []domain.WebSetting) []Response {
	out := make([]Response, 0, len(items))
	for _, i := range items {
		out = append(out, Response{Key: i.Key, Value: i.Value, Type: i.Type})
	}
	return out
}
