# SmartCalendar API 文档（契约）

本文档是 SmartCalendar 前后端唯一接口契约。前端与后端必须严格遵循本文档的路径、参数、校验规则、响应结构与错误码约定。

## 1. 基础约定

### 1.1 Base URL

- 开发环境后端：`http://localhost:8080`
- 所有接口路径以下文档中的 `Path` 为准，默认以 `/api` 为前缀

### 1.2 统一响应格式

成功：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

失败：

```json
{
  "code": 40001,
  "message": "参数校验失败：邮箱格式不正确",
  "data": null
}
```

### 1.3 鉴权

- 使用 JWT Bearer Token
- 请求头：`Authorization: Bearer <token>`
- 若接口标注需要鉴权，未携带 / Token 无效时返回 401 系列错误码
- 用户被禁用（`status=disabled`）时，即使 JWT 未过期也必须被拦截，返回 40301

### 1.4 时间格式

所有时间字段使用 RFC3339 字符串（带时区偏移），示例：`2026-02-25T15:00:00+08:00`。

### 1.5 分页参数

如接口支持分页：

- `page`: 从 1 开始，默认 1
- `page_size`: 1-100，默认 20

统一分页响应结构：

```json
{
  "list": [],
  "page": 1,
  "page_size": 20,
  "total": 0
}
```

## 2. 错误码

| code | 含义 |
|---:|---|
| 0 | success |
| 40001 | 参数校验失败 |
| 40101 | 未登录或 Token 缺失 |
| 40102 | Token 无效或已过期 |
| 40301 | 无权限（含用户被禁用 / 非管理员 / 非创建者操作） |
| 40401 | 资源不存在 |
| 40901 | 资源冲突（如邮箱已注册） |
| 50000 | 服务器内部错误 |

## 3. 数据结构

### 3.1 UserSummary（用户摘要）

```json
{
  "id": 1,
  "nickname": "admin",
  "email": "admin@example.com",
  "avatar": "/upload/avatars/xxx.png",
  "role": "admin",
  "status": "active",
  "created_at": "2026-02-24T10:00:00+08:00",
  "updated_at": "2026-02-24T10:00:00+08:00"
}
```

字段说明：

- `role`: `user` / `admin`
- `status`: `active` / `disabled`

### 3.2 Event（日程）

```json
{
  "id": 100,
  "user_id": 1,
  "title": "产品评审会",
  "type": "work",
  "start_time": "2026-02-25T15:00:00+08:00",
  "end_time": "2026-02-25T17:00:00+08:00",
  "location": "3楼会议室",
  "description": "评审本周版本",
  "created_at": "2026-02-24T10:00:00+08:00",
  "updated_at": "2026-02-24T10:10:00+08:00",
  "is_creator": true,
  "is_collaboration": false,
  "creator": {
    "id": 1,
    "nickname": "admin",
    "email": "admin@example.com",
    "avatar": "/upload/avatars/xxx.png",
    "role": "admin",
    "status": "active",
    "created_at": "2026-02-24T10:00:00+08:00",
    "updated_at": "2026-02-24T10:00:00+08:00"
  },
  "participants": [
    {
      "user_id": 2,
      "user": {
        "id": 2,
        "nickname": "张三",
        "email": "zhangsan@example.com",
        "avatar": "/upload/avatars/a.png",
        "role": "user",
        "status": "active",
        "created_at": "2026-02-24T10:00:00+08:00",
        "updated_at": "2026-02-24T10:00:00+08:00"
      }
    }
  ]
}
```

字段说明：

- `type`: `work` / `life` / `growth`
- `is_creator`: 当前登录用户是否为创建者（用于前端控制编辑/删除/拖拽权限）
- `is_collaboration`: 当前登录用户是否为参与人但非创建者（用于前端展示“协作”标识）
- `participants`: 日程参与人列表（包含用户摘要）

### 3.3 OperationLog（操作记录）

```json
{
  "id": 1000,
  "user_id": 1,
  "action": "update",
  "target_title": "产品评审会",
  "detail": "{\"before\":{\"start_time\":\"...\"},\"after\":{\"start_time\":\"...\"}}",
  "created_at": "2026-02-24T11:00:00+08:00"
}
```

字段说明：

- `action`: `create` / `update` / `delete`
- `detail`: JSON 字符串，包含变更前后摘要（由后端生成）

### 3.4 Notification（通知）

```json
{
  "id": 2000,
  "user_id": 2,
  "type": "invitation",
  "content": "admin 邀请你参加日程《产品评审会》",
  "event_id": 100,
  "is_read": false,
  "created_at": "2026-02-24T10:05:00+08:00"
}
```

