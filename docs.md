# GitHub Star Sync

拉取一个或多个 GitHub 用户的公开星标，自动分类后生成中文 Markdown / HTML。

## 使用

```bash
cp configs/config.example.toml config.toml
# 编辑 [[sources]] 中的用户名

go build -o github-star-sync ./cmd
./github-star-sync -config config.toml
```

生成：`STARRED.md`、`stars/index.html`（可在配置里改路径或只保留一种）。

可选：设置环境变量 `GITHUB_TOKEN`，提高 API 限流（不设也能跑）。

## 配置

```toml
title = "星标收藏"
output_md = "STARRED.md"
output_html = "stars/index.html"

[[sources]]
username = "sky22333"
label = "主号"          # 可选展示名

# 多个账号继续追加 [[sources]]

[classify]
max_categories = 12     # topic 类目上限
min_count = 2           # topic 至少出现几次才成类
fallback = "language"   # 无 topic 时：language | other
sort_within = "stars"   # stars | starred_at | name
```
