import { useState, type ReactNode } from 'react';
import { Link, useLocation } from 'react-router-dom';
import styled from 'styled-components';
import { theme } from '../styles/theme';
import {
  DashboardOutlined,
  ProjectOutlined,
  HistoryOutlined,
  BarChartOutlined,
  WalletOutlined,
  PieChartOutlined,
  FileTextOutlined,
  SettingOutlined,
  SwapOutlined,
  FundOutlined,
  LineChartOutlined,
  ClockCircleOutlined,
  EyeOutlined,
  CalculatorOutlined,
  SafetyOutlined,
  NumberOutlined,
  RobotOutlined,
  MenuOutlined,
  CloseOutlined,
} from '@ant-design/icons';
import { FaBalanceScale } from 'react-icons/fa';

const LayoutContainer = styled.div`
  display: flex;
  min-height: 100vh;
  background: ${theme.colors.background};
  font-family: ${theme.fonts.family};
`;

const Sidebar = styled.div`
  position: fixed;
  top: 0;
  left: 0;
  height: 100vh;
  width: ${theme.layout.sidebarWidth};
  background: ${theme.colors.sidebarBg};
  padding-top: 20px;
  z-index: 1000;
  transition: transform 0.3s ease;

  @media (max-width: ${theme.breakpoints.md}) {
    transform: translateX(-100%);

    &[data-open="true"] {
      transform: translateX(0);
    }
  }
`;

const Brand = styled.div`
  color: #fff;
  font-size: 20px;
  font-weight: bold;
  padding: 10px 20px;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 10px;
`;

const Nav = styled.nav`
  display: flex;
  flex-direction: column;
`;

const NavLink = styled(Link)`
  color: #ecf0f1;
  padding: 12px 20px;
  border-left: 3px solid transparent;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 10px;
  transition: all ${theme.transitions.fast};

  &:hover,
  &.active,
  &[data-active="true"] {
    background: ${theme.colors.sidebarActive};
    border-left-color: ${theme.colors.sidebarBorder};
  }
`;

const Divider = styled.hr`
  border: none;
  border-top: 1px solid #34495e;
  margin: 10px 20px;
`;

const MainContent = styled.div`
  margin-left: ${theme.layout.sidebarWidth};
  padding: 20px;
  flex: 1;
  min-height: 100vh;
  transition: margin-left 0.3s ease;

  @media (max-width: ${theme.breakpoints.md}) {
    margin-left: 0;
    padding: 10px;
  }
`;

const MobileHeader = styled.div`
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 48px;
  background: ${theme.colors.sidebarBg};
  z-index: 999;
  align-items: center;
  padding: 0 16px;
  justify-content: space-between;

  @media (max-width: ${theme.breakpoints.md}) {
    display: flex;
  }
`;

const MobileMenuButton = styled.button`
  background: none;
  border: none;
  color: #fff;
  font-size: 20px;
  cursor: pointer;
  padding: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
`;

const Overlay = styled.div`
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 999;

  @media (max-width: ${theme.breakpoints.md}) {
    display: block;
  }
`;

interface LayoutProps {
  children: ReactNode;
}

