-- 阶段一：因子数据层迁移
-- 执行时间：2026-Q4 第1-2周
-- 版本：v2.7

-- 1. 创建因子数据表
CREATE TABLE IF NOT EXISTS factor_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    factor_name VARCHAR(20) NOT NULL,
    date DATE NOT NULL,
    value DECIMAL(10,6) NOT NULL,
    data_source VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_factor_date ON factor_data(factor_name, date);

-- 2. 创建因子择时信号表
CREATE TABLE IF NOT EXISTS factor_timing_signals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    factor_name VARCHAR(20) NOT NULL,
    signal_date DATE NOT NULL,
    ma_slope_60 DECIMAL(10,6),
    z_score DECIMAL(10,6),
    percentile DECIMAL(5,2),
    signal_strength VARCHAR(20),
    signal_score INTEGER,
    expected_return DECIMAL(10,6),
    confidence DECIMAL(5,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_signal_date ON factor_timing_signals(factor_name, signal_date);

-- 3. 插入默认因子数据（示例）
-- 实际数据需要从Fama-French库导入
INSERT INTO factor_data (factor_name, date, value, data_source) VALUES
('Mkt-RF', '2024-01-01', 0.0052, 'Fama-French Library'),
('SMB', '2024-01-01', 0.0018, 'Fama-French Library'),
('HML', '2024-01-01', 0.0023, 'Fama-French Library'),
('RMW', '2024-01-01', 0.0015, 'Fama-French Library'),
('CMA', '2024-01-01', 0.0012, 'Fama-French Library');

-- 4. 验证迁移
SELECT 'Factor data migration completed successfully' AS status;
SELECT COUNT(*) AS factor_count FROM factor_data;
SELECT COUNT(*) AS signal_count FROM factor_timing_signals;
