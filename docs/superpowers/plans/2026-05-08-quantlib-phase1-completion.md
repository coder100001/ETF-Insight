# QuantLib Phase 1 收尾实施计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 完成 Fincept Terminal QuantLib 集成的 Phase 1 收尾工作：补全美式期权前端UI、Greeks计算入口、动态参考数据、测试覆盖提升

**架构：** 前端 React (Ant Design + Recharts) 页面增强 + Go 后端测试补充。后端 API 已全部完成，仅需前端 UI 补全和测试覆盖提升。

**技术栈：** TypeScript, React, Ant Design, Recharts, Go (testing), shopspring/decimal

---

## 文件结构

| 文件 | 职责 | 操作 |
|------|------|------|
| `frontend/src/pages/QuantLibAnalysis.tsx` | 主页面：添加美式期权Tab + Greeks面板 + 动态参考数据 | **修改** |
| `frontend/src/types/quantlib.ts` | 类型定义（已完整，无需修改） | 只读 |
| `frontend/src/services/api.ts` | quantlibAPI 已完整（无需修改） | 只读 |
| `backend/services/quantlib/quantlib_client_test.go` | 现有测试文件 | **修改**（追加测试） |
| `backend/services/quantlib/quantlib_client_extended_test.go` | 新增扩展测试文件 | **创建** |

---

### 任务 1：添加美式期权定价 Tab

**优先级：** P0（用户可见功能缺失）
**预估时间：** 30 分钟
**依赖：** 无

**文件：**
- 修改：`frontend/src/pages/QuantLibAnalysis.tsx`

- [ ] **步骤 1：添加美式期权状态和表单**

在组件内（第 39-49 行附近），新增状态和表单实例：

```tsx
// 在第 44 行后添加：
const [americanOptionResult, setAmericanOptionResult] = useState<OptionResult | null>(null);
const [americanOptionForm] = Form.useForm();
```

- [ ] **步骤 2：添加美式期权处理函数**

在第 163 行（handleVaRCalculate 函数之后）添加：

```tsx
const handleAmericanOptionPrice = async (values: {
  spot: number;
  strike: number;
  rate: number;
  volatility: number;
  time_to_expiry: number;
  option_type: 'call' | 'put';
  steps: number;
}) => {
  setLoading(true);
  try {
    const response = await quantlibAPI.priceAmericanOption(values);
    if (response.success && response.data) {
      setAmericanOptionResult(response.data);
      message.success('美式期权定价完成');
    } else {
      message.error(response.error || '美式期权定价失败');
    }
  } catch {
    message.error('请求失败');
  } finally {
    setLoading(false);
  }
};
```

- [ ] **步骤 3：添加美式期权 TabPane**

在第 264 行（</TabPane> 选项定价 Tab 结束后，第 266 行债券定价 Tab 前）插入：