字段说明：

- `type`: `reminder` / `invitation` / `change`

## 4. 用户与鉴权

### 4.1 注册

- Method: `POST`
- Path: `/api/auth/register`
- Auth: 无

请求体：

| 字段 | 类型 | 必填 | 校验规则 |
|---|---|---:|---|
| nickname | string | 是 | 1-50 字符 |
| email | string | 是 | 合法邮箱格式，最大 100 |
| password | string | 是 | 6-50 字符 |
| avatar | string | 否 | 头像 URL，最大 500 |

特殊规则：

- 系统首位注册用户必须使用 `nickname=admin`，否则返回 `40001`，message 为 `系统首位用户须以 admin 身份注册`
- 首位用户注册成功后 `role=admin`，后续用户默认 `role=user`
- 邮箱已存在返回 `40901`

请求示例：

```json
{
  "nickname": "admin",
  "email": "admin@example.com",
  "password": "Admin@123",
  "avatar": "/upload/avatars/admin.png"
}
```

响应 `data`：

```json
{
  "token": "jwt-token-string",
  "user": {
    "id": 1,
    "nickname": "admin",
    "email": "admin@example.com",
    "avatar": "/upload/avatars/admin.png",
    "role": "admin",
    "status": "active",
    "created_at": "2026-02-24T10:00:00+08:00",
    "updated_at": "2026-02-24T10:00:00+08:00"
  }
}
```

### 4.2 登录

- Method: `POST`
- Path: `/api/auth/login`
- Auth: 无

请求体：

| 字段 | 类型 | 必填 | 校验规则 |
|---|---|---:|---|
| email | string | 是 | 合法邮箱格式 |
| password | string | 是 | 1-50 字符 |

请求示例：

```json
{
  "email": "admin@example.com",
  "password": "Admin@123"
}
```

响应 `data`：

```json
{
  "token": "jwt-token-string",
  "user": {
    "id": 1,
    "nickname": "admin",
    "email": "admin@example.com",
    "avatar": "/upload/avatars/admin.png",
    "role": "admin",
    "status": "active",
    "created_at": "2026-02-24T10:00:00+08:00",
    "updated_at": "2026-02-24T10:00:00+08:00"
  }
}
```

失败示例（用户被禁用）：

```json
{
  "code": 40301,
  "message": "账号已被禁用",
  "data": null
}
```

### 4.3 获取当前用户信息

- Method: `GET`
- Path: `/api/user/profile`
- Auth: JWT

响应 `data`：UserSummary

### 4.4 更新个人信息

- Method: `PUT`
- Path: `/api/user/profile`
- Auth: JWT

请求体：

| 字段 | 类型 | 必填 | 校验规则 |
|---|---|---:|---|
| nickname | string | 否 | 1-50 字符 |
| avatar | string | 否 | 最大 500 |

请求示例：

```json
{
  "nickname": "张三",
  "avatar": "/upload/avatars/a.png"
}
```

响应 `data`：UserSummary

### 4.5 头像上传

- Method: `POST`
- Path: `/api/upload/avatar`
- Auth: JWT
- Content-Type: `multipart/form-data`

表单字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| file | file | 是 | 头像文件（建议 png/jpg/webp） |

响应 `data`：

```json
{
  "url": "/upload/avatars/20260224_101010_admin.png"
}
```

### 4.6 搜索用户（参与人选择器）

- Method: `GET`
- Path: `/api/users/search`
- Auth: JWT

Query 参数：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| keyword | string | 是 | 1-50 字符，按昵称/邮箱模糊搜索 |
| page | number | 否 | 默认 1 |
| page_size | number | 否 | 默认 20 |

响应 `data`：

```json
{
  "list": [
    {
      "id": 2,
      "nickname": "张三",
      "email": "zhangsan@example.com",
      "avatar": "/upload/avatars/a.png",
      "role": "user",
      "status": "active",
      "created_at": "2026-02-24T10:00:00+08:00",
      "updated_at": "2026-02-24T10:00:00+08:00"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total": 1
}
```

## 5. 管理员模块（admin）

### 5.1 获取所有用户列表

- Method: `GET`
- Path: `/api/admin/users`
- Auth: admin

Query 参数：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| page | number | 否 | 默认 1 |
| page_size | number | 否 | 默认 20 |

响应 `data`：

```json
{
  "list": [
    {
      "id": 1,
      "nickname": "admin",
      "email": "admin@example.com",
      "avatar": "/upload/avatars/admin.png",
      "role": "admin",
      "status": "active",
      "created_at": "2026-02-24T10:00:00+08:00",
      "updated_at": "2026-02-24T10:00:00+08:00"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total": 1
}
```

