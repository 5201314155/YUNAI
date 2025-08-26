# YUNAI 支付系统测试方案

## 文档信息
- **文档名称**: YUNAI 支付系统测试方案
- **版本**: v1.0
- **创建时间**: 2025-08-26
- **作者**: 小云
- **更新时间**: 2025-08-26

## 测试概述

本文档描述了 YUNAI 支付系统的完整测试方案，包括单元测试、集成测试、性能测试和安全测试等。

### 测试目标
- 确保支付功能的正确性和可靠性
- 验证系统的性能和稳定性
- 保证数据的安全性和一致性
- 提供完整的测试覆盖率

## 测试策略

### 测试金字塔
```
        ┌─────────────────┐
        │   E2E Tests     │  10%
        │   (UI Tests)    │
        ├─────────────────┤
        │ Integration     │  20%
        │    Tests        │
        ├─────────────────┤
        │   Unit Tests    │  70%
        │                 │
        └─────────────────┘
```

### 测试类型
1. **单元测试** (70%): 测试单个函数和方法
2. **集成测试** (20%): 测试模块间的交互
3. **端到端测试** (10%): 测试完整的业务流程

## 单元测试

### 测试覆盖范围

#### 1. 卡片管理测试
```go
// TestCardService_BindCard 测试卡片绑定
func TestCardService_BindCard(t *testing.T) {
    tests := []struct {
        name    string
        request *domain.BindCardRequest
        want    *domain.PaymentCard
        wantErr bool
    }{
        {
            name: "绑定标准卡成功",
            request: &domain.BindCardRequest{
                CardNumber:     "8825086242773296",
                CardholderName: "测试用户",
                BankName:       "YUNAI 标准卡",
                BankCode:       "YUNAI",
                CardType:       "debit",
                Email:          "test@example.com",
            },
            want: &domain.PaymentCard{
                CardNumber: "8825********3296",
                BankName:   "YUNAI 标准卡",
                Balance:    0.0,
                Currency:   "CNY",
                IsActive:   true,
                IsFrozen:   false,
            },
            wantErr: false,
        },
        {
            name: "卡号格式错误",
            request: &domain.BindCardRequest{
                CardNumber: "invalid_card_number",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试逻辑
        })
    }
}
```

#### 2. 充值订单测试
```go
// TestPaymentService_CreateRechargeOrder 测试创建充值订单
func TestPaymentService_CreateRechargeOrder(t *testing.T) {
    tests := []struct {
        name     string
        userID   uuid.UUID
        request  *domain.CreateRechargeOrderRequest
        expected *domain.PaymentOrderResponse
        wantErr  bool
    }{
        {
            name:   "创建标准充值订单",
            userID: uuid.New(),
            request: &domain.CreateRechargeOrderRequest{
                Amount:        100.0,
                PaymentMethod: "card",
                PaymentCardID: &uuid.UUID{},
            },
            expected: &domain.PaymentOrderResponse{
                Amount:      100.0,
                CoinsAmount: 1000,
            },
            wantErr: false,
        },
    }
}
```

#### 3. 代充功能测试
```go
// TestPaymentService_ProxyRecharge 测试代充功能
func TestPaymentService_ProxyRecharge(t *testing.T) {
    tests := []struct {
        name        string
        payerUserID uuid.UUID
        request     *domain.ProxyRechargeRequest
        expected    *domain.ProxyRechargeResponse
        wantErr     bool
    }{
        {
            name:        "返利卡代充成功",
            payerUserID: uuid.New(),
            request: &domain.ProxyRechargeRequest{
                TargetUserID:    uuid.New(),
                Amount:          50.0,
                PaymentCardID:   uuid.New(),
                PaymentPassword: "123456",
            },
            expected: &domain.ProxyRechargeResponse{
                OriginalAmount: 50.0,
                ActualAmount:   50.0,
                CoinsAmount:    500,
                BonusCoins:     100, // 20% 返利
                TotalCoins:     600,
            },
            wantErr: false,
        },
    }
}
```

### 测试工具和框架
- **测试框架**: Go 标准库 testing
- **断言库**: testify/assert
- **Mock 工具**: testify/mock
- **数据库**: 内存数据库或测试数据库

## 集成测试

### 测试场景

#### 1. 完整充值流程测试
```go
func TestIntegration_CompleteRechargeFlow(t *testing.T) {
    // 1. 创建测试用户
    user := createTestUser(t)
    
    // 2. 绑定支付卡片
    card := bindTestCard(t, user.ID)
    
    // 3. 给卡片充值
    rechargeCard(t, user.ID, card.ID, 100.0)
    
    // 4. 创建充值订单
    order := createRechargeOrder(t, user.ID, 50.0, card.ID)
    
    // 5. 处理支付
    processPayment(t, order.ID, "123456")
    
    // 6. 验证结果
    verifyWalletBalance(t, user.ID, 500) // 50元 = 500金币
    verifyCardBalance(t, card.ID, 50.0)  // 剩余50元
}
```

#### 2. 代充流程测试
```go
func TestIntegration_ProxyRechargeFlow(t *testing.T) {
    // 1. 创建付款人和目标用户
    payer := createTestUser(t)
    target := createTestUser(t)
    
    // 2. 为付款人绑定返利卡
    bonusCard := bindBonusCard(t, payer.ID)
    rechargeCard(t, payer.ID, bonusCard.ID, 100.0)
    
    // 3. 执行代充
    proxyRecharge(t, payer.ID, target.ID, bonusCard.ID, 50.0)
    
    // 4. 验证结果
    verifyCardBalance(t, bonusCard.ID, 50.0)    // 付款人卡片扣除50元
    verifyWalletBalance(t, target.ID, 600)      // 目标用户获得600金币(含返利)
}
```

