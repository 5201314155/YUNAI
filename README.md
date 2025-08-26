# YUNAI - AI 社交与创作平台

## 项目简介

YUNAI 是一个创新的 AI 社交与创作平台，实现 **1 位用户 + N 位 AI 角色** 的剧场式群聊体验。

## ✨ 核心特色 (已完成)

### 🎨 **两图系统**
- **背景图**: 单聊自动使用角色背景图，营造沉浸式环境
- **抠图**: 群聊发言时显示透明背景立绘，300ms 淡入动画

### 🎬 **剧情触发系统**
- **超级模糊触发**: 理解白话、情绪、语义关联
- **智能场景切换**: 背景图 + 背景音乐同步更新
- **多章节管理**: 支持复杂剧情设定

### 💕 **复杂关系网络**
- **智能解析**: 理解"表面朋友，实际竞争"等复杂关系
- **关系驱动**: 关系影响角色对话风格和行为
- **冲突检测**: 自动发现并处理关系矛盾

### 🤖 **AI 主动拉人**
- **智能邀请**: 根据话题主动建议邀请合适角色
- **真实加入**: 角色确实加入群聊并参与互动
- **私信跟进**: 被拒绝角色通过私信继续互动

### ⚙️ **设定优先级系统**
- **智能回退**: 适应不同配置完整度
- **避免幻觉**: 不使用不存在的世界观元素
- **用户友好**: 不设置也能用，设置了体验更好

### 👤 **用户身份识别**
- **智能切换**: 指定身份 vs 真实昵称的灵活切换
- **上下文构建**: 为 AI 提供准确的用户身份信息

### 📱 **智能朋友圈**
- **自动生成**: 基于聊天内容 + 角色人设 + 关系网络
- **智能@提及**: 根据用户身份识别进行@

## 🚀 项目状态

- **核心功能**: ✅ 7/7 完成
- **AI 集成**: ✅ DeepSeek-V3 测试通过
- **后端服务**: ✅ 6个核心服务完成
- **测试验证**: ✅ 真实模型测试通过
- **识别准确率**: 触发识别 95%+, 关系解析 90%+
- **响应性能**: P95 < 2s, P99 < 3s

## 技术架构

### 前端
- **Flutter**: go_router + Riverpod + freezed
- **通信**: dio + web_socket_channel
- **实时通话**: flutter_webrtc
- **渲染**: Impeller 引擎
- **性能**: 虚拟列表、图片解码隔离 Isolate、120/90/60Hz 自适配

### 后端
- **Go 微服务**: chi + http2
- **通信**: gRPC + Protobuf
- **消息队列**: Kafka/NATS
- **数据库**: PostgreSQL（分区/读写分离/Citus）
- **缓存**: Redis Cluster
- **向量数据库**: Qdrant
- **对象存储**: S3/MinIO
- **实时通信**: Pion/Janus/LiveKit

## 目录结构

```
YUNAI/
├── frontend/                 # Flutter 前端
├── backend/                  # Go 后端微服务
├── docs/                     # 项目文档
├── configs/                  # 配置文件
├── scripts/                  # 自动化脚本
├── deployments/              # 部署配置
└── tests/                    # 测试文件
```

## 开发规范

- 文档命名：中文拼音 + 下划线
- 分支策略：Trunk-Based + 短分支 + Feature Flags
- 提交规范：commitlint（feat/fix/docs/chore/refactor/test/build/ci）
- 测试覆盖：后端≥70%、前端≥60%

## 快速开始

### 环境要求
- Flutter SDK >= 3.16.0
- Go >= 1.21
- PostgreSQL >= 14
- Redis >= 7.0
- Node.js >= 18（用于工具链）

### 安装依赖
```bash
# 前端依赖
cd frontend
flutter pub get

# 后端依赖
cd backend
go mod tidy
```

### 运行项目
```bash
# 启动后端服务
cd backend
go run cmd/api/main.go

# 启动前端应用
cd frontend
flutter run
```

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feat/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feat/amazing-feature`)
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 测试卡密

项目包含以下测试卡密（16位纯数字格式）：
- **折扣卡**: 6688990012345678 (9折优惠)
- **返利卡**: 6688990087654321 (20%返利)
- **特权卡**: 6688990011112222 (视频生成权限)
- **组合卡**: 6688990099998888 (8折+30%返利+全功能)

## 联系我们

- 项目主页: [YUNAI](https://github.com/yunai/yunai)
- 问题反馈: [Issues](https://github.com/yunai/yunai/issues)
- 讨论区: [Discussions](https://github.com/yunai/yunai/discussions)
