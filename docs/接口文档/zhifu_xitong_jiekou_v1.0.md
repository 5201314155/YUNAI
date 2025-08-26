# YUNAI 支付系统接口文档

## 文档信息
- **文档名称**: YUNAI 支付系统接口文档
- **版本**: v1.0
- **创建时间**: 2025-08-26
- **作者**: 小云
- **更新时间**: 2025-08-26

## 接口概述

YUNAI 支付系统提供完整的支付功能，包括卡片管理、充值订单、代充功能、钱包管理等。所有接口都需要 JWT 认证。

### 基础信息
- **Base URL**: `https://api.yunai.com/v1`
- **认证方式**: Bearer Token (JWT)
- **请求格式**: JSON
- **响应格式**: JSON
- **字符编码**: UTF-8

## 通用响应格式

### 成功响应
```json
{
  "code": 200,
  "message": "success",
  "data": {
    // 具体数据
  },
  "timestamp": "2025-08-26T13:30:00Z"
}
```

### 错误响应
```json
{
  "code": 400,
  "message": "请求参数错误",
  "error": "validation_failed",
  "details": {
    "field": "card_number",
    "message": "卡号格式不正确"
  },
  "timestamp": "2025-08-26T13:30:00Z"
}
```

## 1. 卡片管理接口

### 1.1 绑定支付卡片
**接口**: `POST /payment/cards/bind`

**请求参数**:
```json
{
  "card_number": "8825086242773296",
  "cardholder_name": "张三",
  "bank_name": "YUNAI 标准卡",
  "bank_code": "YUNAI",
  "card_type": "debit",
  "is_default": true,
  "email": "user@example.com"
}
```

**响应数据**:
```json
{
  "code": 200,
  "message": "卡片绑定成功",
  "data": {
    "id": "uuid",
    "card_number": "8825********3296",
    "bank_name": "YUNAI 标准卡",
    "cardholder_name": "张三",
    "balance": 0.00,
    "currency": "CNY",
    "is_default": true,
    "is_active": true,
    "is_frozen": false,
    "created_at": "2025-08-26T13:30:00Z"
  }
}
```

### 1.2 获取用户卡片列表
**接口**: `GET /payment/cards`

**响应数据**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "cards": [
      {
        "id": "uuid",
        "card_number": "8825********3296",
        "bank_name": "YUNAI 标准卡",
        "balance": 500.00,
        "currency": "CNY",
        "is_default": true,
        "is_frozen": false,
        "created_at": "2025-08-26T13:30:00Z"
      }
    ],
    "total": 1
  }
}
```

### 1.3 卡片充值
**接口**: `POST /payment/cards/{card_id}/recharge`

**请求参数**:
```json
{
  "amount": 100.00,
  "card_code": "optional_card_code"
}
```

**响应数据**:
```json
{
  "code": 200,
  "message": "充值成功",
  "data": {
    "card_id": "uuid",
    "amount": 100.00,
    "balance_before": 500.00,
    "balance_after": 600.00,
    "transaction_id": "uuid"
  }
}
```

### 1.4 冻结卡片
**接口**: `POST /payment/cards/{card_id}/freeze`

**请求参数**:
```json
{
  "reason": "用户申请冻结",
  "password": "123456"
}
```

### 1.5 解冻卡片
**接口**: `POST /payment/cards/{card_id}/unfreeze`

**请求参数**:
```json
{
  "password": "123456"
}
```

### 1.6 删除卡片
**接口**: `DELETE /payment/cards/{card_id}`

**请求参数**:
```json
{
  "password": "123456",
  "reason": "不再使用"
}
```

## 2. 充值订单接口

### 2.1 创建充值订单
**接口**: `POST /payment/recharge/orders`

**请求参数**:
```json
{
  "amount": 100.00,
  "payment_method": "card",
  "payment_card_id": "uuid",
  "card_code": "optional"
}
```

**响应数据**:
```json
{
  "code": 200,
  "message": "订单创建成功",
  "data": {
    "order_id": "uuid",
    "order_no": "YN20250826130001",
    "amount": 100.00,
    "coins_amount": 1000,
    "payment_url": "https://pay.yunai.com/...",
    "expired_at": "2025-08-26T14:00:00Z"
  }
}
```

### 2.2 获取充值订单详情
**接口**: `GET /payment/recharge/orders/{order_id}`

### 2.3 处理支付
**接口**: `POST /payment/recharge/orders/{order_id}/pay`

**请求参数**:
```json
{
  "payment_password": "123456"
}
```

## 3. 代充功能接口

### 3.1 代充
**接口**: `POST /payment/proxy-recharge`

**请求参数**:
```json
{
  "target_user_id": "uuid",
  "amount": 50.00,
  "payment_card_id": "uuid",
  "message": "代充留言",
  "payment_password": "123456"
}
```

**响应数据**:
```json
{
  "code": 200,
  "message": "代充成功",
  "data": {
    "order_id": "uuid",
    "order_no": "PR20250826130001",
    "target_username": "目标用户",
    "payer_username": "付款用户",
    "original_amount": 50.00,
    "actual_amount": 50.00,
    "discount_amount": 0.00,
    "coins_amount": 500,
    "bonus_coins": 100,
    "total_coins": 600,
    "card_type": "返利卡",
    "message": "代充留言",
    "created_at": "2025-08-26T13:30:00Z"
  }
}
```

### 3.2 获取代充订单列表
**接口**: `GET /payment/proxy-recharge/orders`

**查询参数**:
- `page`: 页码，默认 1
- `limit`: 每页数量，默认 20

## 4. 钱包管理接口

### 4.1 获取钱包信息
**接口**: `GET /wallet`

**响应数据**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "balance": 1500.00,
    "currency": "coins",
    "exchange_rate": 10,
    "daily_limit": 10000.00,
    "daily_spent": 500.00,
    "monthly_limit": 100000.00,
    "monthly_spent": 2000.00
  }
}
```

### 4.2 获取钱包交易记录
**接口**: `GET /wallet/transactions`

**查询参数**:
- `page`: 页码
- `limit`: 每页数量
- `type`: 交易类型过滤

## 5. 系统配置接口

### 5.1 获取充值套餐
**接口**: `GET /payment/packages`

**响应数据**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "packages": [
      {
        "id": "uuid",
        "name": "体验包",
        "amount": 6.00,
        "coins_amount": 60,
        "bonus_coins": 0,
        "discount_rate": 1.0000,
        "is_popular": false
      }
    ]
  }
}
```

### 5.2 获取汇率配置
**接口**: `GET /payment/exchange-rate`

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 422 | 业务逻辑错误 |
| 500 | 服务器内部错误 |

### 业务错误码
| 错误码 | 说明 |
|--------|------|
| 10001 | 卡号格式不正确 |
| 10002 | 卡片余额不足 |
| 10003 | 卡片已冻结 |
| 10004 | 支付密码错误 |
| 10005 | 订单已过期 |
| 10006 | 用户不存在 |

## 接口限流

### 限流规则
- 普通接口: 100 次/分钟
- 支付接口: 10 次/分钟
- 代充接口: 5 次/分钟

### 限流响应
```json
{
  "code": 429,
  "message": "请求过于频繁，请稍后再试",
  "retry_after": 60
}
```

---

**文档维护**: 本文档由小云维护，接口变更需要及时更新此文档。
