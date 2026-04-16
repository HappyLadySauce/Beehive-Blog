package gateway

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/zeromicro/go-zero/core/logx"
)

func auditSuccess(ctx context.Context, action string, fields map[string]any) {
	logx.WithContext(ctx).Infof("audit action=%s result=success %s", action, serializeFields(fields))
}

func auditFailure(ctx context.Context, action string, err error, fields map[string]any) {
	fields = cloneFields(fields)
	fields["error"] = sanitizeError(err)
	logx.WithContext(ctx).Errorf("audit action=%s result=failure %s", action, serializeFields(fields))
}

func auditRPC(ctx context.Context, action, rpcMethod string, startedAt time.Time, err error, fields map[string]any) {
	fields = cloneFields(fields)
	fields["rpc_method"] = rpcMethod
	fields["request_id"] = requestIDFromContext(ctx)
	fields["latency_ms"] = time.Since(startedAt).Milliseconds()
	if err != nil {
		auditFailure(ctx, action, err, fields)
		return
	}
	auditSuccess(ctx, action, fields)
}

func serializeFields(fields map[string]any) string {
	if len(fields) == 0 {
		return ""
	}
	parts := make([]string, 0, len(fields))
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return strings.Join(parts, " ")
}

func cloneFields(fields map[string]any) map[string]any {
	if len(fields) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(fields))
	for k, v := range fields {
		out[k] = v
	}
	return out
}

func sanitizeError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return "unknown_error"
	}
	if utf8.RuneCountInString(msg) > 160 {
		runes := []rune(msg)
		msg = string(runes[:160]) + "..."
	}
	return msg
}

func maskEmail(email string) string {
	email = strings.TrimSpace(email)
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return maskText(email, 2, 2)
	}
	return maskText(parts[0], 2, 1) + "@" + parts[1]
}

func maskAccount(account string) string {
	account = strings.TrimSpace(account)
	if strings.Contains(account, "@") {
		return maskEmail(account)
	}
	return maskText(account, 2, 2)
}

func maskToken(token string) string {
	return maskText(strings.TrimSpace(token), 6, 4)
}

func maskText(v string, keepStart, keepEnd int) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	runes := []rune(v)
	if len(runes) <= keepStart+keepEnd {
		return strings.Repeat("*", len(runes))
	}
	return string(runes[:keepStart]) + strings.Repeat("*", len(runes)-keepStart-keepEnd) + string(runes[len(runes)-keepEnd:])
}
