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
      sortBy: 'symbol',
    };
    setFilters(clearedFilters);
    onFilterChange(clearedFilters);
  }, [onFilterChange]);

  const hasActiveFilters = filters.keyword || filters.strategy || filters.dividendRange || filters.riskLevel;

  const strategyOptions = [
    { value: '', label: '全部策略' },
    { value: 'quality-dividend', label: '质量股息' },
    { value: 'high-yield', label: '高股息收益' },
    { value: 'option-enhanced', label: '期权增强' },
    { value: 'dividend-growth', label: '股息增长' },
    { value: 'tech-growth', label: '科技成长' },
  ];

  const dividendRangeOptions = [
    { value: '', label: '全部股息率' },
    { value: '0-2', label: '0% - 2%' },
    { value: '2-4', label: '2% - 4%' },
    { value: '4-6', label: '4% - 6%' },
    { value: '6+', label: '6%+' },
  ];

  const riskLevelOptions = [
    { value: '', label: '全部风险' },
    { value: 'low', label: '低风险' },
    { value: 'medium', label: '中风险' },
    { value: 'high', label: '高风险' },
  ];

  const sortOptions = [
    { value: 'symbol', label: '代码' },
    { value: 'dividend_yield', label: '股息率' },
    { value: 'total_return', label: '总收益' },
    { value: 'sharpe_ratio', label: '夏普比率' },
    { value: 'volatility', label: '波动率' },
  ];

  return (
    <FilterContainer>
      <FilterRow>
        <FilterGroup style={{ flex: 1, minWidth: '200px' }}>
          <Input
            placeholder="搜索 ETF 代码或名称..."
            prefix={<SearchOutlined />}
            value={filters.keyword}
            onChange={(e) => handleFilterChange('keyword', e.target.value)}
            allowClear
          />
        </FilterGroup>

        <FilterGroup>
          <FilterLabel>策略:</FilterLabel>
          <Select
            style={{ width: 140 }}
            value={filters.strategy}
            onChange={(value) => handleFilterChange('strategy', value)}
            options={strategyOptions}
          />
        </FilterGroup>

        <FilterGroup>
          <FilterLabel>股息率:</FilterLabel>
          <Select
            style={{ width: 130 }}
            value={filters.dividendRange}
            onChange={(value) => handleFilterChange('dividendRange', value)}
            options={dividendRangeOptions}
          />
        </FilterGroup>

        <FilterGroup>
          <FilterLabel>风险:</FilterLabel>
          <Select
            style={{ width: 120 }}
            value={filters.riskLevel}
            onChange={(value) => handleFilterChange('riskLevel', value)}
            options={riskLevelOptions}
          />
        </FilterGroup>

        <FilterGroup>
          <FilterLabel>排序:</FilterLabel>
          <Select
            style={{ width: 130 }}
            value={filters.sortBy}
            onChange={(value) => handleFilterChange('sortBy', value)}
            options={sortOptions}
          />
        </FilterGroup>

        {hasActiveFilters && (
          <Button
            icon={<ClearOutlined />}
            onClick={handleClearFilters}
          >
            清除筛选
          </Button>
        )}
      </FilterRow>

      {hasActiveFilters && (
        <ActiveFilters>
          <FilterLabel>当前筛选:</FilterLabel>
          {filters.keyword && (
            <Tag closable onClose={() => handleFilterChange('keyword', '')}>
              关键词: {filters.keyword}
            </Tag>
          )}
          {filters.strategy && (
            <Tag closable onClose={() => handleFilterChange('strategy', '')}>
              策略: {strategyOptions.find(o => o.value === filters.strategy)?.label}
            </Tag>
          )}
          {filters.dividendRange && (
            <Tag closable onClose={() => handleFilterChange('dividendRange', '')}>
              股息率: {dividendRangeOptions.find(o => o.value === filters.dividendRange)?.label}
            </Tag>
          )}
          {filters.riskLevel && (
            <Tag closable onClose={() => handleFilterChange('riskLevel', '')}>
              风险: {riskLevelOptions.find(o => o.value === filters.riskLevel)?.label}
            </Tag>
          )}
        </ActiveFilters>
      )}
    </FilterContainer>
  );
};

export default ETFFilter;
