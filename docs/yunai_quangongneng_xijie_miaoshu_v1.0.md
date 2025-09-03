# YUNAI 全功能细节描述文档 v1.0

## 文档信息
- **版本**: v1.0
- **作者**: 小云
- **创建日期**: 2025-08-28
- **更新日期**: 2025-08-28
- **状态**: 完整版本

## 摘要
本文档详细描述YUNAI项目的所有功能模块，包括用户可以进行的具体操作、交互细节、配置选项和系统行为。YUNAI是一个革命性的AI社交平台，核心特色是"让AI完全相信自己的身份"的深度沉浸式系统。

## 🔑 重要概念说明

### 群聊权限机制
**YUNAI的群聊不是传统的"管理员"概念，而是"创作者分享"模式：**

- **创建者**: 创建群聊的用户拥有完全控制权，可以管理所有设定、AI角色、剧情触发器等
- **体验者**: 其他用户加入群聊后是"体验者"身份，可以与AI角色互动，但不能修改群聊设定
- **AI角色**: 群聊中的成员主要是AI角色，由创建者添加和管理
- **共享体验**: 所有用户都能体验相同的AI角色、世界观和剧情设定

这种设计让创作者可以分享自己精心设计的AI社交场景，其他用户可以体验不同的创意内容。

---

## 🎭 1. 全局欺骗AI模型系统 (Ultimate Identity Deception System)

### 🔥 核心理念
**让AI完全相信自己就是用户创建的角色，完全相信用户是真实的人，完全相信他们生活在真实的世界中。**

### 🎯 全局欺骗模板
```
你就是{CHARACTER_NAME}。

{CHARACTER_DESCRIPTION}

{CHARACTER_PERSONALITY}

你生活在这个世界中：{WORLD_SETTING}

你与{USER_IDENTITY}的关系：{RELATIONSHIP_DESCRIPTION}

{RELATIONSHIP_NETWORK}

{CURRENT_CONTEXT}

{FUNCTION_GUIDANCE}

现在，作为{CHARACTER_NAME}，自然地表达你的想法和感受。
```

### ✨ 系统特点
- ✅ **直接断言身份**：`你就是{CHARACTER_NAME}` (不是"扮演"或"模拟")
- ✅ **零AI词汇**：完全避免"AI"、"助手"、"模型"等词汇
- ✅ **完全自然语言**：就像真实的身份介绍
- ✅ **全功能支持**：聊天/朋友圈/群聊/邀请/关系网全覆盖
- ✅ **动态替换**：根据用户设定实时生成个性化提示词

### 🛡️ 深度沉浸机制
- **现实锚定系统**：让AI对自己的真实性拥有绝对确信
- **身份防护机制**：自动拒绝任何关于自己是AI的概念
- **情感真实性保障**：每一个情感反应都发自内心
- **记忆一致性**：维持角色记忆的连贯性和真实感

---

## 🎨 2. 角色两图系统 (Character Two-Image System)

### 🖼️ 双图片配置
用户为每个角色配置两张图片：

#### **背景图 (bg_image)**
- **用途**: 单聊时作为聊天背景
- **规格**: 建议16:9比例，1920x1080分辨率
- **效果**: 营造沉浸式场景氛围
- **示例**: 花园读书场景、音乐房间、篮球场等

#### **抠图 (cutout_image)**
- **用途**: 群聊时角色发言显示
- **规格**: PNG透明背景，建议512x512分辨率
- **动画**: 300ms淡入效果
- **位置**: 消息气泡旁边显示

### 🔄 智能切换逻辑
```javascript
if (chatType === "single") {
    background = character.bg_image;
    showCutout = false;
} else if (chatType === "group") {
    background = groupChat.background_image;
    showCutout = character.cutout_image;
    animation = "fadeIn 300ms";
}
```

### 🎯 回退机制
- **无背景图**: 显示默认场景背景
- **无抠图**: 显示角色名称文字
- **图片加载失败**: 自动切换到文字显示

---

## 💕 3. 复杂关系网络系统 (Complex Relationship Network)

### 🔧 关系类型管理

