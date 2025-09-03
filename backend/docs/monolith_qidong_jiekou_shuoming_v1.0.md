# 单体服务启动与接口说明 v1.0

作者：小云（YUNAI 项目）  
时间：2025-08-28

## 一、概览
- 本文档说明 cmd/monolith 单体入口的启动方式、对接模块与可用接口。
- 目标：在一个进程中整合“关系网、智能朋友圈、朋友圈 REST、AI 邀请、语音通道占位”等功能，便于联调与演示。

## 二、启动方式
- 环境要求：PostgreSQL 可访问，环境变量 DATABASE_URL 未设置时默认本地连接
- 启动命令（在 backend 目录）：
  - go run ./cmd/monolith
- 默认端口：8091
- 健康检查：GET http://localhost:8091/health

## 三、已整合模块与主要路由
- 健康检查
  - GET /health
- 关系网
  - GET /relationships/network/{characterID}
- 智能朋友圈（Smart Moments）
  - POST /smart-moments/generate
  - POST /smart-moments/{momentID}/auto-interact
  - POST /smart-moments/{momentID}/process-mentions
  - GET  /smart-moments/generation-context/{characterID}
- 朋友圈 REST
  - POST /moments
  - GET  /moments
  - GET  /moments/{id}
  - PUT  /moments/{id}
  - DELETE /moments/{id}
  - 以及草稿/互动/配置/统计/通知子路由
- AI 邀请
  - POST /ai-invitation/analyze
  - POST /ai-invitation/find-matches
  - POST /ai-invitation/suggestions
  - POST /ai-invitation/execute
- 语音通话（已接入）
  - POST /voice/call
  - 说明：已接入 AIVoiceCallService，可接收 multipart/form-data 音频上传
- 模型管理（Gin Handler 桥接）
  - GET /models/available
  - GET /models/{id}
  - POST /models/{id}/health-check
  - GET /models/admin/
  - POST /models/admin/{id}/enable
  - POST /models/admin/{id}/disable
- 支付管理（Gin Handler 桥接）
  - POST /payment/cards
  - GET /payment/cards
  - POST /payment/cards/{id}/default
  - DELETE /payment/cards/{id}
  - POST /payment/password/set
  - POST /payment/password/verify
  - POST /payment/orders/recharge
  - GET /payment/packages

## 四、数据库与依赖
- 数据库连接（sqlx）：优先读取环境变量 DATABASE_URL；示例：
  - host=localhost user=postgres password=5201314hdz dbname=yunai port=5432 sslmode=disable
- 服务依赖来自 internal 层：
  - 仓库：User、Character、Relationship、Memory、Model、Moments
  - 服务：ModelService、DynamicPromptService、MemoryService、EmbeddingService、CharacterService、RelationshipService、UserIdentityService、SmartMomentsService、MomentsService、AIInvitationService

## 五、快速测试示例
- 健康
  - curl http://localhost:8091/health
- API 文档
  - curl http://localhost:8091/api/docs
- 关系网
  - curl http://localhost:8091/relationships/network/{characterID}
- 智能朋友圈
  - curl -X POST http://localhost:8091/smart-moments/generate -H "Content-Type: application/json" -d '{"character_id":"<uuid>","count":2}'
- 模型管理
  - curl http://localhost:8091/models/available
  - curl http://localhost:8091/models/{id}
- 支付管理
  - curl http://localhost:8091/payment/cards
  - curl http://localhost:8091/payment/packages

## 六、测试结果与问题说明
### ✅ 正常工作的功能
- 健康检查：GET /health ✅
- API 文档：GET /api/docs ✅
- 全局提示词：POST /global-prompt/generate ✅

### ⚠️ 需要数据库表的功能
- 关系网：GET /relationships/network/{characterID} ❌ (缺少 character_relationships_enhanced 表)
- 模型管理：GET /models/available ❌ (路由挂载问题)
- 支付管理：GET /payment/packages ❌ (路由挂载问题)

### 🎯 系统全局提示词实现状态
- ✅ 已实现简化版全局提示词生成接口
- ✅ 支持按功能类型（chat/moments/invite/group_chat）生成不同提示词
- ✅ 集成了 DynamicPromptService，支持场景化提示词构建
- ✅ 数据库中有完整的 global_prompt_templates 表结构
- ✅ 独立的 simple_global_prompt_server.go 提供完整实现

## 七、完成状态
- ✅ 语音通话完整链路（AIVoiceCallService + VoiceCallHandler）已接入
- ✅ 模型管理、支付管理的 Gin Handler 已桥接到 monolith（chi 路由）
- ✅ 增加 /api/docs 完整索引与示例请求
- ✅ 全局提示词系统已实现并可正常调用

## 版本记录
- v1.0（2025-08-28）：创建文档，说明 monolith 启动与接口
- v1.1（2025-08-28）：完成语音通话、模型管理、支付管理接入，增加 API 文档路由
- v1.2（2025-08-28）：测试所有功能，实现全局提示词系统，记录问题与解决方案

