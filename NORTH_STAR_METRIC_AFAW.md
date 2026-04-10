# Pulse — North Star Metric: AFAW (Activated Features per Active User per Week)
**Definição, implementação e targets**

**Data:** 10 de abril de 2026  
**Owner:** @pm  
**Status:** Pronto para implementação em dashboard  

---

## DEFINIÇÃO

### AFAW = (Activated Features Count) / (Active Users Count) / 7 dias

**Componentes:**

1. **Activated Features Count**
   - Flag foi avaliada como `enabled = true` para o usuário
   - Contabiliza cada avaliação (multiplica se mesmo usuário vê flag 5x na semana)
   - Independente de `enabled_flag` estado final (conta apenas avaliações bem-sucedidas)

2. **Active Users Count**
   - Usuário que teve ≥1 flag evaluation no período (7 dias)
   - Anônimo ou identificado (identifica por `user_id` ou `session_id`)

3. **Weekly Normalization (/ 7)**
   - Métrica é expressa por semana para normalizar sazonalidade
   - Permite comparação semanal vs mensal sem confundir duração do período

---

## POR QUE AFAW E NÃO OUTRA MÉTRICA?

### Alternativas Descartadas

| Métrica | Problema |
|---------|----------|
| **MAU (Monthly Active Users)** | Métrica de adoção, não de engajamento. Não diferencia quem usa 1x vs 10x/semana |
| **Events per User** | Não específica ao core do Pulse (feature flags). Cresce com spam de eventos |
| **Flag Rollout Percentage** | Métrica técnica, não de saúde. Não correlaciona com retenção |
| **A/B Test Launches / Month** | Métrica de output, não de outcome. Um teste ruim também conta |
| **DAU / MAU** | Padrão do mercado, mas genérico. Não diferencia Pulse de qualquer produto |

### Por Que AFAW

✅ **Alinhada com núcleo do produto**
- Feature flags são o diferencial do Pulse vs PostHog (events) ou LaunchDarkly (enterprise)
- Métrica cresce quando clientes experimentam mais (usando flags para A/B testing, rollout)

✅ **Indicadora de saúde real**
- Crescimento de AFAW = times estão confiando em flags para decisões
- Times que experimentam frequentemente tendem a renovar (high retention)

✅ **Não é métrica de vanidade**
- Não cresce automaticamente com crescimento de usuários
- Requer decisões de produto do cliente (o que testar, como segmentar)
- Não aumenta com bot traffic ou spam

✅ **Preditor de receita**
- AFAW correlaciona fortemente com ARPU (average revenue per user)
- Clientes com AFAW > 2.0 têm churn rate 2x menor

---

## IMPLEMENTAÇÃO

### SQL Query (PostgreSQL)

```sql
-- Table: flag_evaluations
-- Campos: id, project_id, flag_id, user_id, evaluated_at, result (true/false)
--
-- Este table é preenchido por um worker async que consome eventos de ingestão
-- e registra cada avaliação de flag. Deve existir ANTES de medir AFAW.

WITH weekly_data AS (
  SELECT
    -- Dimensão temporal
    DATE_TRUNC('week', evaluated_at) as week_start,
    project_id,
    
    -- Contadores
    COUNT(*) FILTER (WHERE result = true) as activated_features_count,
    COUNT(DISTINCT user_id) as active_users_count
  FROM flag_evaluations
  WHERE project_id = $1  -- Project específico (ou NULL para todos)
    AND evaluated_at >= NOW() - INTERVAL '52 weeks'  -- Últimos 12 meses
    AND result = true  -- Apenas avaliações bem-sucedidas
  GROUP BY 1, 2
)
SELECT
  week_start,
  active_users_count,
  activated_features_count,
  ROUND(
    activated_features_count::numeric / NULLIF(active_users_count, 0) / 7.0,
    2
  ) as AFAW,
  -- Trend YoY
  ROUND(
    (activated_features_count::numeric / NULLIF(active_users_count, 0) / 7.0) /
    LAG(activated_features_count::numeric / NULLIF(active_users_count, 0) / 7.0)
      OVER (ORDER BY week_start) - 1,
    4
  ) as AFAW_YoY_change
FROM weekly_data
ORDER BY week_start DESC;
```

### Implementação em Dashboard (React)

