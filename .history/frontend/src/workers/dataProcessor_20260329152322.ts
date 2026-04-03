// Web Worker - 数据处理
// 用于处理大量数据计算，避免阻塞主线程

interface DataProcessRequest {
  type: 'calculate-returns' | 'calculate-dividends' | 'process-holdings';
  data: any;
  options?: {
    taxRate?: number;
    initialInvestment?: number;
    years?: number;
  };
  requestId?: string;
}

interface DataProcessResponse {
  requestId: string;
  type: string;
  result: any;
  error?: string;
}

// 处理投资收益计算
function calculateReturns(data: any, options: any) {
  const {
    initialInvestment = 10000,
    taxRate = 0.10,
    years = 10,
    dividendYield = 0.04,
    growthRate = 0.07,
  } = options;

  const results = [];
  let investment = initialInvestment;

  for (let year = 0; year <= years; year++) {
    // 计算股息
    const dividend = investment * dividendYield;
    const dividendAfterTax = dividend * (1 - taxRate);

    // 计算资本增值
    const capitalAppreciation = investment * growthRate;

    // 更新投资价值
    investment += capitalAppreciation + dividendAfterTax;

    results.push({
      year,
      investment: investment,
      capitalAppreciation: capitalAppreciation,
      dividend: dividend,
      dividendAfterTax: dividendAfterTax,
      totalReturn: investment - initialInvestment,
    });
  }

  return results;
}

// 处理持仓数据
function processHoldings(data: any, options: any) {
  const { holdings, taxRate = 0.10 } = options;

  const processed = holdings.map((holding: any) => {
    const {
      investment,
      dividendYield,
      currentPrice,
      shares,
    } = holding;

    // 计算年股息
    const annualDividendBeforeTax = investment * dividendYield;
    const dividendTax = annualDividendBeforeTax * taxRate;
    const annualDividendAfterTax = annualDividendBeforeTax - dividendTax;

    return {
      ...holding,
      annualDividendBeforeTax,
      dividendTax,
      annualDividendAfterTax,
      current_value: shares * currentPrice,
    };
  });

  // 计算总计
  const totals = processed.reduce(
    (acc: any, curr: any) => {
      acc.investment += curr.investment;
      acc.annualDividendBeforeTax += curr.annualDividendBeforeTax;
      acc.dividendTax += curr.dividendTax;
      acc.annualDividendAfterTax += curr.annualDividendAfterTax;
      acc.current_value += curr.current_value;
      return acc;
    },
    {
      investment: 0,
      annualDividendBeforeTax: 0,
      dividendTax: 0,
      annualDividendAfterTax: 0,
      current_value: 0,
    }
  );

  return {
    holdings: processed,
    totals: totals,
  };
}

// 监听消息
self.onmessage = (event: MessageEvent<DataProcessRequest>) => {
  const { type, data, options = {}, requestId = Date.now().toString() } = event.data;

  try {
    let result: any;

    switch (type) {
      case 'calculate-returns':
        result = calculateReturns(data, options);
        break;
      case 'calculate-dividends':
        result = processHoldings(data, options);
        break;
      case 'process-holdings':
        result = processHoldings(data, options);
        break;
      default:
        result = { error: `Unknown type: ${type}` };
    }

    self.postMessage({
      requestId,
      type,
      result,
    } as DataProcessResponse);
  } catch (error) {
    self.postMessage({
      requestId,
      type,
      error: (error as Error).message,
    } as DataProcessResponse);
  }
};

self.onerror = (event: Event) => {
  console.error('Worker error:', event);
};
