package v1

import "time"

// AttachmentResponse is the public metadata shape for an attachment row.
// AttachmentResponse 是附件行的对外元数据结构。
type AttachmentResponse struct {
	ID           int64      `json:"id"`
	OwnerUserID  *int64     `json:"owner_user_id,omitempty"`
	Purpose      string     `json:"purpose"`
	Filename     string     `json:"filename"`
	OriginalName *string    `json:"original_name,omitempty"`
	MimeType     string     `json:"mime_type"`
	Size         int64      `json:"size"`
	StorageType  string     `json:"storage_type"`
	Bucket       *string    `json:"bucket,omitempty"`
	ObjectKey    *string    `json:"object_key,omitempty"`
	LocalPath    *string    `json:"local_path,omitempty"`
	ETag         *string    `json:"etag,omitempty"`
	Checksum     *string    `json:"checksum,omitempty"`
	AccessScope  string     `json:"access_scope"`
	UploadStatus string     `json:"upload_status"`
	Status       string     `json:"status"`
	CategoryIDs  []int64    `json:"category_ids,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// AttachmentListResponse returns cursor-paginated attachments.
// AttachmentListResponse 返回游标分页附件列表。
type AttachmentListResponse struct {
	Items      []AttachmentResponse `json:"items"`
	NextCursor string               `json:"next_cursor,omitempty"`
}

// AttachmentPresignRequest creates a pending remote attachment.
// AttachmentPresignRequest 创建 pending 状态远端附件。
type AttachmentPresignRequest struct {
	StorageType  string  `json:"storage_type" binding:"required,oneof=s3 oss"`
	OwnerUserID  *int64  `json:"owner_user_id,omitempty"`
	Purpose      string  `json:"purpose" binding:"required"`
	Filename     string  `json:"filename" binding:"required,max=255"`
	OriginalName *string `json:"original_name,omitempty"`
	MimeType     string  `json:"mime_type" binding:"required,max=127"`
	Size         int64   `json:"size" binding:"required,min=1"`
	AccessScope  string  `json:"access_scope" binding:"required,oneof=private public"`
	Checksum     *string `json:"checksum,omitempty"`
	CategoryIDs  []int64 `json:"category_ids,omitempty"`
}

// AttachmentPresignResponse returns upload instructions for direct upload.
// AttachmentPresignResponse 返回直传上传指令。
type AttachmentPresignResponse struct {
	Attachment AttachmentResponse `json:"attachment"`
	UploadURL  string             `json:"upload_url"`
	Method     string             `json:"method"`
	Headers    map[string]string  `json:"headers,omitempty"`
	ExpiresAt  time.Time          `json:"expires_at"`
}

// AttachmentCompleteRequest marks a pending remote attachment as ready.
// AttachmentCompleteRequest 将 pending 远端附件标记为 ready。
type AttachmentCompleteRequest struct {
	ETag     *string `json:"etag,omitempty"`
	Checksum *string `json:"checksum,omitempty"`
	Size     *int64  `json:"size,omitempty"`
}

// AttachmentPatchRequest updates mutable attachment metadata.
// AttachmentPatchRequest 更新可变附件元数据。
type AttachmentPatchRequest struct {
	OriginalName *string  `json:"original_name,omitempty"`
	Status       *string  `json:"status,omitempty"`
	AccessScope  *string  `json:"access_scope,omitempty"`
	CategoryIDs  *[]int64 `json:"category_ids,omitempty"`
}

// AttachmentCategoryReplaceRequest replaces all category bindings for one attachment.
// AttachmentCategoryReplaceRequest 替换单个附件的全部分类绑定。
type AttachmentCategoryReplaceRequest struct {
	CategoryIDs []int64 `json:"category_ids"`
}

// AttachmentCategoryResponse is the API shape for one attachment category.
// AttachmentCategoryResponse 是单个附件分类的 API 结构。
type AttachmentCategoryResponse struct {
	ID          int64      `json:"id"`
	ParentID    *int64     `json:"parent_id,omitempty"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description *string    `json:"description,omitempty"`
	Icon        *string    `json:"icon,omitempty"`
	Path        string     `json:"path"`
	Depth       int        `json:"depth"`
	SortOrder   int        `json:"sort_order"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// AttachmentCategoryListResponse returns attachment categories.
// AttachmentCategoryListResponse 返回附件分类列表。
type AttachmentCategoryListResponse struct {
	Items []AttachmentCategoryResponse `json:"items"`
}

// AttachmentCategoryCreateRequest creates a category.
// AttachmentCategoryCreateRequest 创建附件分类。
type AttachmentCategoryCreateRequest struct {
	ParentID    *int64  `json:"parent_id,omitempty"`
	Name        string  `json:"name" binding:"required,max=64"`
	Slug        string  `json:"slug" binding:"required,max=64"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	SortOrder   int     `json:"sort_order"`
	Status      string  `json:"status" binding:"omitempty,oneof=active disabled"`
}

// AttachmentCategoryPatchRequest updates a category.
// AttachmentCategoryPatchRequest 更新附件分类。
type AttachmentCategoryPatchRequest struct {
	ParentID    *int64  `json:"parent_id,omitempty"`
	Name        *string `json:"name,omitempty"`
	Slug        *string `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
	Status      *string `json:"status,omitempty"`
}
