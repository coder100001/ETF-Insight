import React from 'react';
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
  SafetyOutlined,
  InteractionOutlined,
  ExperimentOutlined,
  PlayCircleOutlined,
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

const NavLink = styled(Link)<{ $active?: boolean }>`
  color: #ecf0f1;
  padding: 12px 20px;
  border-left: 3px solid transparent;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 10px;
  transition: all ${theme.transitions.fast};

  &:hover,
  &.active {
    background: ${theme.colors.sidebarHover};
    border-left-color: ${theme.colors.sidebarBorder};
  }

  ${props => props.$active && `
    background: ${theme.colors.sidebarActive};
    border-left-color: ${theme.colors.sidebarBorder};
  `}
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
`;

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const location = useLocation();
  const currentPath = location.pathname;

  const isActive = (path: string) => {
    if (path === '/') {
      return currentPath === '/';
    }
    return currentPath.startsWith(path);
  };

  return (
    <LayoutContainer>
      <Sidebar>
        <Brand>
          <DashboardOutlined />
          ETF工作流
        </Brand>
        <Nav>
          <NavLink to="/" $active={isActive('/')}>
            <BarChartOutlined />
            仪表板
          </NavLink>
          <NavLink to="/workflows" $active={isActive('/workflows')}>
            <ProjectOutlined />
            工作流
          </NavLink>
          <NavLink to="/instances" $active={isActive('/instances')}>
            <HistoryOutlined />
            执行记录
          </NavLink>
          <Divider />
          <NavLink to="/etf-dashboard" $active={isActive('/etf')}>
            <BarChartOutlined />
            ETF分析
          </NavLink>
          <NavLink to="/etf-comparison" $active={isActive('/etf-comparison')}>
            <FaBalanceScale style={{ fontSize: '14px' }} />
            对比分析
          </NavLink>
          <NavLink to="/portfolio-analysis" $active={isActive('/portfolio-analysis')}>
            <WalletOutlined />
            组合分析
          </NavLink>
          <NavLink to="/portfolio-config" $active={isActive('/portfolio-config')}>
            <PieChartOutlined />
            组合配置
          </NavLink>
          <NavLink to="/a-share-portfolio" $active={isActive('/a-share-portfolio')}>
            <FundOutlined />
            A股红利组合
          </NavLink>
          <Divider />
          <NavLink to="/analysis/risk" $active={isActive('/analysis/risk')}>
            <SafetyOutlined />
            风险指标分析
          </NavLink>
          <NavLink to="/analysis/overlap" $active={isActive('/analysis/overlap')}>
            <InteractionOutlined />
            持仓重叠分析
          </NavLink>
          <NavLink to="/analysis/factor" $active={isActive('/analysis/factor')}>
            <ExperimentOutlined />
            因子分析
          </NavLink>
          <NavLink to="/analysis/backtest" $active={isActive('/analysis/backtest')}>
            <PlayCircleOutlined />
            回测分析
          </NavLink>
          <Divider />
          <NavLink to="/operation-logs" $active={isActive('/operation-logs')}>
            <FileTextOutlined />
            操作记录
          </NavLink>
          <NavLink to="/etf-config" $active={isActive('/etf-config')}>
            <SettingOutlined />
            ETF配置
          </NavLink>
          <NavLink to="/exchange-rate" $active={isActive('/exchange-rate')}>
            <SwapOutlined />
            外汇管理
          </NavLink>
        </Nav>
      </Sidebar>

      <MainContent>{children}</MainContent>
    </LayoutContainer>
  );
};

export default Layout;