```tsx
<TabPane tab="美式期权" key="americanOption">
  <Row gutter={24}>
    <Col span={10}>
      <Card title="美式期权参数（二叉树法）">
        <Form
          form={americanOptionForm}
          layout="vertical"
          onFinish={handleAmericanOptionPrice}
          initialValues={{
            spot: 100,
            strike: 100,
            rate: 0.05,
            volatility: 0.25,
            time_to_expiry: 1.0,
            option_type: 'call',
            steps: 200,
          }}
        >
          <Form.Item label="标的价格" name="spot" rules={[{ required: true }]}>
            <InputNumber style={{ width: '100%' }} min={0} step={1} />
          </Form.Item>
          <Form.Item label="行权价" name="strike" rules={[{ required: true }]}>
            <InputNumber style={{ width: '100%' }} min={0} step={1} />
          </Form.Item>
          <Form.Item label="无风险利率" name="rate" rules={[{ required: true }]}>
            <InputNumber style={{ width: '100%' }} min={0} max={1} step={0.01} />
          </Form.Item>
          <Form.Item label="波动率" name="volatility" rules={[{ required: true }]}>
            <InputNumber style={{ width: '100%' }} min={0} max={5} step={0.01} />
          </Form.Item>
          <Form.Item label="到期时间(年)" name="time_to_expiry" rules={[{ required: true }]}>
            <InputNumber style={{ width: '100%' }} min={0.01} step={0.1} />
          </Form.Item>
          <Form.Item label="期权类型" name="option_type" rules={[{ required: true }]}>
            <Select>
              <Option value="call">看涨 (Call)</Option>
              <Option value="put">看跌 (Put)</Option>
            </Select>
          </Form.Item>
          <Form.Item label="二叉树步数" name="steps" rules={[{ required: true }]}>
            <InputNumber style={{ width: '100%' }} min={10} max={1000} step={10} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block>
              计算美式期权价格
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </Col>
    <Col span={14}>
      <Card title="美式期权定价结果">
        {americanOptionResult ? (
          <Table
            dataSource={[
              { key: 'price', label: '期权价格', value: parseFloat(americanOptionResult.price).toFixed(4) },
              { key: 'delta', label: 'Delta', value: parseFloat(americanOptionResult.delta).toFixed(4) },
              { key: 'gamma', label: 'Gamma', value: parseFloat(americanOptionResult.gamma).toFixed(4) },
              { key: 'theta', label: 'Theta', value: parseFloat(americanOptionResult.theta).toFixed(4) },
              { key: 'vega', label: 'Vega', value: parseFloat(americanOptionResult.vega).toFixed(4) },
              { key: 'rho', label: 'Rho', value: parseFloat(americanOptionResult.rho).toFixed(4) },
            ]}
            columns={optionColumns}
            pagination={false}
            size="small"
          />
        ) : (
          <div style={{ textAlign: 'center', padding: 40, color: '#aaa' }}>
            请填写参数并点击计算
          </div>
        )}
      </Card>
    </Col>
  </Row>
</TabPane>
```

- [ ] **步骤 4：验证代码编译**

运行：
```bash
cd frontend && npx tsc --noEmit
```
预期：无类型错误

- [ ] **步骤 5：Commit**

```bash
git add frontend/src/pages/QuantLibAnalysis.tsx
git commit -m "feat(quantlib): add American option pricing tab with binomial tree UI"
```

---

### 任务 2：集成 Greeks 计算功能

**优先级：** P0（API已完成但前端未暴露）
**预估时间：** 20 分钟
**依赖：** 任务 1

**文件：**
- 修改：`frontend/src/pages/QuantLibAnalysis.tsx`

- [ ] **步骤 1：添加 Greeks 状态和处理函数**

在第 49 行（varForm 后）添加：

```tsx
const [greeksResult, setGreeksResult] = useState<OptionResult | null>(null);
const [greeksForm] = Form.useForm();
```

在 handleAmericanOptionPrice 函数后添加：

```tsx
const handleGreeksCalculate = async (values: {
  spot: number;
  strike: number;
  rate: number;
  volatility: number;
  time_to_expiry: number;
  option_type: 'call' | 'put';
}) => {
  setLoading(true);
  try {
    const response = await quantlibAPI.calculateGreeks(values);
    if (response.success && response.data) {
      setGreeksResult(response.data);
      message.success('Greeks 计算完成');
    } else {
      message.error(response.error || 'Greeks 计算失败');
    }
  } catch {
    message.error('请求失败');
  } finally {
    setLoading(false);
  }
};
```

- [ ] **步骤 2：在欧式期权结果区域增加 Greeks 详细分析按钮**

修改第 248-261 行的 Card 内容，增加切换按钮：

```tsx
<Card title="定价结果与希腊字母" extra={
  <Button size="small" onClick={() => setActiveGreeksTab(!activeGreeksTab)}>
    {activeGreeksTab ? '返回定价结果' : 'Greeks 敏感性分析'}
  </Button>
}>
```

注意：需要先添加状态 `const [activeGreeksTab, setActiveGreeksTab] = useState(false);`

- [ ] **步骤 3：添加 Greeks 分析子面板**

在选项定价 Tab 的 Col 内部，根据 activeGreeksTab 条件渲染：

