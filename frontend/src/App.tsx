import { lazy, Suspense, type FC } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, App as AntdApp, Spin } from 'antd';
import { theme } from './styles/theme';
import { ErrorBoundary } from './components/ErrorBoundary';
import './App.css';

// 路由级代码分割：页面按需加载，首屏只加载当前路由的页面 chunk
const Dashboard = lazy(() => import('./pages/Dashboard'));
const ETFDashboard = lazy(() => import('./pages/ETFDashboard'));
const PortfolioAnalysis = lazy(() => import('./pages/PortfolioAnalysis'));
const ETFComparison = lazy(() => import('./pages/ETFComparison'));
const ETFDetail = lazy(() => import('./pages/ETFDetail'));
const PortfolioConfig = lazy(() => import('./pages/PortfolioConfig'));
const ETFConfig = lazy(() => import('./pages/ETFConfig'));
const ExchangeRate = lazy(() => import('./pages/ExchangeRate'));
const ASharePortfolio = lazy(() => import('./pages/ASharePortfolio'));
const TechnicalAnalysis = lazy(() => import('./pages/TechnicalAnalysis'));
const RiskAnalysis = lazy(() => import('./pages/RiskAnalysis'));
const PortfolioOptimization = lazy(() => import('./pages/PortfolioOptimization'));
const DCACalculator = lazy(() => import('./pages/DCACalculator'));
const InvestmentStrategy = lazy(() => import('./pages/InvestmentStrategy'));

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

// 页面切换时的加载占位
const PageLoading = () => (
  <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
    <Spin size="large" />
  </div>
);

const App: FC = () => {
  return (
    <ConfigProvider theme={antdTheme}>
      <AntdApp>
        <Router>
          <ErrorBoundary>
            <Suspense fallback={<PageLoading />}>
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
            </Suspense>
          </ErrorBoundary>
        </Router>
      </AntdApp>
    </ConfigProvider>
  );
};

export default App;
