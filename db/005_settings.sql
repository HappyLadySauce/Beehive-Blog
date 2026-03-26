-- ============================================
-- 5. 系统设置相关表
-- ============================================

-- 系统设置表
CREATE TABLE IF NOT EXISTS settings (
    id BIGSERIAL PRIMARY KEY COMMENT '设置ID',
    key VARCHAR(100) NOT NULL UNIQUE COMMENT '设置键名',
    value TEXT COMMENT '设置值',
    "group" VARCHAR(50) NOT NULL DEFAULT 'general' COMMENT '设置分组: general-常规, seo-SEO, smtp-邮件, comment-评论, security-安全',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间'
);

CREATE INDEX idx_settings_group ON settings("group");
CREATE INDEX idx_settings_key ON settings(key);

COMMENT ON TABLE settings IS '系统设置表';

-- 友情链接表
CREATE TABLE IF NOT EXISTS links (
    id BIGSERIAL PRIMARY KEY COMMENT '友链ID',
    name VARCHAR(50) NOT NULL COMMENT '网站名称',
    url VARCHAR(500) NOT NULL COMMENT '网站URL',
    description VARCHAR(255) COMMENT '网站描述',
    logo VARCHAR(500) COMMENT '网站Logo',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序权重',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间'
);

CREATE INDEX idx_links_is_enabled ON links(is_enabled);

COMMENT ON TABLE links IS '友情链接表';

-- 操作日志表
CREATE TABLE IF NOT EXISTS operation_logs (
    id BIGSERIAL PRIMARY KEY COMMENT '日志ID',
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL COMMENT '用户ID',
    username VARCHAR(50) COMMENT '用户名',
    action VARCHAR(50) NOT NULL COMMENT '操作类型: create/update/delete/login等',
    object_type VARCHAR(50) COMMENT '操作对象类型: article/user/comment等',
    object_id VARCHAR(50) COMMENT '操作对象ID',
    detail TEXT COMMENT '操作详情',
    ip VARCHAR(50) COMMENT 'IP地址',
    user_agent VARCHAR(500) COMMENT '用户代理',
    status VARCHAR(20) NOT NULL DEFAULT 'success' COMMENT '操作状态: success/failed',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
);

CREATE INDEX idx_operation_logs_user_id ON operation_logs(user_id);
CREATE INDEX idx_operation_logs_action ON operation_logs(action);
CREATE INDEX idx_operation_logs_created_at ON operation_logs(created_at);

COMMENT ON TABLE operation_logs IS '操作日志表';

-- 数据备份表
CREATE TABLE IF NOT EXISTS backups (
    id BIGSERIAL PRIMARY KEY COMMENT '备份ID',
    name VARCHAR(100) NOT NULL COMMENT '备份名称',
    file_path VARCHAR(500) NOT NULL COMMENT '文件路径',
    file_size BIGINT COMMENT '文件大小(字节)',
    type VARCHAR(20) NOT NULL DEFAULT 'manual' CHECK (type IN ('manual', 'auto')) COMMENT '备份类型: manual-手动, auto-自动',
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL COMMENT '创建者ID',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
);

CREATE INDEX idx_backups_type ON backups(type);
CREATE INDEX idx_backups_created_at ON backups(created_at);

COMMENT ON TABLE backups IS '数据备份记录表';

-- 主题表
CREATE TABLE IF NOT EXISTS themes (
    id BIGSERIAL PRIMARY KEY COMMENT '主题ID',
    name VARCHAR(50) NOT NULL COMMENT '主题名称',
    slug VARCHAR(50) NOT NULL UNIQUE COMMENT '主题标识',
    description VARCHAR(255) COMMENT '主题描述',
    author VARCHAR(50) COMMENT '作者',
    version VARCHAR(20) COMMENT '版本号',
    path VARCHAR(255) NOT NULL COMMENT '主题路径',
    screenshot VARCHAR(500) COMMENT '预览图URL',
    is_active BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否激活',
    is_system BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否系统主题',
    config JSONB COMMENT '主题配置(JSON格式)',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间'
);

CREATE INDEX idx_themes_slug ON themes(slug);
CREATE INDEX idx_themes_is_active ON themes(is_active);

COMMENT ON TABLE themes IS '主题表';

-- 菜单表
CREATE TABLE IF NOT EXISTS menus (
    id BIGSERIAL PRIMARY KEY COMMENT '菜单ID',
    name VARCHAR(50) NOT NULL COMMENT '菜单名称',
    location VARCHAR(50) NOT NULL COMMENT '菜单位置: header/footer/sidebar',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序权重',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间'
);

CREATE INDEX idx_menus_location ON menus(location);

COMMENT ON TABLE menus IS '菜单表';

-- 菜单项表
CREATE TABLE IF NOT EXISTS menu_items (
    id BIGSERIAL PRIMARY KEY COMMENT '菜单项ID',
    menu_id BIGINT NOT NULL REFERENCES menus(id) ON DELETE CASCADE COMMENT '菜单ID',
    parent_id BIGINT REFERENCES menu_items(id) ON DELETE CASCADE COMMENT '父菜单项ID',
    name VARCHAR(50) NOT NULL COMMENT '菜单项名称',
    type VARCHAR(20) NOT NULL CHECK (type IN ('link', 'page', 'category', 'tag')) COMMENT '类型: link-链接, page-页面, category-分类, tag-标签',
    target_id VARCHAR(50) COMMENT '关联资源ID',
    url VARCHAR(500) COMMENT '链接地址(类型为link时使用)',
    icon VARCHAR(50) COMMENT '图标',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序权重',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间'
);

CREATE INDEX idx_menu_items_menu_id ON menu_items(menu_id);
CREATE INDEX idx_menu_items_parent_id ON menu_items(parent_id);

COMMENT ON TABLE menu_items IS '菜单项表';

-- 独立页面表
CREATE TABLE IF NOT EXISTS pages (
    id BIGSERIAL PRIMARY KEY COMMENT '页面ID',
    title VARCHAR(200) NOT NULL COMMENT '页面标题',
    slug VARCHAR(100) NOT NULL UNIQUE COMMENT 'URL别名',
    content TEXT NOT NULL COMMENT '页面内容(Markdown)',
    status VARCHAR(20) NOT NULL DEFAULT 'published' CHECK (status IN ('draft', 'published', 'archived', 'private')) COMMENT '状态: draft-草稿, published-已发布, archived-已归档, private-私密',
    is_in_menu BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否显示在菜单中',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序权重',
    view_count BIGINT NOT NULL DEFAULT 0 COMMENT '浏览次数',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP COMMENT '删除时间(软删除)'
);

CREATE INDEX idx_pages_slug ON pages(slug);
CREATE INDEX idx_pages_status ON pages(status);
CREATE INDEX idx_pages_deleted_at ON pages(deleted_at);

COMMENT ON TABLE pages IS '独立页面表';