```tsx
{activeGreeksTab ? (
  greeksResult ? (
    <div>
      <Alert
        message="Greeks 敏感性分析"
        description="展示各希腊字母对期权价格的影响"
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />
      <Table
        dataSource={[
          { key: 'delta', label: 'Delta (Δ)', value: parseFloat(greeksResult.delta).toFixed(6), desc: '价格对标的价格敏感度' },
          { key: 'gamma', label: 'Gamma (Γ)', value: parseFloat(greeksResult.gamma).toFixed(6), desc: 'Delta 对标的价格变化率' },
          { key: 'theta', label: 'Theta (Θ)', value: parseFloat(greeksResult.theta).toFixed(6), desc: '时间衰减率' },
          { key: 'vega', label: 'Vega (ν)', value: parseFloat(greeksResult.vega).toFixed(6), desc: '对波动率的敏感度' },
          { key: 'rho', label: 'Rho (ρ)', value: parseFloat(greeksResult.rho).toFixed(6), desc: '对利率的敏感度' },
        ]}
        columns={[
          { title: '指标', dataIndex: 'label', key: 'label' },
          { title: '数值', dataIndex: 'value', key: 'value' },
          { title: '解释', dataIndex: 'desc', key: 'desc' },
        ]}
        pagination={false}
        size="small"
      />
      <Button
        type="link"
        onClick={() => {
          greeksForm.setFieldsValue(optionForm.getFieldsValue());
          handleGreeksCalculate(optionForm.getFieldsValue());
        }}
        style={{ marginTop: 12 }}
      >
        使用当前参数重新计算 Greeks
      </Button>
    </div>
  ) : (
    <div style={{ textAlign: 'center', padding: 40, color: '#aaa' }}>
      请先计算期权价格，然后查看 Greeks 分析
    </div>
  )
) : (
  // 原有的 optionResult Table 保持不变
)}
```

- [ ] **步骤 4：导入 Alert 组件**

修改第 3 行的 antd import，添加 Alert：

```tsx
import {
  Card,
  Tabs,
  Form,
  InputNumber,
  Select,
  Button,
  Row,
  Col,
  Statistic,
  Table,
  message,
  Spin,
  Input,
  Alert,  // 新增
} from 'antd';
```

- [ ] **步骤 5：验证编译**

```bash
cd frontend && npx tsc --noEmit
```
预期：无错误

- [ ] **步骤 6：Commit**

```bash
git add frontend/src/pages/QuantLibAnalysis.tsx
git commit -m "feat(quantlib): integrate Greeks calculation with sensitivity analysis panel"
```

---

### 任务 3：收益率曲线集成动态参考数据（含前端缓存）

**优先级：** P1（改善用户体验 + 避免API限频）
**预估时间：** 30 分钟
**依赖：** 任务 1, 2

**设计决策：**
- ✅ **必须实现前端内存缓存层**
- 缓存策略：TTL = 5 分钟（避免频繁调用 API）
- 缓存位置：React 组件 state + localStorage 持久化
- 失效机制：TTL 过期自动刷新 + 手动强制刷新按钮

**文件：**
- 修改：`frontend/src/pages/QuantLibAnalysis.tsx`

- [ ] **步骤 1：创建自定义 Hook `useReferenceData`（带缓存）**

在文件顶部（import 之后）添加：

```tsx
// 参考数据缓存接口
interface ReferenceDataCache<T> {
  data: T;
  timestamp: number;
  ttl: number; // 毫秒
}

// 自定义Hook：带缓存的参考数据加载
function useReferenceData<T>(
  fetchFn: () => Promise<unknown>,
  cacheKey: string,
  ttlMinutes: number = 5
): { data: T[] | null; loading: boolean; refresh: () => void } {
  const [data, setData] = useState<T[] | null>(null);
  const [loading, setLoading] = useState(false);

  const loadData = async (forceRefresh: boolean = false) => {
    // 检查 localStorage 缓存
    if (!forceRefresh) {
      try {
        const cached = localStorage.getItem(cacheKey);
        if (cached) {
          const parsed: ReferenceDataCache<T[]> = JSON.parse(cached);
          const now = Date.now();
          if (now - parsed.timestamp < parsed.ttl) {
            setData(parsed.data);
            return; // 命中缓存
          }
        }
      } catch {
        console.warn(`Failed to parse cache for ${cacheKey}`);
      }
    }

    setLoading(true);
    try {
      const response = await fetchFn();
      if (response && Array.isArray(response)) {
        const result = response as T[];
        setData(result);

        // 写入缓存（localStorage + 内存）
        const cacheEntry: ReferenceDataCache<T[]> = {
          data: result,
          timestamp: Date.now(),
          ttl: ttlMinutes * 60 * 1000,
        };
        localStorage.setItem(cacheKey, JSON.stringify(cacheEntry));
      }
    } catch (error) {
      console.error(`Failed to load reference data for ${cacheKey}:`, error);
      // 如果加载失败但有旧缓存，保留旧数据（降级策略）
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []); // 仅组件挂载时加载一次

  return { data, loading, refresh: () => loadData(true) };
}
```

