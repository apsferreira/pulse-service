package handler

// checkout_webhook_handler.go — recebe confirmação de pagamento do checkout-service
// e atualiza o plano do projeto Pulse correspondente.
// Rota: POST /api/v1/webhooks/checkout (protegida por X-Service-Token)

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

// CheckoutWebhookHandler processa confirmações de pagamento vindas do checkout-service.
type CheckoutWebhookHandler struct {
	projectRepo    *repository.ProjectRepository
	licenseHandler *LicenseHandler
	serviceToken   string
}

func NewCheckoutWebhookHandler(projectRepo *repository.ProjectRepository, licenseHandler *LicenseHandler, serviceToken string) *CheckoutWebhookHandler {
	return &CheckoutWebhookHandler{
		projectRepo:    projectRepo,
		licenseHandler: licenseHandler,
		serviceToken:   serviceToken,
	}
}

// RegisterRoutes registra a rota de webhook de checkout.
func (h *CheckoutWebhookHandler) RegisterRoutes(app *fiber.App) {
	app.Post("/api/v1/webhooks/checkout",
		middleware.ServiceTokenMiddleware(h.serviceToken),
		h.HandleCheckoutConfirmation,
	)
}

// checkoutPayload é o payload esperado do checkout-service.
type checkoutPayload struct {
	Event         string `json:"event"`
	OrderID       string `json:"order_id"`
	ProductSlug   string `json:"product_slug"`
	PlanSlug      string `json:"plan_slug"`
	ProjectID     string `json:"project_id"`
	UserID        string `json:"user_id"`
	CustomerEmail string `json:"customer_email"`
}

// planSlugToInternal mapeia slugs do checkout para planos internos do Pulse.
var planSlugToInternal = map[string]string{
	"pulse-starter":     "starter",
	"pulse-growth":      "growth",
	"pulse-scale":       "pro",
	"pulse-self-hosted": "pro",
}

// HandleCheckoutConfirmation processa o evento payment.confirmed do checkout-service.
// POST /api/v1/webhooks/checkout
func (h *CheckoutWebhookHandler) HandleCheckoutConfirmation(c *fiber.Ctx) error {
	var payload checkoutPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload inválido"})
	}

	if payload.Event != "payment.confirmed" {
		// Evento não tratado — retorna 200 para evitar reenvio desnecessário
		return c.JSON(fiber.Map{"ok": true, "skipped": true})
	}

	if payload.ProjectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project_id obrigatório"})
	}
	if payload.OrderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "order_id obrigatório"})
	}

	newPlan, ok := planSlugToInternal[payload.PlanSlug]
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "plan_slug não reconhecido para o Pulse: " + payload.PlanSlug,
		})
	}

	if err := h.projectRepo.UpdatePlan(c.Context(), payload.ProjectID, newPlan, payload.OrderID); err != nil {
		log.Printf("[checkout_webhook] UpdatePlan project_id=%s plan=%s order=%s: %v",
			payload.ProjectID, newPlan, payload.OrderID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "falha ao atualizar plano"})
	}

	log.Printf("[checkout_webhook] plano atualizado project_id=%s plan=%s order=%s",
		payload.ProjectID, newPlan, payload.OrderID)

	// Para planos self-hosted, gera a licença de instalação
	if payload.PlanSlug == "pulse-self-hosted" && h.licenseHandler != nil {
		email := payload.CustomerEmail
		if email == "" {
			log.Printf("[checkout_webhook] self-hosted sem customer_email order=%s — licença não gerada", payload.OrderID)
		} else {
			if licErr := h.licenseHandler.GenerateForOrder(c, payload.OrderID, email); licErr != nil {
				// Não é fatal — o plano já foi atualizado; licença pode ser gerada manualmente
				log.Printf("[checkout_webhook] GenerateForOrder order=%s: %v", payload.OrderID, licErr)
			}
		}
	}

	return c.JSON(fiber.Map{"ok": true})
}
