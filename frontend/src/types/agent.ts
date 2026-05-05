export type AgentCategory =
  | 'legendary_investor'
  | 'hedge_fund'
  | 'geopolitics'
  | 'macro_economic'
  | 'technical';

export interface AgentInfo {
  id: string;
  name: string;
  category: AgentCategory;
  description: string;
  system_prompt_preview: string;
}

export interface AgentRunRequest {
  agent_id: string;
  query: string;
  context?: Record<string, unknown>;
  llm_provider: string;
  model: string;
  temperature?: number;
  max_tokens?: number;
}

export interface AgentRunResponse {
  agent_id: string;
  agent_name: string;
  response: string;
  model: string;
  tokens_used: number;
  duration_ms: number;
}

export interface AgentTeamRequest {
  agent_ids: string[];
  query: string;
  rounds?: number;
  llm_provider: string;
  model: string;
}

export interface AgentTeamResponse {
  query: string;
  rounds: Array<{
    round: number;
    responses: AgentRunResponse[];
  }>;
  synthesis: string;
}