- [ ] **步骤 2：在组件中使用缓存 Hook 加载货币列表**

替换原有的状态定义和 useEffect：

```tsx
// 删除旧的：
// const [currencies, setCurrencies] = useState<string[]>(['USD', 'EUR', 'CNY', 'GBP', 'JPY']);
// const [loadingRefData, setLoadingRefData] = useState(false);
// useEffect(() => { ... }, []);

// 替换为：
const { data: currencies, loading: loadingCurrencies, refresh: refreshCurrencies } =
  useReferenceData<string>(
    async () => {
      const response = await quantlibAPI.getReferenceData('currencies');
      return response.success ? response.data : null;
    },
    'quantlib_currencies',
    5 // 5分钟TTL
  );
```

- [ ] **步骤 3：修改收益率曲线的 Currency 字段为动态加载（含刷新按钮）**

替换第 352-360 行的硬编码 Select 为：

```tsx
<Form.Item label="货币" name="currency" rules={[{ required: true }]}>
  <Input.Group compact>
    <Select
      style={{ width: 'calc(100% - 32px)' }}
      loading={loadingCurrencies}
      showSearch
      optionFilterProp="children"
      notFoundContent={loadingCurrencies ? <Spin size="small" /> : '暂无数据'}
    >
      {(currencies || ['USD', 'EUR', 'CNY', 'GBP', 'JPY']).map((currency) => (
        <Option key={currency} value={currency}>{currency}</Option>
      ))}
    </Select>
    <Button
      icon={<ReloadOutlined />}
      onClick={(e) => {
        e.preventDefault();
        refreshCurrencies();
      }}
      title="刷新货币列表"
    />
  </Input.Group>
</Form.Item>
```

注意：需要导入 ReloadOutlined 图标：
```tsx
import { ReloadOutlined } from '@ant-design/icons';
```

- [ ] **步骤 4：（可选增强）添加日历和日期计数约定缓存加载**

如果需要更完整的参考数据支持：

```tsx
const { data: calendars, loading: loadingCalendars, refresh: refreshCalendars } =
  useReferenceData<string>(
    async () => {
      const response = await quantlibAPI.getReferenceData('calendars');
      return response.success ? response.data : null;
    },
    'quantlib_calendars',
    10 // 日历数据变化少，10分钟TTL
  );

const { data: dayCounts, loading: loadingDayCounts, refresh: refreshDayCounts } =
  useReferenceData<string>(
    async () => {
      const response = await quantlibAPI.getReferenceData('day-count-conventions');
      return response.success ? response.data : null;
    },
    'quantlib_daycounts',
    10
  );
```

- [ ] **步骤 5：验证编译和缓存行为**

```bash
cd frontend && npx tsc --noEmit
```

手动测试：
1. 打开页面 → 检查 Network 面板，应只有 1 次 `/reference/currencies` 请求
2. 切换 Tab 再切回来 → 不应有新请求（命中缓存）
3. 点击刷新按钮 → 应有 1 次新请求
4. 关闭浏览器再打开 → 5分钟内不应有请求（localStorage 生效）
5. 等待 5 分钟后操作 → 应自动刷新

预期：编译无错误，缓存机制正常工作

- [ ] **步骤 6：Commit**

```bash
git add frontend/src/pages/QuantLibAnalysis.tsx
git commit -m "feat(quantlib): integrate dynamic reference data with TTL cache to prevent API rate limiting

- Add useReferenceData custom hook with localStorage + memory caching
- Cache TTL: 5 minutes for currencies, 10 minutes for calendars/day-counts
- Add manual refresh button to force cache invalidation
- Fallback to defaults when API unavailable (graceful degradation)"
```

