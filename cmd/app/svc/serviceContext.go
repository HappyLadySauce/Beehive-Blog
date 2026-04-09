package svc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/models"
	"github.com/HappyLadySauce/Beehive-Blog/pkg/mailer"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"k8s.io/klog/v2"

	"github.com/HappyLadySauce/Beehive-Blog/cmd/app/options"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config options.Options
	DB     *gorm.DB
	Redis  *redis.Client
	// Mailer 为 nil 时表示 SMTP 未配置，调用方应跳过邮件发送。
	Mailer *mailer.SMTPMailer
}

// NewServiceContext creates a new ServiceContext
// 创建一个新的服务上下文
// 初始化数据库连接和 Redis 连接
func NewServiceContext(c options.Options) (*ServiceContext, error) {
	// 构建 PostgreSQL DSN（使用 keyword/value 格式）
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s connect_timeout=%d",
		c.DatabaseOptions.Host,
		c.DatabaseOptions.Username,
		c.DatabaseOptions.Password,
		c.DatabaseOptions.Database,
		c.DatabaseOptions.Port,
		c.DatabaseOptions.SSLMode,
		c.DatabaseOptions.TimeZone,
		c.DatabaseOptions.ConnectTimeoutSeconds,
	)
	if c.DatabaseOptions.SSLMode == "disable" {
		klog.Warning("PostgreSQL is using sslmode=disable; this is insecure for non-local environments")
	}

	// 初始化数据库连接
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// 获取底层 sql.DB 并配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		// best-effort：db 已创建但无法取到底层 sql.DB 时，尝试关闭连接池避免泄漏
		_ = closeGormConnPool(db)
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	// 设置连接池参数
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// 使用模型自动迁移数据库结构
	if c.DatabaseOptions.AutoMigrate {
		if err := autoMigrateModels(db); err != nil {
			_ = sqlDB.Close()
			return nil, fmt.Errorf("failed to auto migrate models: %w", err)
		}
	} else {
		klog.Info("Database auto migration is disabled by configuration")
	}

	if err := ensureDefaultCategory(db); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("failed to ensure default category: %w", err)
	}

	if err := ensureDefaultStoragePolicy(db, c.StorageOptions.UploadDir); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("failed to ensure default storage policy: %w", err)
	}

	// 初始化 Redis 连接
	redisOptions := &redis.Options{
		Addr:         c.RedisOptions.RedisHost,
		Password:     c.RedisOptions.RedisPass,
		DB:           c.RedisOptions.RedisDB,
		DialTimeout:  time.Duration(c.RedisOptions.DialTimeoutSeconds) * time.Second,
		ReadTimeout:  time.Duration(c.RedisOptions.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(c.RedisOptions.WriteTimeoutSeconds) * time.Second,
	}
	if c.RedisOptions.EnableTLS {
		redisOptions.TLSConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: c.RedisOptions.InsecureSkipVerify,
		}
		if c.RedisOptions.InsecureSkipVerify {
			klog.Warning("Redis TLS cert verification is disabled; this is insecure for production")
		}
	}
	client := redis.NewClient(redisOptions)
	// 测试 Redis 连接是否成功
	redisPingCtx, cancel := context.WithTimeout(context.Background(), time.Duration(c.RedisOptions.ConnectTimeoutSeconds)*time.Second)
	defer cancel()
	if _, err := client.Ping(redisPingCtx).Result(); err != nil {
		_ = client.Close()
		_ = sqlDB.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	svcCtx := &ServiceContext{
		Config: c,
		DB:     db,
		Redis:  client,
	}

	// 尝试从 settings 表加载 SMTP 配置并初始化 Mailer；失败时仅记录警告，不阻断启动。
	if m, err := loadMailerFromDB(db); err != nil {
		klog.Warningf("SMTP mailer init skipped: %v", err)
	} else {
		svcCtx.Mailer = m
	}

	return svcCtx, nil
}

// RebuildMailer 从 settings 表重新加载 SMTP 配置并替换 Mailer（设置更新后调用）。
func (s *ServiceContext) RebuildMailer() {
	m, err := loadMailerFromDB(s.DB)
	if err != nil {
		klog.Warningf("RebuildMailer: %v", err)
		s.Mailer = nil
		return
	}
	s.Mailer = m
	klog.InfoS("SMTP mailer rebuilt successfully")
}

