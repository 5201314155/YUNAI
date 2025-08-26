# YUNAI 项目结构总览

## 📁 目录结构

```
YUNAI/
├── .augment/                          # Augment 配置（不可修改）
│   └── rules/
│       ├── 设定.md                    # 身份设定规范
│       └── 项目结构功能.md             # 项目功能规范
├── README.md                          # 项目说明
├── PROJECT_STRUCTURE.md               # 项目结构说明（本文件）
├── frontend/                          # Flutter 前端
│   ├── lib/
│   │   ├── features/                  # 功能模块
│   │   │   ├── auth/                  # 认证模块
│   │   │   │   ├── screens/           # 页面
│   │   │   │   ├── components/        # 组件
│   │   │   │   ├── states/            # 状态管理
│   │   │   │   └── repository/        # 数据仓库
│   │   │   ├── chat/                  # 聊天模块
│   │   │   ├── character/             # 角色模块
│   │   │   ├── media/                 # 媒体模块
│   │   │   ├── moments/               # 朋友圈模块
│   │   │   ├── wallet/                # 钱包模块
│   │   │   └── settings/              # 设置模块
│   │   ├── core/                      # 核心功能
│   │   │   ├── network/               # 网络层
│   │   │   ├── storage/               # 存储层
│   │   │   ├── utils/                 # 工具类
│   │   │   ├── constants/             # 常量定义
│   │   │   └── theme/                 # 主题配置
│   │   └── shared/                    # 共享模块
│   │       ├── widgets/               # 共享组件
│   │       ├── models/                # 数据模型
│   │       ├── services/              # 共享服务
│   │       └── providers/             # 状态提供者
│   ├── assets/                        # 资源文件
│   ├── test/                          # 测试文件
│   └── pubspec.yaml                   # Flutter 依赖配置
├── backend/                           # Go 后端
│   ├── cmd/                           # 服务入口
│   │   ├── auth-svc/                  # 认证服务
│   │   ├── user-svc/                  # 用户服务
│   │   ├── character-svc/             # 角色服务
│   │   ├── chat-svc/                  # 聊天服务
│   │   ├── media-svc/                 # 媒体服务
│   │   ├── relation-svc/              # 关系服务
│   │   ├── orchestrator-svc/          # 编排服务
│   │   ├── provider-hub-svc/          # 模型中心服务
│   │   ├── notification-svc/          # 通知服务
│   │   └── wallet-svc/                # 钱包服务
│   ├── internal/                      # 内部包
│   │   ├── transport/                 # 传输层
│   │   │   ├── http/                  # HTTP 传输
│   │   │   ├── ws/                    # WebSocket 传输
│   │   │   └── grpc/                  # gRPC 传输
│   │   ├── service/                   # 业务逻辑层
│   │   ├── repository/                # 数据访问层
│   │   ├── domain/                    # 领域模型
│   │   └── jobs/                      # 后台任务
│   ├── pkg/                           # 公共包
│   │   ├── auth/                      # 认证工具
│   │   ├── validate/                  # 验证工具
│   │   ├── obs/                       # 对象存储
│   │   ├── logger/                    # 日志工具
│   │   ├── database/                  # 数据库工具
│   │   ├── redis/                     # Redis 工具
│   │   ├── queue/                     # 消息队列
│   │   ├── websocket/                 # WebSocket 工具
│   │   └── grpc/                      # gRPC 工具
│   ├── api/                           # API 定义
│   ├── configs/                       # 配置文件
│   ├── scripts/                       # 脚本文件
│   ├── deployments/                   # 部署配置
│   ├── docs/                          # 后端文档
│   └── go.mod                         # Go 模块配置
├── docs/                              # 项目文档
│   ├── jiagou_zonglan/                # 架构总览
│   │   └── yunai_jichu_sheji_v1.0.md  # 基础架构设计
│   ├── jiekou_wendang/                # 接口文档
│   │   └── auth_wallet_v1.0.md        # 认证钱包接口
│   ├── mokuai_sheji/                  # 模块设计
│   ├── shuju_jiegou/                  # 数据结构
│   ├── ceshi_fangan/                  # 测试方案
│   ├── bushu_zhinan/                  # 部署指南
│   ├── yunwei_shouce/                 # 运维手册
│   ├── guanli_beifen/                 # 备份与恢复
│   └── guize_hegui/                   # 安全合规
├── configs/                           # 全局配置
│   ├── models.json                    # 模型配置
│   ├── feature_flags.json             # 功能开关配置
│   └── workflows.yaml                 # 工作流配置
├── scripts/                           # 自动化脚本
├── deployments/                       # 部署配置
└── tests/                             # 集成测试
```