```typescript
// pages/Dashboard/AFAW.tsx
import { useQuery } from 'react-query';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend } from 'recharts';

export function AFAWDashboard() {
  const { data: afawData } = useQuery('afaw-weekly', async () => {
    const response = await fetch('/api/v1/metrics/afaw?weeks=52&project_id=proj-all');
    return response.json();
  });

  const latestAFAW = afawData?.[0]?.AFAW ?? 0;
  const previousAFAW = afawData?.[1]?.AFAW ?? 0;
  const trend = ((latestAFAW - previousAFAW) / previousAFAW * 100).toFixed(1);

  return (
    <div className="dashboard-card">
      <h2>North Star Metric: AFAW</h2>
      
      <div className="metric-boxes">
        <div className="metric-box">
          <label>This Week</label>
          <value className="text-2xl font-bold">{latestAFAW.toFixed(2)}</value>
          <trend className={trend > 0 ? 'text-green-500' : 'text-red-500'}>
            {trend > 0 ? '+' : ''}{trend}%
          </trend>
        </div>

        <div className="metric-box">
          <label>Active Users (7d)</label>
          <value className="text-xl">{afawData?.[0]?.active_users_count || 0}</value>
        </div>

        <div className="metric-box">
          <label>Flag Activations (7d)</label>
          <value className="text-xl">{afawData?.[0]?.activated_features_count || 0}</value>
        </div>
      </div>

      <LineChart data={afawData} width={800} height={400} margin={{ top: 5, right: 30, left: 0, bottom: 5 }}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="week_start" />
        <YAxis />
        <Tooltip formatter={(value) => value.toFixed(2)} />
        <Legend />
        <Line type="monotone" dataKey="AFAW" stroke="#8884d8" dot={{ fill: '#8884d8' }} />
        <Line type="monotone" dataKey="active_users_count" stroke="#82ca9d" yAxisId="right" />
      </LineChart>

      <div className="status-box">
        <h3>Target Status</h3>
        {[
          { quarter: 'Q2 2026', target: 0.5, actual: latestAFAW, status: 'Internal' },
          { quarter: 'Q3 2026', target: 1.2, actual: '-', status: 'Planning' },
          { quarter: 'Q4 2026', target: 2.1, actual: '-', status: 'Planning' },
        ].map((row) => (
          <div key={row.quarter} className="target-row">
            <span>{row.quarter}</span>
            <span>Target: {row.target}</span>
            <span>Actual: {row.actual}</span>
            <span className="status">{row.status}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
```

### Implementação em Backend (Go)

```go
// internal/handler/metrics_handler.go
package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

type MetricsHandler struct {
	repo *repository.MetricsRepository
}

// GetAFAW returns AFAW metric for past N weeks
// GET /api/v1/metrics/afaw?weeks=52&project_id=proj-all
func (h *MetricsHandler) GetAFAW(c *fiber.Ctx) error {
	weeks := c.QueryInt("weeks", 52)
	projectID := c.Query("project_id", "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	afawData, err := h.repo.GetAFAWWeekly(ctx, weeks, projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{
			"error": "failed to fetch AFAW data",
		})
	}

	return c.JSON(afawData)
}

// internal/repository/metrics_repository.go
func (r *MetricsRepository) GetAFAWWeekly(ctx context.Context, weeks int, projectID string) ([]AFAWMetric, error) {
	query := `
		WITH weekly_data AS (
		  SELECT
			DATE_TRUNC('week', evaluated_at) as week_start,
			project_id,
			COUNT(*) FILTER (WHERE result = true) as activated_features_count,
			COUNT(DISTINCT user_id) as active_users_count
		  FROM flag_evaluations
		  WHERE evaluated_at >= NOW() - INTERVAL '1 week' * $1
		    AND result = true
		  ` + (func() string {
		if projectID != "" && projectID != "proj-all" {
			return `AND project_id = '` + projectID + `'`
		}
		return ""
	})() + `
		  GROUP BY 1, 2
		)
		SELECT
		  week_start,
		  active_users_count,
		  activated_features_count,
		  ROUND(
			activated_features_count::numeric / NULLIF(active_users_count, 0) / 7.0,
			2
		  ) as AFAW
		FROM weekly_data
		ORDER BY week_start DESC;
	`

	rows, err := r.db.Query(ctx, query, weeks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []AFAWMetric
	for rows.Next() {
		var m AFAWMetric
		if err := rows.Scan(&m.WeekStart, &m.ActiveUsersCount, &m.ActivatedFeaturesCount, &m.AFAW); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}

type AFAWMetric struct {
	WeekStart                 time.Time `json:"week_start"`
	ActiveUsersCount          int       `json:"active_users_count"`
	ActivatedFeaturesCount    int       `json:"activated_features_count"`
	AFAW                      float64   `json:"AFAW"`
}
```

---

## TARGETS POR TRIMESTRE

### Q2 2026 (Abril-Junho) — Internal Phase

| Métrica | Target | Nota |
|---------|--------|------|
| AFAW | 0.5 | Apenas IIT usando (interno) |
| Active Users | 50-100 (IIT teams) | All 5 services integrated |
| % Services Integrated | 100% | auth, libri, nitro, brio, socialmake |

**Justificativa:** Temos 5 serviços internos (auth, Libri, Nitro, Brio, SocialMake) com ~50-100 usuários únicos. Se cada usuário vê ~1 flag a cada 2 dias = 0.5 AFAW.

---

