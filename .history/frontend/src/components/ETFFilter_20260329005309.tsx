import React, { useState, useCallback } from 'react';
import styled from 'styled-components';
import { Input, Select, Space, Button, Tag } from 'antd';
import { SearchOutlined, FilterOutlined, ClearOutlined } from '@ant-design/icons';
import { theme } from '../styles/theme';

const FilterContainer = styled.div`
  background: ${theme.colors.surface};
  border-radius: ${theme.borderRadius.lg};
  padding: ${theme.spacing.lg};
  margin-bottom: ${theme.spacing.lg};
  box-shadow: ${theme.shadows.card};
`;

const FilterRow = styled.div`
  display: flex;
  flex-wrap: wrap;
  gap: ${theme.spacing.md};
  align-items: center;

  @media (max-width: 768px) {
    flex-direction: column;
    align-items: stretch;
  }
`;

const FilterGroup = styled.div`
  display: flex;
  align-items: center;
  gap: ${theme.spacing.sm};

  @media (max-width: 768px) {
    width: 100%;
  }
`;

const FilterLabel = styled.span`
  color: ${theme.colors.textSecondary};
  font-size: ${theme.fonts.size.sm};
  white-space: nowrap;
`;

const ActiveFilters = styled.div`
  margin-top: ${theme.spacing.md};
  padding-top: ${theme.spacing.md};
  border-top: 1px solid ${theme.colors.divider};
  display: flex;
  flex-wrap: wrap;
  gap: ${theme.spacing.sm};
  align-items: center;
`;

export interface FilterState {
  keyword: string;
  strategy: string;
  dividendRange: string;
  riskLevel: string;
  sortBy: string;
}

interface ETFFilterProps {
  onFilterChange: (filters: FilterState) => void;
  initialFilters?: Partial<FilterState>;
}

const ETFFilter: React.FC<ETFFilterProps> = ({ onFilterChange, initialFilters }) => {
  const [filters, setFilters] = useState<FilterState>({
    keyword: initialFilters?.keyword || '',
    strategy: initialFilters?.strategy || '',
    dividendRange: initialFilters?.dividendRange || '',
    riskLevel: initialFilters?.riskLevel || '',
    sortBy: initialFilters?.sortBy || 'symbol',
  });

  const handleFilterChange = useCallback((key: keyof FilterState, value: string) => {
    const newFilters = { ...filters, [key]: value };
    setFilters(newFilters);
    onFilterChange(newFilters);
  }, [filters, onFilterChange]);

  const handleClearFilters = useCallback(() => {
    const clearedFilters: FilterState = {
      keyword: '',
      strategy: '',
      dividendRange: '',
      riskLevel: '',
