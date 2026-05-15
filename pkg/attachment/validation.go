package attachment

import (
	"fmt"
	"strings"
)

const (
	MaxUploadBytes    = 2 * 1024 * 1024 * 1024
	PresignTTLSeconds = 15 * 60
)

// ValidateCommon checks shared attachment metadata constraints.
// ValidateCommon 校验附件元数据的通用约束。
func ValidateCommon(ownerUserID *int64, purpose string, mimeType string, size int64, accessScope string) error {
	if !PurposeKnown(purpose) {
		return fmt.Errorf("%w: invalid purpose", ErrInvalid)
	}
	if ownerUserID == nil && purpose != PurposeSystem {
		return fmt.Errorf("%w: owner_user_id is required for non-system attachments", ErrInvalid)
	}
	if strings.TrimSpace(mimeType) == "" {
		return fmt.Errorf("%w: mime_type is required", ErrInvalid)
	}
	if purpose == PurposeAvatar && !strings.HasPrefix(mimeType, "image/") {
		return fmt.Errorf("%w: avatar attachments must be images", ErrInvalid)
	}
	if size <= 0 || size > MaxUploadBytes {
		return fmt.Errorf("%w: attachment size must be between 1 and %d bytes", ErrInvalid, MaxUploadBytes)
	}
	if !AccessScopeKnown(accessScope) {
		return fmt.Errorf("%w: invalid access_scope", ErrInvalid)
	}
	return nil
}

// RequireAdmin rejects non-admin attachment management callers.
// RequireAdmin 拒绝非管理员的附件管理调用方。
func RequireAdmin(actor Actor) error {
	if actor.Role != RoleAdmin {
		return ErrForbidden
	}
	return nil
}

// PurposeKnown reports whether v is a supported attachment purpose.
// PurposeKnown 判断 v 是否为受支持的附件用途。
func PurposeKnown(v string) bool {
	switch v {
	case PurposeAvatar, PurposeContent, PurposeSystem, PurposeOther:
		return true
	default:
		return false
	}
}

// AccessScopeKnown reports whether v is a supported access scope.
// AccessScopeKnown 判断 v 是否为受支持的访问范围。
func AccessScopeKnown(v string) bool {
	switch v {
	case AccessPrivate, AccessPublic:
		return true
	default:
		return false
	}
}

// StatusKnown reports whether v is a supported attachment status.
// StatusKnown 判断 v 是否为受支持的附件状态。
func StatusKnown(v string) bool {
	switch v {
	case StatusActive, StatusHidden, StatusArchived:
		return true
	default:
		return false
	}
}

// CategoryStatusKnown reports whether v is a supported category status.
// CategoryStatusKnown 判断 v 是否为受支持的分类状态。
func CategoryStatusKnown(v string) bool {
	switch v {
	case CategoryStatusActive, CategoryStatusDisabled:
		return true
	default:
		return false
	}
}
