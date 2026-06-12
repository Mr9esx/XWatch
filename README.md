# XWatch

X (Twitter) 推文订阅与推送工具。配合 [Hermes Agent](https://github.com/NousResearch/hermes-agent) 实现对话式管理。

## 功能

- **推文订阅** — 订阅 X 用户，定时检查新推文
- **自动翻译** — 非中文推文自动翻译为中文（通过 Hermes LLM）
- **消息推送** — 新推文通过 Hermes 推送到 Telegram/微信等平台
- **对话管理** — 通过 Hermes 自然语言对话管理订阅、调整设置
- **推文总结** — 让 AI 总结指定用户或所有用户的推文动态
- **代理支持** — 支持 SOCKS5/HTTP 代理

## 架构

```
Hermes Agent (对话交互 + 定时调度 + 翻译 + 总结)
    ↕ 调用 CLI（skill 自带二进制）
XWatch CLI (Go, 推文抓取 + 去重 + 存储)
    ↕
X GraphQL API / SQLite
```

XWatch 作为一个自包含的 Hermes Skill 发布——SKILL.md 操作指南 + Go 二进制 + cron 脚本全部打包在一起。安装只需拷贝一个目录。

## 快速开始

### 1. 编译

```bash
cd xwatch
go build -o hermes/xwatch-skill/scripts/xwatch ./cmd/xwatch/
```

编译结果直接输出到 skill 目录内。

### 2. 安装到 Hermes

```bash
cp -r hermes/xwatch-skill/ ~/.hermes/skills/xwatch/
```

一条命令，完成安装。Hermes 会自动发现 xwatch skill。

### 3. 初始化（在 Hermes 对话中或命令行）

```bash
# 命令行方式
~/.hermes/skills/xwatch/scripts/xwatch init
~/.hermes/skills/xwatch/scripts/xwatch config set x_auth_token <your_auth_token>
~/.hermes/skills/xwatch/scripts/xwatch config set x_ct0 <your_ct0_token>

# 或直接在 Hermes 对话中说：
# "帮我初始化 xwatch"
```

### 4. 配置代理（如需要）

```bash
~/.hermes/skills/xwatch/scripts/xwatch config set proxy socks5://user:pass@host:port
```

### 5. 配置定时检查（在 Hermes 对话中）

```
/cron add "every 5m" "如果脚本输出是 [SILENT]，直接回复 [SILENT]。否则将非中文推文翻译为中文，逐条推送。" --script ~/.hermes/skills/xwatch/scripts/check.sh --name "xwatch" --deliver telegram,weixin
```

## Skill 目录结构

```
~/.hermes/skills/xwatch/
├── SKILL.md              # Hermes 操作指南（LLM 读取）
└── scripts/
    ├── xwatch            # Go 二进制（自带）
    └── check.sh          # cron 检查脚本
```

## CLI 命令

| 命令 | 说明 |
|------|------|
| `xwatch init` | 初始化配置 |
| `xwatch check [--force]` | 检查新推文（供 cron 调用） |
| `xwatch search <query> [--json]` | 搜索 X 用户 |
| `xwatch subscribe <name> [--interval N]` | 订阅用户 |
| `xwatch unsubscribe <name>` | 取消订阅 |
| `xwatch list [--json]` | 列出订阅 |
| `xwatch tweets [--user X] [--since 24h]` | 查询缓存推文 |
| `xwatch config get/set/list` | 配置管理 |

## 对话交互示例

```
你：帮我订阅 Elon Musk
AI：我帮你搜索了一下，找到以下用户：
    1. @elonmusk (Elon Musk) — 200.5M followers
    2. @elonmusk_bot ...
    你想订阅哪个？

你：第一个
AI：已订阅 @elonmusk，后续有新推文会第一时间通知你。

--- 几分钟后 ---

AI：@elonmusk 于 2026-06-11 18:30 发布了新推文：
    「SpaceX 星舰第七次试飞成功」
    原文：SpaceX Starship Flight 7 success...
    链接：https://x.com/elonmusk/status/123456789

你：总结一下 @elonmusk 今天的推文
AI：@elonmusk 今天共发了 5 条推文，主要关于...
```

## 配置项

| Key | 说明 | 默认值 |
|-----|------|--------|
| `default_check_interval` | 默认检查间隔（秒） | 300 |
| `proxy` | 代理地址 | (空) |
| `x_auth_token` | X auth_token Cookie | (空) |
| `x_ct0` | X ct0 Cookie | (空) |
| `hermes_deliver_target` | Hermes 推送目标 | telegram |

## 获取 X Cookie

1. 用浏览器登录 https://x.com
2. 打开开发者工具 (F12) -> Application -> Cookies -> https://x.com
3. 找到 `auth_token` 和 `ct0` 的值
4. 通过 `xwatch config set` 写入

## License

MIT
