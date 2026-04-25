-- 阶段四：插件架构层迁移
-- 执行时间：2027-Q1 第1-2周
-- 版本：v2.8

-- 1. 创建插件注册表
CREATE TABLE IF NOT EXISTS plugin_registries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_name VARCHAR(100) UNIQUE NOT NULL,
    plugin_type VARCHAR(50) NOT NULL,
    version VARCHAR(20),
    input_schema TEXT,
    output_schema TEXT,
    dependencies TEXT,
    description VARCHAR(500),
    author VARCHAR(100),
    documentation VARCHAR(200),
    status VARCHAR(20) DEFAULT 'active',
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_plugin_type ON plugin_registries(plugin_type);

-- 2. 创建插件配置表
CREATE TABLE IF NOT EXISTS plugin_configurations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id INTEGER NOT NULL,
    portfolio_id INTEGER,
    config_name VARCHAR(100),
    parameters TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plugin_id) REFERENCES plugin_registries(id),
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id)
);

-- 3. 创建插件执行日志表
CREATE TABLE IF NOT EXISTS plugin_execution_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plugin_id INTEGER NOT NULL,
    config_id INTEGER,
    execution_id VARCHAR(50) UNIQUE NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration INTEGER,
    input_data TEXT,
    output_data TEXT,
    status VARCHAR(20),
    error_message TEXT,
    memory_usage INTEGER,
    cpu_usage REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (plugin_id) REFERENCES plugin_registries(id),
    FOREIGN KEY (config_id) REFERENCES plugin_configurations(id)
);

CREATE INDEX IF NOT EXISTS idx_execution_time ON plugin_execution_logs(start_time);

-- 4. 创建模型基准对比矩阵表
CREATE TABLE IF NOT EXISTS model_benchmark_matrices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    comparison_name VARCHAR(100) NOT NULL,
    alpha_plugin_id INTEGER,
    optimizer_plugin_id INTEGER,
    risk_plugin_id INTEGER,
    backtest_window INTEGER DEFAULT 3,
    rebalance_frequency VARCHAR(20),
    transaction_cost DECIMAL(5,4),
    total_return DECIMAL(10,6),
    annual_return DECIMAL(10,6),
    volatility DECIMAL(10,6),
    sharpe_ratio DECIMAL(10,6),
    max_drawdown DECIMAL(10,6),
    calmar_ratio DECIMAL(10,6),
    rolling_1y_win_rate DECIMAL(5,2),
    tail_dependency_index DECIMAL(10,6),
    detailed_metrics TEXT,
    comparison_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (alpha_plugin_id) REFERENCES plugin_registries(id),
    FOREIGN KEY (optimizer_plugin_id) REFERENCES plugin_registries(id),
    FOREIGN KEY (risk_plugin_id) REFERENCES plugin_registries(id)
);

-- 5. 创建策略实验表
CREATE TABLE IF NOT EXISTS strategy_experiments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    experiment_name VARCHAR(100) NOT NULL,
    description VARCHAR(500),
    benchmark_matrix_id INTEGER,
    allocation_ratio DECIMAL(5,2) DEFAULT 20.00,
    experiment_result TEXT,
    is_successful BOOLEAN,
    success_criteria TEXT,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    status VARCHAR(20) DEFAULT 'running',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (benchmark_matrix_id) REFERENCES model_benchmark_matrices(id)
);

-- 6. 插入默认插件注册数据
INSERT INTO plugin_registries (plugin_name, plugin_type, version, description, author, status) VALUES
('FamaFrenchAlpha', 'AlphaGenerator', '1.0.0', 'Fama-French factor-based alpha generator', 'ETF-Insight Team', 'active'),
('MPTOptimizer', 'PortfolioOptimizer', '1.0.0', 'Modern Portfolio Theory optimizer', 'ETF-Insight Team', 'active'),
('RiskParityOptimizer', 'PortfolioOptimizer', '1.0.0', 'Risk parity portfolio optimizer', 'ETF-Insight Team', 'active'),
('BlackLittermanOptimizer', 'PortfolioOptimizer', '1.0.0', 'Black-Litterman portfolio optimizer', 'ETF-Insight Team', 'active'),
('CVaRRiskModel', 'RiskModel', '1.0.0', 'Conditional Value at Risk model', 'ETF-Insight Team', 'active');

-- 7. 验证迁移
SELECT 'Plugin architecture layer migration completed successfully' AS status;
SELECT COUNT(*) AS plugin_count FROM plugin_registries;
SELECT COUNT(*) AS experiment_count FROM strategy_experiments;
