-- identity.user_identities: external provider identities bound to local users.
-- identity.user_identities：绑定到本地用户的外部身份提供方身份。
CREATE TABLE identity.user_identities (
  id BIGSERIAL PRIMARY KEY,

  user_id BIGINT NOT NULL
    CONSTRAINT fk_user_identities_user
    REFERENCES identity.users (id)
    ON DELETE CASCADE,

  provider VARCHAR(32) NOT NULL,
  provider_subject VARCHAR(128) NOT NULL,
  email_at_bind VARCHAR(320) NULL,

  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL,

  CONSTRAINT chk_user_identities_provider_not_empty CHECK (provider <> ''),
  CONSTRAINT chk_user_identities_provider_subject_not_empty CHECK (provider_subject <> '')
);

CREATE UNIQUE INDEX ux_user_identities_provider_subject
  ON identity.user_identities (provider, provider_subject)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_user_identities_user_id
  ON identity.user_identities (user_id)
  WHERE deleted_at IS NULL;

COMMENT ON TABLE identity.user_identities IS
  'External provider identity bindings. / 外部身份提供方身份绑定。';
COMMENT ON COLUMN identity.user_identities.provider_subject IS
  'Stable provider-side subject, for GitHub this is the numeric user ID. / 提供方侧稳定主体；GitHub 使用数字用户 ID。';