## 🎯 核心特性

### 产品定位
- **YUNAI = AI 社交与创作平台**
- 1 位用户 + N 位角色的剧场式群聊
- 朋友圈、通话、媒体生成功能
- 关系网驱动行为与剧情触发

### 技术栈
- **前端**: Flutter + Riverpod + go_router
- **后端**: Go 微服务 + gRPC + PostgreSQL + Redis
- **实时通信**: WebSocket + WebRTC
- **AI 能力**: 多模型接入 + 动态路由
- **存储**: S3/MinIO + Qdrant 向量数据库

## 📋 开发规范

### 命名规范
- **文档**: 中文拼音 + 下划线 (例: `liaotian_jiekou_wendang_v1.0.md`)
- **目录**: 中文拼音 + 下划线 (例: `jiagou_zonglan/`)
- **代码**: 遵循各语言标准 (Go: snake_case, Dart: camelCase)

### 文档分类
- `jiagou_zonglan` - 架构总览
- `jiekou_wendang` - 接口文档  
- `mokuai_sheji` - 模块设计
- `shuju_jiegou` - 数据结构
- `ceshi_fangan` - 测试方案
- `bushu_zhinan` - 部署指南
- `yunwei_shouce` - 运维手册
- `guanli_beifen` - 备份与恢复
- `guize_hegui` - 安全合规

### 功能开关
所有功能通过 `feature_flags.json` 控制，支持：
- 用户类型限制 (basic/vip/creator/admin)
- 灰度发布 (rollout_percentage)
- 环境控制 (dev/stg/prod)
- 紧急开关 (emergency_switches)

### 模型管理
通过 `models.json` 统一管理：
- 多提供商接入 (OpenAI/Stability/Runway/ElevenLabs)
- 能力映射 (chat/txt2img/img2video/tts)
- 动态参数表单 (JSON Schema)
- 健康检查与降级链

## 🚀 快速开始

### 环境要求
- Flutter SDK >= 3.16.0
- Go >= 1.21
- PostgreSQL >= 14
- Redis >= 7.0

### 启动开发环境
```bash
# 后端服务
cd backend
go mod tidy
go run cmd/auth-svc/main.go

# 前端应用
cd frontend
flutter pub get
flutter run
```

### 配置说明
1. 复制配置模板到对应环境
2. 修改数据库连接信息
3. 配置第三方 API 密钥
4. 启动相关服务

## 📝 重要说明

### 不可修改文件
- `.augment/rules/` 下的所有文档为项目规范，**严禁修改**
- 所有开发必须严格遵循规范文档中的标准

### 卡密系统
- 格式: 16位纯数字 (如: 6688990012345678)
- 前缀: 668899 (YUNAI专用前缀)
- 支持折扣、返利、特权解锁
- 金币汇率: 1元 = 10金币 (可配置)

### 权限体系
- **basic**: 普通用户，基础功能
- **vip**: 会员用户，高级功能
- **creator**: 创作者，专业工具
- **admin**: 管理员，全部权限

## 📞 联系方式

如有问题请联系项目团队或查看相关文档。
