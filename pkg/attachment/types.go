package attachment

import (
	"errors"
	"io"

	"github.com/HappyLadySauce/Beehive-Blog/pkg/attachment/driver"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

const (
	RoleAdmin = "admin"

	PurposeAvatar  = "avatar"
	PurposeContent = "content"
	PurposeSystem  = "system"
	PurposeOther   = "other"

	AccessPrivate = "private"
	AccessPublic  = "public"

	UploadPending = "pending"
	UploadReady   = "ready"
	UploadFailed  = "failed"

	StatusActive   = "active"
	StatusHidden   = "hidden"
	StatusArchived = "archived"

	CategoryStatusActive   = "active"
	CategoryStatusDisabled = "disabled"
)

var (
	ErrForbidden = errors.New("attachment operation is forbidden")
	ErrNotFound  = errors.New("attachment resource not found")
	ErrInvalid   = errors.New("attachment request is invalid")
	ErrConflict  = errors.New("attachment resource conflict")
)

// Actor is the authenticated caller used for authorization decisions.
// Actor 表示用于授权判断的已认证调用方。
type Actor struct {
	UID  int64
	Role string
}

// LocalUploadInput describes a multipart upload accepted by an admin.
// LocalUploadInput 描述管理员提交的 multipart 上传。
type LocalUploadInput struct {
	OwnerUserID    *int64
	Purpose        string
	Filename       string
	OriginalName   *string
	MimeType       string
	Size           int64
	Reader         io.Reader
	AccessScope    string
	CategoryIDs    []int64
	StorageMountID *int64
}

// RemotePresignInput describes a remote direct-upload request.
// RemotePresignInput 描述远端直传请求。
type RemotePresignInput struct {
	OwnerUserID    *int64
	Purpose        string
	Filename       string
	OriginalName   *string
	MimeType       string
	Size           int64
	AccessScope    string
	Checksum       *string
	CategoryIDs    []int64
	StorageMountID *int64
}

// PresignOutput returns the created pending row and upload instructions.
// PresignOutput 返回已创建的 pending 行与上传指令。
type PresignOutput struct {
	Attachment model.Attachment
	Upload     driver.PresignResult
}

// CompleteInput describes upload confirmation metadata.
// CompleteInput 描述上传确认元数据。
type CompleteInput struct {
	ETag     *string
	Checksum *string
	Size     *int64
}

// PatchInput describes mutable attachment fields.
// PatchInput 描述附件可变字段。
type PatchInput struct {
	OriginalName *string
	Status       *string
	AccessScope  *string
	CategoryIDs  *[]int64
}

// ListInput describes admin attachment listing filters.
// ListInput 描述管理员附件列表过滤条件。
type ListInput struct {
	OwnerUserID *int64
	Purpose     string
	Status      string
	CategoryID  *int64
	CursorID    int64
	Limit       int
}

// ContentResult describes how a client should receive attachment content.
// ContentResult 描述客户端获取附件内容的方式。
type ContentResult struct {
	Attachment  model.Attachment
	LocalPath   string
	RedirectURL string
}

// CategoryCreateInput describes a new category.
// CategoryCreateInput 描述新分类。
type CategoryCreateInput struct {
	ParentID    *int64
	Name        string
	Slug        string
	Description *string
	Icon        *string
	SortOrder   int
	Status      string
}

// CategoryPatchInput describes category updates.
// CategoryPatchInput 描述分类更新。
type CategoryPatchInput struct {
	ParentID    *int64
	Name        *string
	Slug        *string
	Description *string
	Icon        *string
	SortOrder   *int
	Status      *string
}