// loadMailerFromDB 从 settings 表 group=smtp 读取配置并构建 SMTPMailer。
func loadMailerFromDB(db *gorm.DB) (*mailer.SMTPMailer, error) {
	if db == nil {
		return nil, errors.New("nil db")
	}
	var rows []models.Setting
	if err := db.Where(`"group" = ?`, models.SettingGroupSMTP).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query smtp settings: %w", err)
	}
	kv := make(map[string]string, len(rows))
	for _, r := range rows {
		kv[r.Key] = r.Value
	}
	cfg := mailer.ConfigFromSMTPSettings(kv)
	if !cfg.IsValid() {
		return nil, errors.New("smtp settings not fully configured")
	}
	return mailer.New(cfg)
}

func autoMigrateModels(db *gorm.DB) error {
	modelsToMigrate := []interface{}{
		&models.User{},
		&models.UserLevel{},
		&models.Category{},
		&models.Tag{},
		&models.Article{},
		&models.ArticleVersion{},
		&models.ArticleTag{},
		&models.ArticleLike{},
		&models.UserFavorite{},
		&models.ArticleViewLog{},
		&models.Comment{},
		&models.CommentLike{},
		&models.AttachmentGroup{},
		&models.StoragePolicy{},
		&models.Attachment{},
		&models.ArticleAttachment{},
		&models.Setting{},
		&models.Link{},
		&models.OperationLog{},
		&models.Backup{},
		&models.Theme{},
		&models.Menu{},
		&models.MenuItem{},
		&models.Page{},
		&models.Notification{},
		&models.NotificationSetting{},
		&models.Subscription{},
		&models.Webhook{},
		&models.WebhookLog{},
	}

	for _, model := range modelsToMigrate {
		if err := db.AutoMigrate(model); err != nil {
			return err
		}
	}

	klog.InfoS("Database auto migration completed", "modelCount", len(modelsToMigrate))
	return nil
}

const defaultCategorySlug = "default"

// ensureDefaultStoragePolicy 保证存在一条 type=local、is_default=true 的本地存储策略（幂等）。
func ensureDefaultStoragePolicy(db *gorm.DB, uploadPath string) error {
	if db == nil {
		return errors.New("nil db")
	}
	var existing models.StoragePolicy
	err := db.Where("type = ? AND is_default = true", models.StoragePolicyLocal).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if uploadPath == "" {
		uploadPath = "uploads"
	}
	policy := &models.StoragePolicy{
		Name:       "本地存储",
		Type:       models.StoragePolicyLocal,
		IsDefault:  true,
		UploadPath: uploadPath,
		SortOrder:  0,
	}
	if err := db.Create(policy).Error; err != nil {
		low := strings.ToLower(err.Error())
		if strings.Contains(low, "duplicate") || strings.Contains(low, "unique") {
			return nil
		}
		return err
	}
	klog.InfoS("Default storage policy ensured", "id", policy.ID, "uploadPath", uploadPath)
	return nil
}

// ensureDefaultCategory 保证存在 slug=default 的兜底分类（幂等；与 db/007_seed.sql 一致）。
func ensureDefaultCategory(db *gorm.DB) error {
	if db == nil {
		return errors.New("nil db")
	}
	var existing models.Category
	err := db.Where("slug = ?", defaultCategorySlug).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	category := &models.Category{
		Name:        "默认分类",
		Slug:        defaultCategorySlug,
		Description: "系统初始化自动创建的默认分类",
		SortOrder:   0,
	}
	if err := db.Create(category).Error; err != nil {
		low := strings.ToLower(err.Error())
		if strings.Contains(low, "duplicate") || strings.Contains(low, "unique") {
			return nil
		}
		return err
	}
	klog.InfoS("Default category ensured", "slug", defaultCategorySlug, "id", category.ID)
	return nil
}

// closeGormConnPool best-effort closes gorm underlying connection pool.
func closeGormConnPool(db *gorm.DB) error {
	if db == nil || db.ConnPool == nil {
		return nil
	}
	closer, ok := db.ConnPool.(interface{ Close() error })
	if !ok {
		return nil
	}
	return closer.Close()
}

// Close closes the ServiceContext and releases resources
// 关闭一个服务上下文并释放资源
func (s *ServiceContext) Close() error {
	var errs []error

	// 关闭数据库连接
	if s.DB != nil {
		if sqlDB, err := s.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close database: %w", err))
			}
		}
	}

	// 关闭 Redis 连接
	if s.Redis != nil {
		if err := s.Redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close redis: %w", err))
		}
	}

	return errors.Join(errs...)
}