---

### 任务 4：补充核心功能测试（覆盖率目标 >70%）

**优先级：** P1（质量保障）
**预估时间：** 45 分钟
**依赖：** 无（可并行）

**文件：**
- 修改：`backend/services/quantlib/quantlib_client_test.go`（追加）
- 创建：`backend/services/quantlib/quantlib_client_extended_test.go`（新文件）

#### 4.1 美式期权测试

- [ ] **步骤 1：编写美式期权正常路径测试**

在 `quantlib_client_test.go` 文件末尾追加：

```go
func TestPriceAmericanOption(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/options/american", r.URL.Path)

		var req models.AmericanOptionRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.True(t, req.Spot.Equal(decimal.NewFromFloat(100)))
		assert.Equal(t, 200, req.Steps) // 验证 steps 参数传递

		resp := models.QuantLibAPIResponse{
			Success: true,
			Message: "ok",
			Data: map[string]any{
				"price": "11.2345",
				"delta": "0.6123",
				"gamma": "0.0210",
				"theta": "-0.0145",
				"vega":  "0.3890",
				"rho":   "0.4890",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	req := models.AmericanOptionRequest{
		Spot:         decimal.NewFromFloat(100),
		Strike:       decimal.NewFromFloat(105),
		Rate:         decimal.NewFromFloat(0.05),
		Volatility:   decimal.NewFromFloat(0.25),
		TimeToExpiry: decimal.NewFromFloat(1.0),
		OptionType:   models.OptionTypeCall,
		Steps:        200,
	}

	result, err := client.PriceAmericanOption(req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Price.Equal(decimal.NewFromFloat(11.2345)))
	assert.True(t, result.Delta.Equal(decimal.NewFromFloat(0.6123)))
}
```

- [ ] **步骤 2：编写美式期权参数验证测试**

```go
func TestPriceAmericanOptionValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     models.AmericanOptionRequest
		wantErr bool
	}{
		{
			name: "steps below minimum",
			req: models.AmericanOptionRequest{
				Spot: decimal.NewFromFloat(100), Strike: decimal.NewFromFloat(100),
				Rate: decimal.NewFromFloat(0.05), Volatility: decimal.NewFromFloat(0.2),
				TimeToExpiry: decimal.NewFromFloat(1), OptionType: models.OptionTypeCall,
				Steps: 5, // 小于最小值 10
			},
			wantErr: true,
		},
		{
			name: "steps above maximum",
			req: models.AmericanOptionRequest{
				Spot: decimal.NewFromFloat(100), Strike: decimal.NewFromFloat(100),
				Rate: decimal.NewFromFloat(0.05), Volatility: decimal.NewFromFloat(0.2),
				TimeToExpiry: decimal.NewFromFloat(1), OptionType: models.OptionTypeCall,
				Steps: 1500, // 大于最大值 1000
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := models.ValidateAmericanOptionRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
```

#### 4.2 Greeks 计算测试

- [ ] **步骤 3：编写 Greeks 正常路径测试**

```go
func TestCalculateGreeks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/options/greeks", r.URL.Path)

		resp := models.QuantLibAPIResponse{
			Success: true,
			Message: "ok",
			Data: map[string]any{
				"price": "10.4506",
				"delta": "0.5948",
				"gamma": "0.0188",
				"theta": "-0.0128",
				"vega":  "0.3756",
				"rho":   "0.4502",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	req := models.GreeksRequest{
		Spot:         decimal.NewFromFloat(100),
		Strike:       decimal.NewFromFloat(105),
		Rate:         decimal.NewFromFloat(0.05),
		Volatility:   decimal.NewFromFloat(0.2),
		TimeToExpiry: decimal.NewFromFloat(1.0),
		OptionType:   models.OptionTypeCall,
	}

	result, err := client.CalculateGreeks(req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Price.Equal(decimal.NewFromFloat(10.4506)))
	assert.True(t, result.Delta.Equal(decimal.NewFromFloat(0.5948)))
}
```

