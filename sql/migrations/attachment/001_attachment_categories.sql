CREATE SCHEMA IF NOT EXISTS attachment;

-- attachment.categories: hierarchical taxonomy for attachments using a
-- materialized-path tree (parent_id self-FK + path + depth).
-- Maintained by triggers below; GORM should only write parent_id.
-- Run after 000_attachment_attachments.sql (basename sort: 000 before 001).
-- attachment.categories：基于物化路径的附件分类树（parent_id 自引用 + path + depth）。
-- path/depth 由下方触发器维护，GORM 仅写入 parent_id。
-- 须在 000_attachment_attachments.sql 之后执行（按文件名前缀 000 < 001）。
CREATE TABLE attachment.categories (
  id              BIGSERIAL PRIMARY KEY,

  -- Self-referencing parent. RESTRICT prevents hard-deleting non-empty subtrees;
  -- soft-delete (deleted_at) does not block business workflows.
  -- 自引用父分类。RESTRICT 阻止物理删除非空子树；软删（deleted_at）不影响业务流。
  parent_id       BIGINT NULL
    CONSTRAINT fk_attachment_categories_parent
    REFERENCES attachment.categories (id)
    ON DELETE RESTRICT,

  -- Display fields. / 展示字段。
  name            VARCHAR(64)  NOT NULL,
  slug            VARCHAR(64)  NOT NULL,
  description     TEXT         NULL,
  icon            VARCHAR(64)  NULL,

  -- Materialized path for O(log n) subtree queries: '/1/3/7/'.
  -- Trigger-managed: never write directly from the application layer.
  -- 物化路径，便于子树前缀扫描，例如 '/1/3/7/'。由触发器维护，应用层禁止直接写入。
  path            TEXT         NOT NULL,
  depth           INT          NOT NULL DEFAULT 0,

  -- Manual ordering among siblings (lower first).
  -- 同级排序权重（数值小者靠前）。
  sort_order      INT          NOT NULL DEFAULT 0,

  -- Lifecycle status (soft-delete remains orthogonal via deleted_at).
  -- 生命周期状态（软删仍由 deleted_at 独立判定）。
  status          VARCHAR(16)  NOT NULL DEFAULT 'active',

  -- GORM-standard timestamps. / GORM 标准时间字段。
  created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  deleted_at      TIMESTAMPTZ  NULL,

  CONSTRAINT chk_attachment_categories_status
    CHECK (status IN ('active', 'disabled')),
  CONSTRAINT chk_attachment_categories_depth
    CHECK (depth >= 0),
  CONSTRAINT chk_attachment_categories_slug_format
    CHECK (slug ~ '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$'),
  CONSTRAINT chk_attachment_categories_path_format
    CHECK (path ~ '^(/[0-9]+)+/$'),
  CONSTRAINT chk_attachment_categories_no_self_parent
    CHECK (parent_id IS NULL OR parent_id <> id)
);

-- Globally unique slug among live rows; allows reuse after soft-delete.
-- 活跃行内 slug 全局唯一；软删后可被复用。
CREATE UNIQUE INDEX ux_attachment_categories_slug
  ON attachment.categories (slug)
  WHERE deleted_at IS NULL;

-- Sibling-name uniqueness for non-root nodes (case-insensitive).
-- 非根分类下同级名称唯一（不区分大小写）。
CREATE UNIQUE INDEX ux_attachment_categories_sibling_name_child
  ON attachment.categories (parent_id, lower(name))
  WHERE deleted_at IS NULL AND parent_id IS NOT NULL;

-- Sibling-name uniqueness for roots (parent_id NULL must be handled separately).
-- 根级分类名称唯一（parent_id 为空时无法走上方索引，单独建一条）。
CREATE UNIQUE INDEX ux_attachment_categories_sibling_name_root
  ON attachment.categories (lower(name))
  WHERE deleted_at IS NULL AND parent_id IS NULL;

-- Ordered child enumeration under a given parent.
-- 按父节点检索有序子分类。
CREATE INDEX idx_attachment_categories_parent_sort
  ON attachment.categories (parent_id, sort_order, id)
  WHERE deleted_at IS NULL;

-- Subtree prefix scans: WHERE path LIKE '/1/3/%' will use this index
-- because text_pattern_ops bypasses collation-dependent ordering.
-- 子树前缀查询：text_pattern_ops 让 LIKE 'prefix%' 稳定走索引（不依赖排序规则）。
CREATE INDEX idx_attachment_categories_path
  ON attachment.categories (path text_pattern_ops)
  WHERE deleted_at IS NULL;

-- Status index limited to non-default rows to keep the index lean.
-- 仅对非默认状态建索引，控制索引体积。
CREATE INDEX idx_attachment_categories_status
  ON attachment.categories (status)
  WHERE deleted_at IS NULL AND status <> 'active';

-- Audit / cleanup queries on soft-deleted rows.
-- 用于审计或清理软删行的索引。
CREATE INDEX idx_attachment_categories_deleted_at
  ON attachment.categories (deleted_at)
  WHERE deleted_at IS NOT NULL;

