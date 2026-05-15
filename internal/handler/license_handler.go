package handler

// license_handler.go — endpoints de geração e validação de licenças self-hosted.
//
// POST /api/v1/license/validate  — sem autenticação (chamado pelo Pulse self-hosted na init)
// POST /api/v1/license/generate  — requer X-Service-Token (chamado pelo webhook do checkout)

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/institutoitinerante/pulse-service/internal/middleware"
	"github.com/institutoitinerante/pulse-service/internal/repository"
)

// LicenseHandler gerencia endpoints de licença self-hosted.
type LicenseHandler struct {
	repo         *repository.LicenseRepository
	serviceToken string
}

// NewLicenseHandler cria um novo LicenseHandler.
func NewLicenseHandler(repo *repository.LicenseRepository, serviceToken string) *LicenseHandler {
	return &LicenseHandler{repo: repo, serviceToken: serviceToken}
}

// RegisterRoutes registra as rotas de licença.
func (h *LicenseHandler) RegisterRoutes(app *fiber.App) {
	// Validação não requer auth — chamada pelo Pulse self-hosted na inicialização
	app.Post("/api/v1/license/validate", h.Validate)
	// Geração requer service token — chamada pelo webhook de checkout
	app.Post("/api/v1/license/generate",
		middleware.ServiceTokenMiddleware(h.serviceToken),
		h.Generate,
	)
}

// validateLicenseRequest é o payload de validação.
type validateLicenseRequest struct {
	Key      string `json:"key"`
	Hostname string `json:"hostname"`
}

// generateLicenseRequest é o payload de geração.
type generateLicenseRequest struct {
	OrderID       string `json:"order_id"`
	CustomerEmail string `json:"customer_email"`
	CustomerName  string `json:"customer_name"`
	ValidDays     int    `json:"valid_days"`
}

// Validate godoc
// POST /api/v1/license/validate — sem autenticação.
// Chamado pelo Pulse self-hosted na inicialização para verificar se a licença é válida.
func (h *LicenseHandler) Validate(c *fiber.Ctx) error {
	var req validateLicenseRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload inválido"})
	}
	if req.Key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "key é obrigatória"})
	}

	lic, err := h.repo.GetByKey(c.Context(), req.Key)
	if err != nil {
		log.Printf("[license] GetByKey error key=%s: %v", req.Key, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro interno"})
	}
	if lic == nil {
		return c.Status(fiber.StatusOK).JSON(domain.LicenseValidateResponse{
			Valid:   false,
			Message: "Licença não encontrada",
		})
	}

	if lic.RevokedAt != nil {
		return c.Status(fiber.StatusOK).JSON(domain.LicenseValidateResponse{
			Valid:   false,
			Message: "Licença revogada",
		})
	}

	now := time.Now().UTC()
	if lic.ValidUntil.Before(now) {
		return c.Status(fiber.StatusOK).JSON(domain.LicenseValidateResponse{
			Valid:      false,
			ValidUntil: lic.ValidUntil,
			DaysLeft:   0,
			Message:    "Licença expirada",
		})
	}

	// Licença válida — ativa (idempotente) e retorna dados
	if req.Hostname != "" {
		if actErr := h.repo.Activate(c.Context(), req.Key, req.Hostname); actErr != nil {
			log.Printf("[license] Activate key=%s hostname=%s: %v", req.Key, req.Hostname, actErr)
			// Não é fatal — continua respondendo como válida
		}
	}

	daysLeft := int(math.Ceil(lic.ValidUntil.Sub(now).Hours() / 24))

	return c.Status(fiber.StatusOK).JSON(domain.LicenseValidateResponse{
		Valid:      true,
		Plan:       lic.Plan,
		ValidUntil: lic.ValidUntil,
		DaysLeft:   daysLeft,
	})
}

// Generate godoc
// POST /api/v1/license/generate — requer X-Service-Token.
// Gera uma nova licença self-hosted após confirmação de pagamento.
func (h *LicenseHandler) Generate(c *fiber.Ctx) error {
	var req generateLicenseRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload inválido"})
	}
	if req.CustomerEmail == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "customer_email é obrigatório"})
	}
	if req.OrderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "order_id é obrigatório"})
	}
	if req.ValidDays <= 0 {
		req.ValidDays = 365 // padrão: 1 ano
	}

	key, err := generateLicenseKey()
	if err != nil {
		log.Printf("[license] Generate: generateLicenseKey error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao gerar chave"})
	}

	lic := domain.LicenseKey{
		Key:           key,
		Plan:          "self-hosted",
		CustomerEmail: req.CustomerEmail,
		CustomerName:  req.CustomerName,
		OrderID:       req.OrderID,
		ValidUntil:    time.Now().UTC().AddDate(0, 0, req.ValidDays),
	}

	created, err := h.repo.Create(c.Context(), lic)
	if err != nil {
		log.Printf("[license] Generate: Create error order=%s: %v", req.OrderID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "erro ao salvar licença"})
	}

	log.Printf("[license] gerada order_id=%s plan=%s valid_until=%s", req.OrderID, created.Plan, created.ValidUntil.Format(time.RFC3339))

	return c.Status(fiber.StatusCreated).JSON(created)
}

// GenerateForOrder cria uma licença self-hosted a partir de um pedido do checkout.
// Método auxiliar chamado internamente pelo checkout_webhook_handler.
func (h *LicenseHandler) GenerateForOrder(ctx *fiber.Ctx, orderID, customerEmail string) error {
	key, err := generateLicenseKey()
	if err != nil {
		return fmt.Errorf("license: GenerateForOrder: %w", err)
	}

	lic := domain.LicenseKey{
		Key:           key,
		Plan:          "self-hosted",
		CustomerEmail: customerEmail,
		OrderID:       orderID,
		ValidUntil:    time.Now().UTC().AddDate(1, 0, 0), // 1 ano
	}

	created, err := h.repo.Create(ctx.Context(), lic)
	if err != nil {
		return fmt.Errorf("license: GenerateForOrder: %w", err)
	}

	log.Printf("[license] gerada via webhook order_id=%s key=%s valid_until=%s",
		orderID, created.Key, created.ValidUntil.Format(time.RFC3339))
	return nil
}

// generateLicenseKey gera uma chave no formato pulse_lic_ + 40 chars hex.
func generateLicenseKey() (string, error) {
	b := make([]byte, 20) // 20 bytes = 40 chars hex
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "pulse_lic_" + hex.EncodeToString(b), nil
}