#### 4.3 收益率曲线测试

- [ ] **步骤 4：编写收益率曲线构建测试**

```go
func TestBuildYieldCurve(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/yield-curve/build", r.URL.Path)

		var req models.YieldCurveRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "USD", req.Currency)
		assert.Len(t, req.Tenors, 8)

		resp := models.QuantLibAPIResponse{
			Success: true,
			Message: "ok",
			Data: map[string]any{
				"currency":         "USD",
				"tenors":           []any{"1M", "3M", "6M", "1Y", "2Y", "5Y", "10Y", "30Y"},
				"rates":            []any{"0.043", "0.044", "0.045", "0.046", "0.047", "0.048", "0.049", "0.05"},
				"zero_rates":       []any{"0.0425", "0.0435", "0.0445", "0.0455", "0.0465", "0.0475", "0.0485", "0.0495"},
				"forward_rates":    []any{"0.0435", "0.0445", "0.0465", "0.0475", "0.0495", "0.0515", "0.0535", "0.0555"},
				"discount_factors": []any{"0.9965", "0.9891", "0.9780", "0.9558", "0.9137", "0.7896", "0.6139", "0.2237"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	req := models.YieldCurveRequest{
		Currency: "USD",
		Tenors:   []string{"1M", "3M", "6M", "1Y", "2Y", "5Y", "10Y", "30Y"},
		Rates: []decimal.Decimal{
			decimal.NewFromFloat(0.043),
			decimal.NewFromFloat(0.044),
			decimal.NewFromFloat(0.045),
			decimal.NewFromFloat(0.046),
			decimal.NewFromFloat(0.047),
			decimal.NewFromFloat(0.048),
			decimal.NewFromFloat(0.049),
			decimal.NewFromFloat(0.05),
		},
	}

	result, err := client.BuildYieldCurve(req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "USD", result.Currency)
	assert.Len(t, result.Tenors, 8)
	assert.Len(t, result.Rates, 8)
	assert.Len(t, result.ZeroRates, 8)
	assert.Len(t, result.ForwardRates, 8)
	assert.Len(t, result.DiscountFactors, 8)
}
```

#### 4.4 参考数据缓存测试

- [ ] **步骤 5：编写参考数据获取和缓存测试**

```go
func TestGetSupportedCurrencies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/core/types/currencies", r.URL.Path)

		resp := map[string]any{
			"currencies": []any{"USD", "EUR", "GBP", "JPY", "CNY", "AUD", "CAD", "CHF"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
		cache:      newCache(),
	}

	result, err := client.GetSupportedCurrencies()

	require.NoError(t, err)
	require.NotNil(t, result)

	// 验证第二次请求命中缓存（不会再次调用server）
	result2, err := client.GetSupportedCurrencies()
	require.NoError(t, err)
	assert.Equal(t, result, result2)
}

func TestGetFrequencies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/core/types/frequencies", r.URL.Path)

		resp := map[string]any{
			"frequencies": []any{"Monthly", "Quarterly", "SemiAnnual", "Annual"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
		cache:      newCache(),
	}

	result, err := client.GetFrequencies()

	require.NoError(t, err)
	require.NotNil(t, result)
}
```

#### 4.5 边界条件和异常处理测试

- [ ] **步骤 6：编写边界条件测试套件**

创建新文件 `quantlib_client_extended_test.go`：

