package comments

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	v1 "github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/api/v1"
	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/model"
)

func parseContentID(ctx *gin.Context) (int64, bool) {
	raw := ctx.Param("id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		common.Fail(ctx, common.NewBadRequest("invalid content id", fmt.Errorf("parse: %w", err)))
		return 0, false
	}
	return id, true
}

func toCommentItem(row model.Comment) v1.CommentItem {
	return v1.CommentItem{
		ID:        row.ID,
		ContentID: row.ContentID,
		Nickname:  row.Nickname,
		Website:   row.Website,
		Body:      row.Body,
		CreatedAt: row.CreatedAt,
	}
}

func normalizeEmailHash(email string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email))))
	return hex.EncodeToString(sum[:])
}

func cleanOptionalWebsite(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	value := strings.TrimSpace(*raw)
	if value == "" {
		return nil, nil
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, common.NewBadRequest("invalid website URL", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, common.NewBadRequest("invalid website URL", fmt.Errorf("unsupported scheme"))
	}
	return &value, nil
}

func mapCommentFirstError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return common.NewNotFound("content not found", err)
	}
	return common.NewInternal("failed to fetch content", err)
}

func nowUTC() time.Time {
	return time.Now().UTC()
}
