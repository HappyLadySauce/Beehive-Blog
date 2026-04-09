-- ============================================
-- 7. 初始数据
-- ============================================

-- 默认管理员账号
-- 用户名: Beehive
-- 密码:   Beehive
-- 使用 pgcrypto 的 bcrypt 生成兼容应用登录校验的密码哈希。
CREATE EXTENSION IF NOT EXISTS pgcrypto;
INSERT INTO users (username, nickname, email, password, role, status, level)
VALUES (
    'Beehive',
    'Beehive',
    'beehive@local',
    crypt('Beehive', gen_salt('bf', 12)),
    'admin',
    'active',
    6
)
ON CONFLICT (username) DO UPDATE
SET
    password = EXCLUDED.password,
    role = 'admin',
    status = 'active',
    nickname = EXCLUDED.nickname;

-- 用户等级配置初始数据
INSERT INTO user_levels (level, name, required_exp, required_days, required_articles, description) VALUES
(1, '初来乍到', 0, 0, 0, '注册即获得'),
(2, '博客新手', 100, 7, 10, '注册满7天 或 有效阅读10篇文章'),
(3, '活跃读者', 300, 30, 30, '注册满30天 或 有效阅读30篇文章'),
(4, '资深访客', 600, 90, 60, '注册满90天 或 有效阅读60篇文章'),
(5, '博客达人', 1000, 180, 100, '注册满180天 或 有效阅读100篇文章'),
(6, '资深博主', 2000, 365, 200, '注册满365天 或 有效阅读200篇文章');

-- 默认分类
INSERT INTO categories (name, slug, description, sort_order)
VALUES ('默认分类', 'default', '系统初始化自动创建的默认分类', 0)
ON CONFLICT (slug) DO NOTHING;

-- 默认存储策略
INSERT INTO storage_policies (name, type, is_default, base_url, upload_path, sort_order) VALUES
('本地存储', 'local', TRUE, '/uploads', 'uploads', 1);

-- 系统设置初始数据
INSERT INTO settings (key, value, "group") VALUES
-- 常规设置
('site_name', 'Beehive Blog', 'general'),
('site_description', '一个简洁优雅的个人博客系统', 'general'),
('site_keywords', '博客,技术,分享', 'general'),
('site_logo', '', 'general'),
('site_favicon', '', 'general'),
('footer_text', '© 2026 Beehive Blog. All rights reserved.', 'general'),
('icp_number', '', 'general'),
('allow_register', 'true', 'general'),
('comment_need_review', 'true', 'general'),

-- SEO设置
('seo_title', 'Beehive Blog', 'seo'),
('seo_description', '一个简洁优雅的个人博客系统', 'seo'),
('seo_keywords', '博客,技术,分享', 'seo'),
('allow_indexing', 'true', 'seo'),
('robots_txt', 'User-agent: *\nAllow: /', 'seo'),

-- SMTP设置
('smtp_host', '', 'smtp'),
('smtp_port', '587', 'smtp'),
('smtp_encryption', 'tls', 'smtp'),
('smtp_username', '', 'smtp'),
('smtp_password', '', 'smtp'),
('smtp_from_email', '', 'smtp'),
('smtp_from_name', 'Beehive Blog', 'smtp'),
('smtp_enabled', 'false', 'smtp'),

-- 评论设置
('comment_enabled', 'true', 'comment'),
('comment_need_login', 'false', 'comment'),
('comment_max_length', '2000', 'comment'),
('comment_allow_guest', 'true', 'comment'),

-- 安全设置
('rate_limit_enabled', 'true', 'security'),
('rate_limit_requests', '100', 'security'),
('rate_limit_window', '60', 'security'),
('login_max_attempts', '5', 'security'),
('login_lockout_duration', '3600', 'security'),

-- Hexo 同步行为（路径 hexo_dir 仅来自服务端 YAML）
('hexo.auto_sync', 'false', 'hexo'),
('hexo.clean_args', '', 'hexo'),
('hexo.generate_args', '', 'hexo'),
('hexo.rebuild_after_auto_sync', 'false', 'hexo');

-- 默认菜单
INSERT INTO menus (name, location, sort_order, is_enabled) VALUES
('顶部导航', 'header', 1, TRUE),
('底部导航', 'footer', 1, TRUE);

-- 获取菜单ID并插入默认菜单项
INSERT INTO menu_items (menu_id, name, type, url, sort_order, is_enabled)
SELECT m.id, '首页', 'link', '/', 1, TRUE
FROM menus m WHERE m.location = 'header';

INSERT INTO menu_items (menu_id, name, type, url, sort_order, is_enabled)
SELECT m.id, '分类', 'link', '/categories', 2, TRUE
FROM menus m WHERE m.location = 'header';

INSERT INTO menu_items (menu_id, name, type, url, sort_order, is_enabled)
SELECT m.id, '标签', 'link', '/tags', 3, TRUE
FROM menus m WHERE m.location = 'header';

INSERT INTO menu_items (menu_id, name, type, url, sort_order, is_enabled)
SELECT m.id, '关于', 'link', '/about', 4, TRUE
FROM menus m WHERE m.location = 'header';

-- 底部菜单
INSERT INTO menu_items (menu_id, name, type, url, sort_order, is_enabled)
SELECT m.id, '首页', 'link', '/', 1, TRUE
FROM menus m WHERE m.location = 'footer';

INSERT INTO menu_items (menu_id, name, type, url, sort_order, is_enabled)
SELECT m.id, 'RSS', 'link', '/rss', 2, TRUE
FROM menus m WHERE m.location = 'footer';

INSERT INTO menu_items (menu_id, name, type, url, sort_order, is_enabled)
SELECT m.id, '站点地图', 'link', '/sitemap.xml', 3, TRUE
FROM menus m WHERE m.location = 'footer';

-- 默认主题
INSERT INTO themes (name, slug, description, author, version, path, is_active, is_system) VALUES
('默认主题', 'default', 'Beehive Blog 默认主题', 'Beehive', '1.0.0', 'themes/default', TRUE, TRUE);