-- =====================================================================
-- Path / depth maintenance triggers
-- 物化路径与深度维护触发器
-- =====================================================================

-- BEFORE INSERT: derive path/depth from parent. Relies on the fact that
-- BIGSERIAL defaults are evaluated before BEFORE-row triggers run, so
-- NEW.id is already populated.
-- 插入前：由父行推导 path/depth；BIGSERIAL 默认值在 BEFORE 触发器之前生成，NEW.id 已可用。
CREATE OR REPLACE FUNCTION attachment.fn_categories_set_path_before_insert()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
  parent_path  TEXT;
  parent_depth INT;
BEGIN
  IF NEW.parent_id IS NULL THEN
    NEW.depth := 0;
    NEW.path  := '/' || NEW.id::text || '/';
  ELSE
    SELECT c.path, c.depth
      INTO parent_path, parent_depth
      FROM attachment.categories c
     WHERE c.id = NEW.parent_id;

    IF parent_path IS NULL THEN
      RAISE EXCEPTION 'attachment.categories.parent_id=% not found', NEW.parent_id
        USING ERRCODE = 'foreign_key_violation';
    END IF;

    NEW.depth := parent_depth + 1;
    NEW.path  := parent_path || NEW.id::text || '/';
  END IF;
  RETURN NEW;
END;
$$;

CREATE TRIGGER trg_categories_before_insert
  BEFORE INSERT ON attachment.categories
  FOR EACH ROW
  EXECUTE PROCEDURE attachment.fn_categories_set_path_before_insert();

-- BEFORE UPDATE OF parent_id: detect cycles, then recompute self path/depth.
-- Descendant cascade is deferred to the AFTER trigger so we never recurse
-- into this BEFORE handler.
-- parent_id 变更前：检测环路并重新计算自身 path/depth；后代级联放在 AFTER 触发器，避免递归。
CREATE OR REPLACE FUNCTION attachment.fn_categories_recompute_self_on_reparent()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
  parent_path  TEXT;
  parent_depth INT;
BEGIN
  IF NEW.parent_id IS NULL THEN
    NEW.depth := 0;
    NEW.path  := '/' || NEW.id::text || '/';
    RETURN NEW;
  END IF;

  SELECT c.path, c.depth
    INTO parent_path, parent_depth
    FROM attachment.categories c
   WHERE c.id = NEW.parent_id;

  IF parent_path IS NULL THEN
    RAISE EXCEPTION 'attachment.categories.parent_id=% not found', NEW.parent_id
      USING ERRCODE = 'foreign_key_violation';
  END IF;

  -- Cycle guard: the new parent must not live inside the moving subtree.
  -- 防环：新父节点不得位于当前子树内。
  IF parent_path LIKE OLD.path || '%' THEN
    RAISE EXCEPTION
      'cycle detected: cannot move attachment.categories.id=% under its descendant id=%',
      NEW.id, NEW.parent_id
      USING ERRCODE = 'check_violation';
  END IF;

  NEW.depth := parent_depth + 1;
  NEW.path  := parent_path || NEW.id::text || '/';
  RETURN NEW;
END;
$$;

CREATE TRIGGER trg_categories_before_update_parent
  BEFORE UPDATE OF parent_id ON attachment.categories
  FOR EACH ROW
  WHEN (NEW.parent_id IS DISTINCT FROM OLD.parent_id)
  EXECUTE PROCEDURE attachment.fn_categories_recompute_self_on_reparent();

-- AFTER UPDATE OF parent_id: rewrite path/depth on every descendant in one
-- pass. The WHEN clause limits firing to genuine reparenting events so the
-- cascading UPDATEs (which only touch path/depth, not parent_id) don't
-- re-enter this trigger.
-- parent_id 变更后：一条 SQL 同步重写所有后代的 path/depth；WHEN 约束保证级联回写不会再次触发本逻辑。
CREATE OR REPLACE FUNCTION attachment.fn_categories_cascade_path_to_descendants()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  UPDATE attachment.categories
     SET path       = NEW.path || substring(path FROM char_length(OLD.path) + 1),
         depth      = depth + (NEW.depth - OLD.depth),
         updated_at = NOW()
   WHERE path LIKE OLD.path || '%'
     AND id <> NEW.id;
  RETURN NULL;
END;
$$;

CREATE TRIGGER trg_categories_after_update_parent
  AFTER UPDATE OF parent_id ON attachment.categories
  FOR EACH ROW
  WHEN (NEW.parent_id IS DISTINCT FROM OLD.parent_id)
  EXECUTE PROCEDURE attachment.fn_categories_cascade_path_to_descendants();

