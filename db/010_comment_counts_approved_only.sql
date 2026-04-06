-- ============================================
-- 10. 评论计数仅统计已审核(approved)评论
-- 替换 008_triggers.sql 中的函数与触发器定义
-- 已有库可单独执行本文件；新库通过 init.sql 在 008 之后加载
-- ============================================

-- 更新文章评论计数（仅 approved）
CREATE OR REPLACE FUNCTION update_article_comment_count()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF NEW.status = 'approved' THEN
      UPDATE articles SET comment_count = comment_count + 1 WHERE id = NEW.article_id;
    END IF;
    RETURN NEW;
  ELSIF TG_OP = 'DELETE' THEN
    IF OLD.status = 'approved' THEN
      UPDATE articles SET comment_count = comment_count - 1 WHERE id = OLD.article_id;
    END IF;
    RETURN OLD;
  ELSIF TG_OP = 'UPDATE' THEN
    IF OLD.status = 'approved' AND NEW.status <> 'approved' THEN
      UPDATE articles SET comment_count = comment_count - 1 WHERE id = OLD.article_id;
    ELSIF OLD.status <> 'approved' AND NEW.status = 'approved' THEN
      UPDATE articles SET comment_count = comment_count + 1 WHERE id = NEW.article_id;
    ELSIF OLD.status = 'approved' AND NEW.status = 'approved' AND OLD.article_id IS DISTINCT FROM NEW.article_id THEN
      UPDATE articles SET comment_count = comment_count - 1 WHERE id = OLD.article_id;
      UPDATE articles SET comment_count = comment_count + 1 WHERE id = NEW.article_id;
    END IF;
    RETURN NEW;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 更新用户评论计数（仅 approved 且 user_id 非空）
CREATE OR REPLACE FUNCTION update_user_comment_count()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF NEW.status = 'approved' AND NEW.user_id IS NOT NULL THEN
      UPDATE users SET comment_count = comment_count + 1 WHERE id = NEW.user_id;
    END IF;
    RETURN NEW;
  ELSIF TG_OP = 'DELETE' THEN
    IF OLD.status = 'approved' AND OLD.user_id IS NOT NULL THEN
      UPDATE users SET comment_count = comment_count - 1 WHERE id = OLD.user_id;
    END IF;
    RETURN OLD;
  ELSIF TG_OP = 'UPDATE' THEN
    IF OLD.status = 'approved' AND NEW.status <> 'approved' THEN
      IF OLD.user_id IS NOT NULL THEN
        UPDATE users SET comment_count = comment_count - 1 WHERE id = OLD.user_id;
      END IF;
    ELSIF OLD.status <> 'approved' AND NEW.status = 'approved' THEN
      IF NEW.user_id IS NOT NULL THEN
        UPDATE users SET comment_count = comment_count + 1 WHERE id = NEW.user_id;
      END IF;
    ELSIF OLD.status = 'approved' AND NEW.status = 'approved' AND OLD.user_id IS DISTINCT FROM NEW.user_id THEN
      IF OLD.user_id IS NOT NULL THEN
        UPDATE users SET comment_count = comment_count - 1 WHERE id = OLD.user_id;
      END IF;
      IF NEW.user_id IS NOT NULL THEN
        UPDATE users SET comment_count = comment_count + 1 WHERE id = NEW.user_id;
      END IF;
    END IF;
    RETURN NEW;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_article_comment_count_trigger ON comments;
CREATE TRIGGER update_article_comment_count_trigger
  AFTER INSERT OR UPDATE OR DELETE ON comments
  FOR EACH ROW EXECUTE FUNCTION update_article_comment_count();

DROP TRIGGER IF EXISTS update_user_comment_count_trigger ON comments;
CREATE TRIGGER update_user_comment_count_trigger
  AFTER INSERT OR UPDATE OR DELETE ON comments
  FOR EACH ROW EXECUTE FUNCTION update_user_comment_count();
