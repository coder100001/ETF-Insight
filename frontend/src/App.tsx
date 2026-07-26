import type { FC } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, App as AntdApp } from 'antd';
import { theme } from './styles/theme';
import { ErrorBoundary } from './components/ErrorBoundary';
import Dashboard from './pages/Dashboard';
import ETFDashboard from './pages/ETFDashboard';
import PortfolioAnalysis from './pages/PortfolioAnalysis';
import ETFComparison from './pages/ETFComparison';
import ETFDetail from './pages/ETFDetail';
import PortfolioConfig from './pages/PortfolioConfig';
import ETFConfig from './pages/ETFConfig';
import ExchangeRate from './pages/ExchangeRate';
import ASharePortfolio from './pages/ASharePortfolio';
import TechnicalAnalysis from './pages/TechnicalAnalysis';
import RiskAnalysis from './pages/RiskAnalysis';
import PortfolioOptimization from './pages/PortfolioOptimization';
import DCACalculator from './pages/DCACalculator';
import InvestmentStrategy from './pages/InvestmentStrategy';
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
          <ErrorBoundary>
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
              <Route path="/dca-calculator" element={<DCACalculator />} />

              <Route path="/investment-strategy" element={<InvestmentStrategy />} />

              <Route path="/exchange-rate" element={<ExchangeRate />} />

              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </ErrorBoundary>
        </Router>
      </AntdApp>
    </ConfigProvider>
  );
};

export default App;
