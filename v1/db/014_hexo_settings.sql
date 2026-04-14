-- Hexo 设置默认行（存量库升级；与 007_seed 中 hexo 段一致，幂等）
INSERT INTO settings (key, value, "group") VALUES
('hexo.auto_sync', 'false', 'hexo'),
('hexo.clean_args', '', 'hexo'),
('hexo.generate_args', '', 'hexo'),
('hexo.rebuild_after_auto_sync', 'false', 'hexo')
ON CONFLICT (key) DO NOTHING;
