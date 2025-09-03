#!/bin/bash

# YUNAI 项目测试启动脚本

echo "🚀 YUNAI 项目全功能测试启动器"
echo "================================================================================"
echo ""

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go 环境"
    exit 1
fi

echo "✅ Go 环境检查通过"

# 检查服务器是否运行
echo "🔍 检查服务器状态..."
if curl -s http://localhost:8092/health > /dev/null 2>&1; then
    echo "✅ 服务器已运行"
else
    echo "📡 服务器未运行，正在启动..."
    
    # 启动服务器
    echo "🔧 启动 YUNAI 服务器..."
    go run cmd/monolith/main.go &
    SERVER_PID=$!
    
    echo "⏳ 等待服务器启动..."
    sleep 10
    
    # 再次检查服务器状态
    if curl -s http://localhost:8092/health > /dev/null 2>&1; then
        echo "✅ 服务器启动成功 (PID: $SERVER_PID)"
    else
        echo "❌ 服务器启动失败"
        kill $SERVER_PID 2>/dev/null
        exit 1
    fi
fi

echo ""
echo "🧪 开始执行全功能测试..."
echo ""

# 运行测试
go run test_all_features.go

echo ""
echo "📋 测试完成！"

# 如果启动了服务器，询问是否关闭
if [ ! -z "$SERVER_PID" ]; then
    echo ""
    read -p "是否关闭测试服务器? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "🔴 关闭服务器..."
        kill $SERVER_PID
        echo "✅ 服务器已关闭"
    else
        echo "ℹ️  服务器继续运行 (PID: $SERVER_PID)"
        echo "   使用 'kill $SERVER_PID' 手动关闭"
    fi
fi

echo ""
echo "🎉 测试流程完成！"
