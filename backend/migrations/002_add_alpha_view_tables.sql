-- 阶段二：Alpha观点层迁移
-- 执行时间：2026-Q4 第3-4周
-- 版本：v2.7

-- 1. 创建Alpha观点表
CREATE TABLE IF NOT EXISTS alpha_views (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_id INTEGER,
    asset_symbol VARCHAR(20) NOT NULL,
    view_return DECIMAL(10,6) NOT NULL,
    confidence DECIMAL(5,2) NOT NULL,
    view_type VARCHAR(20),
    view_method VARCHAR(50),
    generated_at TIMESTAMP NOT NULL,
    valid_until TIMESTAMP,
    status VARCHAR(20) DEFAULT 'active',
    source_factor VARCHAR(20),
    factor_loading DECIMAL(10,6),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id)
);

CREATE INDEX IF NOT EXISTS idx_view_asset ON alpha_views(asset_symbol, generated_at);
CREATE INDEX IF NOT EXISTS idx_view_portfolio_status ON alpha_views(portfolio_id, status);

-- 2. 创建Alpha观点表现表
CREATE TABLE IF NOT EXISTS alpha_view_performances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    view_id INTEGER NOT NULL,
    actual_return DECIMAL(10,6),
    prediction_error DECIMAL(10,6),
    is_validated BOOLEAN DEFAULT FALSE,
    validation_date TIMESTAMP,
    is_correct BOOLEAN,
    rolling_win_rate DECIMAL(5,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (view_id) REFERENCES alpha_views(id)
);

-- 3. 创建Black-Litterman配置表
CREATE TABLE IF NOT EXISTS black_litterman_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_id INTEGER UNIQUE,
    risk_aversion DECIMAL(10,6),
    prior_type VARCHAR(20),
    prior_weights TEXT,
    implied_returns TEXT,
    omega_method VARCHAR(20),
    omega_matrix TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    last_calculated TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id)
);

-- 4. 创建BL后验收益表
CREATE TABLE IF NOT EXISTS bl_posterior_returns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    config_id INTEGER NOT NULL,
    calculation_date TIMESTAMP NOT NULL,
    posterior_returns TEXT,
    posterior_weights TEXT,
    posterior_cov TEXT,
    num_views INTEGER,
    view_impact DECIMAL(10,6),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (config_id) REFERENCES black_litterman_configs(id)
);

CREATE INDEX IF NOT EXISTS idx_posterior_date ON bl_posterior_returns(config_id, calculation_date);

-- 5. 验证迁移
SELECT 'Alpha view layer migration completed successfully' AS status;
SELECT COUNT(*) AS alpha_view_count FROM alpha_views;
SELECT COUNT(*) AS bl_config_count FROM black_litterman_configs;
