-- attachment.file_nodes: virtual file-system nodes for Studio file browsing.
-- Expresses a directory / file tree within a storage mount. Path and depth
-- are maintained by application code (triggers deferred to a later phase).
-- Business attachments MAY reference a file_node_id; unattached files
-- (no attachment row) are also representable.
-- attachment.file_nodes：虚拟文件系统节点，用于 Studio 文件浏览。
-- 表达存储挂载项内的目录 / 文件层级。路径与深度由应用层维护
-- （触发器延后实现）。业务附件可选引用 file_node_id；
-- 无附件行的独立文件亦可表达。
CREATE TABLE attachment.file_nodes (
    id               BIGSERIAL    PRIMARY KEY,
    parent_id        BIGINT,
    storage_mount_id BIGINT       NOT NULL,
    node_type        VARCHAR(16)  NOT NULL,
    name             VARCHAR(255) NOT NULL,
    path             TEXT         NOT NULL,
    full_path        TEXT         NOT NULL,
    depth            INT          NOT NULL DEFAULT 0,
    sort_order       INT          NOT NULL DEFAULT 0,
    status           VARCHAR(16)  NOT NULL DEFAULT 'active',
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ,

    CONSTRAINT chk_file_node_type   CHECK (node_type IN ('directory', 'file')),
    CONSTRAINT chk_file_node_status CHECK (status IN ('active', 'hidden', 'archived')),
    CONSTRAINT fk_file_node_parent  FOREIGN KEY (parent_id)
        REFERENCES attachment.file_nodes (id),
    CONSTRAINT fk_file_node_mount   FOREIGN KEY (storage_mount_id)
        REFERENCES attachment.storage_mounts (id)
);

-- Unique full_path per mount for active rows.
-- 活跃行内同一 mount 下 full_path 唯一。
CREATE UNIQUE INDEX ux_file_nodes_mount_full_path
    ON attachment.file_nodes (storage_mount_id, full_path)
    WHERE deleted_at IS NULL;

-- Active sibling names must be unique under the same parent.
-- 同一父节点下的活跃同级名称必须唯一。
CREATE UNIQUE INDEX ux_file_nodes_active_sibling_name
    ON attachment.file_nodes (storage_mount_id, parent_id, name)
    WHERE deleted_at IS NULL AND parent_id IS NOT NULL;

-- Active root names are unique per mount; NULL parent_id needs a separate index.
-- 根节点 parent_id 为 NULL，需要单独约束同一 mount 下根节点名称唯一。
CREATE UNIQUE INDEX ux_file_nodes_active_root_name
    ON attachment.file_nodes (storage_mount_id, name)
    WHERE deleted_at IS NULL AND parent_id IS NULL;

-- Fast child lookup for tree traversal.
-- 树遍历的子节点快速查找。
CREATE INDEX idx_file_nodes_parent
    ON attachment.file_nodes (parent_id)
    WHERE deleted_at IS NULL;

COMMENT ON COLUMN attachment.file_nodes.node_type IS
    'directory | file. / 目录或文件。';
COMMENT ON COLUMN attachment.file_nodes.path IS
    'Materialized path like /1/9/20/ for efficient subtree queries. / 物化路径，如 /1/9/20/，用于高效子树查询。';
COMMENT ON COLUMN attachment.file_nodes.full_path IS
    'User-visible path like /images/avatar/a.png. / 用户可见路径，如 /images/avatar/a.png。';
COMMENT ON COLUMN attachment.file_nodes.status IS
    'active | hidden | archived. / 活跃、隐藏或归档。';
