import type { FC } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, App as AntdApp } from 'antd';
import { theme } from './styles/theme';
import Dashboard from './pages/Dashboard';
import ETFDashboard from './pages/ETFDashboard';
import PortfolioAnalysis from './pages/PortfolioAnalysis';
import ETFComparison from './pages/ETFComparison';
import ETFDetail from './pages/ETFDetail';
import WorkflowList from './pages/WorkflowList';
import InstanceList from './pages/InstanceList';
import PortfolioConfig from './pages/PortfolioConfig';
import OperationLogs from './pages/OperationLogs';
import ETFConfig from './pages/ETFConfig';
import ExchangeRate from './pages/ExchangeRate';
import ASharePortfolio from './pages/ASharePortfolio';
import './App.css';

// 配置Ant Design主题 - 匹配Django模板风格
const antdTheme = {
  token: {
    colorPrimary: theme.colors.primary,
    colorSuccess: theme.colors.success,
    colorWarning: theme.colors.warning,
    colorError: theme.colors.danger,
    colorInfo: theme.colors.info,
    borderRadius: 4,
    fontFamily: theme.fonts.family,
  },
  components: {
    Button: {
      borderRadius: 4,
    },
    Card: {
      borderRadius: 4,
    },
    Table: {
      borderRadius: 4,
    },
  },
};

const App: FC = () => {
  return (
    <ConfigProvider theme={antdTheme}>
      <AntdApp>
        <Router>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/dashboard" element={<Dashboard />} />
            
            {/* ETF相关路由 */}
            <Route path="/etf-dashboard" element={<ETFDashboard />} />
            <Route path="/etf-market" element={<ETFDashboard />} />
            <Route path="/etf-comparison" element={<ETFComparison />} />
            <Route path="/etf-detail/:symbol" element={<ETFDetail />} />
            <Route path="/etf-config" element={<ETFConfig />} />
            
            {/* 投资组合路由 */}
            <Route path="/portfolio-analysis" element={<PortfolioAnalysis />} />
            <Route path="/portfolio-config" element={<PortfolioConfig />} />
            <Route path="/a-share-portfolio" element={<ASharePortfolio />} />
            
            {/* 工作流路由 */}
            <Route path="/workflows" element={<WorkflowList />} />
            <Route path="/instances" element={<InstanceList />} />
            
            {/* 其他路由 */}
            <Route path="/operation-logs" element={<OperationLogs />} />
            <Route path="/exchange-rate" element={<ExchangeRate />} />
            
            {/* 默认重定向 */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </Router>
      </AntdApp>
    </ConfigProvider>
  );
}

export default App;