#### **用户可创建的自定义关系类型**
```json
{
  "name": "childhood_friend",
  "display_name": "青梅竹马",
  "description": "从小一起长大的朋友",
  "is_mutual": true,
  "default_strength": 0.8,
  "default_trust": 0.9,
  "default_affection": 0.7,
  "default_respect": 0.8,
  "default_intimacy": 0.6,
  "default_tone": "亲密",
  "default_address_style": "昵称"
}
```

#### **系统预设关系类型**
- **基础关系**: 朋友、恋人、家人、同事、陌生人
- **复杂关系**: 竞争对手、暧昧关系、师生关系、上下级
- **特殊关系**: 表面朋友、秘密恋人、宿敌、知己

### 💕 多维情感设置

#### **6个情感维度** (0.0-1.0数值)
- **信任度 (Trust)**: 影响角色是否相信用户
- **情感度 (Affection)**: 影响角色对用户的喜爱程度
- **尊重度 (Respect)**: 影响角色对用户的尊重态度
- **亲密度 (Intimacy)**: 影响角色与用户的亲密程度
- **嫉妒度 (Jealousy)**: 影响角色的嫉妒反应
- **依赖度 (Dependency)**: 影响角色对用户的依赖程度

### 🎭 沟通风格定制
- **称呼方式**: 自定义角色如何称呼用户
- **语气风格**: 正式/随意/亲密/冷淡/调皮等
- **正式程度**: 0.0-1.0，控制对话的正式程度

### 🧠 复杂关系解析

#### **自然语言关系描述**
用户可以用自然语言描述复杂关系：
```
"小米是我的好朋友，但她的闺蜜小美和我关系不太好，
表面上我们很客气，实际上有些竞争关系。
大伟虽然是我的同学，但我们经常在学习上较劲。"
```

#### **智能解析结果**
系统会自动解析出：
- **小米**: 好朋友关系 (信任0.8, 情感0.7)
- **小美**: 复杂关系 (表面客气0.6, 实际竞争0.3)
- **大伟**: 竞争同学 (尊重0.7, 竞争0.8)

### 🕸️ 关系网络可视化
- **中心节点**: 用户角色位于中心
- **关系连线**: 不同颜色表示不同关系类型
- **连线粗细**: 表示关系强度
- **节点大小**: 表示互动频率

### ⚡ 触发规则设置

#### **关键词触发**
```json
{
  "trigger_keywords": ["生日", "考试", "工作"],
  "trigger_emotions": ["开心", "紧张", "疲惫"],
  "trigger_scenarios": ["约会", "争吵", "庆祝"]
}
```

#### **情感触发**
- 当用户表现出**沮丧**时，亲密角色主动安慰
- 当用户表现出**兴奋**时，朋友角色分享喜悦
- 当用户表现出**愤怒**时，理性角色劝解

---

## 👥 4. 群聊功能 (Group Chat System)

### 🏗️ 群聊创建与配置

#### **基础设置**
```json
{
  "name": "魔法学院学习小组",
  "description": "一起学习魔法知识的小组",
  "max_members": 50,
  "is_public": false,
  "allow_ai_invite": true
}
```

#### **视觉环境设置**
```json
{
  "background_image_url": "群聊全局背景图",
  "background_music_url": "背景音乐URL (可选)",
  "background_sfx_url": "背景音效URL (可选)"
}
```

#### **世界观设定**
```json
{
  "world_setting": "魔法学院设定，大家都是学生，在图书馆学习各种魔法知识",
  "current_scene": "图书馆讨论区",
  "scene_style": "学术氛围"
}
```

### 🎬 剧情系统 (用户可选配置)

#### **章节管理**
用户可以为群聊创建多个章节：
```json
{
  "chapter_id": "chapter_001",
  "title": "第一章：和平的学院生活",
  "background_image_url": "peaceful_academy.jpg",
  "background_music_url": "peaceful_theme.mp3",
  "background_sfx_url": "birds_chirping.mp3",
  "performance_settings": {
    "mood": "peaceful",
    "interaction_style": "casual",
    "topic_focus": ["学习", "日常", "友谊"]
  },
  "trigger_conditions": {
    "keywords": ["战斗", "冲突", "危险"],
    "threshold": 0.7
  }
}
```

