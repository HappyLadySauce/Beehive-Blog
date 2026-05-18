package attachments

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/types/common"
)

func parseIDParam(ctx *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(ctx.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		common.Fail(ctx, common.NewBadRequest("invalid id", err))
		return 0, false
	}
	return id, true
}

func optionalInt64Query(ctx *gin.Context, key string) (*int64, error) {
	value := strings.TrimSpace(ctx.Query(key))
	if value == "" {
		return nil, nil
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n <= 0 {
		return nil, fmt.Errorf("%s must be a positive integer", key)
	}
	return &n, nil
}

func optionalInt64Form(ctx *gin.Context, key string) (*int64, error) {
	value := strings.TrimSpace(ctx.PostForm(key))
	if value == "" {
		return nil, nil
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n <= 0 {
		return nil, fmt.Errorf("%s must be a positive integer", key)
	}
	return &n, nil
}

func int64ListForm(ctx *gin.Context, key string) ([]int64, error) {
	values := ctx.PostFormArray(key)
	if len(values) == 0 {
		values = strings.Split(ctx.PostForm(key), ",")
	}
	out := make([]int64, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil || n <= 0 {
			return nil, fmt.Errorf("%s must contain positive integers", key)
		}
		out = append(out, n)
	}
	return out, nil
}

func optionalCursor(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("cursor must be a positive integer")
	}
	return n, nil
}

func optionalLimit(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 20, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 || n > 200 {
		return 0, fmt.Errorf("limit must be between 1 and 200")
	}
	return n, nil
}

func optionalPage(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 1, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("page must be a positive integer")
	}
	return n, nil
}

func optionalPageSize(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 20, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 || n > 200 {
		return 0, fmt.Errorf("page_size must be between 1 and 200")
	}
	return n, nil
}

func hasPageQuery(ctx *gin.Context) bool {
	_, ok := ctx.GetQuery("page")
	return ok
}

func defaultString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
