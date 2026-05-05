import type { FC } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, App as AntdApp } from 'antd';
import { theme } from './styles/theme';
import Dashboard from './pages/Dashboard';
import ETFDashboard from './pages/ETFDashboard';
import PortfolioAnalysis from './pages/PortfolioAnalysis';
import ETFComparison from './pages/ETFComparison';
import ETFDetail from './pages/ETFDetail';
import PortfolioConfig from './pages/PortfolioConfig';
import OperationLogs from './pages/OperationLogs';
import ETFConfig from './pages/ETFConfig';
import ExchangeRate from './pages/ExchangeRate';
import ASharePortfolio from './pages/ASharePortfolio';
import TechnicalAnalysis from './pages/TechnicalAnalysis';
import RiskAnalysis from './pages/RiskAnalysis';
import PortfolioOptimization from './pages/PortfolioOptimization';
import FactorAnalysis from './pages/FactorAnalysis';
import FactorTiming from './pages/FactorTiming';
import AlphaViews from './pages/AlphaViews';
import BlackLittermanConfig from './pages/BlackLittermanConfig';
import RiskBudget from './pages/RiskBudget';
import QuantLibAnalysis from './pages/QuantLibAnalysis';
import AIAgents from './pages/AIAgents';
import './App.css';

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

            <Route path="/etf-dashboard" element={<ETFDashboard />} />
            <Route path="/etf-market" element={<ETFDashboard />} />
            <Route path="/etf-comparison" element={<ETFComparison />} />
            <Route path="/etf-detail/:symbol" element={<ETFDetail />} />
            <Route path="/etf-config" element={<ETFConfig />} />

            <Route path="/portfolio-analysis" element={<PortfolioAnalysis />} />
            <Route path="/portfolio-config" element={<PortfolioConfig />} />
            <Route path="/a-share-portfolio" element={<ASharePortfolio />} />

            <Route path="/technical-analysis" element={<TechnicalAnalysis />} />
            <Route path="/risk-analysis" element={<RiskAnalysis />} />
            <Route path="/portfolio-optimization" element={<PortfolioOptimization />} />
            <Route path="/factor-analysis" element={<FactorAnalysis />} />
            <Route path="/factor-timing" element={<FactorTiming />} />
            <Route path="/alpha-views" element={<AlphaViews />} />
            <Route path="/black-litterman" element={<BlackLittermanConfig />} />
            <Route path="/risk-budget" element={<RiskBudget />} />
            <Route path="/quantlib" element={<QuantLibAnalysis />} />
            <Route path="/ai-agents" element={<AIAgents />} />

            <Route path="/operation-logs" element={<OperationLogs />} />
            <Route path="/exchange-rate" element={<ExchangeRate />} />

            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </Router>
      </AntdApp>
    </ConfigProvider>
  );
};

export default App;
