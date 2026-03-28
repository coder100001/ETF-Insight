import React from 'react';
import styled from 'styled-components';
import { theme } from '../styles/theme';

const Card = styled.div<{ $borderColor?: string }>`
  background: ${theme.colors.surface};
  border-radius: ${theme.borderRadius.lg};
  padding: ${theme.spacing.lg};
  box-shadow: ${theme.shadows.card};
  transition: all ${theme.transitions.normal};
  border-left: 4px solid ${props => props.$borderColor || theme.colors.primary};

  &:hover {
    box-shadow: ${theme.shadows.hover};
    transform: translateY(-2px);
  }
`;

const Label = styled.div`
  font-size: ${theme.fonts.size.sm};
  color: ${theme.colors.textSecondary};
  margin-bottom: ${theme.spacing.sm};
`;

const Value = styled.div<{ $color?: string }>`
  font-size: ${theme.fonts.size['2xl']};
  font-weight: ${theme.fonts.weight.bold};
  color: ${props => props.$color || theme.colors.textPrimary};
  font-family: ${theme.fonts.familyMono};
`;

const Change = styled.div<{ $isUp?: boolean }>`
  font-size: ${theme.fonts.size.sm};
  color: ${props => props.$isUp ? theme.colors.up : theme.colors.down};
  margin-top: ${theme.spacing.xs};
  font-weight: ${theme.fonts.weight.medium};
`;

const IconWrapper = styled.div<{ $bgColor?: string }>`
  width: 48px;
  height: 48px;
  border-radius: ${theme.borderRadius.md};
  background: ${props => props.$bgColor || theme.colors.primaryLight};
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  margin-bottom: ${theme.spacing.md};
`;

interface StatCardProps {
  label: string;
  value: string | number;
  change?: string;
  isUp?: boolean;
  icon?: React.ReactNode;
  iconBg?: string;
  borderColor?: string;
  valueColor?: string;
}

const StatCard: React.FC<StatCardProps> = ({
  label,
  value,
  change,
  isUp,
  icon,
  iconBg,
  borderColor,
  valueColor,
}) => {
  return (
    <Card $borderColor={borderColor}>
      {icon && <IconWrapper $bgColor={iconBg}>{icon}</IconWrapper>}
      <Label>{label}</Label>
      <Value $color={valueColor}>{value}</Value>
      {change && <Change $isUp={isUp}>{change}</Change>}
    </Card>
  );
};

export default StatCard;
