import { useState, useEffect } from 'react';
import styled from 'styled-components';
import { Card, Table, Button, Badge, Space, App, Modal, InputNumber, Select, Descriptions, Statistic, Row, Col } from 'antd';
import { SwapOutlined, ReloadOutlined, EditOutlined, HistoryOutlined, CalculatorOutlined } from '@ant-design/icons';
import axios from 'axios';

import Layout from '../components/Layout';
import { theme } from '../styles/theme';
import { exchangeRateAPI } from '../services/api';

const PageHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;

  h2 {
    margin: 0;
    font-size: ${theme.fonts.size['2xl']};
    color: ${theme.colors.textPrimary};
    display: flex;
    align-items: center;
    gap: 10px;
  }
`;

const StatsRow = styled.div`
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 20px;

  @media (max-width: ${theme.breakpoints.lg}) {
    grid-template-columns: repeat(2, 1fr);
  }

  @media (max-width: ${theme.breakpoints.sm}) {
    grid-template-columns: 1fr;
  }
`;

const StatCard = styled.div<{ $borderColor?: string }>`
  background: ${theme.colors.surface};
  padding: 20px;
  border-radius: ${theme.borderRadius.md};
  box-shadow: ${theme.shadows.card};
  border-left: 4px solid ${props => props.$borderColor || theme.colors.primary};

  h3 {
    font-size: ${theme.fonts.size['2xl']};
    margin: 0 0 5px 0;
    color: ${props => props.$borderColor || theme.colors.primary};
  }

  p {
    color: ${theme.colors.textSecondary};
    margin: 0;
    font-size: ${theme.fonts.size.base};
  }
`;

const StyledTable = styled(Table)`
  .ant-table-thead > tr > th {
    background: ${theme.colors.background};
    font-weight: ${theme.fonts.weight.semibold};
  }

  .ant-table-tbody > tr:hover > td {
    background: #f8f9fa;
  }