-- =====================================================================
-- attachment.attachment_categories: many-to-many join between attachments
-- and categories. Hard-delete on either side cascades; soft-delete leaves
-- join rows untouched and is filtered at query time via JOIN predicates.
-- attachment.attachment_categories：附件与分类的多对多联结表。任一方物理删除时级联清理；
-- 软删不动联结行，查询时通过 JOIN 上的 deleted_at IS NULL 过滤。
-- =====================================================================
CREATE TABLE attachment.attachment_categories (
  attachment_id BIGINT      NOT NULL
    CONSTRAINT fk_attachment_categories_join_attachment
    REFERENCES attachment.attachments (id)
    ON DELETE CASCADE,
  category_id   BIGINT      NOT NULL
    CONSTRAINT fk_attachment_categories_join_category
    REFERENCES attachment.categories (id)
    ON DELETE CASCADE,

  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT pk_attachment_categories_join
    PRIMARY KEY (attachment_id, category_id)
);

-- Reverse lookup: list attachments belonging to a given category.
-- 反向查询：按分类列出归属附件。
CREATE INDEX idx_attachment_categories_join_category
  ON attachment.attachment_categories (category_id, attachment_id);

-- =====================================================================
-- Bilingual table / column / function / trigger comments
-- 双语表 / 列 / 函数 / 触发器注释
-- =====================================================================

COMMENT ON TABLE attachment.categories IS
  'Hierarchical attachment taxonomy with materialized path. / 基于物化路径的附件分类层级表。';

COMMENT ON COLUMN attachment.categories.parent_id IS
  'Self-FK to parent category; NULL for roots. RESTRICT on hard delete. / 自引用父分类外键，根分类为 NULL；物理删除走 RESTRICT。';
COMMENT ON COLUMN attachment.categories.name IS
  'Display name; unique among siblings on live rows (case-insensitive). / 展示名；活跃行内同级唯一（不区分大小写）。';
COMMENT ON COLUMN attachment.categories.slug IS
  'URL-safe key, globally unique among live rows. / URL 友好键，活跃行内全局唯一。';
COMMENT ON COLUMN attachment.categories.description IS
  'Optional description. / 可选描述。';
COMMENT ON COLUMN attachment.categories.icon IS
  'Optional UI icon identifier. / 可选 UI 图标标识。';
COMMENT ON COLUMN attachment.categories.path IS
  'Materialized path like ''/1/3/7/''; trigger-managed, do not write from app. / 物化路径形如 ''/1/3/7/''；由触发器维护，应用层禁止直接写入。';
COMMENT ON COLUMN attachment.categories.depth IS
  'Tree depth (root=0); trigger-managed. / 树深度（根为 0），由触发器维护。';
COMMENT ON COLUMN attachment.categories.sort_order IS
  'Manual ordering among siblings, lower first. / 同级排序权重，数值小者靠前。';
COMMENT ON COLUMN attachment.categories.status IS
  'Lifecycle: active | disabled. Soft-delete is always deleted_at. / 生命周期：active | disabled；软删始终用 deleted_at。';
COMMENT ON COLUMN attachment.categories.created_at IS
  'Row creation timestamp, maintained by GORM CreatedAt. / 行创建时间，由 GORM CreatedAt 维护。';
COMMENT ON COLUMN attachment.categories.updated_at IS
  'Row last-update timestamp, maintained by GORM UpdatedAt. / 行最近更新时间，由 GORM UpdatedAt 维护。';
COMMENT ON COLUMN attachment.categories.deleted_at IS
  'Soft-deletion timestamp aligned with gorm.DeletedAt. / 与 gorm.DeletedAt 对齐的软删时间戳。';

COMMENT ON FUNCTION attachment.fn_categories_set_path_before_insert() IS
  'Derives path/depth from parent_id on INSERT. / 插入时由 parent_id 推导 path 与 depth。';
COMMENT ON FUNCTION attachment.fn_categories_recompute_self_on_reparent() IS
  'Validates non-cyclic reparent and recomputes self path/depth. / 校验非环路并重算自身 path 与 depth。';
COMMENT ON FUNCTION attachment.fn_categories_cascade_path_to_descendants() IS
  'Rewrites descendants path/depth after a reparent. / 重新挂载后批量重写后代 path 与 depth。';

COMMENT ON TRIGGER trg_categories_before_insert ON attachment.categories IS
  'Populate path/depth before INSERT. / 插入前填充 path 与 depth。';
COMMENT ON TRIGGER trg_categories_before_update_parent ON attachment.categories IS
  'Cycle check + self path/depth recompute on parent_id change. / parent_id 变更时检环并重算自身 path 与 depth。';
COMMENT ON TRIGGER trg_categories_after_update_parent ON attachment.categories IS
  'Cascade path/depth rewrite to descendants after parent_id change. / parent_id 变更后向后代级联重写 path 与 depth。';

COMMENT ON TABLE attachment.attachment_categories IS
  'Many-to-many join between attachments and categories. / 附件与分类的多对多联结表。';
COMMENT ON COLUMN attachment.attachment_categories.attachment_id IS
  'FK to attachment.attachments(id); ON DELETE CASCADE on hard delete. / 指向附件表的外键；物理删除级联。';
COMMENT ON COLUMN attachment.attachment_categories.category_id IS
  'FK to attachment.categories(id); ON DELETE CASCADE on hard delete. / 指向分类表的外键；物理删除级联。';
COMMENT ON COLUMN attachment.attachment_categories.created_at IS
  'Association creation timestamp. / 关联建立时间。';
