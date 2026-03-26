-- Beehive-Blog 数据库初始化脚本
-- 数据库: PostgreSQL
-- 版本: 1.0

-- ============================================
-- 1. 用户相关表
-- ============================================

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY COMMENT '用户ID',
    username VARCHAR(20) NOT NULL UNIQUE COMMENT '用户名',
    nickname VARCHAR(50) COMMENT '昵称',
    email VARCHAR(100) UNIQUE COMMENT '邮箱',
    password VARCHAR(255) NOT NULL COMMENT '密码(bcrypt加密)',
    avatar VARCHAR(500) COMMENT '头像URL',
    role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('guest', 'user', 'admin')) COMMENT '角色: guest-访客, user-普通用户, admin-管理员',
    status VARCHAR(20) NOT NULL DEFAULT 'inactive' CHECK (status IN ('active', 'inactive', 'disabled', 'deleted')) COMMENT '状态: active-正常, inactive-未激活, disabled-禁用, deleted-已删除',
    level INT NOT NULL DEFAULT 1 COMMENT '用户等级(1-6)',
    experience INT NOT NULL DEFAULT 0 COMMENT '当前经验值',
    comment_count INT NOT NULL DEFAULT 0 COMMENT '评论数量',
    article_view_count INT NOT NULL DEFAULT 0 COMMENT '有效文章阅读数量',
    last_login_at TIMESTAMP COMMENT '最后登录时间',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP COMMENT '删除时间(软删除)'
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- 用户等级配置表
CREATE TABLE IF NOT EXISTS user_levels (
    id BIGSERIAL PRIMARY KEY COMMENT '等级ID',
    level INT NOT NULL UNIQUE COMMENT '等级(1-6)',
    name VARCHAR(50) NOT NULL COMMENT '等级名称',
    required_exp INT NOT NULL COMMENT '所需经验值',
    required_days INT COMMENT '所需注册天数',
    required_articles INT COMMENT '所需有效阅读文章数',
    description VARCHAR(255) COMMENT '等级描述',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间'
);

