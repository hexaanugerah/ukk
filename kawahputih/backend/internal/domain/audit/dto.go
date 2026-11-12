package audit

import (
	"encoding/json"
	"time"

	"github.com/KawahPutih/backend/pkg/apperror"
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/go-playground/validator/v10"
)

// AuditLogResponse adalah representasi API dari satu audit log.
type AuditLogResponse struct {
	ID         string                 `json:"id"`
	ActorID    string                 `json:"actor_id"`
	ActorRole  string                 `json:"actor_role"`
	Action     string                 `json:"action"`
	TargetType string                 `json:"target_type"`
	TargetID   string                 `json:"target_id"`
	Status     string                 `json:"status"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

func toAuditLogResponse(log *AuditLog) AuditLogResponse {
	var meta map[string]interface{}
	if log.Metadata != "" {
		_ = json.Unmarshal([]byte(log.Metadata), &meta)
	}
	return AuditLogResponse{
		ID:         log.ID,
		ActorID:    log.ActorID,
		ActorRole:  log.ActorRole,
		Action:     log.Action,
		TargetType: log.TargetType,
		TargetID:   log.TargetID,
		Status:     log.Status,
		Metadata:   meta,
		IPAddress:  log.IPAddress,
		UserAgent:  log.UserAgent,
		CreatedAt:  log.CreatedAt,
	}
}

var validate = validator.New()

func ValidateStruct(s interface{}) []response.FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	var out []response.FieldError
	for _, fe := range err.(validator.ValidationErrors) {
		out = append(out, response.FieldError{Field: fe.Field(), Message: fe.Field() + " is invalid"})
	}
	return out
}

// ListRequest adalah parameter query untuk daftar audit log.
type ListRequest struct {
	Page       int    `json:"page" form:"page"`
	Limit      int    `json:"limit" form:"limit"`
	ActorID    string `json:"actor_id" form:"actor_id"`
	Action     string `json:"action" form:"action"`
	TargetType string `json:"target_type" form:"target_type"`
	TargetID   string `json:"target_id" form:"target_id"`
	Status     string `json:"status" form:"status"`
	From       string `json:"from" form:"from"` // YYYY-MM-DD
	To         string `json:"to" form:"to"`     // YYYY-MM-DD
}

// StatsRequest adalah parameter query untuk statistik audit.
type StatsRequest struct {
	From string `json:"from" form:"from"`
	To   string `json:"to" form:"to"`
}

var (
	ErrNotFound = apperror.NotFound("Data audit log tidak ditemukan.").WithCode("AUDIT_001")
)

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