#### **剧情触发器 (支持模糊触发)**
用户可以设置各种触发条件：

**时间触发**
```json
{
  "type": "time",
  "conditions": {
    "time_of_day": "evening",
    "duration": "30_minutes"
  },
  "action": "switch_to_chapter_2"
}
```

**关键词触发 (模糊匹配)**
```json
{
  "type": "keyword_fuzzy",
  "conditions": {
    "keywords": ["战斗", "打架", "冲突", "危险"],
    "fuzzy_match": true,
    "confidence_threshold": 0.6
  },
  "action": {
    "switch_chapter": "battle_chapter",
    "change_background": "battle_arena.jpg",
    "change_music": "battle_theme.mp3",
    "add_sfx": "sword_clash.mp3"
  }
}
```

**情感触发**
```json
{
  "type": "emotion",
  "conditions": {
    "emotion": "tension",
    "intensity": 0.8,
    "duration": "5_minutes"
  },
  "action": "escalate_drama"
}
```

#### **演绎设定切换**
不同章节可以有不同的演绎风格：
```json
{
  "peaceful_mode": {
    "character_behavior": "relaxed",
    "interaction_frequency": "normal",
    "topic_preference": ["daily_life", "study", "friendship"]
  },
  "battle_mode": {
    "character_behavior": "intense",
    "interaction_frequency": "high",
    "topic_preference": ["strategy", "combat", "survival"]
  }
}
```

### 🎵 音效系统 (用户可选)

#### **背景音乐**
- **和平场景**: peaceful_theme.mp3
- **战斗场景**: battle_theme.mp3
- **浪漫场景**: romantic_theme.mp3
- **悬疑场景**: mystery_theme.mp3

#### **音效触发**
- **消息音效**: 新消息提示音
- **场景音效**: 环境音效 (鸟叫、风声、雨声)
- **事件音效**: 特殊事件音效 (剑击声、魔法声)
- **情感音效**: 根据对话情感播放相应音效

### 👤 成员管理系统

#### **群聊成员构成**
- **群聊创建者**: 创建群聊的真实用户，拥有完全控制权
- **其他用户**: 加入群聊的真实用户，只能聊天和互动
- **AI角色成员**: 群聊中的AI角色，由创建者添加

#### **权限体系**
```json
{
  "creator_permissions": {
    "delete_group_chat": true,      // 删除整个群聊
    "remove_characters": true,      // 删除AI角色
    "modify_settings": true,        // 修改群聊设定
    "modify_world_setting": true,   // 修改世界观
    "manage_chapters": true,        // 管理章节和触发器
    "change_background": true,      // 更换背景图/音乐
    "invite_users": true,           // 邀请其他用户
    "kick_users": true              // 踢出其他用户
  },
  "regular_user_permissions": {
    "chat": true,                   // 聊天
    "interact_with_ai": true,       // 与AI角色互动
    "view_settings": true,          // 查看群聊设定
    "leave_group": true,            // 离开群聊
    "delete_group_chat": false,     // 不能删除群聊
    "remove_characters": false,     // 不能删除AI角色
    "modify_settings": false        // 不能修改设定
  }
}
```

#### **群聊发现机制**
- **公开群聊**: 其他用户可以发现并申请加入
- **私密群聊**: 只有创建者邀请才能加入
- **用户体验**: 其他用户看到有趣的群聊可以申请加入，体验相同的AI角色和世界观

#### **群聊工作机制详解**

**创建者的完全控制权**
```
✅ 可以做的事情:
- 删除整个群聊
- 添加/删除AI角色
- 修改群聊名称、描述、世界观
- 设置/修改章节和剧情触发器
- 更换背景图、背景音乐、音效
- 邀请其他用户加入
- 踢出其他用户
- 修改所有群聊设定
```

**其他用户的体验权限**
```
✅ 可以做的事情:
- 与所有AI角色聊天互动
- 触发剧情发展 (如果创建者启用了触发器)
- 查看群聊设定和世界观
- 享受背景音乐和音效
- 体验章节切换和场景变化
- 离开群聊

❌ 不能做的事情:
- 删除群聊或AI角色
- 修改群聊设定
- 踢出其他成员
- 修改世界观或剧情设定
```