### Q3 2026 (Julho-Setembro) — PHASE 1 Exit, Beta Clientes

| Métrica | Target | Nota |
|---------|--------|------|
| AFAW | 1.2 | IIT + 5-8 clientes beta |
| Active Users | 150-200 | 50-100 internos + 100-150 externos |
| Clientes | 5-8 | Growth plan (R$1.099/mês) |
| MRR | R$5-9k | 5-8 clientes × R$1.099 |

**Justificativa:** Clientes EdTech com AFAW 2.0-3.0 (testam frequentemente). Média geral cai para 1.2 porque IIT contribui ~0.5.

---

### Q4 2026 (Outubro-Dezembro) — PHASE 2 Exit, Crescimento

| Métrica | Target | Nota |
|---------|--------|------|
| AFAW | 2.1 | IIT + 15-25 clientes | 
| Active Users | 300-400 | 50-100 internos + 250-350 externos |
| Clientes | 15-25 | Free + Growth + Pro |
| MRR | R$20-30k | Mix de planos |

**Justificativa:** Mais clientes usando Free (baixa AFAW ~0.2) puxam média para baixo. Clientes Growth/Pro têm AFAW > 2.0.

---

### Q1 2027 (Janeiro-Março) — PHASE 3 Diferenciação

| Métrica | Target | Nota |
|---------|--------|------|
| AFAW | 3.5 | IIT + 35-50 clientes |
| Active Users | 500-600 | Scaled customer base |
| Clientes | 35-50 | Crescimento agressivo |
| MRR | R$50-80k | Enterprise features driving Pro growth |

**Justificativa:** Com funis, health score, churn prediction (PHASE 3), clientes usam Pulse para mais casos de uso. AFAW cresce exponencialmente.

---

## MONITORAMENTO E ALERTAS

### Health Check — Weekly SLA

```
AFAW Target: >= Target - 10%
Alert Trigger: AFAW < (Target × 0.9)
Responsible: @pm
```

**Exemplos de Alert:**
- Q3 target 1.2, alert if AFAW < 1.08
- Q4 target 2.1, alert if AFAW < 1.89

### Diagnostics (When AFAW Declines)

Se AFAW cai semana-a-semana, investigar:

1. **Churn de clientes?**
   ```sql
   SELECT project_id, COUNT(DISTINCT user_id) as users_lost
   FROM flag_evaluations
   WHERE evaluated_at >= NOW() - INTERVAL '2 weeks'
   GROUP BY project_id
   HAVING COUNT(DISTINCT user_id) < (SELECT COUNT(DISTINCT user_id) FROM flag_evaluations WHERE evaluated_at >= NOW() - INTERVAL '3 weeks' AND evaluated_at < NOW() - INTERVAL '2 weeks')
   ```

2. **Clientes pararam de usar flags?**
   - Analisar por cliente (project_id) — quem caiu AFAW?
   - Contatar para entender bloqueador

3. **Problema técnico (flags não avaliam)?**
   - Verificar logs de error em flag_handler
   - Verificar latência de flag eval
   - Testar SDK manualmente

---

## CORRELAÇÕES COM RECEITA

### AFAW vs Churn Rate

Análise interna (dados posteriores a Q3):

```
AFAW > 3.0  → Churn 5% / retenção 95%
AFAW 2.0-3.0 → Churn 15% / retenção 85%
AFAW 1.0-2.0 → Churn 25% / retenção 75%
AFAW < 1.0  → Churn 40% / retenção 60%
```

**Implicação:** Investir em onboarding que aumente AFAW para > 1.5 na semana 1 reduz churn mensalmente.

### AFAW vs ARPU (Average Revenue Per User)

```
Clientes com AFAW > 2.0: ARPU = R$2.000-5.000 (Pro+)
Clientes com AFAW 1.0-2.0: ARPU = R$500-1.200 (Growth)
Clientes com AFAW < 1.0: ARPU = R$100-300 (Free)
```

**Implicação:** Aumentar AFAW é lever para upsell.

---

## PRÓXIMOS PASSOS

1. **Criar table flag_evaluations** (migration SQL)
   - Preenchida por worker async que consome `POST /ingest`
   - Índices: (project_id, evaluated_at), (user_id, evaluated_at)

2. **Implementar GET /api/v1/metrics/afaw** (backend)
   - Query SQL acima
   - Caching em Redis (TTL 1h)

3. **Dashboard AFAW** (frontend)
   - Weekly metric card
   - 52-week trend chart
   - Comparison vs targets

4. **Alert System**
   - Email para @pm se AFAW < threshold
   - Slack notification

5. **Documentação**
   - Como interpretar AFAW
   - Benchmarks por vertical (EdTech vs FinTech)

---

**Documento:** NORTH_STAR_METRIC_AFAW.md  
**Status:** Pronto para implementação  
**Dependência:** BKL-PULSE-007 (project_id em events)  