### 5.2 禁用 / 启用用户

- Method: `PUT`
- Path: `/api/admin/users/:id/status`
- Auth: admin

Path 参数：

| 参数 | 类型 | 必填 |
|---|---|---:|
| id | number | 是 |

请求体：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| status | string | 是 | `active` / `disabled` |

请求示例：

```json
{
  "status": "disabled"
}
```

响应 `data`：UserSummary

### 5.3 重置用户密码

- Method: `PUT`
- Path: `/api/admin/users/:id/reset-password`
- Auth: admin

Path 参数：

| 参数 | 类型 | 必填 |
|---|---|---:|
| id | number | 是 |

响应 `data`：

```json
{
  "user_id": 2,
  "new_password": "Smart@123"
}
```

## 6. 日程模块

### 6.1 新建日程

- Method: `POST`
- Path: `/api/events`
- Auth: JWT

请求体：

| 字段 | 类型 | 必填 | 校验规则 |
|---|---|---:|---|
| title | string | 是 | 1-100 字符 |
| type | string | 是 | `work` / `life` / `growth` |
| start_time | string | 是 | RFC3339 |
| end_time | string | 是 | RFC3339，且必须晚于 start_time |
| participant_ids | number[] | 否 | 参与人 user_id 列表（可为空数组） |
| location | string | 否 | 最大 200 |
| description | string | 否 | 最大 500 |

请求示例：

```json
{
  "title": "产品评审会",
  "type": "work",
  "start_time": "2026-02-25T15:00:00+08:00",
  "end_time": "2026-02-25T17:00:00+08:00",
  "participant_ids": [2, 3],
  "location": "3楼会议室",
  "description": "评审本周版本"
}
```

响应 `data`：Event

副作用（后端必须保证）：

- 自动写入 OperationLog（`action=create`）
- 为每个参与人生成 `invitation` 通知（创建者本人不生成 invitation）

### 6.2 查询日程（自建 + 协作）

- Method: `GET`
- Path: `/api/events`
- Auth: JWT

Query 参数：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| type | string | 否 | `work` / `life` / `growth` |
| start | string | 否 | RFC3339，范围开始（含） |
| end | string | 否 | RFC3339，范围结束（不含） |

响应 `data`：

```json
{
  "list": [
    {
      "id": 100,
      "user_id": 1,
      "title": "产品评审会",
      "type": "work",
      "start_time": "2026-02-25T15:00:00+08:00",
      "end_time": "2026-02-25T17:00:00+08:00",
      "location": "3楼会议室",
      "description": "评审本周版本",
      "created_at": "2026-02-24T10:00:00+08:00",
      "updated_at": "2026-02-24T10:10:00+08:00",
      "is_creator": true,
      "is_collaboration": false,
      "creator": {
        "id": 1,
        "nickname": "admin",
        "email": "admin@example.com",
        "avatar": "/upload/avatars/admin.png",
        "role": "admin",
        "status": "active",
        "created_at": "2026-02-24T10:00:00+08:00",
        "updated_at": "2026-02-24T10:00:00+08:00"
      },
      "participants": []
    }
  ]
}
```

说明：

- 返回范围内所有与当前用户相关的日程：
  - `user_id = 当前用户`
  - 或当前用户在 `participants` 中

### 6.3 获取日程详情

- Method: `GET`
- Path: `/api/events/:id`
- Auth: JWT

响应 `data`：Event

### 6.4 更新日程（仅创建者）

- Method: `PUT`
- Path: `/api/events/:id`
- Auth: JWT

权限：

- 仅创建者可更新，否则返回 `40301`

请求体（允许部分更新，但后端需校验 end_time > start_time 的最终结果）：

| 字段 | 类型 | 必填 | 校验规则 |
|---|---|---:|---|
| title | string | 否 | 1-100 |
| type | string | 否 | `work` / `life` / `growth` |
| start_time | string | 否 | RFC3339 |
| end_time | string | 否 | RFC3339 |
| participant_ids | number[] | 否 | 替换为新的参与人列表 |
| location | string | 否 | 最大 200 |
| description | string | 否 | 最大 500 |

响应 `data`：Event

副作用（后端必须保证）：

- 自动写入 OperationLog（`action=update`）
- 通知所有参与人生成 `change` 通知（更新后参与人集合，以更新后的为准；不包含创建者）

### 6.5 删除日程（仅创建者）

- Method: `DELETE`
- Path: `/api/events/:id`
- Auth: JWT

权限：

- 仅创建者可删除，否则返回 `40301`