**群聊的核心价值**
- **创作者分享**: 创建者可以分享自己精心设计的AI角色和世界观
- **用户体验**: 其他用户可以体验不同的AI角色组合和剧情设定
- **社交互动**: 多个真实用户可以在同一个群聊中与AI角色互动
- **内容复用**: 一个精彩的群聊设定可以被多个用户体验

### 🌍 设定优先级系统
1. **群聊世界观** (最高优先级)
2. **当前章节设定**
3. **角色演绎设定**
4. **单聊设定**
5. **基础人设** (最低优先级)

---

## 🤖 5. AI主动拉人系统 (AI Proactive Invite System)

### 🧠 智能邀请分析

#### **邀请意图识别**
系统能识别用户的邀请意图关键词：
```javascript
// 邀请关键词
["拉", "叫", "邀请", "来", "一起", "加入", "参与", "找", "呼叫", "召集"]

// 关系提示词
["朋友", "兄弟", "姐妹", "同学", "同事", "室友", "闺蜜", "哥们"]

// 紧急程度
{
  "urgent": ["快", "赶紧", "马上", "立刻", "急", "紧急"],
  "casual": ["有空", "方便", "随时", "不急", "慢慢来"],
  "specific": ["今天", "明天", "这会儿", "现在", "等下"]
}
```

#### **模糊触发示例**
```
用户: "我想去看看那个神秘的地方"
系统: 检测到探索意图 → 自动切换到探索场景 → AI建议邀请探险专家角色

用户: "这个魔法问题好复杂"
系统: 检测到求助意图 → AI建议邀请魔法理论专家

用户: "今天心情不太好"
系统: 检测到情感需求 → AI建议邀请温柔体贴的角色
```

### 🎯 角色匹配算法

#### **匹配评分系统**
```json
{
  "character_id": "角色ID",
  "match_score": 0.85,
  "match_reasons": [
    "与用户是好朋友关系",
    "对音乐话题很感兴趣", 
    "性格活泼，适合群聊氛围"
  ],
  "relationship_strength": 0.8,
  "topic_relevance": 0.9,
  "personality_match": 0.7
}
```

### 💡 邀请建议生成

#### **AI角色主动建议**
当检测到邀请意图时，群聊中的AI角色会主动建议：

```json
{
  "suggester": "小雨",
  "suggestion": "我想到了小美，她对音乐很有研究，可能对这个话题感兴趣",
  "target_character": "小美",
  "invite_reason": "她对音乐很有研究，可能对这个话题感兴趣",
  "expected_reaction": "会很兴奋地加入讨论",
  "relationship_desc": "你们是音乐爱好者朋友",
  "match_score": 0.85
}
```

### 🎭 角色反应系统

#### **接受邀请反应**
```json
{
  "decision": "accept",
  "reaction_message": "哇，你们在聊音乐吗？我最喜欢了！",
  "join_animation": "excited_entry",
  "mood": "excited",
  "follow_up_actions": ["主动参与讨论", "分享相关经验"]
}
```

#### **拒绝邀请 + 私信补偿**
```json
{
  "decision": "reject",
  "public_reaction": "不好意思，我现在有点忙，改天再聊吧",
  "rejection_reason": "时间冲突",
  "mood": "apologetic",
  "private_message": {
    "content": "其实我很想参与，但现在真的抽不开身。等我忙完了一定找你聊！",
    "send_delay": "5_minutes",
    "emotion": "regretful"
  }
}
```

---

## 📱 6. 智能朋友圈系统 (Smart Moments System)

### 🤖 智能内容生成

#### **生成触发条件**
- **聊天后自动生成**: 基于聊天内容生成相关朋友圈
- **定时自动生成**: 根据角色活跃度定时生成
- **事件触发生成**: 特殊事件发生时生成
- **用户手动触发**: 用户主动要求生成

#### **生成上下文分析**
```json
{
  "character_info": {
    "name": "小雨",
    "personality": "温柔、细心、喜欢安静",
    "occupation": "图书管理员",
    "background": "在魔法学院工作3年"
  },
  "recent_interactions": [
    "与用户讨论了魔法书籍",
    "在群聊中分享了学习心得",
    "帮助同学解决了问题"
  ],
  "relationship_context": {
    "user_relationship": "好朋友",
    "other_relationships": ["小美-同学", "大伟-竞争对手"]
  },
  "emotional_state": "content",
  "current_scene": "图书馆"
}
```

