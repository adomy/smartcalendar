# SmartCalendar

SmartCalendar 是一个支持智能日程管理的全栈项目，包含用户体系、日程协作、通知提醒、操作记录与 AI 辅助创建/修改/删除日程能力。前端基于 React + Ant Design + FullCalendar，后端基于 Gin + Gorm + SQLite，并通过 Eino 接入豆包 Ark 模型实现意图识别。

## 功能特性
- 用户注册与登录（JWT 鉴权）
- 日程创建、修改、删除与参与人协作
- 通知中心（邀请、变更、提醒）与未读统计
- 操作记录查询
- AI 自然语言日程处理（确认后执行）
- 管理员用户管理（启用/禁用/重置密码）

## 目录结构
- backend/：后端服务（Gin + Gorm + SQLite + Eino）
- frontend/：前端应用（React + Vite + Ant Design + FullCalendar）
- docs/：接口文档

## 环境要求
- Go 1.20+（建议使用项目现有版本）
- Node.js 18/20/22+

## 启动后端
```bash
cd backend
GOTOOLCHAIN=local go run -buildvcs=false .
```

健康检查：
```bash
curl http://localhost:8080/health
```

## 启动前端
```bash
cd frontend
npm install
npm run dev
```

## 前后端联调
- 默认使用 Vite 代理转发 `/api` 到 `http://localhost:8080`
- 若需要直接访问后端，可在 `frontend/.env.local` 配置：
```
VITE_API_BASE=http://localhost:8080/api
```

## 后端环境变量
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

## AI 使用说明
AI 接口：`POST /api/ai/chat`  
流程：
1. 用户输入自然语言，模型识别意图（创建/修改/删除）
2. 返回摘要与候选日程，等待确认
3. 用户确认后执行操作

修改/删除时若匹配到多条日程，会返回候选列表并要求指定日程 ID。

## 接口文档
详见 [docs/api-docs.md](file:///Users/bytedance/Projects/godev/aiproj/smartcalendar/docs/api-docs.md)

## 账号与权限
首位注册用户必须使用昵称 admin，自动成为管理员。

## 备注
后端与前端均包含 README，分别说明更细的模块与使用细节。
