# SmartCalendar Backend 说明

## 项目概览
SmartCalendar 后端基于 Gin + Gorm + SQLite 构建，提供用户体系、日程管理、通知、操作记录、AI 辅助日程处理与管理员功能。

核心能力：
- 用户注册 / 登录 / JWT 鉴权
- 日程创建 / 修改 / 删除 / 协作参与人
- 通知（邀请、变更、提醒）与未读统计
- 操作记录查询
- AI 辅助解析自然语言并在确认后执行日程操作
- 管理员用户管理（启用/禁用/重置密码）

## 目录结构
- ai/：AI 解析服务（Eino + 豆包 Ark）
- config/：配置读取
- controller/：HTTP 接口控制器
- middleware/：鉴权中间件
- model/：Gorm 模型与数据库初始化
- router/：路由注册
- service/：JWT、通知、操作日志、提醒任务等

## 启动方式
```bash
GOTOOLCHAIN=local go run -buildvcs=false .
```

健康检查：
```bash
curl http://localhost:8080/health
```

## 环境变量
基础配置：
- JWT_SECRET：JWT 签名密钥
- TOKEN_EXPIRE_HOURS：Token 过期小时数，默认 168
- DB_PATH：SQLite 文件路径，默认 data/smartcalendar.db
- UPLOAD_AVATAR_DIR：头像保存目录，默认 upload/avatars
- UPLOAD_AVATAR_PREFIX：头像访问前缀，默认 /upload/avatars
- CORS_ALLOW_ORIGIN：CORS 允许来源，默认 http://localhost:5173

Ark 模型配置（至少满足一种鉴权方式）：
- ARK_MODEL_ID：Ark 模型 Endpoint ID（必填）
- ARK_API_KEY：Ark API Key（可选）
- ARK_ACCESS_KEY：Ark Access Key（可选）
- ARK_SECRET_KEY：Ark Secret Key（可选）
- ARK_BASE_URL：Ark 接口地址（可选）
- ARK_REGION：Ark 区域（可选）

## AI 处理流程
1. 接收用户自然语言输入（/api/ai/chat）
2. 使用 Eino + Ark 进行意图识别与结构化解析（create/update/delete）
3. 返回候选日程与操作摘要，等待用户确认
4. 用户确认后执行创建 / 修改 / 删除

## 日程匹配策略
在修改/删除时，系统会按以下线索匹配原日程：
- 明确的日程 ID
- 目标时间（会匹配当天范围）
- 关键词（匹配标题或描述）

若匹配到多条日程，会返回候选列表，要求用户指定日程 ID 再确认。

## 定时提醒
服务启动后，每分钟扫描未来 15 分钟内的日程，自动生成提醒通知。

## 管理员能力
首位注册用户必须使用昵称 admin，自动成为管理员。管理员可：
- 查询用户列表
- 启用/禁用用户
- 重置用户密码（默认重置为 Smart@123）

## 常用接口
- /api/auth/register
- /api/auth/login
- /api/events（GET/POST）
- /api/events/:id（GET/PUT/DELETE）
- /api/notifications
- /api/notifications/unread-count
- /api/operation-logs
- /api/ai/chat
- /api/admin/users