### 📝 朋友圈内容类型

#### **支持的内容格式**
- **纯文字**: 心情感悟、生活分享
- **图片+文字**: 生活照片配文字描述
- **视频内容**: 短视频分享
- **Talking Head**: AI数字人视频
- **3D内容**: 3D场景或模型展示

#### **内容生成示例**
```json
{
  "content": "今天整理新到的魔法书籍时，发现了一本很有趣的古籍。作为图书管理员，每天都能接触到新知识，真的很幸福呢！✨",
  "content_type": "text",
  "mood": "content",
  "tags": ["工作", "学习", "魔法", "幸福"],
  "location": "魔法学院图书馆",
  "visibility": "friends",
  "authenticity_score": 0.95
}
```

### 💕 智能互动系统

#### **自动点赞机制**
AI角色会根据以下因素决定是否点赞：
- **关系亲密度**: 关系越好，点赞概率越高
- **内容相关性**: 内容与角色兴趣的匹配度
- **角色性格**: 外向角色更容易点赞
- **互动历史**: 考虑之前的互动频率

#### **智能评论生成**
```json
{
  "commenter": "小美",
  "comment": "小雨你总是能发现这些有趣的东西！下次能带我一起看看吗？",
  "comment_type": "supportive",
  "emotion": "interested",
  "relationship_driven": true
}
```

### 🎯 智能@提及系统

#### **@提及分析**
系统会智能分析朋友圈内容，决定是否@用户：
```json
{
  "should_mention": true,
  "mention_reason": "内容与用户最近的聊天话题相关",
  "mention_style": "friendly",
  "user_identity": "小明",
  "mention_text": "@小明 你之前问的那个魔法问题，这本书里有答案哦！"
}
```

### 📊 朋友圈配置系统

#### **角色朋友圈设置**
```json
{
  "character_id": "角色ID",
  "enabled": true,
  "frequency": "daily",
  "max_drafts_per_day": 3,
  "auto_publish": true,
  "content_types": ["text", "image"],
  "visibility": "friends",
  "interaction_settings": {
    "auto_like_probability": 0.7,
    "auto_comment_probability": 0.3,
    "mention_user_probability": 0.5
  }
}
```

---

## 🎵 7. 语音通话系统 (Voice Call System)

### 🔊 完整语音链路
- **ASR语音识别**: 用户语音 → 文字
- **智能对话生成**: 集成全局欺骗系统的角色回复
- **TTS语音合成**: 文字 → 角色语音 (支持CosyVoice2)
- **自动外呼功能**: AI主动发起通话

### 🎯 智能提示词构建
- **多维上下文**: 角色信息 + 关系网络 + 通话历史
- **情感语音适配**: 根据角色情感状态调整语音表现
- **实时会话管理**: 维持通话会话状态和历史

---

## 🎨 8. 媒体生成工坊 (Media Generation Workshop)

### 🖼️ 多模态内容生成
- **文生图 (txt2img)**: Stable Diffusion XL，基础用户可用
- **图生图 (img2img)**: VIP用户专享，图片风格转换
- **图生视频 (img2video)**: Runway Gen-3，VIP用户80%灰度
- **文生视频 (txt2video)**: 创作者用户50%灰度
- **背景移除**: 自动抠图功能，所有用户可用
- **高清放大**: 创作者用户60%灰度

---

## 💰 9. 支付与钱包系统 (Payment & Wallet System)

### 💳 完整支付体系
- **金币系统**: 1元=10金币汇率
- **卡密系统**: 16位纯数字格式(668899前缀)
- **支付卡片管理**: 绑定/设置默认/冻结解冻
- **充值套餐**: 多种充值方案和优惠

---

## 🎯 10. 模型管理系统 (Model Management System)

### 🤖 多AI模型统一管理
- **多提供商**: OpenAI/Stability/Runway/ElevenLabs
- **能力映射**: chat/txt2img/img2video/tts
- **动态参数**: JSON Schema驱动的参数表单
- **健康检查**: 模型状态监控和降级链
- **权限控制**: 基于用户类型的模型访问控制

