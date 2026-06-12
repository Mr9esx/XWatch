# CLI 操作场景

## 首次设置

当用户说"帮我设置 xwatch"、"初始化"、"配置账号"、"设置 X 账号"时：

**步骤 1：初始化**

执行 `~/.hermes/skills/xwatch/scripts/xwatch init`，确认数据目录和数据库创建成功。

**步骤 2：检查现有配置**

执行 `~/.hermes/skills/xwatch/scripts/xwatch config list --json`，查看是否已有 auth_token 和 ct0。如果已配置，告知用户并询问是否需要更新。

**步骤 3：引导获取 Cookie**

告诉用户如何获取 X 认证信息：

> 需要从浏览器获取你的 X 登录凭证。请按以下步骤操作：
> 1. 用浏览器登录 https://x.com
> 2. 按 F12 打开开发者工具
> 3. 切换到 Application（应用） 标签页
> 4. 左侧找到 Cookies -> https://x.com
> 5. 分别复制 `auth_token` 和 `ct0` 的值发给我

**步骤 4：接收并写入配置**

用户发来 token 值后，依次执行：
```bash
~/.hermes/skills/xwatch/scripts/xwatch config set x_auth_token <用户提供的值>
~/.hermes/skills/xwatch/scripts/xwatch config set x_ct0 <用户提供的值>
```

如果用户需要代理，也一并设置：
```bash
~/.hermes/skills/xwatch/scripts/xwatch config set proxy <代理地址>
```

如有 Cloudflare 验证，可额外设置 `cf_clearance` Cookie（可选）：
```bash
~/.hermes/skills/xwatch/scripts/xwatch config set cf_clearance <值>
```

**认证说明：** xwatch 只需 `x_auth_token` + `x_ct0`，不需要 `bearer_token`。Token 过期后需用户手动从浏览器重新复制并 `config set` 更新。如遇到 `verify_credentials` 相关错误，说明二进制版本过旧，需重新编译安装。运行 `xwatch version` 应显示 `auth check: cookie-only`。

**步骤 5：验证配置**

执行一次测试搜索来验证认证是否有效：
```bash
~/.hermes/skills/xwatch/scripts/xwatch search "twitter" --json
```

- 成功返回搜索结果：告知"账号配置成功，已通过验证！"
- 返回错误：引导重新检查 token、代理、Cookie 是否过期，然后重试。

## 订阅管理

**订阅用户** — 当用户说"订阅 xxx"、"关注 xxx"、"帮我看 xxx 的推文"时：

1. `~/.hermes/skills/xwatch/scripts/xwatch search "xxx" --json`
2. 多个结果时展示候选列表（@screen_name、昵称、followers、简介），让用户选择
3. `~/.hermes/skills/xwatch/scripts/xwatch subscribe <screen_name> [--tags <主题>]`

**取消订阅** — 当用户说"取消订阅 xxx"、"不看 xxx 了"时：

`~/.hermes/skills/xwatch/scripts/xwatch unsubscribe <screen_name>`

**查看订阅** — 当用户问"我订阅了谁"、"列表"时：

`~/.hermes/skills/xwatch/scripts/xwatch list [--tag <主题>] [--json]`

## 标签管理

当用户说"把 xxx 标为黄金"、"这是搞 AI 的账号"时：

```bash
~/.hermes/skills/xwatch/scripts/xwatch tag add <screen_name> 黄金
~/.hermes/skills/xwatch/scripts/xwatch subscribe <screen_name> --tags 黄金   # 订阅时直接打标
~/.hermes/skills/xwatch/scripts/xwatch list --tag 黄金 --json               # 查看某主题下的账号
~/.hermes/skills/xwatch/scripts/xwatch tag list                             # 查看所有主题
```

## 查看与拉取推文

**拉取新推文：**

```bash
~/.hermes/skills/xwatch/scripts/xwatch check [--tag <主题>] [--force]
```

**查询缓存推文：**

```bash
~/.hermes/skills/xwatch/scripts/xwatch tweets [--user <screen_name>] [--tag <主题>] [--since 24h] [--limit 100] [--json]
```

推文的排版和分析方法见 [推送排版规范](push-format.md) 和 [总结分析方法论](analysis-guide.md)。

## 设置与账号池

**调整设置** — 当用户说"把检查频率改成 30 秒"、"修改代理"等：

```bash
~/.hermes/skills/xwatch/scripts/xwatch config set <key> <value>
~/.hermes/skills/xwatch/scripts/xwatch config list
```

常见配置项：
- `default_check_interval` — 默认检查间隔（秒）
- `proxy` — 代理地址（如 socks5://host:port）

**账号池** — 多个 X 账号轮流发起请求，降低限流/封禁风险。池中有 active 账号时自动使用池模式，池为空则回退到 config 中的单账号。

```bash
~/.hermes/skills/xwatch/scripts/xwatch account add <名称> --auth-token <token> --csrf-token <ct0> [--cf-clearance <值>]
~/.hermes/skills/xwatch/scripts/xwatch account list --json
~/.hermes/skills/xwatch/scripts/xwatch account remove <名称>
```

账号池特性：
- 轮转调度：每次请求自动切换到下一个可用账号
- 自动降级：429 限流时标记冷却 15 分钟，自动切换下一个
- 认证失效保护：401/403 时标记禁用 24 小时
- 冷却恢复：冷却到期后自动恢复使用
- 全池耗尽提示：所有账号不可用时返回最早恢复时间
