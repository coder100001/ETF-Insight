import { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Select,
  Button,
  Input,
  Tag,
  Spin,
  message,
  Typography,
  Space,
  List,
} from 'antd';
import {
  RobotOutlined,
  SendOutlined,
  TeamOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { agentAPI } from '../services/api';
import type { AgentInfo, AgentRunResponse, AgentTeamResponse } from '../types/agent';
import Layout from '../components/Layout';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;
const { Option } = Select;

const categoryColors: Record<string, string> = {
  legendary_investor: 'gold',
  hedge_fund: 'blue',
  geopolitics: 'red',
  macro_economic: 'green',
  technical: 'purple',
};

const categoryLabels: Record<string, string> = {
  legendary_investor: '投资大师',
  hedge_fund: '对冲基金',
  geopolitics: '地缘政治',
  macro_economic: '宏观经济',
  technical: '技术分析',
};

const llmOptions = [
  { provider: 'openai', model: 'gpt-4o-mini', label: 'GPT-4o Mini' },
  { provider: 'openai', model: 'gpt-4o', label: 'GPT-4o' },
  { provider: 'deepseek', model: 'deepseek-chat', label: 'DeepSeek Chat' },
  { provider: 'ollama', model: 'llama3.1', label: 'Ollama (本地)' },
];

const AIAgents = () => {
  const [agents, setAgents] = useState<AgentInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [agentsLoading, setAgentsLoading] = useState(true);
  const [selectedAgent, setSelectedAgent] = useState<string>('');
  const [query, setQuery] = useState('');
  const [llmConfig, setLlmConfig] = useState(llmOptions[0]);
  const [result, setResult] = useState<AgentRunResponse | null>(null);
  const [teamResult, setTeamResult] = useState<AgentTeamResponse | null>(null);
  const [selectedTeam, setSelectedTeam] = useState<string[]>([]);
  const [mode, setMode] = useState<'single' | 'team'>('single');

  useEffect(() => {
    loadAgents();
  }, []);

  const loadAgents = async () => {
    setAgentsLoading(true);
    try {
      const resp = await agentAPI.discover();
      if (resp.success && resp.data) {
        setAgents(resp.data);
      } else {
        message.error('获取Agent列表失败');
      }
    } catch {
      message.error('Agent服务不可用，请确认服务已启动 (port 8091)');
    } finally {
      setAgentsLoading(false);
    }
  };

  const handleRun = async () => {
    if (!query.trim()) {
      message.warning('请输入分析问题');
      return;
    }

    setLoading(true);
    setResult(null);
    setTeamResult(null);

    try {
      if (mode === 'single' && selectedAgent) {
        const resp = await agentAPI.run({
          agent_id: selectedAgent,
          query,
          llm_provider: llmConfig.provider,
          model: llmConfig.model,
        });
        if (resp.success && resp.data) {
          setResult(resp.data);
        } else {
          message.error(resp.error || '执行失败');
        }
      } else if (mode === 'team' && selectedTeam.length >= 2) {
        const resp = await agentAPI.runTeam({
          agent_ids: selectedTeam,
          query,
          rounds: 1,
          llm_provider: llmConfig.provider,
          model: llmConfig.model,
        });
        if (resp.success && resp.data) {
          setTeamResult(resp.data);
        } else {
          message.error(resp.error || '团队执行失败');
        }
      } else {
        message.warning('请选择Agent');
      }
    } catch {
      message.error('请求失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Layout>
      <Title level={2}>
        <RobotOutlined /> AI 投资分析助手
      </Title>
      <Paragraph type="secondary">
        基于多Agent协作的智能投资分析平台，支持投资大师视角分析和多Agent辩论
      </Paragraph>

      <Row gutter={24}>
        <Col span={8}>
          <Card title="Agent 列表" loading={agentsLoading} style={{ marginBottom: 16 }}>
            <Space direction="vertical" style={{ width: '100%' }}>
              <Select
                placeholder="选择模式"
                value={mode}
                onChange={setMode}
                style={{ width: '100%' }}
              >
                <Option value="single">单Agent分析</Option>
                <Option value="team">多Agent团队</Option>
              </Select>

              {mode === 'single' ? (
                <Select
                  placeholder="选择一个Agent"
                  value={selectedAgent}
                  onChange={setSelectedAgent}
                  style={{ width: '100%' }}
                >
                  {agents.map((a) => (
                    <Option key={a.id} value={a.id}>
                      <Tag color={categoryColors[a.category]}>{categoryLabels[a.category]}</Tag>
                      {a.name}
                    </Option>
                  ))}
                </Select>
              ) : (
                <Select
                  mode="multiple"
                  placeholder="选择2-5个Agent"
                  value={selectedTeam}
                  onChange={setSelectedTeam}
                  style={{ width: '100%' }}
                  maxCount={5}
                >
                  {agents.map((a) => (
                    <Option key={a.id} value={a.id}>
                      <Tag color={categoryColors[a.category]}>{categoryLabels[a.category]}</Tag>
                      {a.name}
                    </Option>
                  ))}
                </Select>
              )}

              <Select
                value={llmConfig.model}
                onChange={(val) => setLlmConfig(llmOptions.find((o) => o.model === val) || llmOptions[0])}
                style={{ width: '100%' }}
              >
                {llmOptions.map((o) => (
                  <Option key={o.model} value={o.model}>{o.label}</Option>
                ))}
              </Select>
            </Space>
          </Card>

          <Card title="Agent 详情" size="small">
            {agents.filter((a) => mode === 'single' ? a.id === selectedAgent : selectedTeam.includes(a.id)).map((a) => (
              <div key={a.id} style={{ marginBottom: 12 }}>
                <Text strong>{a.name}</Text>
                <br />
                <Tag color={categoryColors[a.category]}>{categoryLabels[a.category]}</Tag>
                <br />
                <Text type="secondary" style={{ fontSize: 12 }}>{a.description}</Text>
              </div>
            ))}
            {mode === 'single' && !selectedAgent && <Text type="secondary">请先选择一个Agent</Text>}
            {mode === 'team' && selectedTeam.length < 2 && <Text type="secondary">请至少选择2个Agent</Text>}
          </Card>
        </Col>

        <Col span={16}>
          <Card
            title={mode === 'single' ? '单Agent分析' : '多Agent团队辩论'}
            extra={
              <Button
                type="primary"
                icon={<SendOutlined />}
                onClick={handleRun}
                loading={loading}
                disabled={mode === 'single' ? !selectedAgent : selectedTeam.length < 2}
              >
                {mode === 'team' ? <><TeamOutlined /> 团队分析</> : '开始分析'}
              </Button>
            }
          >
            <TextArea
              rows={3}
              placeholder="输入你的投资分析问题，例如：分析苹果公司(AAPL)的投资价值"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onPressEnter={(e) => {
                if (e.ctrlKey) handleRun();
              }}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>Ctrl+Enter 发送</Text>
          </Card>

          {loading && (
            <Card style={{ marginTop: 16, textAlign: 'center' }}>
              <Spin size="large" />
              <div style={{ marginTop: 16 }}>
                <Text>Agent 正在分析中...</Text>
              </div>
            </Card>
          )}

          {result && mode === 'single' && (
            <Card
              title={
                <Space>
                  <UserOutlined />
                  {result.agent_name}
                  <Tag>{result.model}</Tag>
                  <Text type="secondary">{result.tokens_used} tokens · {result.duration_ms}ms</Text>
                </Space>
              }
              style={{ marginTop: 16 }}
            >
              <div style={{ whiteSpace: 'pre-wrap', lineHeight: 1.8 }}>
                {result.response}
              </div>
            </Card>
          )}

          {teamResult && mode === 'team' && (
            <>
              {teamResult.rounds.map((round) => (
                <Card
                  key={round.round}
                  title={`第 ${round.round} 轮`}
                  style={{ marginTop: 16 }}
                >
                  <List
                    dataSource={round.responses}
                    renderItem={(resp: AgentRunResponse) => (
                      <List.Item>
                        <List.Item.Meta
                          title={<Space><UserOutlined /> {resp.agent_name} <Tag>{resp.model}</Tag></Space>}
                          description={
                            <div style={{ whiteSpace: 'pre-wrap', lineHeight: 1.6 }}>
                              {resp.response}
                            </div>
                          }
                        />
                      </List.Item>
                    )}
                  />
                </Card>
              ))}

              <Card
                title="综合分析"
                style={{ marginTop: 16, borderColor: '#1890ff' }}
              >
                <div style={{ whiteSpace: 'pre-wrap', lineHeight: 1.8 }}>
                  {teamResult.synthesis}
                </div>
              </Card>
            </>
          )}
        </Col>
      </Row>
    </Layout>
  );
};

export default AIAgents;
