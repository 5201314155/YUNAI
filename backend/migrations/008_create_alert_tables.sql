-- 创建告警相关表

-- 告警规则表
CREATE TABLE IF NOT EXISTS alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metric VARCHAR(100) NOT NULL,           -- 监控指标名称
    operator VARCHAR(10) NOT NULL,          -- 比较操作符 (>, <, =, >=, <=, !=)
    threshold DECIMAL(10,2) NOT NULL,       -- 阈值
    duration INTEGER DEFAULT 300,           -- 持续时间(秒)
    severity VARCHAR(20) NOT NULL,          -- 严重程度 (low, medium, high, critical)
    enabled BOOLEAN DEFAULT true,           -- 是否启用
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 告警记录表
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    rule_name VARCHAR(255) NOT NULL,
    metric VARCHAR(100) NOT NULL,
    value DECIMAL(10,2) NOT NULL,           -- 触发时的值
    threshold DECIMAL(10,2) NOT NULL,       -- 阈值
    severity VARCHAR(20) NOT NULL,          -- 严重程度
    status VARCHAR(20) NOT NULL,            -- 状态 (firing, resolved)
    message TEXT,                           -- 告警消息
    fired_at TIMESTAMP WITH TIME ZONE NOT NULL,
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 系统指标表
CREATE TABLE IF NOT EXISTS system_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(15,4) NOT NULL,
    metric_type VARCHAR(50),                -- counter, gauge, histogram
    labels JSONB,                          -- 标签数据
    collected_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 用户事件表 (用于长期存储)
CREATE TABLE IF NOT EXISTS user_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_name VARCHAR(100) NOT NULL,
    properties JSONB,                      -- 事件属性
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 业务指标表
CREATE TABLE IF NOT EXISTS business_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(15,4) NOT NULL,
    metric_type VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(date, metric_name)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_alert_rules_enabled ON alert_rules(enabled);
CREATE INDEX IF NOT EXISTS idx_alert_rules_metric ON alert_rules(metric);

CREATE INDEX IF NOT EXISTS idx_alerts_rule_id ON alerts(rule_id);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_fired_at ON alerts(fired_at);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);

CREATE INDEX IF NOT EXISTS idx_system_metrics_name ON system_metrics(metric_name);
CREATE INDEX IF NOT EXISTS idx_system_metrics_collected_at ON system_metrics(collected_at);

CREATE INDEX IF NOT EXISTS idx_user_events_user_id ON user_events(user_id);
CREATE INDEX IF NOT EXISTS idx_user_events_event_name ON user_events(event_name);
CREATE INDEX IF NOT EXISTS idx_user_events_created_at ON user_events(created_at);

CREATE INDEX IF NOT EXISTS idx_business_metrics_date ON business_metrics(date);
CREATE INDEX IF NOT EXISTS idx_business_metrics_name ON business_metrics(metric_name);

-- 插入默认告警规则
INSERT INTO alert_rules (name, description, metric, operator, threshold, duration, severity, enabled) VALUES
('API响应时间过高', '当API平均响应时间超过1秒时触发告警', 'api_response_time', '>', 1000.0, 300, 'high', true),
('API错误率过高', '当API错误率超过5%时触发告警', 'api_error_rate', '>', 5.0, 180, 'medium', true),
('日活用户数过低', '当日活用户数低于100时触发告警', 'daily_active_users', '<', 100.0, 600, 'low', true),
('系统CPU使用率过高', '当系统CPU使用率超过80%时触发告警', 'cpu_usage', '>', 80.0, 300, 'critical', true),
('内存使用率过高', '当系统内存使用率超过70%时触发告警', 'memory_usage', '>', 70.0, 300, 'high', true),
('磁盘空间不足', '当磁盘使用率超过90%时触发告警', 'disk_usage', '>', 90.0, 600, 'critical', true),
('Redis连接异常', '当Redis连接失败时触发告警', 'redis_health', '<', 1.0, 60, 'critical', true),
('数据库连接异常', '当数据库连接失败时触发告警', 'database_health', '<', 1.0, 60, 'critical', true)
ON CONFLICT DO NOTHING;

-- 创建触发器更新updated_at字段
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_alert_rules_updated_at BEFORE UPDATE ON alert_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_alerts_updated_at BEFORE UPDATE ON alerts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 添加注释
COMMENT ON TABLE alert_rules IS '告警规则配置表';
COMMENT ON TABLE alerts IS '告警记录表';
COMMENT ON TABLE system_metrics IS '系统指标表';
COMMENT ON TABLE user_events IS '用户事件表';
COMMENT ON TABLE business_metrics IS '业务指标表';

COMMENT ON COLUMN alert_rules.metric IS '监控指标名称，如api_response_time, cpu_usage等';
COMMENT ON COLUMN alert_rules.operator IS '比较操作符: >, <, =, >=, <=, !=';
COMMENT ON COLUMN alert_rules.threshold IS '告警阈值';
COMMENT ON COLUMN alert_rules.duration IS '持续时间，单位秒';
COMMENT ON COLUMN alert_rules.severity IS '严重程度: low, medium, high, critical';

COMMENT ON COLUMN alerts.status IS '告警状态: firing(触发中), resolved(已解决)';
COMMENT ON COLUMN alerts.fired_at IS '告警触发时间';
COMMENT ON COLUMN alerts.resolved_at IS '告警解决时间';
