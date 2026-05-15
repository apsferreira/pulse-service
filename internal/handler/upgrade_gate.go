package handler

import "github.com/gofiber/fiber/v2"

// UpgradeGateResponse é a resposta padronizada retornada quando um recurso não está disponível no plano atual.
// O frontend usa o campo code=="upgrade_required" para exibir o modal de upgrade.
type UpgradeGateResponse struct {
	Error      string `json:"error"`
	Code       string `json:"code"`        // sempre "upgrade_required"
	Plan       string `json:"plan"`        // plano atual do projeto
	UpgradeURL string `json:"upgrade_url"` // URL direta para o checkout do plano requerido
	Message    string `json:"message"`     // mensagem amigável com preço e benefícios
}

var upgradeURLs = map[string]string{
	"starter": "https://checkout-api.institutoitinerante.com.br/checkout?plan=pulse-starter",
	"growth":  "https://checkout-api.institutoitinerante.com.br/checkout?plan=pulse-growth",
	"pro":     "https://checkout-api.institutoitinerante.com.br/checkout?plan=pulse-scale",
}

var upgradeMessages = map[string]string{
	"starter": "Disponível no plano Starter (R$399/mês). Comece com 14 dias grátis.",
	"growth":  "Disponível no plano Growth (R$1.099/mês). Inclui A/B testing e Health Score.",
	"pro":     "Disponível no plano Scale (R$2.399/mês). Projetos e eventos ilimitados.",
}

// upgradeGate retorna HTTP 402 com payload padronizado para o frontend exibir o modal de upgrade.
// currentPlan é o plano atual do projeto; requiredPlan é o plano mínimo para desbloquear o recurso.
func upgradeGate(c *fiber.Ctx, currentPlan, requiredPlan string) error {
	return c.Status(fiber.StatusPaymentRequired).JSON(UpgradeGateResponse{
		Error:      "Recurso não disponível no seu plano atual",
		Code:       "upgrade_required",
		Plan:       currentPlan,
		UpgradeURL: upgradeURLs[requiredPlan],
		Message:    upgradeMessages[requiredPlan],
	})
}

// isPaidPlan retorna true se o plano é pago (starter ou superior).
func isPaidPlan(plan string) bool {
	switch plan {
	case "starter", "growth", "pro", "enterprise", "internal":
		return true
	}
	return false
}

// isUnlimitedPlan retorna true para planos sem gating de quota ou features.
// enterprise e internal têm acesso irrestrito — self-hosted ou contrato especial.
func isUnlimitedPlan(plan string) bool {
	return plan == "pro" || plan == "enterprise" || plan == "internal"
}
