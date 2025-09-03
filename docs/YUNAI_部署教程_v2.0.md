# 🚀 YUNAI 完整部署教程 v2.0

## 📋 目录
- [系统要求](#系统要求)
- [环境准备](#环境准备)
- [数据库部署](#数据库部署)
- [后端服务部署](#后端服务部署)
- [前端部署](#前端部署)
- [功能测试](#功能测试)
- [常见问题](#常见问题)

## 🔧 系统要求

### 硬件要求
- **CPU**: 4核心以上
- **内存**: 8GB以上
- **存储**: 50GB以上可用空间
- **网络**: 稳定的互联网连接

### 软件要求
- **操作系统**: Windows 10/11, Ubuntu 20.04+, macOS 12+
- **Go**: 1.21+
- **Node.js**: 18+
- **PostgreSQL**: 14+
- **Git**: 最新版本

## 🛠️ 环境准备

### 1. 安装PostgreSQL

#### Windows
```bash
# 下载并安装PostgreSQL 14+
# https://www.postgresql.org/download/windows/
# 安装时设置密码为: 5201314hdz
```

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo -u postgres psql -c "ALTER USER postgres PASSWORD '5201314hdz';"
```

#### macOS
```bash
brew install postgresql
brew services start postgresql
psql postgres -c "ALTER USER postgres PASSWORD '5201314hdz';"
```

### 2. 安装Go语言

#### Windows
```bash
# 下载并安装Go 1.21+
# https://golang.org/dl/
```

#### Ubuntu/Debian
```bash
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### macOS
```bash
brew install go
```

### 3. 安装Node.js

#### Windows
```bash
# 下载并安装Node.js 18+
# https://nodejs.org/
```

#### Ubuntu/Debian
```bash
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs
```

#### macOS
```bash
brew install node
```

## 🗄️ 数据库部署

### 方法一：一键部署脚本（推荐）

#### Windows
```bash
cd backend/scripts
deploy_yunai_database.bat
```

#### Linux/Mac
```bash
cd backend/scripts
chmod +x deploy_yunai_database.sh
./deploy_yunai_database.sh
```

### 方法二：手动部署

#### 1. 创建数据库
```bash
# 设置密码环境变量
export PGPASSWORD=5201314hdz  # Linux/Mac
set PGPASSWORD=5201314hdz     # Windows

# 创建数据库
createdb -h localhost -U postgres -E UTF8 yunai
```

#### 2. 执行建表脚本
```bash
cd backend/scripts
psql -h localhost -U postgres -d yunai -f yunai_complete_database_setup.sql
```

#### 3. 验证部署
```bash
psql -h localhost -U postgres -d yunai -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';"
```

## 🖥️ 后端服务部署

### 1. 安装Go依赖
```bash
cd backend
go mod tidy
go mod download
```

### 2. 配置环境变量
```bash
# 创建 .env 文件
cat > .env << EOF
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=5201314hdz
DB_NAME=yunai

# 服务器配置
SERVER_PORT=8081
JWT_SECRET=yunai_jwt_secret_key_2025

# AI模型配置
DEEPSEEK_API_KEY=your_deepseek_api_key
OPENAI_API_KEY=your_openai_api_key
ANTHROPIC_API_KEY=your_anthropic_api_key

# 文件上传配置
UPLOAD_PATH=./uploads
MAX_FILE_SIZE=10MB
EOF
```

### 3. 启动后端服务
```bash
cd backend
go run yunai_core_server.go
```

### 4. 验证后端服务
```bash
# 检查服务状态
curl http://localhost:8081/health

# 访问API文档
# http://localhost:8081/swagger/index.html
```

## 🌐 前端部署

### 1. 安装前端依赖
```bash
cd frontend
npm install
# 或者使用yarn
yarn install
```

### 2. 配置前端环境
```bash
# 创建 .env.local 文件
cat > .env.local << EOF
NEXT_PUBLIC_API_BASE_URL=http://localhost:8081
NEXT_PUBLIC_WS_URL=ws://localhost:8081/ws
EOF
```

### 3. 启动前端服务
```bash
cd frontend
npm run dev
# 或者
yarn dev
```

### 4. 访问前端应用
```
http://localhost:3000
```

## 🧪 功能测试

### 快速验证
```bash
cd backend
go run test_complete_system.go
```

### 详细测试流程
1. **用户注册登录测试**
2. **AI角色创建测试**
3. **用户身份识别测试**
4. **全局欺诈提示词测试**
5. **复杂关系网络测试**
6. **群聊功能测试**
7. **剧情触发测试**
8. **AI主动拉人测试**
9. **朋友圈功能测试**
10. **钱包支付测试**
11. **通知系统测试**
12. **语音通话测试**

## 🔧 常见问题

### Q1: PostgreSQL连接失败
**解决方案:**
```bash
# 检查PostgreSQL服务状态
sudo systemctl status postgresql  # Linux
brew services list | grep postgresql  # macOS

# 重启PostgreSQL服务
sudo systemctl restart postgresql  # Linux
brew services restart postgresql  # macOS
```

### Q2: 数据库权限问题
**解决方案:**
```bash
# 修改PostgreSQL配置
sudo nano /etc/postgresql/14/main/pg_hba.conf
# 将 local all postgres peer 改为 local all postgres md5
sudo systemctl restart postgresql
```

### Q3: Go模块下载失败
**解决方案:**
```bash
# 设置Go代理
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
```

### Q4: 端口被占用
**解决方案:**
```bash
# 查找占用端口的进程
netstat -tulpn | grep :8081  # Linux
lsof -i :8081  # macOS
netstat -ano | findstr :8081  # Windows

# 终止进程
kill -9 <PID>  # Linux/Mac
taskkill /PID <PID> /F  # Windows
```

## 📊 数据库表结构说明

### 核心表
- **users**: 用户基础信息
- **characters**: AI角色信息
- **user_identity_contexts**: 用户身份识别
- **character_relationships_enhanced**: 复杂关系网络
- **group_chats**: 群聊信息
- **moments**: 朋友圈内容

### 功能表
- **global_prompt_templates**: 全局提示词模板
- **story_chapters**: 剧情章节
- **ai_invitations**: AI邀请记录
- **voice_calls**: 语音通话记录
- **wallets**: 用户钱包
- **notifications**: 通知系统

### 配置表
- **ai_models**: AI模型配置
- **system_configs**: 系统配置
- **world_settings**: 世界观设定
- **sensitive_words**: 敏感词库

## 🎯 部署检查清单

- [ ] PostgreSQL服务正常运行
- [ ] 数据库yunai创建成功
- [ ] 所有数据表创建完成（约20+个表）
- [ ] 初始化数据插入成功
- [ ] Go后端服务启动成功（端口8081）
- [ ] 前端服务启动成功（端口3000）
- [ ] API文档可访问
- [ ] 基础功能测试通过

## 🚀 生产环境部署

### Docker部署（推荐）
```bash
# 构建镜像
docker build -t yunai-backend ./backend
docker build -t yunai-frontend ./frontend

# 启动服务
docker-compose up -d
```

### 云服务器部署
1. 配置反向代理（Nginx）
2. 设置SSL证书
3. 配置域名解析
4. 设置防火墙规则
5. 配置监控和日志

## 📞 技术支持

如果在部署过程中遇到问题，请：
1. 检查日志文件
2. 验证配置参数
3. 查看常见问题解决方案
4. 联系技术支持团队

---

**🎉 祝您部署成功！开始体验YUNAI的强大功能吧！**