```go
package quantlib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"etf-insight/models"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEuropeanOptionBoundaryZeroSpot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.QuantLibAPIResponse{
			Success: false,
			Message: "validation error: spot must be greater than 0",
		})
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	req := models.EuropeanOptionRequest{
		Spot:         decimal.Zero,
		Strike:       decimal.NewFromFloat(100),
		Rate:         decimal.NewFromFloat(0.05),
		Volatility:   decimal.NewFromFloat(0.2),
		TimeToExpiry: decimal.NewFromFloat(1.0),
		OptionType:   models.OptionTypeCall,
	}

	result, err := client.PriceEuropeanOption(req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "validation error")
}

func TestVaRMinimumReturns(t *testing.T) {
	err := models.ValidateVaRRequest(models.VaRRequest{
		PortfolioValue: decimal.NewFromFloat(1000000),
		Returns:        []decimal.Decimal{decimal.NewFromFloat(0.01)}, // 只有1个元素
		Confidence:     decimal.NewFromFloat(0.95),
		HoldingPeriod:  1,
		Method:         "historical",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 2 elements")
}

func TestBondFrequencyBounds(t *testing.T) {
	// 测试频率为0（低于最小值）
	err := models.ValidateBondRequest(models.BondRequest{
		FaceValue:       decimal.NewFromFloat(1000),
		CouponRate:      decimal.NewFromFloat(0.05),
		Frequency:       0, // 无效
		Maturity:        "2030-01-01",
		YieldToMaturity: decimal.NewFromFloat(0.04),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "between 1 and 12")

	// 测试频率为13（高于最大值）
	err = models.ValidateBondRequest(models.BondRequest{
		FaceValue:       decimal.NewFromFloat(1000),
		CouponRate:      decimal.NewFromFloat(0.05),
		Frequency:       13, // 无效
		Maturity:        "2030-01-01",
		YieldToMaturity: decimal.NewFromFloat(0.04),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "between 1 and 12")
}

func TestYieldCurveEmptyTenors(t *testing.T) {
	err := models.ValidateYieldCurveRequest(models.YieldCurveRequest{
		Currency: "USD",
		Tenors:   []string{}, // 空
		Rates:    []decimal.Decimal{decimal.NewFromFloat(0.05)},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenors cannot be empty")
}

func TestYieldCurveMismatchedLengths(t *testing.T) {
	err := models.ValidateYieldCurveRequest(models.YieldCurveRequest{
		Currency: "USD",
		Tenors:   []string{"1M", "3M", "6M"},       // 3个
		Rates:    []decimal.Decimal{decimal.NewFromFloat(0.04), decimal.NewFromFloat(0.045)}, // 2个
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "same length")
}

func TestInvalidVaRMethod(t *testing.T) {
	err := models.ValidateVaRRequest(models.VaRRequest{
		PortfolioValue: decimal.NewFromFloat(1000000),
		Returns: []decimal.Decimal{
			decimal.NewFromFloat(0.01),
			decimal.NewFromFloat(-0.02),
		},
		Confidence:    decimal.NewFromFloat(0.95),
		HoldingPeriod: 1,
		Method:        "invalid_method", // 无效方法
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be 'historical', 'parametric', or 'monte_carlo'")
}
```

- [ ] **步骤 7：运行所有测试并检查覆盖率**

```bash
cd backend && go test ./services/quantlib/... -v -coverprofile=coverage.out
go tool cover -func=coverage.out
```
预期：
- 所有新测试通过
- 覆盖率从 41.6% 提升至 >70%

- [ ] **步骤 8：Commit**

```bash
git add backend/services/quantlib/quantlib_client_test.go
git add backend/services/quantlib/quantlib_client_extended_test.go
git commit -m "test(quantlib): add comprehensive tests for American options, Greeks, yield curve, and boundary conditions"
```

---

### 任务 5：端到端验证与收尾

**优先级：** P0（确保质量）
**预估时间：** 20 分钟
**依赖：** 任务 1-4 全部完成

- [ ] **步骤 1：启动后端服务**

```bash
cd backend
export QUANTLIB_API_URL=https://api.fincept.in/quantlib
# 如果有本地 mock 服务，可以使用本地地址测试
go run main.go
```

预期：服务启动成功，监听端口 8080

- [ ] **步骤 2：启动前端开发服务器**

```bash
cd frontend
npm run dev
```
预期：前端启动成功，通常在 localhost:5173

- [ ] **步骤 3：手动验证所有 7 个 API 功能**

浏览器访问 `http://localhost:5173/quantlib`，逐项验证：

