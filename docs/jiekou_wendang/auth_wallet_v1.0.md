# 认证与钱包接口文档

## 文档信息
- **版本**: v1.0
- **作者**: 小云
- **日期**: 2025-08-26
- **摘要**: YUNAI 认证服务和钱包服务的 API 接口文档

## 目录
1. [认证接口](#认证接口)
2. [钱包接口](#钱包接口)
3. [卡密接口](#卡密接口)
4. [错误码定义](#错误码定义)

## 认证接口

### 用户注册
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "string",
  "email": "string",
  "password": "string",
  "invite_code": "string?"
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "uuid",
    "username": "string",
    "email": "string",
    "user_type": "basic",
    "created_at": "2025-08-26T10:00:00Z"
  }
}
```

### 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "string",
  "password": "string",
  "totp_code": "string?"
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "jwt_token",
    "refresh_token": "jwt_token",
    "expires_in": 3600,
    "user": {
      "user_id": "uuid",
      "username": "string",
      "user_type": "basic"
    }
  }
}
```

### 刷新令牌
```http
POST /api/v1/auth/refresh
Authorization: Bearer {refresh_token}
```

## 钱包接口

### 获取钱包余额
```http
GET /api/v1/wallet/balance
Authorization: Bearer {access_token}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "uuid",
    "balance": 1000,
    "currency": "金币",
    "exchange_rate": 10,
    "updated_at": "2025-08-26T10:00:00Z"
  }
}
```

### 充值记录
```http
GET /api/v1/wallet/recharge/history
Authorization: Bearer {access_token}
Query Parameters:
- page: int (default: 1)
- limit: int (default: 20)
- start_date: string (YYYY-MM-DD)
- end_date: string (YYYY-MM-DD)
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 100,
    "page": 1,
    "limit": 20,
    "records": [
      {
        "recharge_id": "uuid",
        "amount": 100,
        "currency": "金币",
        "payment_method": "alipay",
        "status": "completed",
        "created_at": "2025-08-26T10:00:00Z"
      }
    ]
  }
}
```

### 消费记录
```http
GET /api/v1/wallet/consumption/history
Authorization: Bearer {access_token}
Query Parameters:
- page: int (default: 1)
- limit: int (default: 20)
- service_type: string (chat|media|call)
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 50,
    "page": 1,
    "limit": 20,
    "records": [
      {
        "consumption_id": "uuid",
        "service_type": "chat",
        "model_name": "gpt-4",
        "amount": 10,
        "description": "群聊对话消费",
        "created_at": "2025-08-26T10:00:00Z"
      }
    ]
  }
}
```

## 卡密接口

### 兑换卡密
```http
POST /api/v1/wallet/redeem
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "card_code": "YUNAI-XXXX-XXXX-XXXX"
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "card_type": "discount",
    "original_amount": 100,
    "discount_rate": 0.9,
    "final_amount": 90,
    "bonus_amount": 20,
    "total_received": 120,
    "privileges": ["txt2video", "hd_upscaler"],
    "redeemed_at": "2025-08-26T10:00:00Z"
  }
}
```

### 卡密验证
```http
POST /api/v1/wallet/card/validate
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "card_code": "YUNAI-XXXX-XXXX-XXXX"
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "valid": true,
    "card_type": "discount",
    "value": 100,
    "discount_rate": 0.9,
    "bonus_rate": 0.2,
    "privileges": ["txt2video"],
    "expires_at": "2025-12-31T23:59:59Z"
  }
}
```

## 错误码定义

| 错误码 | 错误信息 | 描述 |
|--------|----------|------|
| 0 | success | 成功 |
| 1001 | invalid_request | 请求参数无效 |
| 1002 | unauthorized | 未授权访问 |
| 1003 | forbidden | 权限不足 |
| 1004 | not_found | 资源不存在 |
| 2001 | user_not_found | 用户不存在 |
| 2002 | invalid_credentials | 用户名或密码错误 |
| 2003 | account_locked | 账户已锁定 |
| 2004 | totp_required | 需要双因子认证 |
| 2005 | invalid_totp | 双因子认证码错误 |
| 3001 | insufficient_balance | 余额不足 |
| 3002 | invalid_card_code | 卡密无效 |
| 3003 | card_expired | 卡密已过期 |
| 3004 | card_already_used | 卡密已使用 |
| 3005 | recharge_failed | 充值失败 |
| 5000 | internal_error | 服务器内部错误 |

## 附录

### 认证流程说明
1. 用户注册后获得 basic 权限
2. 登录成功获得 access_token 和 refresh_token
3. access_token 有效期 1 小时，refresh_token 有效期 30 天
4. 支持 TOTP 双因子认证（可选）

### 钱包计费规则
- 默认汇率：1 元 = 10 金币
- 卡密折扣：按金额打折
- 卡密返利：额外赠送金币
- 特权解锁：开启高级功能权限

### 更新记录
| 版本 | 日期 | 作者 | 更新内容 |
|------|------|------|----------|
| v1.0 | 2025-08-26 | 小云 | 初始版本，认证和钱包接口 |
