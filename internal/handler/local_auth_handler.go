package handler

// LocalAuthHandler implementa login/logout/me para auth local (PULSE_AUTH_MODE=local).
// Usado em instâncias self-hosted sem auth-service externo.
//
// Endpoints:
//
//	POST /api/v1/auth/login   — email+senha, retorna JWT
//	POST /api/v1/auth/logout  — stateless (client descarta token)
//	GET  /api/v1/auth/me      — retorna usuário autenticado (Bearer token)
//	POST /api/v1/auth/setup   — cria primeiro usuário admin (só quando tabela vazia)
//
// Hashing de senha: HMAC-SHA256(senha, salt) com salt aleatório de 32 bytes.
// O hash armazenado tem formato "salt_hex:hmac_hex".
// Nota: para maior segurança em produção, considere adicionar golang.org/x/crypto
// para usar bcrypt — requer apenas `go get golang.org/x/crypto` e atualizar go.mod.

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/institutoitinerante/pulse-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LocalAuthHandler lida com autenticação local por email+senha.
type LocalAuthHandler struct {
	db        *pgxpool.Pool
	jwtSecret string
}

// NewLocalAuthHandler cria um LocalAuthHandler.
// jwtSecret deve ser não-vazio; o config.Load() já aplica o fallback para ServiceToken.
func NewLocalAuthHandler(db *pgxpool.Pool, jwtSecret string) *LocalAuthHandler {
	return &LocalAuthHandler{db: db, jwtSecret: jwtSecret}
}

// RegisterRoutes registra as rotas de auth local no app Fiber.
func (h *LocalAuthHandler) RegisterRoutes(app *fiber.App) {
	app.Post("/api/v1/auth/login", h.Login)
	app.Post("/api/v1/auth/logout", h.Logout)
	app.Get("/api/v1/auth/me", h.Me)
	app.Post("/api/v1/auth/setup", h.Setup)
}

// loginRequest é o body esperado em POST /api/v1/auth/login.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// setupRequest é o body esperado em POST /api/v1/auth/setup.
type setupRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// Login godoc
// POST /api/v1/auth/login
// Verifica email+senha contra a tabela pulse_users e retorna JWT (exp 24h).
func (h *LocalAuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "body inválido"})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "email e password são obrigatórios"})
	}
	if len(req.Email) > 255 || len(req.Password) > 255 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "email ou password excedem 255 caracteres"})
	}

	// Busca usuário por email — prepared statement (R1)
	var (
		id           string
		name         string
		passwordHash string
		role         string
	)
	err := h.db.QueryRow(
		c.UserContext(),
		`SELECT id, name, password_hash, role FROM pulse_users WHERE email = $1`,
		req.Email,
	).Scan(&id, &name, &passwordHash, &role)
	if err != nil {
		// Resposta genérica para não vazar se o email existe (timing attack mitigation)
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{Error: "credenciais inválidas"})
	}

	if !verifyPasswordHash(req.Password, passwordHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{Error: "credenciais inválidas"})
	}

	token, err := issueJWT(id, req.Email, name, role, h.jwtSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao gerar token"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":    id,
			"email": req.Email,
			"name":  name,
			"role":  role,
		},
	})
}

// Logout godoc
// POST /api/v1/auth/logout
// Stateless — o client deve descartar o token do lado dele.
func (h *LocalAuthHandler) Logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "logout realizado"})
}

// Me godoc
// GET /api/v1/auth/me
// Valida o Bearer JWT e retorna os dados do usuário autenticado.
func (h *LocalAuthHandler) Me(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{Error: "Authorization Bearer obrigatório"})
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	claims, ok := parseJWT(tokenStr, h.jwtSecret)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(domain.ErrorResponse{Error: "token inválido ou expirado"})
	}

	return c.JSON(fiber.Map{
		"id":    claims["sub"],
		"email": claims["email"],
		"name":  claims["name"],
		"role":  claims["role"],
	})
}

// Setup godoc
// POST /api/v1/auth/setup
// Cria o primeiro usuário admin. Retorna 403 se já existir qualquer usuário.
// Não requer autenticação — proteção via verificação de tabela vazia.
func (h *LocalAuthHandler) Setup(c *fiber.Ctx) error {
	var req setupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "body inválido"})
	}

	// Validação de input (R3)
	if req.Email == "" || req.Name == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "email, name e password são obrigatórios"})
	}
	if len(req.Email) > 255 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "email excede 255 caracteres"})
	}
	if len(req.Name) > 255 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "name excede 255 caracteres"})
	}
	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "password deve ter no mínimo 8 caracteres"})
	}
	if len(req.Password) > 255 {
		return c.Status(fiber.StatusBadRequest).JSON(domain.ErrorResponse{Error: "password excede 255 caracteres"})
	}

	// Verifica se já existe algum usuário — prepared statement (R1)
	var count int
	if err := h.db.QueryRow(
		c.UserContext(),
		`SELECT COUNT(*) FROM pulse_users`,
	).Scan(&count); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao verificar usuários"})
	}
	if count > 0 {
		return c.Status(fiber.StatusForbidden).JSON(domain.ErrorResponse{
			Error: "setup já realizado — use login para acessar",
		})
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao processar senha"})
	}

	var id string
	if err := h.db.QueryRow(
		c.UserContext(),
		`INSERT INTO pulse_users (email, name, password_hash, role)
		 VALUES ($1, $2, $3, 'admin')
		 RETURNING id`,
		req.Email, req.Name, hash,
	).Scan(&id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(domain.ErrorResponse{Error: "erro ao criar usuário"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":      id,
		"email":   req.Email,
		"name":    req.Name,
		"role":    "admin",
		"message": "usuário admin criado com sucesso",
	})
}

// --- helpers de hashing ---

// hashPassword gera "saltHex:hmacHex" usando HMAC-SHA256 com salt aleatório de 32 bytes.
// O salt é único por senha, evitando ataques de rainbow table.
func hashPassword(password string) (string, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	saltHex := hex.EncodeToString(salt)
	mac := hmac.New(sha256.New, salt)
	mac.Write([]byte(password))
	hashHex := hex.EncodeToString(mac.Sum(nil))
	return saltHex + ":" + hashHex, nil
}

// verifyPasswordHash verifica se password bate com o hash armazenado.
// Usa comparação em tempo constante para evitar timing attacks.
func verifyPasswordHash(password, stored string) bool {
	parts := strings.SplitN(stored, ":", 2)
	if len(parts) != 2 {
		return false
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, salt)
	mac.Write([]byte(password))
	expected := hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(expected), []byte(parts[1])) == 1
}

// --- helpers de JWT ---

// issueJWT gera um JWT HS256 com claims padrão e expiração de 24h.
// Implementação sem dependências externas, igual à usada em middleware/auth.go.
func issueJWT(id, email, name, role, secret string) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	exp := time.Now().Add(24 * time.Hour).Unix()
	payloadMap := map[string]interface{}{
		"sub":   id,
		"email": email,
		"name":  name,
		"role":  role,
		"exp":   exp,
		"iat":   time.Now().Unix(),
	}
	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	signingInput := header + "." + payload
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + sig, nil
}

// parseJWT valida a assinatura e a expiração de um JWT HS256, retornando os claims.
// Implementação sem dependências externas, coerente com middleware/auth.go.
func parseJWT(tokenStr, secret string) (map[string]interface{}, bool) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, false
	}

	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expectedSig), []byte(parts[2])) {
		return nil, false
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, false
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, false
	}

	// Verifica expiração
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, false
		}
	}

	return claims, true
}