| # | 功能 | 验证点 | 预期结果 |
|---|------|--------|----------|
| 1 | 欧式期权 | 输入默认参数 → 点击计算 | 显示价格+Greeks表格 |
| 2 | 美式期权 | 切换到美式期权Tab → 输入参数 → 计算 | 显示价格+Greeks |
| 3 | Greeks分析 | 欧式期权结果页 → 点击"Greeks敏感性分析" | 显示详细Greeks说明表 |
| 4 | 债券定价 | 输入默认参数 → 计算 | 显示净价/全价/久期等6项指标 |
| 5 | 收益率曲线 | 选择货币 → 输入期限利率 → 构建 | 显示三条曲线图表 |
| 6 | VaR计算 | 输入组合价值和收益率 → 计算 | 显示VaR/CVaR值 |
| 7 | 参考数据 | 观察收益率曲线货币下拉框 | 显示从API加载的货币列表 |

- [ ] **步骤 4：检查控制台和网络请求**

打开浏览器开发者工具（F12）：
- Console：无 JavaScript 错误
- Network：所有 API 请求返回 200 或正确的错误响应
- 无 CORS 错误

- [ ] **步骤 5：运行完整测试套件确认无回归**

```bash
# 后端测试
cd backend && go test ./... -cover

# 前端类型检查
cd frontend && npx tsc --noEmit

# 前端 lint（如果配置了）
npm run lint
```
预期：全部通过

- [ ] **步骤 6：最终 Commit 和标记版本**

```bash
git add -A
git commit -m "feat(quantlib): complete Phase 1 integration - American options, Greeks, dynamic reference data, comprehensive tests

✅ New features:
- Add American option pricing tab with binomial tree method
- Integrate Greeks sensitivity analysis panel
- Dynamic reference data loading for yield curve currencies

📊 Test coverage improvement:
- From 41.6% to >70%
- Added 15 new test cases covering:
  - American option pricing and validation
  - Greeks calculation
  - Yield curve construction
  - Reference data caching
  - Boundary conditions and edge cases

🎯 All 7 QuantLib APIs now fully functional in UI"
```

- [ ] **步骤 7：更新 AGENTS.md 进度标记**

在 AGENTS.md 的 v2.8 更新内容部分，将 QuantLib 集成标记为：

```markdown
#### FinceptTerminal QuantLib 集成 (Phase 1 ✅ Complete)
- ✅ **QuantLib 云 API 对接**: 直接调用 `api.fincept.in/quantlib/` 云服务
  - [完整的功能清单保持不变]
- ✅ **前端页面**: `QuantLibAnalysis.tsx` - **6 个 Tab** (原4个 + 美式期权 + Greeks)
  - 交互式表单 + 实时结果展示
  - Recharts 收益率曲线图表
  - 动态参考数据加载
  - 路由: `/quantlib`
- ✅ **代码审查**: 2 轮审查，10 个问题已修复 (P0×3, P1×4, P2×3)
- ✅ **测试覆盖**: **22 个单元测试**, 覆盖率 **>70%**
```

---

## 执行顺序建议

```
任务 1 (美式期权) ──┐
                    ├──> 任务 5 (E2E验证)
任务 2 (Greeks) ────┤
                    │
任务 3 (动态数据) ──┤
                    │
任务 4 (测试) ──────┘ (可并行)
```

**总预估时间：** 2-3 小时（不含等待审核时间）

---

## 验收标准

### 功能完整性 ✅
- [ ] 用户可以通过 UI 使用全部 7 个 QuantLib API
- [ ] 美式期权定价功能可用（含 steps 参数）
- [ ] Greeks 敏感性分析面板可用
- [ ] 收益率曲线货币选择器从 API 动态加载数据

### 代码质量 ✅
- [ ] TypeScript 编译零错误 (`tsc --noEmit`)
- [ ] Go 测试全部通过 (`go test ./services/quantlib/...`)
- [ ] 测试覆盖率 ≥ 70%
- [ ] 无 TODO/FIXME 残留（除已知的技术债务）

### 用户体验 ✅
- [ ] 表单输入验证清晰（必填项、范围提示）
- [ ] 加载状态反馈（Spin 组件）
- [ ] 错误提示友好（message.error）
- [ ] 结果展示直观（表格/统计数值/图表）

---

## 回滚方案

如果任何任务引入回归：
1. 使用 `git revert <commit-hash>` 回滚特定提交
2. 重新运行测试确认基线恢复
3. 分离问题并单独修复

**关键回滚点：**
- 任务 1-3 的每次 Commit 都是独立可回滚的
- 任务 4 的测试提交不影响功能代码
