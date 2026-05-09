-- identity.user_credentials: password hashes for local authentication, one active per user.
-- identity.user_credentials：本地认证的密码哈希，每个活跃用户仅一条。
-- Must run after 010_identity_users.sql.
-- 必须在 010_identity_users.sql 之后执行。
-- IF NOT EXISTS pairs with versioned -force re-apply (same rationale as 010).
-- 与 versioned 模式下 -force 重跑策略一致，见 010 文件头注释。
CREATE TABLE IF NOT EXISTS identity.user_credentials (
  id BIGSERIAL PRIMARY KEY,

  user_id BIGINT NOT NULL
    CONSTRAINT fk_user_credentials_user
    REFERENCES identity.users (id)
    ON DELETE CASCADE,

  password_hash VARCHAR(255) NOT NULL,

  -- GORM-standard timestamps. / GORM 标准时间字段。
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL,

  -- One active credential row per user; a soft-deleted row allows password reset history.
  -- 每个活跃用户仅一条凭证行；软删旧行可保留密码修改历史。
  CONSTRAINT chk_user_credentials_password_hash_not_empty CHECK (password_hash <> '')
);

-- Unique index: only one active credential per user.
-- 唯一索引：每个用户仅有一个活跃凭证。
CREATE UNIQUE INDEX IF NOT EXISTS ux_user_credentials_user_id
  ON identity.user_credentials (user_id)
  WHERE deleted_at IS NULL;

COMMENT ON TABLE identity.user_credentials IS
  'Password hashes for local username/password login. / 本地用户名/密码登录的密码哈希。';
COMMENT ON COLUMN identity.user_credentials.user_id IS
  'FK to identity.users. / 外键指向 identity.users。';
COMMENT ON COLUMN identity.user_credentials.password_hash IS
  'bcrypt hash of the plaintext password. / 明文密码的 bcrypt 哈希。';
COMMENT ON COLUMN identity.user_credentials.deleted_at IS
  'Soft-deletion permits password change history; only rows with NULL deleted_at are active. / 软删保留改密历史；仅 deleted_at IS NULL 行为活跃凭证。';

-- Default admin password: Admin@123 (bcrypt cost 12, matches pkg/auth/passwd.DefaultCost).
-- 默认管理员密码：Admin@123（bcrypt cost 12，与 pkg/auth/passwd.DefaultCost 一致）。
INSERT INTO identity.user_credentials (user_id, password_hash, created_at, updated_at)
SELECT u.id,
  '$2a$12$fKPFyp6NfF/paV.nMbBdPujwpX9XeS6RnYICqKlm2G5sYlmO57BPO',
  NOW(),
  NOW()
FROM identity.users u
WHERE u.username = 'admin'
  AND u.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM identity.user_credentials c
    WHERE c.user_id = u.id AND c.deleted_at IS NULL
  );
