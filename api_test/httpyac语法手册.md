这是一份简要的 **httpyac** 语法手册。

**httpyac** 是目前 VS Code 中功能最强大的 `.http` 文件执行工具，它完全兼容标准 `.http` 语法，并增加了**断言（Test）**、**变量捕获**和**脚本支持**。

本文档面向“够用 + 不踩坑”的日常开发场景；当你遇到登录/302/会话保持等问题时，优先参考本文的 **重定向** 与 **CookieJar** 小节。

---

### 1. 基础结构 (Basic)

使用 `###` 分隔不同的请求。

```http
### 获取用户信息
GET https://api.example.com/users/1
Authorization: Bearer {{token}}
Content-Type: application/json
```

---

### 2. 变量定义 (Variables)

httpyac 支持多种定义变量的方式。

#### 文件内定义

使用 `@` 符号定义局部变量：

```http
@hostname = http://localhost:3000
@userId = 1001

GET {{hostname}}/users/{{userId}}
```

#### 环境变量定义 (Environment Variables)

使用 `{{$dotenv <variable_name>}}` 定义环境变量：

```http
@base_url = {{$dotenv BASE_URL}}
@token = {{$dotenv TOKEN}}
```

---

### 3. 断言 (Assertions) - **核心功能**

这是 httpyac 区别于普通 REST Client 的地方。它支持简写断言和 JS 断言。
详见: https://httpyac.github.io/guide/assert.html

#### 简写断言 (Shorthand)

使用 `??` 开头，非常适合快速检查状态码或 JSON 字段。

```http
GET https://httpbin.org/json

# @test 状态码必须是 200
?? status == 200

# @test 检查 body 中的字段
?? js response.parsedBody.slideshow.author == Yours Truly
```

#### JavaScript 断言 (Complex)

它也支持更复杂的断言逻辑（NodeJS 断言库/第三方断言库）。

```http

GET https://httpbin.org/json
{{
  const { equal } = require('assert');
  test('status code 200', () => {
    equal(response.statusCode, 200);
  });
}}

```

---

### 4. 请求链与变量提取 (Chaining)

如何将 **请求 A** 的结果（如 Token）传给 **请求 B**？

#### 第一步：给请求命名并提取数据

使用 `# @name` 给请求命名，httpyac 会自动将响应存储在这个名字下。

```http
### 1. 登录
# @name loginReq
POST https://api.example.com/login
Content-Type: application/json

{ "username": "admin" }
```

#### 第二步：在后续请求中引用

直接使用 `{{请求名.body.字段}}`。

```http
### 2. 使用 Token
GET https://api.example.com/profile
# 直接引用上一个请求的响应
Authorization: Bearer {{loginReq.body.token}}
```

---

### 5. 实用元数据 (Metadata)

httpyac 使用 `# @关键字` 来控制行为：

| 关键字           | 作用                                        | 示例                    |
| :--------------- | :------------------------------------------ | :---------------------- |
| `# @name <id>`   | 给请求命名，用于后续引用                    | `# @name myLogin`       |
| `# @no-log`      | 不在输出窗口打印此请求的日志 (保护敏感数据) | `# @no-log`             |
| `# @noRedirect`  | 禁止自动跟随 30x 重定向（登录/302 断言必备） | `# @noRedirect`         |
| `# @noCookieJar` | 禁用 CookieJar（一般不建议）                | `# @noCookieJar`        |
| `# @loop <list>` | 循环发送请求                                | `# @loop for 10`        |
| `# @sleep <ms>`  | 请求后暂停多少毫秒                          | `# @sleep 1000`         |
| `# @forceRef`    | 强制重新执行依赖的请求                      | `# @forceRef`           |
| `# @note <text>` | 在执行请求前显示确认对话框                  | `# @note Are you sure?` |

---

### 6. 重定向(302/301)与 `# @noRedirect`（非常常用）

httpyac 默认会跟随重定向。比如登录接口常见行为是 **302 到某个页面**：

- 如果你要断言“登录返回 302”，或者要读取 **302 响应上的 header（如 `set-cookie`）**，需要加 `# @noRedirect`
- 如果你不关心中间 302，只关心最终页面（通常 200），可以不加

