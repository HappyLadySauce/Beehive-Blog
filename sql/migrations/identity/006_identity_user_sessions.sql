-- identity.user_sessions: server-side refresh token sessions for revocation and rotation.
-- identity.user_sessions：服务端 refresh token 会话，用于撤销与轮换。
CREATE TABLE identity.user_sessions (
  id BIGSERIAL PRIMARY KEY,

  user_id BIGINT NOT NULL
    CONSTRAINT fk_user_sessions_user
    REFERENCES identity.users (id)
    ON DELETE CASCADE,

  refresh_token_hash CHAR(64) NOT NULL,
  refresh_jti VARCHAR(64) NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ NULL,
  revoked_reason VARCHAR(64) NULL,
  rotated_at TIMESTAMPTZ NULL,
  created_ip VARCHAR(64) NOT NULL DEFAULT '',
  user_agent VARCHAR(512) NOT NULL DEFAULT '',
  last_used_at TIMESTAMPTZ NULL,

  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT chk_user_sessions_refresh_hash_not_empty CHECK (refresh_token_hash <> ''),
  CONSTRAINT chk_user_sessions_refresh_jti_not_empty CHECK (refresh_jti <> '')
);

CREATE UNIQUE INDEX ux_user_sessions_refresh_jti
  ON identity.user_sessions (refresh_jti);

CREATE INDEX idx_user_sessions_user_active
  ON identity.user_sessions (user_id, expires_at)
  WHERE revoked_at IS NULL AND rotated_at IS NULL;

CREATE INDEX idx_user_sessions_revoked_at
  ON identity.user_sessions (revoked_at)
  WHERE revoked_at IS NOT NULL;

CREATE INDEX idx_user_sessions_rotated_at
  ON identity.user_sessions (rotated_at)
  WHERE rotated_at IS NOT NULL;

COMMENT ON TABLE identity.user_sessions IS
  'Server-side refresh token session state. / 服务端 refresh token 会话状态。';
COMMENT ON COLUMN identity.user_sessions.refresh_token_hash IS
  'SHA-256 hash of the refresh JWT; plaintext tokens are never stored. / refresh JWT 的 SHA-256 哈希；不存储明文令牌。';
COMMENT ON COLUMN identity.user_sessions.refresh_jti IS
  'JWT ID bound to the active refresh token for replay detection. / 绑定当前 refresh token 的 JWT ID，用于重放检测。';
COMMENT ON COLUMN identity.user_sessions.rotated_at IS
  'Set when this refresh token has been exchanged for a newer session. / 当前 refresh token 已换取新会话时设置。';
