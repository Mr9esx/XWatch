---
name: xwatch
description: X (Twitter) 推文订阅、推送与分析工具。管理推文订阅、查看动态、翻译推文、总结分析。
version: 1.0.0
metadata:
  hermes:
    tags: [xwatch, twitter, social-media, subscription, translation]
    category: social-media
---

# XWatch — X 推文订阅与推送

XWatch 是一个 X (Twitter) 推文监控工具，已打包在此 skill 内。

**二进制路径：** `~/.hermes/skills/xwatch/scripts/xwatch`

下文中所有 `xwatch` 命令均指此路径，执行时请使用完整路径。

## CLI 命令参考

```
xwatch search <query> [--json]         # 搜索 X 用户
xwatch subscribe <screen_name> [--tags T] [--interval N] [--json]  # 订阅用户
xwatch unsubscribe <screen_name> [--json]  # 取消订阅
xwatch list [--tag T] [--json]           # 列出所有订阅
xwatch tag add|remove|set|list [--json]  # 管理主题标签
xwatch tweets [--user X] [--tag T] [--since 24h] [--limit 100] [--json]  # 查询缓存推文
xwatch config get|set|list [--json]    # 配置管理
xwatch account add|remove|list [--json] # 账号池管理
xwatch check [--tag T] [--force]       # 拉取并检查新推文
xwatch init                           # 首次初始化
xwatch version [--json]               # 查看版本（排查认证问题用）
```

所有命令加 `--json` 返回结构化 JSON，便于解析。

## 子文档

- **[CLI 操作场景](cli-scenarios.md)** — 首次设置、订阅管理、标签管理、设置与账号池
- **[推送排版规范](push-format.md)** — 推文推送的格式、分组、样式规则
- **[总结分析方法论](analysis-guide.md)** — 推文总结、翻译、关键观点提炼

## 注意事项

- `config list` 中的 `hermes_deliver_target` 已废弃，xwatch 不读取，请忽略，勿据此推断推送渠道
- 推文数据来自本地缓存，只包含订阅后抓取到的内容
- 如果返回空结果，建议用户扩大时间范围或确认是否已订阅
- 推文原文可能是英文或其他语言，总结和推送时应翻译为中文
- screen_name 参数不需要 @ 前缀
