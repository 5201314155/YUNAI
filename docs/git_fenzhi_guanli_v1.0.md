# YUNAI Git 分支管理策略 v1.0

**文档版本**: v1.0  
**创建日期**: 2025-01-26  
**作者**: 小云 (YUNAI 核心开发工程师)  
**更新记录**: 初版创建，建立完整的分支管理策略  

---

## 🌿 分支结构

### 📋 **主要分支**

#### 🏠 **main** - 稳定主分支
- **用途**: 生产环境代码，稳定版本
- **保护**: 只能通过 PR 合并，不允许直接推送
- **内容**: 已完成并测试通过的功能
- **当前状态**: ✅ 7大核心功能完整实现

#### 🔧 **develop** - 开发集成分支  
- **用途**: 集成各个功能分支，预发布测试
- **合并来源**: 各个 feature 分支
- **合并目标**: main 分支（发布时）
- **当前状态**: ✅ 已创建并推送到远程

#### 🚀 **feature/*** - 功能开发分支
- **命名规范**: `feature/功能名称`
- **生命周期**: 功能开发完成后合并到 develop
- **隔离性**: 每个功能独立开发，互不影响

---

## 🎯 当前分支规划

### ✅ **已创建分支**

| 分支名 | 用途 | 状态 | 负责功能 |
|--------|------|------|----------|
| `main` | 稳定主分支 | ✅ 活跃 | 7大核心功能 |
| `develop` | 开发集成 | ✅ 活跃 | 新功能集成 |
| `feature/rtc-system` | 通话系统 | ✅ 开发中 | WebRTC + 虚拟人通话 |

### 📋 **计划创建分支**

| 分支名 | 用途 | 优先级 | 预计时间 |
|--------|------|--------|----------|
| `feature/media-workshop` | 媒体工坊 | 高 | 3-4周 |
| `feature/frontend-flutter` | Flutter前端 | 高 | 4-6周 |
| `feature/voice-call` | 语音通话 | 中 | 2-3周 |
| `feature/virtual-human` | 虚拟人系统 | 中 | 3-4周 |
| `feature/monitoring` | 监控面板 | 低 | 1-2周 |

---

## 🔄 工作流程

### 📝 **开发新功能流程**

1. **创建功能分支**
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/新功能名称
   ```

2. **开发功能**
   ```bash
   # 正常开发，提交代码
   git add .
   git commit -m "feat: 添加新功能"
   git push origin feature/新功能名称
   ```

3. **合并到开发分支**
   ```bash
   git checkout develop
   git pull origin develop
   git merge feature/新功能名称
   git push origin develop
   ```

4. **发布到主分支**
   ```bash
   git checkout main
   git pull origin main
   git merge develop
   git push origin main
   ```

### 🔒 **分支保护规则**

#### **main 分支保护**
- ❌ 禁止直接推送
- ✅ 必须通过 Pull Request
- ✅ 需要代码审查
- ✅ 需要通过所有测试

#### **develop 分支规则**
- ✅ 允许直接推送（开发阶段）
- ✅ 功能分支合并点
- ✅ 集成测试分支

#### **feature 分支规则**
- ✅ 完全自由开发
- ✅ 可以随时推送
- ✅ 不影响其他分支

---

## 🎯 当前开发计划

### 🚀 **Phase 1: 通话系统 (当前)**
**分支**: `feature/rtc-system`  
**时间**: 3-4周  
**功能**:
- WebRTC 语音/视频通话
- 虚拟人通话系统
- TTS/STT 语音处理
- 实时字幕和摘要

### 🎨 **Phase 2: 媒体工坊**
**分支**: `feature/media-workshop`  
**时间**: 3-4周  
**功能**:
- 文生图/图生图
- 图生视频/文生视频
- 抠图/修复/动画
- 参考图/参考视频

### 📱 **Phase 3: Flutter 前端**
**分支**: `feature/frontend-flutter`  
**时间**: 4-6周  
**功能**:
- 移动端 UI 实现
- 两图系统前端
- 通话界面
- 媒体生成界面

---

## 📊 分支管理命令

### 🔍 **查看分支状态**
```bash
# 查看所有分支
git branch -a

# 查看当前分支
git branch

# 查看远程分支
git branch -r
```

### 🔄 **分支切换**
```bash
# 切换到主分支
git checkout main

# 切换到开发分支  
git checkout develop

# 切换到功能分支
git checkout feature/rtc-system

# 创建并切换到新分支
git checkout -b feature/新功能名称
```

### 🔗 **分支同步**
```bash
# 拉取远程更新
git pull origin 分支名

# 推送本地分支
git push origin 分支名

# 推送并设置上游分支
git push -u origin 分支名
```

### 🔀 **分支合并**
```bash
# 合并功能分支到开发分支
git checkout develop
git merge feature/功能名称

# 合并开发分支到主分支
git checkout main  
git merge develop
```

---

## 🚨 注意事项

### ⚠️ **重要规则**
1. **永远不要直接在 main 分支开发**
2. **功能开发完成后及时合并到 develop**
3. **定期同步远程分支避免冲突**
4. **提交信息要清晰描述功能**

### 🔧 **冲突解决**
```bash
# 如果出现合并冲突
git status                    # 查看冲突文件
# 手动编辑冲突文件
git add 冲突文件名
git commit -m "resolve: 解决合并冲突"
```

### 📝 **提交信息规范**
```bash
feat: 新功能
fix: 修复bug  
docs: 文档更新
style: 代码格式
refactor: 重构
test: 测试相关
chore: 构建/工具相关
```

---

## 🎉 优势总结

### ✅ **分支管理优势**
1. **安全隔离** - 新功能开发不影响稳定代码
2. **并行开发** - 多个功能可以同时开发
3. **版本控制** - 可以随时回退到稳定版本
4. **代码审查** - 通过 PR 进行代码质量控制
5. **发布管理** - 可以控制功能发布时机

### 🚀 **当前状态**
- ✅ **main 分支**: 7大核心功能稳定运行
- ✅ **develop 分支**: 准备集成新功能
- ✅ **feature/rtc-system**: 准备开发通话系统
- ✅ **远程同步**: 所有分支已推送到 GitHub

---

**🌿 现在我们可以安全地在 `feature/rtc-system` 分支开发通话功能，完全不会影响 main 分支的稳定代码！**