示例：

```http
### 登录（希望看到 302，而不是跟随到 200 页面）
# @noRedirect
POST https://api.example.com/login
Content-Type: application/x-www-form-urlencoded

username=admin&password=123

?? status == 302
```

---

### 7. CookieJar（会话保持的最佳实践）

httpyac 自带 CookieJar：同一个 `.http` 文件里，后续请求会自动携带前面响应里设置的 Cookie（只要你没有 `# @noCookieJar`）。

这意味着：

- **大多数情况下你不需要手动维护 `Cookie: ...` 头**
- “变量未定义 / Cookie 丢失 / 重定向循环”的问题通常也会少很多

---

### 8. 脚本(Scripting)与“导出变量”

httpyac 支持在请求前/后写 NodeJS 脚本（用 `{{` 和 `}}` 包起来）。

- 请求前脚本：放在请求行之前，会在发请求前执行
- 请求后脚本：放在请求体之后，会在收到响应后执行
- 在脚本里使用 `exports.xxx = ...` 可以把值导出为变量，供后续 `{{xxx}}` 使用

示例：从响应头里拿 `set-cookie` 并导出一个变量（仅用于调试/少数场景；一般建议直接用 CookieJar）

```http
# @noRedirect
POST https://api.example.com/login
Content-Type: application/x-www-form-urlencoded

username=admin&password=123

{{
  const raw = response.headers?.['set-cookie'];
  const setCookie = Array.isArray(raw) ? raw[0] : (typeof raw === 'string' ? raw : '');
  exports.session_cookie = setCookie.split(';')[0];
}}
```

---

### 9. 常用响应对象字段（写断言/脚本会用到）

在断言/脚本中常见可用字段（命名以 httpyac 的 `response` 为准）：

- `response.statusCode`：状态码
- `response.headers`：响应头（注意：`set-cookie` 可能是数组）
- `response.parsedBody`：解析后的 body（JSON 会被解析成对象；HTML/文本则按实现可能为空或为字符串）

---

### 10. 表单登录 + Session Cookie 的通用示例

很多后台管理系统采用“表单登录 + session cookie”的方式。推荐做法：

1) 登录请求使用 `# @noRedirect`，断言 302（确保真的登录成功，而不是悄悄被跟随/被重定向到登录页）
2) 后续需要登录的页面/接口不要手写 `Cookie:`，交给 CookieJar

示例（把路径替换成你的系统实际路径）：

```http
@baseUrl = http://localhost:8080
@adminUser = admin
@adminPass = change-me

### 登录
# @noRedirect
POST {{baseUrl}}/login
# @no-log
Content-Type: application/x-www-form-urlencoded

username={{adminUser}}&password={{adminPass}}

?? status == 302

### 登录后访问需要权限的页面（CookieJar 会自动带上会话）
# @noRedirect
GET {{baseUrl}}/admin

?? status == 200
```

---

### 11. 常见问题排查（建议先看这里）

1) **断言 302 失败，变成了 200**
   - 你大概率没有加 `# @noRedirect`，httpyac 已经自动跟随重定向到最终页面了

2) **一直跳回 `/admin/login` 或 “Redirected 10 times”**
   - 登录失败（用户名/密码不对），或你手动拼了错误的 `Cookie` 头导致会话无效
   - 建议：先只执行“登录”请求，看是否确实返回 302，并检查响应里是否有 `set-cookie`
   - 建议：不要手写 `Cookie:`，让 CookieJar 处理

3) **取不到 `set-cookie`**
   - 你没加 `# @noRedirect`（跟随之后拿到的是最终页面的响应头）
   - 或服务端没有在该响应上设置 cookie（登录失败/未命中正确接口）

---

### 12. 执行方式（VS Code / CLI）

VS Code（vscode-httpyac 插件）：

- 打开 `.http` 文件，把光标放在某个请求块里执行
- 或选择执行整个文件（不同版本 UI 文案可能略有不同）

CLI（适合 CI/批量跑）：

```bash
# 执行文件内所有请求
httpyac path/to/your.http --all

# 从指定行附近开始执行（调试脚本很有用）
httpyac path/to/your.http -l 1
```
