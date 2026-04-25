-- 阶段三：风险预算层迁移
-- 执行时间：2026-Q4 第5-6周
-- 版本：v2.7

-- 1. 创建风险预算配置表
CREATE TABLE IF NOT EXISTS risk_budget_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_id INTEGER UNIQUE,
    stock_cvar_budget DECIMAL(5,2) DEFAULT 40.00,
    bond_cvar_budget DECIMAL(5,2) DEFAULT 10.00,
    commodity_cvar_budget DECIMAL(5,2) DEFAULT 20.00,
    cash_cvar_budget DECIMAL(5,2) DEFAULT 5.00,
    use_var_constraint BOOLEAN DEFAULT FALSE,
    stock_var_budget DECIMAL(5,2),
    bond_var_budget DECIMAL(5,2),
    min_skewness DECIMAL(10,6),
    max_drawdown DECIMAL(5,2),
    cvar_confidence DECIMAL(5,4) DEFAULT 0.95,
    var_confidence DECIMAL(5,4) DEFAULT 0.95,
    is_active BOOLEAN DEFAULT TRUE,
    effective_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id)
);

-- 2. 创建蒙特卡洛模拟表
CREATE TABLE IF NOT EXISTS monte_carlo_simulations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_id INTEGER NOT NULL,
    simulation_date TIMESTAMP NOT NULL,
    num_paths INTEGER DEFAULT 10000,
    time_steps INTEGER DEFAULT 252,
    time_horizon INTEGER DEFAULT 252,
    confidence_level DECIMAL(5,4) DEFAULT 0.95,
    var_95 DECIMAL(10,6),
    var_99 DECIMAL(10,6),
    cvar_95 DECIMAL(10,6),
    cvar_99 DECIMAL(10,6),
    mean_return DECIMAL(10,6),
    std_dev DECIMAL(10,6),
    skewness DECIMAL(10,6),
    kurtosis DECIMAL(10,6),
    simulation_result TEXT,
    cache_expiry TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id)
);

CREATE INDEX IF NOT EXISTS idx_simulation_portfolio ON monte_carlo_simulations(portfolio_id, simulation_date);

-- 3. 创建风险贡献表
CREATE TABLE IF NOT EXISTS risk_contributions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    simulation_id INTEGER NOT NULL,
    asset_id INTEGER NOT NULL,
    asset_symbol VARCHAR(20),
    weight DECIMAL(5,2),
    cvar_contribution DECIMAL(10,6),
    marginal_cvar DECIMAL(10,6),
    cvar_percentage DECIMAL(5,2),
    budget_limit DECIMAL(5,2),
    budget_usage DECIMAL(5,2),
    budget_deviation DECIMAL(5,2),
    calculation_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (simulation_id) REFERENCES monte_carlo_simulations(id),
    FOREIGN KEY (asset_id) REFERENCES assets(id)
);

-- 4. 创建风险预算执行记录表
CREATE TABLE IF NOT EXISTS risk_budget_executions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_id INTEGER NOT NULL,
    execution_date TIMESTAMP NOT NULL,
    total_cvar DECIMAL(10,6),
    stock_cvar_usage DECIMAL(5,2),
    bond_cvar_usage DECIMAL(5,2),
    is_budget_satisfied BOOLEAN,
    violation_details TEXT,
    optimization_method VARCHAR(50),
    iterations INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id)
);

CREATE INDEX IF NOT EXISTS idx_execution_date ON risk_budget_executions(portfolio_id, execution_date);

-- 5. 验证迁移
SELECT 'Risk budget layer migration completed successfully' AS status;
SELECT COUNT(*) AS risk_budget_count FROM risk_budget_configs;
SELECT COUNT(*) AS simulation_count FROM monte_carlo_simulations;
