import React from 'react';
import styled from 'styled-components';
import { theme } from '../styles/theme';
import type { ETFData } from '../types';
import { ArrowUpOutlined, ArrowDownOutlined, LineChartOutlined } from '@ant-design/icons';

const Card = styled.div`
  background: ${theme.colors.surface};
  border-radius: ${theme.borderRadius.lg};
  box-shadow: ${theme.shadows.card};
  padding: ${theme.spacing.lg};
  transition: all ${theme.transitions.normal};
  cursor: pointer;
  border: 1px solid transparent;

  &:hover {
    box-shadow: ${theme.shadows.hover};
    border-color: ${theme.colors.primaryLight};
    transform: translateY(-2px);
  }
`;

const Header = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: ${theme.spacing.md};
`;

const Symbol = styled.div`
  font-size: ${theme.fonts.size.xl};
  font-weight: ${theme.fonts.weight.bold};
  color: ${theme.colors.textPrimary};
`;

const Tag = styled.span<{ $isUp: boolean }>`
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: ${theme.borderRadius.sm};
  font-size: ${theme.fonts.size.sm};
  font-weight: ${theme.fonts.weight.semibold};
  background: ${props => props.$isUp ? theme.colors.upBg : theme.colors.downBg};
  color: ${props => props.$isUp ? theme.colors.up : theme.colors.down};
`;

const PriceSection = styled.div`
  margin-bottom: ${theme.spacing.lg};
`;

const CurrentPrice = styled.div`
  font-size: ${theme.fonts.size['3xl']};
  font-weight: ${theme.fonts.weight.bold};
  color: ${theme.colors.textPrimary};
  font-family: ${theme.fonts.familyMono};
`;

const PriceChange = styled.div<{ $isUp: boolean }>`
  display: flex;
  align-items: center;
  gap: ${theme.spacing.sm};
  margin-top: ${theme.spacing.xs};
  font-size: ${theme.fonts.size.base};
  color: ${props => props.$isUp ? theme.colors.up : theme.colors.down};
  font-weight: ${theme.fonts.weight.medium};
`;

const InfoGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: ${theme.spacing.md};
  padding-top: ${theme.spacing.md};
  border-top: 1px solid ${theme.colors.divider};
`;

const InfoItem = styled.div`
  display: flex;
  flex-direction: column;
  gap: 2px;
`;

const InfoLabel = styled.span`
  font-size: ${theme.fonts.size.xs};
  color: ${theme.colors.textTertiary};
`;

const InfoValue = styled.span<{ $color?: string }>`
  font-size: ${theme.fonts.size.sm};
  font-weight: ${theme.fonts.weight.medium};
  color: ${props => props.$color || theme.colors.textSecondary};
  font-family: ${theme.fonts.familyMono};
`;

const StrategyTag = styled.div`
  margin-top: ${theme.spacing.md};
  padding: ${theme.spacing.sm} ${theme.spacing.md};
  background: ${theme.colors.background};
  border-radius: ${theme.borderRadius.md};
  font-size: ${theme.fonts.size.xs};
  color: ${theme.colors.textSecondary};
  text-align: center;
`;

const ActionButton = styled.button`
  width: 100%;
  margin-top: ${theme.spacing.md};
  padding: ${theme.spacing.sm} ${theme.spacing.md};
  background: ${theme.colors.primaryLight};
  color: ${theme.colors.primary};
  border: none;
  border-radius: ${theme.borderRadius.md};
  font-size: ${theme.fonts.size.sm};
  font-weight: ${theme.fonts.weight.medium};
  cursor: pointer;
  transition: all ${theme.transitions.fast};
  display: flex;
  align-items: center;
  justify-content: center;
  gap: ${theme.spacing.xs};

  &:hover {
    background: ${theme.colors.primary};
    color: ${theme.colors.textInverse};
  }
`;

interface StockCardProps {
  etf: ETFData;
  onClick?: (etf: ETFData) => void;
  onDetailClick?: (etf: ETFData) => void;
}

const StockCard: React.FC<StockCardProps> = ({ etf, onClick, onDetailClick }) => {
  const isUp = etf.change_percent >= 0;
  
  const handleClick = () => {
    onClick?.(etf);
  };

  const handleDetailClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    onDetailClick?.(etf);
  };

  return (
    <Card onClick={handleClick}>
      <Header>
        <Symbol>{etf.symbol}</Symbol>
        <Tag $isUp={isUp}>
          {isUp ? <ArrowUpOutlined /> : <ArrowDownOutlined />}
          {Math.abs(etf.change_percent).toFixed(2)}%
        </Tag>
      </Header>

      <PriceSection>
        <CurrentPrice>${etf.current_price.toFixed(2)}</CurrentPrice>
        <PriceChange $isUp={isUp}>
          {isUp ? '+' : ''}{etf.change.toFixed(2)} ({isUp ? '+' : ''}{etf.change_percent.toFixed(2)}%)
        </PriceChange>
      </PriceSection>

      <InfoGrid>
        <InfoItem>
          <InfoLabel>股息率</InfoLabel>
          <InfoValue $color={theme.colors.warning}>
            {etf.dividend_yield ? `${etf.dividend_yield.toFixed(2)}%` : '-'}
          </InfoValue>
        </InfoItem>
        <InfoItem>
          <InfoLabel>年化波动率</InfoLabel>
          <InfoValue>{etf.volatility?.toFixed(2) ?? '-'}%</InfoValue>
        </InfoItem>
        <InfoItem>
          <InfoLabel>年度收益</InfoLabel>
          <InfoValue $color={(etf.total_return ?? 0) >= 0 ? theme.colors.up : theme.colors.down}>
            {(etf.total_return ?? 0) >= 0 ? '+' : ''}{etf.total_return?.toFixed(2) ?? '-'}%
          </InfoValue>
        </InfoItem>
        <InfoItem>
          <InfoLabel>夏普比率</InfoLabel>
          <InfoValue>{etf.sharpe_ratio?.toFixed(2) ?? '-'}</InfoValue>
        </InfoItem>
      </InfoGrid>

      <StrategyTag>{etf.info?.strategy ?? '-'}</StrategyTag>

      <ActionButton onClick={handleDetailClick}>
        <LineChartOutlined />
        查看详情
      </ActionButton>
    </Card>
  );
};

export default StockCard;