const Layout = ({ children }: LayoutProps) => {
  const location = useLocation();
  const currentPath = location.pathname;
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const isActive = (path: string) => {
    if (path === '/') {
      return currentPath === '/';
    }
    return currentPath.startsWith(path);
  };

  const toggleSidebar = () => setSidebarOpen(!sidebarOpen);
  const closeSidebar = () => setSidebarOpen(false);

  return (
    <LayoutContainer>
      <MobileHeader>
        <MobileMenuButton onClick={toggleSidebar}>
          {sidebarOpen ? <CloseOutlined /> : <MenuOutlined />}
        </MobileMenuButton>
        <Brand style={{ margin: 0, padding: 0, fontSize: '16px' }}>
          <DashboardOutlined />
          ETF Insight
        </Brand>
        <div style={{ width: 40 }} />
      </MobileHeader>

      {sidebarOpen && <Overlay onClick={closeSidebar} />}

      <Sidebar data-open={sidebarOpen}>
        <Brand>
          <DashboardOutlined />
          ETF Insight
        </Brand>
        <Nav>
          <NavLink to="/" data-active={isActive('/')} onClick={closeSidebar}>
            <BarChartOutlined />
            仪表板
          </NavLink>
          <NavLink to="/workflows" data-active={isActive('/workflows')} onClick={closeSidebar}>
            <ProjectOutlined />
            工作流
          </NavLink>
          <NavLink to="/instances" data-active={isActive('/instances')} onClick={closeSidebar}>
            <HistoryOutlined />
            执行记录
          </NavLink>
          <Divider />
          <NavLink to="/etf-dashboard" data-active={isActive('/etf')} onClick={closeSidebar}>
            <BarChartOutlined />
            ETF分析
          </NavLink>
          <NavLink to="/etf-comparison" data-active={isActive('/etf-comparison')} onClick={closeSidebar}>
            <FaBalanceScale style={{ fontSize: '14px' }} />
            对比分析
          </NavLink>
          <NavLink to="/portfolio-analysis" data-active={isActive('/portfolio-analysis')} onClick={closeSidebar}>
            <WalletOutlined />
            组合分析
          </NavLink>
          <NavLink to="/portfolio-config" data-active={isActive('/portfolio-config')} onClick={closeSidebar}>
            <PieChartOutlined />
            组合配置
          </NavLink>
          <NavLink to="/a-share-portfolio" data-active={isActive('/a-share-portfolio')} onClick={closeSidebar}>
            <FundOutlined />
            A股红利组合
          </NavLink>
          <NavLink to="/portfolio-optimization" data-active={isActive('/portfolio-optimization')} onClick={closeSidebar}>
            <PieChartOutlined />
            组合优化
          </NavLink>
          <NavLink to="/factor-analysis" data-active={isActive('/factor-analysis')} onClick={closeSidebar}>
            <LineChartOutlined />
            因子分析
          </NavLink>
          <NavLink to="/factor-timing" data-active={isActive('/factor-timing')} onClick={closeSidebar}>
            <ClockCircleOutlined />
            因子择时
          </NavLink>
          <NavLink to="/alpha-views" data-active={isActive('/alpha-views')} onClick={closeSidebar}>
            <EyeOutlined />
            Alpha观点
          </NavLink>
          <NavLink to="/black-litterman" data-active={isActive('/black-litterman')} onClick={closeSidebar}>
            <CalculatorOutlined />
            BL模型配置
          </NavLink>
          <NavLink to="/risk-budget" data-active={isActive('/risk-budget')} onClick={closeSidebar}>
            <SafetyOutlined />
            风险预算
          </NavLink>
          <NavLink to="/quantlib" data-active={isActive('/quantlib')} onClick={closeSidebar}>
            <NumberOutlined />
            QuantLib 分析
          </NavLink>
          <NavLink to="/ai-agents" data-active={isActive('/ai-agents')} onClick={closeSidebar}>
            <RobotOutlined />
            AI Agent
          </NavLink>
          <Divider />
          <NavLink to="/operation-logs" data-active={isActive('/operation-logs')} onClick={closeSidebar}>
            <FileTextOutlined />
            操作记录
          </NavLink>
          <NavLink to="/etf-config" data-active={isActive('/etf-config')} onClick={closeSidebar}>
            <SettingOutlined />
            ETF配置
          </NavLink>
          <NavLink to="/exchange-rate" data-active={isActive('/exchange-rate')} onClick={closeSidebar}>
            <SwapOutlined />
            外汇管理
          </NavLink>
        </Nav>
      </Sidebar>

      <MainContent style={{ paddingTop: '0' }}>
        <div style={{ height: '48px', display: 'none' }} className="mobile-spacer" />
        {children}
      </MainContent>
    </LayoutContainer>
  );
};

export default Layout;
