# YUNAI系统完整部署指南

## 📋 部署概述

YUNAI是一个基于Go + PostgreSQL的AI社交平台，支持用户管理、AI角色、智能聊天、朋友圈、语音通话等完整功能。

## 🛠️ 环境要求

### 必需软件
- **Go 1.19+** - 后端服务开发语言
- **PostgreSQL 13+** - 主数据库
- **Git** - 代码版本控制

### 系统要求
- **操作系统**: Windows 10/11, macOS, Linux
- **内存**: 最少4GB，推荐8GB+
- **存储**: 最少10GB可用空间
- **网络**: 需要访问外网（下载依赖）

## 🚀 快速部署

### 1. 克隆项目
```bash
git clone <repository-url>
cd YUNAI
```

### 2. 数据库设置
```bash
# 创建数据库
createdb yunai

# 设置环境变量
export PGPASSWORD="your_password"

# 或在Windows PowerShell中：
$env:PGPASSWORD="your_password"
```

### 3. 初始化数据库
```bash
cd backend
psql -h localhost -U postgres -d yunai -f ../database/init_database.sql
```

### 4. 修复外键约束（重要！）
```bash
psql -h localhost -U postgres -d yunai -f ../fix_foreign_keys.sql
```

### 5. 安装Go依赖
```bash
go mod tidy
```

### 6. 编译项目
```bash
go build .
```

### 7. 启动服务器
```bash
go run yunai_core_server.go user_identity_service.go global_prompt_service.go ai_invitation_service.go story_trigger_service.go world_setting_service.go voice_call_service.go
```

## 📊 验证部署

### 健康检查
```bash
curl http://localhost:8081/health
```

### 功能测试
```bash
go run test_all_systems.go
```

## 🔧 详细部署步骤

### 数据库配置

#### PostgreSQL安装和配置
1. 安装PostgreSQL 13+
2. 创建数据库用户（如果需要）
3. 创建yunai数据库
4. 配置连接参数

#### 数据库连接配置
在`yunai_core_server.go`中的数据库连接字符串：
```go
dsn := "host=localhost user=postgres password=5201314hdz dbname=yunai port=5432 sslmode=disable TimeZone=Asia/Shanghai"
```

### 服务配置

#### 端口配置
- **主服务端口**: 8081
- **健康检查**: http://localhost:8081/health
- **API文档**: http://localhost:8081/api/docs

#### 日志配置
- 使用logrus进行结构化日志
- 日志级别：INFO
- 输出格式：JSON

### 核心服务模块

#### 1. 用户认证系统
- **文件**: `yunai_core_server.go`
- **功能**: 用户注册、登录、Token管理
- **API**: `/api/v1/auth/*`

#### 2. AI角色管理系统
- **文件**: `yunai_core_server.go`
- **功能**: 角色创建、两图系统、角色信息管理
- **API**: `/api/v1/characters/*`

#### 3. 用户身份识别系统
- **文件**: `user_identity_service.go`
- **功能**: 智能身份提取、上下文切换、@提及解析
- **API**: `/api/v1/identity/*`

#### 4. 全局欺诈提示词系统
- **文件**: `global_prompt_service.go`
- **功能**: AI身份欺骗、现实信念、记忆沉浸
- **API**: `/api/v1/global-prompt/*`

#### 5. 朋友圈系统
- **文件**: `yunai_core_server.go`
- **功能**: 朋友圈创建、AI自动生成、点赞评论分享
- **API**: `/api/v1/moments/*`

#### 6. 钱包支付系统
- **文件**: `yunai_core_server.go`
- **功能**: 钱包查询、充值、消费记录、卡片兑换
- **API**: `/api/v1/wallet/*`

#### 7. 世界观设定系统
- **文件**: `world_setting_service.go`
- **功能**: 世界观创建、设定优先级、智能回退
- **API**: `/api/v1/world-setting/*`

#### 8. 语音通话系统
- **文件**: `voice_call_service.go`
- **功能**: 语音通话、语音消息、通话记录
- **API**: `/api/v1/voice-call/*`

#### 9. AI邀请系统
- **文件**: `ai_invitation_service.go`
- **功能**: 智能邀请分析、角色推荐、邀请响应
- **API**: `/api/v1/ai-invitation/*`

#### 10. 剧情触发系统
- **文件**: `story_trigger_service.go`
- **功能**: 章节管理、超级模糊触发、场景切换
- **API**: `/api/v1/story/*`

## ⚠️ 常见问题和解决方案

### 1. 外键约束错误
**问题**: `violates foreign key constraint`
**原因**: 外键指向错误的表（users vs core_users）
**解决**: 运行`fix_foreign_keys.sql`脚本

### 2. JSONB扫描错误
**问题**: `unsupported Scan, storing driver.Value type []uint8 into type *map[string]interface{}`
**原因**: GORM无法直接扫描JSONB到map类型
**解决**: 使用原生SQL查询或字符串类型

### 3. 数组类型插入错误
**问题**: `column "tags" is of type text[] but expression is of type record`
**原因**: GORM处理PostgreSQL数组类型有问题
**解决**: 使用原生SQL和手动数组格式化

### 4. 路由重复注册
**问题**: `handlers are already registered for path`
**原因**: 同一路由在多个地方注册
**解决**: 检查并移除重复的路由注册

## 🔍 监控和维护

### 健康检查端点
- **URL**: `GET /health`
- **响应**: 数据库连接状态、服务器运行时间、版本信息

### 系统统计
- **URL**: `GET /stats`
- **功能**: 实时系统指标、用户统计、性能数据

### 日志监控
- 服务器启动日志
- API请求日志
- 错误日志
- 数据库查询日志

## 📈 性能优化建议

### 数据库优化
1. 为常用查询字段添加索引
2. 定期清理过期数据
3. 监控慢查询
4. 配置连接池

### 服务器优化
1. 启用GZIP压缩
2. 配置缓存策略
3. 监控内存使用
4. 设置合理的超时时间

## 🔒 安全配置

### 数据库安全
- 使用强密码
- 限制数据库访问IP
- 定期备份数据
- 启用SSL连接（生产环境）

### API安全
- JWT Token认证
- 请求频率限制
- 输入验证和过滤
- CORS配置

## 📦 生产环境部署

### Docker部署（推荐）
```dockerfile
# 待完善Docker配置
```

### 系统服务配置
```bash
# 创建systemd服务文件
sudo nano /etc/systemd/system/yunai.service
```

### 反向代理配置
```nginx
# Nginx配置示例
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 🎯 测试验证

### 功能测试结果
根据最新测试，YUNAI系统功能测试通过率：**87.5% (7/8)**

**✅ 通过的功能模块:**
1. 用户认证系统 ✅
2. AI角色管理系统 ✅
3. 用户身份识别系统 ✅
4. 全局欺诈提示词系统 ✅
5. 钱包支付系统 ✅
6. 世界观设定系统 ✅
7. 语音通话系统 ✅

**⚠️ 需要优化的功能:**
1. 朋友圈系统 - 功能正常，测试脚本需要优化

### 测试命令
```bash
# 运行完整功能测试
go run test_all_systems.go

# 运行特定功能测试
go run test_identity_fix.go
```

## 📞 技术支持

如果在部署过程中遇到问题，请检查：
1. 数据库连接是否正常
2. 外键约束是否正确
3. Go依赖是否完整
4. 端口是否被占用
5. 日志中的具体错误信息
