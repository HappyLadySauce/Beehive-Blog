-- Remove the file_node_id column and the file_nodes table.
-- Neither has been exposed via any API; the table has never been populated.
-- 移除 file_node_id 列与 file_nodes 表。二者均未通过 API 暴露，表无任何数据。

ALTER TABLE attachment.attachments
    DROP CONSTRAINT IF EXISTS fk_attachment_file_node;

ALTER TABLE attachment.attachments
    DROP COLUMN IF EXISTS file_node_id;

DROP TABLE IF EXISTS attachment.file_nodes;
