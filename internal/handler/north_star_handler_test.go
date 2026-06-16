package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/valyala/fasthttp"
)

// testFiberCtx cria um *fiber.Ctx funcional para testes unitários que precisam
// de um context.Context (ex: h.db.QueryRow(c.Context(), ...)).
func testFiberCtx() *fiber.Ctx {
	app := fiber.New()
	return app.AcquireCtx(&fasthttp.RequestCtx{})
}

// ─── mock pgx.Row ─────────────────────────────────────────────────────────

type mockRow struct {
	val int64
}

func (m *mockRow) Scan(dest ...any) error {
	*dest[0].(*int64) = m.val
	return nil
}

// ─── mock dbPool ───────────────────────────────────────────────────────────

// mockSequentialDB retorna valores na sequência pré-definida.
// A ordem das chamadas em brioResolutionRate é:
//   current total, current resolved,
//   prev total, prev resolved,
//   spark[7] total, spark[7] resolved, ... (x8)
type mockSequentialDB struct {
	vals []int64
	pos  int
}

func (m *mockSequentialDB) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	if m.pos >= len(m.vals) {
		return &mockRow{val: 0}
	}
	v := m.vals[m.pos]
	m.pos++
	return &mockRow{val: v}
}

// ─── Testes ────────────────────────────────────────────────────────────────

// TestBrioResolutionRate_DenominadorFiltraCorretamente verifica que apenas
// eventos de conversa são contados no denominador. Com 3 conversas e 2 resolvidas,
// a taxa deve ser 66% (não ~2% que resultaria se 100 page_views entrassem).
func TestBrioResolutionRate_DenominadorFiltraCorretamente(t *testing.T) {
	now := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)

	// 10 períodos (current + prev + 8 sparkline), cada um com total + resolved
	type call struct {
		total    int64
		resolved int64
	}

	periods := []call{
		{total: 3, resolved: 2}, // current: 3 convs, 2 resolvidas → 66%
		{total: 5, resolved: 3}, // prev: 5 convs, 3 resolvidas → 60%
		{total: 2, resolved: 1}, // spark[7]
		{total: 2, resolved: 1}, // spark[6]
		{total: 2, resolved: 1}, // spark[5]
		{total: 2, resolved: 1}, // spark[4]
		{total: 2, resolved: 1}, // spark[3]
		{total: 2, resolved: 1}, // spark[2]
		{total: 2, resolved: 1}, // spark[1]
		{total: 2, resolved: 1}, // spark[0]
	}

	expected := make([]int64, 10)
	for i, p := range periods {
		if p.total == 0 {
			expected[i] = 0
		} else {
			expected[i] = (p.resolved * 100) / p.total
		}
	}

	// Monta slice plano com 20 valores na ordem de chamada
	allVals := make([]int64, 0, 20)
	for _, p := range periods {
		allVals = append(allVals, p.total, p.resolved)
	}

	db := &mockSequentialDB{vals: allVals}
	h := &NorthStarHandler{db: db, serviceToken: ""}
	cur, prev, spark := h.brioResolutionRate(testFiberCtx(), now)

	// Verifica current
	if cur != expected[0] {
		t.Errorf("current: esperado %d, recebido %d (se 0, denominador está contando demais)", expected[0], cur)
	}

	// Verifica prev
	if prev != expected[1] {
		t.Errorf("prev: esperado %d, recebido %d", expected[1], prev)
	}

	// Verifica sparkline
	for i, v := range spark {
		// sparkline[0] corresponde a expected[2] (primeiro período após current+prev)
		if v != expected[i+2] {
			t.Errorf("sparkline[%d]: esperado %d, recebido %d", i, expected[i+2], v)
		}
	}
}

// TestBrioResolutionRate_EventosMistos_SoConversasContam verifica o cenário
// crítico: 100 page_views + 3 conversas.started existem no banco.
// O mock retorna apenas 3 no denominador (conversas) — se o denominador
// contasse page_views, a taxa seria ~2% e não 66%.
func TestBrioResolutionRate_EventosMistos_SoConversasContam(t *testing.T) {
	now := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)

	// Apenas 1 período (current) com total=3 (conversas), resolved=2
	// Os 100 page_views no banco real seriam ignorados pelo filtro
	db := &mockSequentialDB{
		vals: []int64{
			3, 2, // current: total=3, resolved=2
			3, 2, // prev: idem
			2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, // sparkline
		},
	}

	h := &NorthStarHandler{db: db, serviceToken: ""}
	cur, _, _ := h.brioResolutionRate(testFiberCtx(), now)

	// Se o denominador contasse page_views (ex: 103), cur seria (2*100)/103 = 1
	// Com denominador correto (3), cur = (2*100)/3 = 66
	if cur != 66 {
		t.Errorf("brioResolutionRate=%d, esperado 66 (denominador provavelmente contando eventos não-conversa)", cur)
	}
}

// TestBrioResolutionRate_ZeroConversas_NaoPanic retorna 0 quando não há eventos.
func TestBrioResolutionRate_ZeroConversas_NaoPanic(t *testing.T) {
	now := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)

	db := &mockSequentialDB{
		vals: make([]int64, 20), // todos zeros
	}

	h := &NorthStarHandler{db: db, serviceToken: ""}
	cur, prev, spark := h.brioResolutionRate(testFiberCtx(), now)

	if cur != 0 {
		t.Errorf("current com zero eventos: esperado 0, recebido %d", cur)
	}
	if prev != 0 {
		t.Errorf("prev com zero eventos: esperado 0, recebido %d", prev)
	}
	for i, v := range spark {
		if v != 0 {
			t.Errorf("sparkline[%d] com zero eventos: esperado 0, recebido %d", i, v)
		}
	}
}

// TestBrioResolutionRate_EndpointHTTP_Retorna200 verifica que o endpoint HTTP
// responde com 200 e inclui a métrica brio no JSON.
func TestBrioResolutionRate_EndpointHTTP_Retorna200(t *testing.T) {
	db := &mockSequentialDB{
		vals: make([]int64, 20),
	}
	h := &NorthStarHandler{db: db, serviceToken: "test-token"}

	// Usa RegisterRoutes para testar via HTTP
	app := fiber.New()
	h.RegisterRoutes(app)

	req := httptest.NewRequest("GET", "/api/v1/north-star", nil)
	req.Header.Set("X-Service-Token", "test-token")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("erro ao chamar endpoint: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("north-star deveria retornar 200, recebeu %d", resp.StatusCode)
	}

	var body struct {
		Metrics []map[string]any `json:"metrics"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("erro ao decodificar JSON: %v", err)
	}
	foundBrio := false
	for _, m := range body.Metrics {
		if m["product"] == "brio" {
			foundBrio = true
			break
		}
	}
	if !foundBrio {
		t.Error("resposta não contém métrica do produto brio")
	}
}