### 数据库集成测试
- 使用真实的 PostgreSQL 测试数据库
- 每个测试用例使用独立的事务
- 测试完成后自动回滚数据

## 性能测试

### 测试指标
- **响应时间**: API 平均响应时间 < 200ms
- **吞吐量**: 支持 1000+ QPS
- **并发用户**: 支持 10000+ 并发用户
- **数据库性能**: 查询响应时间 < 100ms

### 测试场景

#### 1. 压力测试
```bash
# 使用 Apache Bench 进行压力测试
ab -n 10000 -c 100 -H "Authorization: Bearer $TOKEN" \
   -T "application/json" \
   -p recharge_request.json \
   http://localhost:8080/api/v1/payment/recharge/orders
```

#### 2. 负载测试
```yaml
# 使用 k6 进行负载测试
scenarios:
  recharge_load_test:
    executor: ramping-vus
    startVUs: 0
    stages:
      - duration: 2m
        target: 100
      - duration: 5m
        target: 100
      - duration: 2m
        target: 200
      - duration: 5m
        target: 200
      - duration: 2m
        target: 0
```

#### 3. 并发测试
```go
func TestConcurrency_ProxyRecharge(t *testing.T) {
    const numGoroutines = 100
    const numRequests = 10
    
    var wg sync.WaitGroup
    results := make(chan error, numGoroutines*numRequests)
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < numRequests; j++ {
                err := performProxyRecharge()
                results <- err
            }
        }()
    }
    
    wg.Wait()
    close(results)
    
    // 验证结果
    errorCount := 0
    for err := range results {
        if err != nil {
            errorCount++
        }
    }
    
    assert.Equal(t, 0, errorCount, "并发测试不应该有错误")
}
```

## 安全测试

### 测试内容

#### 1. 认证授权测试
```go
func TestSecurity_Authentication(t *testing.T) {
    tests := []struct {
        name       string
        token      string
        expectCode int
    }{
        {"有效Token", validToken, 200},
        {"无效Token", "invalid_token", 401},
        {"过期Token", expiredToken, 401},
        {"无Token", "", 401},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resp := makeRequestWithToken(tt.token)
            assert.Equal(t, tt.expectCode, resp.StatusCode)
        })
    }
}
```

#### 2. 输入验证测试
```go
func TestSecurity_InputValidation(t *testing.T) {
    maliciousInputs := []string{
        "<script>alert('xss')</script>",
        "'; DROP TABLE users; --",
        "../../../etc/passwd",
        "{{7*7}}",
    }
    
    for _, input := range maliciousInputs {
        resp := makeRequestWithInput(input)
        assert.Equal(t, 400, resp.StatusCode, "应该拒绝恶意输入")
    }
}
```

#### 3. 权限控制测试
```go
func TestSecurity_Authorization(t *testing.T) {
    // 用户A尝试访问用户B的卡片
    userA := createTestUser(t)
    userB := createTestUser(t)
    cardB := bindTestCard(t, userB.ID)
    
    // 用户A尝试操作用户B的卡片，应该被拒绝
    err := rechargeCardAsUser(userA.ID, cardB.ID, 100.0)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "权限不足")
}
```

## 测试数据管理

### 测试数据准备
```go
// TestDataBuilder 测试数据构建器
type TestDataBuilder struct {
    db *sql.DB
}

func (b *TestDataBuilder) CreateUser(username string) *domain.User {
    user := &domain.User{
        ID:       uuid.New(),
        Username: username,
        Email:    username + "@test.com",
        IsActive: true,
    }
    // 插入数据库
    return user
}

func (b *TestDataBuilder) CreateCard(userID uuid.UUID, cardType string) *domain.PaymentCard {
    generator, _ := auth.NewYunaiCardGeneratorByType(cardType)
    cardNumber, _ := generator.GenerateYunaiCard()
    
    card := &domain.PaymentCard{
        ID:         uuid.New(),
        UserID:     userID,
        CardNumber: generator.MaskYunaiCard(cardNumber),
        BankName:   "YUNAI " + cardType,
        Balance:    0.0,
        IsActive:   true,
    }
    // 插入数据库
    return card
}
```

### 测试数据清理
```go
func cleanupTestData(t *testing.T, db *sql.DB) {
    tables := []string{
        "wallet_transactions",
        "card_transactions", 
        "proxy_recharge_orders",
        "recharge_orders",
        "payment_cards",
        "wallets",
        "users",
    }
    
    for _, table := range tables {
        _, err := db.Exec("DELETE FROM " + table + " WHERE created_at > NOW() - INTERVAL '1 hour'")
        assert.NoError(t, err)
    }
}
```

## 测试执行

### 本地测试
```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/service

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### CI/CD 集成
```yaml
# GitHub Actions 配置
name: Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:14
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
    - uses: actions/checkout@v2
    - uses: actions/setup-go@v2
      with:
        go-version: 1.21
    
    - name: Run tests
      run: |
        go test -v -coverprofile=coverage.out ./...
        go tool cover -func=coverage.out
```

## 测试报告

### 覆盖率要求
- **总体覆盖率**: > 80%
- **核心业务逻辑**: > 90%
- **支付相关功能**: > 95%

### 测试报告格式
- 测试执行总结
- 覆盖率统计
- 性能测试结果
- 安全测试结果
- 问题和建议

---

**文档维护**: 本文档由小云维护，测试方案需要随着功能迭代持续更新。
