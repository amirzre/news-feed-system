DROP TRIGGER IF EXISTS update_posts_updated_at ON posts;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_posts_category_published;
DROP INDEX IF EXISTS idx_posts_created_at;
DROP INDEX IF EXISTS idx_posts_category;
DROP INDEX IF EXISTS idx_posts_source;
DROP INDEX IF EXISTS idx_posts_published_at;

DROP TABLE IF EXISTS posts;