` as typeof Table;

interface ExchangeRate {
  id: number;
  from_currency: string;
  to_currency: string;
  rate: number;
  previous_rate: number;
  change_percent: number;
  updated_at: string;
  data_source: string;
  source_type: string;
}

interface DataSourceStatus {
  name: string;
  available: boolean;
  response_time: string;
  success_rate: string;
  rate_limit: number;
  api_key: string;
}

const ExchangeRatePage: React.FC = () => {
  const { message } = App.useApp();
  const [rates, setRates] = useState<ExchangeRate[]>([]);
  const [loading, setLoading] = useState(false);
  const [lastUpdated, setLastUpdated] = useState<string>('');
  const [dataSourceStatus, setDataSourceStatus] = useState<DataSourceStatus | null>(null);

  // 编辑弹窗状态
  const [editVisible, setEditVisible] = useState(false);
  const [editingRate, setEditingRate] = useState<ExchangeRate | null>(null);

  // 历史弹窗状态
  const [historyVisible, setHistoryVisible] = useState(false);
  const [historyRate, setHistoryRate] = useState<ExchangeRate | null>(null);

  // 汇率转换状态
  const [convertAmount, setConvertAmount] = useState<number>(1000);
  const [convertFrom, setConvertFrom] = useState<string>('USD');
  const [convertTo, setConvertTo] = useState<string>('CNY');
  const [convertResult, setConvertResult] = useState<{ result: number; rate: number } | null>(null);
  const [convertLoading, setConvertLoading] = useState(false);

  useEffect(() => {
    fetchRates();
    fetchDataSourceStatus();
    // 每5分钟自动刷新一次
    const interval = setInterval(() => {
      fetchRates();
      fetchDataSourceStatus();
    }, 5 * 60 * 1000);
    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const fetchRates = async () => {
    setLoading(true);
    try {
      const response = await axios.get('/api/exchange-rates');
      if (response.data.success && response.data.data) {
        setRates(response.data.data);
        setLastUpdated(new Date().toLocaleString('zh-CN'));
      } else {
        message.error('获取汇率数据失败');
      }
    } catch (error) {
      message.error('获取汇率数据失败: ' + (error as Error).message);
    } finally {
      setLoading(false);
    }
  };

  const fetchDataSourceStatus = async () => {
    try {
      const response = await axios.get('/api/exchange-rates/datasource-status');
      if (response.data.success && response.data.data) {
        setDataSourceStatus(response.data.data.current);
      }
    } catch (error) {
      console.error('获取数据源状态失败:', error);
    }
  };

  // 获取指定货币对的汇率显示
  const getRateDisplay = (from: string, to: string): string => {
    const rate = rates.find(r =>
      r.from_currency === from && r.to_currency === to
    );
    if (rate) {
      const change = rate.change_percent;
      const sign = change >= 0 ? '+' : '';
      return `${rate.rate.toFixed(4)} (${sign}${change.toFixed(2)}%)`;
    }
    return '暂无数据';
  };

  // 获取数据源显示名称
  const getDataSourceDisplay = (name: string): string => {
    const displayNames: Record<string, string> = {
      'openexchange': 'Open Exchange Rates',
      'currencyapi': 'CurrencyAPI',
      'frankfurter': 'Frankfurter',
      'fallback': '本地缓存'
    };
    return displayNames[name] || name;
  };

  const handleRefresh = () => {
    fetchRates();
    fetchDataSourceStatus();
    message.success('汇率数据已更新');
  };

  const handleEdit = (record: ExchangeRate) => {
    setEditingRate(record);
    setEditVisible(true);
  };

  const handleHistory = (record: ExchangeRate) => {
    setHistoryRate(record);
    setHistoryVisible(true);
  };

  const handleConvert = async () => {
    if (!convertAmount || convertAmount <= 0) {
      message.warning('请输入有效金额');
      return;
    }
    setConvertLoading(true);
    try {
      const response = await exchangeRateAPI.convert(convertAmount, convertFrom, convertTo);
      if (response.success && response.data) {
        setConvertResult({ result: response.data.result, rate: response.data.rate });
      } else {
        message.error('汇率转换失败');
      }
    } catch {
      message.error('汇率转换失败，请检查网络连接');
    } finally {
      setConvertLoading(false);
    }
  };

  const handleSwapCurrency = () => {
    setConvertFrom(convertTo);
    setConvertTo(convertFrom);
    setConvertResult(null);
  };

  const columns: import('antd').TableProps<ExchangeRate>['columns'] = [
    {
      title: '货币对',
      dataIndex: 'from_currency',
      key: 'pair',
      render: (_, record) => (
        <strong>{record.from_currency}/{record.to_currency}</strong>
      ),
    },
    {
      title: '当前汇率',
      dataIndex: 'rate',
      key: 'rate',
      align: 'center' as const,
      render: (rate) => <strong>{(rate as number).toFixed(4)}</strong>,
    },
    {
      title: '涨跌',
      dataIndex: 'change_percent',
      key: 'change_percent',
      align: 'center' as const,
      render: (value) => (
        <Badge
          count={`${(value as number) >= 0 ? '+' : ''}${(value as number).toFixed(2)}%`}
          style={{
            backgroundColor: (value as number) >= 0 ? theme.colors.success : theme.colors.danger,
          }}
        />
      ),
    },
    {
      title: '数据源',
      dataIndex: 'data_source',
      key: 'data_source',
      align: 'center' as const,
      render: (source: string) => (
        <Badge
          count={source || 'unknown'}
          style={{
            backgroundColor: source === 'demo' ? '#ff4d4f' : source === 'fallback' ? '#faad14' : '#52c41a'
          }}
        />
      ),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      align: 'center' as const,
    },
    {
      title: '操作',
      key: 'action',
      align: 'center' as const,
      render: (_: unknown, record: ExchangeRate) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>编辑</Button>
          <Button size="small" icon={<HistoryOutlined />} onClick={() => handleHistory(record)}>历史</Button>
        </Space>
      ),
    },
  ];

  return (
    <Layout>
      <PageHeader>
        <div style={{ display: 'flex', alignItems: 'center', gap: '15px' }}>
          <h2 style={{ margin: 0 }}>
            <SwapOutlined />
            外汇管理
          </h2>
          {dataSourceStatus && (
            <Badge
              count={`数据源: ${getDataSourceDisplay(dataSourceStatus.name)}`}
              style={{
                backgroundColor: dataSourceStatus.available ? '#52c41a' : '#ff4d4f',
                fontSize: '12px'
              }}
            />
          )}
        </div>
        <Button type="primary" icon={<ReloadOutlined />} onClick={handleRefresh}>
          刷新汇率
        </Button>
      </PageHeader>

      <StatsRow>
        <StatCard $borderColor={theme.colors.primary}>
          <h3>USD/CNY</h3>
          <p>{getRateDisplay('USD', 'CNY')}</p>
        </StatCard>
        <StatCard $borderColor={theme.colors.success}>
          <h3>EUR/CNY</h3>
          <p>{getRateDisplay('EUR', 'CNY')}</p>
        </StatCard>
        <StatCard $borderColor={theme.colors.warning}>
          <h3>GBP/CNY</h3>
          <p>{getRateDisplay('GBP', 'CNY')}</p>
        </StatCard>
        <StatCard $borderColor={theme.colors.info}>
          <h3>JPY/CNY</h3>
          <p>{getRateDisplay('JPY', 'CNY')}</p>
        </StatCard>
      </StatsRow>

      <Card
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <CalculatorOutlined />
            <span>汇率转换</span>
          </div>
        }
        style={{ boxShadow: theme.shadows.card, marginBottom: 20 }}
      >
        <Row gutter={16} align="middle">
          <Col span={5}>
            <InputNumber
              style={{ width: '100%' }}
              value={convertAmount}
              onChange={(v) => { setConvertAmount(v ?? 0); setConvertResult(null); }}
              min={0}
              precision={2}
              placeholder="输入金额"
            />
          </Col>
          <Col span={4}>
            <Select
              style={{ width: '100%' }}
              value={convertFrom}
              onChange={(v) => { setConvertFrom(v); setConvertResult(null); }}
              options={[...new Set(rates.map(r => [r.from_currency, r.to_currency]).flat())].map(c => ({ label: c, value: c }))}
            />
          </Col>
          <Col span={2} style={{ textAlign: 'center' }}>
            <Button
              icon={<SwapOutlined />}
              onClick={handleSwapCurrency}
              type="text"
            />
          </Col>
          <Col span={4}>
            <Select
              style={{ width: '100%' }}
              value={convertTo}
              onChange={(v) => { setConvertTo(v); setConvertResult(null); }}
              options={[...new Set(rates.map(r => [r.from_currency, r.to_currency]).flat())].map(c => ({ label: c, value: c }))}
            />
          </Col>
          <Col span={3}>
            <Button type="primary" icon={<SwapOutlined />} onClick={handleConvert} loading={convertLoading}>
              转换
            </Button>
          </Col>
          <Col span={6}>
            {convertResult && (
              <Statistic
                title={`${convertAmount} ${convertFrom} =`}
                value={convertResult.result}
                precision={2}
                suffix={convertTo}
                valueStyle={{ color: theme.colors.primary }}
              />
            )}
          </Col>
        </Row>
      </Card>

      <Card
        style={{ boxShadow: theme.shadows.card }}
        title={
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span>汇率列表</span>
            {lastUpdated && (
              <span style={{ fontSize: '12px', color: theme.colors.textSecondary }}>
                最后更新: {lastUpdated}
              </span>
            )}
          </div>
        }
      >
        <StyledTable
          dataSource={rates}
          columns={columns}
          rowKey="id"
          pagination={false}
          loading={loading}
        />
      </Card>

      {/* 编辑弹窗 */}
      <Modal
        title={`编辑汇率 - ${editingRate?.from_currency}/${editingRate?.to_currency}`}
        open={editVisible}
        onCancel={() => setEditVisible(false)}
        footer={[
          <Button key="cancel" onClick={() => setEditVisible(false)}>取消</Button>,
          <Button key="save" type="primary" onClick={() => {
            message.success('汇率信息已保存');
            setEditVisible(false);
          }}>保存</Button>,
        ]}
      >
        {editingRate && (
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="货币对">
              {editingRate.from_currency}/{editingRate.to_currency}
            </Descriptions.Item>
            <Descriptions.Item label="当前汇率">
              {editingRate.rate.toFixed(4)}
            </Descriptions.Item>
            <Descriptions.Item label="前次汇率">
              {editingRate.previous_rate?.toFixed(4) ?? '-'}
            </Descriptions.Item>
            <Descriptions.Item label="涨跌幅">
              <span style={{ color: editingRate.change_percent >= 0 ? theme.colors.success : theme.colors.danger }}>
                {editingRate.change_percent >= 0 ? '+' : ''}{editingRate.change_percent?.toFixed(2) ?? 0}%
              </span>
            </Descriptions.Item>
            <Descriptions.Item label="数据源">
              {editingRate.data_source || '未知'}
            </Descriptions.Item>
            <Descriptions.Item label="更新时间">
              {editingRate.updated_at}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>

      {/* 历史记录弹窗 */}
      <Modal
        title={`汇率历史 - ${historyRate?.from_currency}/${historyRate?.to_currency}`}
        open={historyVisible}
        onCancel={() => setHistoryVisible(false)}
        footer={[
          <Button key="close" onClick={() => setHistoryVisible(false)}>关闭</Button>,
        ]}
        width={600}
      >
        {historyRate && (
          <div>
            <Row gutter={16} style={{ marginBottom: 16 }}>
              <Col span={8}>
                <Statistic
                  title="当前汇率"
                  value={historyRate.rate}
                  precision={4}
                  valueStyle={{ color: theme.colors.primary }}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="前次汇率"
                  value={historyRate.previous_rate ?? 0}
                  precision={4}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="涨跌幅"
                  value={historyRate.change_percent ?? 0}
                  precision={2}
                  suffix="%"
                  valueStyle={{ color: (historyRate.change_percent ?? 0) >= 0 ? theme.colors.success : theme.colors.danger }}
                  prefix={(historyRate.change_percent ?? 0) >= 0 ? '+' : ''}
                />
              </Col>
            </Row>
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label="货币对">
                {historyRate.from_currency} → {historyRate.to_currency}
              </Descriptions.Item>
              <Descriptions.Item label="数据源">
                {historyRate.data_source || '未知'}
              </Descriptions.Item>
              <Descriptions.Item label="来源类型">
                {historyRate.source_type || '实时'}
              </Descriptions.Item>
              <Descriptions.Item label="最后更新">
                {historyRate.updated_at}
              </Descriptions.Item>
            </Descriptions>
          </div>
        )}
      </Modal>
    </Layout>
  );
};

export default ExchangeRatePage;