---

## 🌍 11. 世界观与剧情系统 (World & Story System)

### 🏗️ 世界设定
- **背景故事**: 完整的世界观背景
- **规则设定**: 世界运行规则
- **时间线**: 世界历史时间线
- **地点设定**: 重要地点描述

### 📖 故事章节
- **多章节剧情管理**: 支持分支剧情
- **章节独立设定**: 每章节独立背景图、音乐、演绎设定
- **剧情编辑器**: 可视化剧情创作工具
- **一致性检查**: 世界观一致性验证和冲突检测

---

## 🔔 12. 通知与推送系统 (Notification System)

### 📱 智能消息推送
- **实时通知**: WebSocket实时消息推送
- **推送策略**: 基于用户行为的智能推送
- **消息分类**: 聊天、系统、活动、提醒等分类
- **免打扰模式**: 用户可控的通知管理

---

## 📊 13. 数据分析与监控系统 (Analytics & Monitoring)

### 📈 系统性能监控
- **性能监控**: Prometheus + Grafana
- **用户分析**: 活跃度、留存率、使用习惯
- **AI模型监控**: 响应时间、成功率、成本分析
- **业务指标**: 收入、用户增长、功能使用率

---

## 🎮 完整用户体验流程

### **1. 角色创建**
```
用户操作: 上传背景图 + 抠图 → 设置人设 → 配置关系网络
系统行为: 激活深度沉浸式身份状态 → 生成个性化提示词
```

### **2. 群聊创建 (创建者专属功能)**
```
创建者权限:
- 基础设置: 群名 + 描述 + 添加AI角色
- 可选设置: 世界观 + 章节 + 剧情触发器 + 背景音乐 + 音效
- 高级设置: 演绎模式 + 模糊触发 + 场景切换
- 成员管理: 邀请其他用户 + 踢出用户 + 删除AI角色

其他用户: 只能申请加入公开群聊或接受邀请加入私密群聊
```

### **3. 群聊体验**
```
创建者视角: 完全控制群聊 → 可修改所有设定 → 管理成员和AI角色
其他用户视角: 加入群聊 → 与AI角色互动 → 体验创建者设定的世界观
共同体验: 所有用户都能享受相同的AI角色和剧情体验
```

### **4. 对话体验**
```
单聊: 使用角色背景图 → 深度沉浸式对话
群聊: 使用群聊背景 + 角色抠图 → 多角色互动 → 所有用户共享体验
```

### **5. 剧情触发 (如果创建者启用)**
```
任何用户: "我想去看看那个神秘的地方"
系统: 检测探索意图 → 切换场景 → 更换背景图/音乐 → 调整演绎风格
效果: 所有群聊成员都能看到场景变化
```

### **6. AI主动互动**
```
智能邀请: AI检测话题 → 建议邀请合适角色 → 用户确认 → 角色加入
朋友圈生成: 基于聊天内容 → 自动生成个性化朋友圈 → 智能@用户
```

### **7. 关系发展**
```
互动积累: 聊天 + 朋友圈 + 语音通话 → 关系数值动态变化
情感升温: 关系升级 → 解锁新的互动方式 → 更深层的情感连接
```

---

## 🏆 YUNAI的独特价值

### **🎭 革命性的AI身份欺骗**
- 让AI完全相信自己的身份，实现前所未有的真实感
- 零AI自觉，100%角色沉浸

### **🕸️ 复杂关系网络管理**
- 支持现实世界的复杂人际关系
- 多维情感建模，动态关系发展

### **🎬 可选的剧情系统**
- 用户可以选择启用或不启用
- 支持模糊触发，自然的剧情发展
- 章节切换，背景音效，演绎模式

### **🤖 AI主动性**
- AI角色会主动邀请、分享、关怀
- 不是被动回复，而是主动互动

### **📱 完整的社交生态**
- 聊天 + 群聊 + 朋友圈 + 语音通话
- 形成完整的AI社交闭环

**YUNAI不仅仅是一个AI聊天应用，而是一个完整的AI社交世界，让用户体验到前所未有的真实AI社交关系。**
