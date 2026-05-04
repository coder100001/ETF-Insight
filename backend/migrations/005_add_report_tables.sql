-- 阶段五：报告系统迁移
-- 执行时间：2027-Q1 第3-4周
-- 版本：v2.9

-- 1. 创建报告模板表
CREATE TABLE IF NOT EXISTS report_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    config TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_report_category ON report_templates(category);

-- 2. 创建报告章节表
CREATE TABLE IF NOT EXISTS report_sections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    type VARCHAR(50),
    content TEXT,
    order INTEGER DEFAULT 0,
    required BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (template_id) REFERENCES report_templates(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_section_template ON report_sections(template_id);

-- 3. 创建报告参数表
CREATE TABLE IF NOT EXISTS report_parameters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    label VARCHAR(255),
    type VARCHAR(50),
    required BOOLEAN DEFAULT TRUE,
    default TEXT,
    options TEXT,
    description TEXT,
    order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (template_id) REFERENCES report_templates(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_param_template ON report_parameters(template_id);

-- 4. 创建生成的报告表
CREATE TABLE IF NOT EXISTS generated_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INTEGER,
    title VARCHAR(255),
    format VARCHAR(20),
    file_path VARCHAR(500),
    file_size INTEGER,
    status VARCHAR(20) DEFAULT 'pending',
    error_message TEXT,
    data TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (template_id) REFERENCES report_templates(id)
);

CREATE INDEX IF NOT EXISTS idx_report_template ON generated_reports(template_id);
CREATE INDEX IF NOT EXISTS idx_report_status ON generated_reports(status);
CREATE INDEX IF NOT EXISTS idx_report_created ON generated_reports(created_at);

-- 5. 插入默认报告模板数据
INSERT INTO report_templates (name, description, category, is_default) VALUES
('投资组合分析报告', '全面的投资组合分析报告，包含收益、风险、配置等指标', 'portfolio', TRUE),
('风险分析报告', '深度风险分析报告，包含VaR、回撤、相关性分析等', 'risk', TRUE),
('ETF对比报告', '多只ETF对比分析报告，包含持仓重叠、费率、表现对比', 'etf_comparison', TRUE),
('市场分析周报', '每周市场分析报告，包含指数表现、ETF资金流向等', 'market', TRUE);

-- 6. 插入默认报告章节数据
INSERT INTO report_sections (template_id, title, type, order, required) VALUES
(1, '执行摘要', 'executive_summary', 1, TRUE),
(1, '组合概览', 'metric', 2, TRUE),
(1, '收益分析', 'chart', 3, TRUE),
(1, '风险指标', 'table', 4, TRUE),
(1, '资产配置', 'chart', 5, TRUE),
(1, '行业分布', 'chart', 6, TRUE),
(1, '持仓分析', 'table', 7, TRUE),
(1, '情景分析', 'chart', 8, FALSE),
(1, '建议与结论', 'text', 9, TRUE),
(2, '风险概览', 'metric', 1, TRUE),
(2, '波动率分析', 'chart', 2, TRUE),
(2, 'VaR/CVaR分析', 'table', 3, TRUE),
(2, '回撤分析', 'chart', 4, TRUE),
(2, '相关性分析', 'chart', 5, TRUE),
(2, '压力测试', 'table', 6, FALSE),
(2, '风险归因', 'table', 7, TRUE),
(2, '风险建议', 'text', 8, TRUE),
(3, '对比概览', 'metric', 1, TRUE),
(3, '基本信息对比', 'table', 2, TRUE),
(3, '业绩表现对比', 'chart', 3, TRUE),
(3, '风险指标对比', 'table', 4, TRUE),
(3, '持仓对比', 'table', 5, TRUE),
(3, '费用对比', 'chart', 6, TRUE),
(3, '重叠度分析', 'chart', 7, TRUE),
(3, '选择建议', 'text', 8, TRUE),
(4, '本周市场回顾', 'executive_summary', 1, TRUE),
(4, '主要指数表现', 'chart', 2, TRUE),
(4, 'ETF资金流向', 'table', 3, TRUE),
(4, '行业轮动', 'chart', 4, TRUE),
(4, '重要新闻与事件', 'text', 5, FALSE),
(4, '下周展望', 'text', 6, TRUE);

-- 7. 插入默认报告参数数据
INSERT INTO report_parameters (template_id, name, label, type, required, order) VALUES
(1, 'portfolio_id', '投资组合', 'select', TRUE, 1),
(1, 'date_range', '时间范围', 'date_range', TRUE, 2),
(1, 'benchmark', '基准指数', 'select', FALSE, 3),
(2, 'portfolio_id', '投资组合', 'select', TRUE, 1),
(2, 'confidence_level', '置信水平', 'select', TRUE, 2),
(2, 'lookback_period', '回溯期', 'select', TRUE, 3),
(3, 'etf_symbols', 'ETF列表', 'multi_select', TRUE, 1),
(3, 'date_range', '时间范围', 'date_range', TRUE, 2),
(4, 'week_end', '周末日期', 'date', TRUE, 1),
(4, 'market_regions', '市场区域', 'multi_select', TRUE, 2);

-- 8. 验证迁移
SELECT 'Report system migration completed successfully' AS status;
SELECT COUNT(*) AS template_count FROM report_templates;
SELECT COUNT(*) AS section_count FROM report_sections;
SELECT COUNT(*) AS parameter_count FROM report_parameters;
