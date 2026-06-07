import { useState, useEffect } from 'react';
import { financialConfigAPI } from '../services/api';
import type { FinancialConfig } from '../types/common';

const DEFAULT_CONFIG: FinancialConfig = {
  risk_free_rate: 0.0435,
  trading_days_year: 252,
  default_currency: 'USD',
};

export function useFinancialConfig() {
  const [config, setConfig] = useState<FinancialConfig>(DEFAULT_CONFIG);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    financialConfigAPI.get()
      .then(res => {
        const data = res?.data || res;
        if (data && typeof data === 'object' && 'risk_free_rate' in data) {
          setConfig(data as FinancialConfig);
        }
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  return { config, loading, setConfig };
}