响应 `data`：

```json
{
  "deleted": true
}
```

副作用（后端必须保证）：

- 自动写入 OperationLog（`action=delete`）
- 通知所有参与人生成 `change` 通知（不包含创建者）

## 7. 操作记录模块

### 7.1 查询当前用户操作记录

- Method: `GET`
- Path: `/api/operation-logs`
- Auth: JWT

Query 参数：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| action | string | 否 | `create` / `update` / `delete` |
| page | number | 否 | 默认 1 |
| page_size | number | 否 | 默认 20 |

响应 `data`：

```json
{
  "list": [
    {
      "id": 1000,
      "user_id": 1,
      "action": "update",
      "target_title": "产品评审会",
      "detail": "{\"before\":{\"start_time\":\"...\"},\"after\":{\"start_time\":\"...\"}}",
      "created_at": "2026-02-24T11:00:00+08:00"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total": 1
}
```

## 8. 通知模块

### 8.1 获取通知列表

- Method: `GET`
- Path: `/api/notifications`
- Auth: JWT

Query 参数：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| is_read | boolean | 否 | `true` / `false` |
| page | number | 否 | 默认 1 |
| page_size | number | 否 | 默认 20 |

响应 `data`：

```json
{
  "list": [
    {
      "id": 2000,
      "user_id": 2,
      "type": "reminder",
      "content": "您的日程《产品评审会》将在 15 分钟后开始",
      "event_id": 100,
      "is_read": false,
      "created_at": "2026-02-25T14:45:00+08:00"
    }
  ],
  "page": 1,
  "page_size": 20,
  "total": 1
}
```

### 8.2 获取未读通知数量

- Method: `GET`
- Path: `/api/notifications/unread-count`
- Auth: JWT

响应 `data`：

```json
{
  "count": 3
}
```

### 8.3 标记单条通知为已读

- Method: `PUT`
- Path: `/api/notifications/:id/read`
- Auth: JWT

响应 `data`：Notification

### 8.4 全部标为已读

- Method: `PUT`
- Path: `/api/notifications/read-all`
- Auth: JWT

响应 `data`：

```json
{
  "updated": 3
}
```

## 9. AI 模块

### 9.1 自然语言日程操作

- Method: `POST`
- Path: `/api/ai/chat`
- Auth: JWT

请求体：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| message | string | 是 | 用户自然语言输入，1-1000 字符 |
| confirm_id | string | 否 | 当上次返回 `need_confirm` 时携带 |
| confirm | boolean | 否 | 当用户点击“确认执行”时置为 `true` |

请求示例：

```json
{
  "message": "明天下午3点到5点在3楼会议室开产品评审会，邀请张三和李四"
}
```

响应 `data`：

```json
{
  "status": "success",
  "intent": "create",
  "result": "已为你创建日程：产品评审会 2026-02-25 15:00-17:00",
  "event": {
    "id": 100,
    "user_id": 1,
    "title": "产品评审会",
    "type": "work",
    "start_time": "2026-02-25T15:00:00+08:00",
    "end_time": "2026-02-25T17:00:00+08:00",
    "location": "3楼会议室",
    "description": "评审本周版本",
    "created_at": "2026-02-24T10:00:00+08:00",
    "updated_at": "2026-02-24T10:00:00+08:00",
    "is_creator": true,
    "is_collaboration": false,
    "creator": {
      "id": 1,
      "nickname": "admin",
      "email": "admin@example.com",
      "avatar": "/upload/avatars/admin.png",
      "role": "admin",
      "status": "active",
      "created_at": "2026-02-24T10:00:00+08:00",
      "updated_at": "2026-02-24T10:00:00+08:00"
    },
    "participants": []
  }
}
```

当需要用户确认时：

```json
{
  "status": "need_confirm",
  "intent": "create",
  "result": "我理解你想创建日程：产品评审会 2026-02-25 15:00-17:00（地点：3楼会议室，参与人：张三、李四）。是否确认创建？",
  "confirm_id": "c_20260224_xxx",
  "proposal": {
    "title": "产品评审会",
    "type": "work",
    "start_time": "2026-02-25T15:00:00+08:00",
    "end_time": "2026-02-25T17:00:00+08:00",
    "location": "3楼会议室",
    "participant_keywords": ["张三", "李四"],
    "description": ""
  }
}
```

前端点击“确认执行”后再次调用同一接口：

```json
{
  "message": "确认",
  "confirm_id": "c_20260224_xxx",
  "confirm": true
}
```

失败示例：

```json
{
  "code": 40001,
  "message": "参数校验失败：无法解析时间范围",
  "data": null
}
```

