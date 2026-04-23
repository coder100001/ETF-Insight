import type { FC } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, App as AntdApp } from 'antd';
import { theme } from './styles/theme';
import Login from './pages/Login';
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
import TechnicalAnalysis from './pages/TechnicalAnalysis';
import RiskAnalysis from './pages/RiskAnalysis';
import PortfolioOptimization from './pages/PortfolioOptimization';
import FactorAnalysis from './pages/FactorAnalysis';
import ProtectedRoute from './components/ProtectedRoute';
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
            <Route path="/login" element={<Login />} />

            <Route path="/" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
            <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />

            <Route path="/etf-dashboard" element={<ProtectedRoute><ETFDashboard /></ProtectedRoute>} />
            <Route path="/etf-market" element={<ProtectedRoute><ETFDashboard /></ProtectedRoute>} />
            <Route path="/etf-comparison" element={<ProtectedRoute><ETFComparison /></ProtectedRoute>} />
            <Route path="/etf-detail/:symbol" element={<ProtectedRoute><ETFDetail /></ProtectedRoute>} />
            <Route path="/etf-config" element={<ProtectedRoute><ETFConfig /></ProtectedRoute>} />

            <Route path="/portfolio-analysis" element={<ProtectedRoute><PortfolioAnalysis /></ProtectedRoute>} />
            <Route path="/portfolio-config" element={<ProtectedRoute><PortfolioConfig /></ProtectedRoute>} />
            <Route path="/a-share-portfolio" element={<ProtectedRoute><ASharePortfolio /></ProtectedRoute>} />

            <Route path="/workflows" element={<ProtectedRoute><WorkflowList /></ProtectedRoute>} />
            <Route path="/instances" element={<ProtectedRoute><InstanceList /></ProtectedRoute>} />

            <Route path="/technical-analysis" element={<ProtectedRoute><TechnicalAnalysis /></ProtectedRoute>} />
            <Route path="/risk-analysis" element={<ProtectedRoute><RiskAnalysis /></ProtectedRoute>} />
            <Route path="/portfolio-optimization" element={<ProtectedRoute><PortfolioOptimization /></ProtectedRoute>} />
            <Route path="/factor-analysis" element={<ProtectedRoute><FactorAnalysis /></ProtectedRoute>} />

            <Route path="/operation-logs" element={<ProtectedRoute><OperationLogs /></ProtectedRoute>} />
            <Route path="/exchange-rate" element={<ProtectedRoute><ExchangeRate /></ProtectedRoute>} />

            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </Router>
      </AntdApp>
    </ConfigProvider>
  );
}

export default App;
