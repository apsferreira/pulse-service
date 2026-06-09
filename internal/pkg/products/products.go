// Package products centraliza a normalização e validação do campo `product`
// usado em todos os eventos do pulse-service.
//
// fix(#64): SDKs podiam enviar 'Nitro' (maiúscula) enquanto filtros SQL
// consultam `WHERE product = 'nitro'` (lowercase). Resultado: sparklines e
// métricas voltavam zeradas silenciosamente. A normalização passa a ocorrer
// na ingestão (lowercase + trim) e produtos fora da allowlist geram alerta.
package products

import (
	"log"
	"os"
	"strings"
	"sync"
)

// defaultKnownProducts é o conjunto canônico de produtos do ecossistema IIT.
// Pode ser sobrescrito via PULSE_KNOWN_PRODUCTS (CSV) para self-hosted.
var defaultKnownProducts = []string{
	"libri",
	"nitro",
	"brio",
	"socialmake",
	"storemake",
	"pulse",
	"default",
}

var (
	knownMu      sync.RWMutex
	knownLoaded  bool
	knownProduct map[string]bool
)

func loadKnown() {
	knownMu.Lock()
	defer knownMu.Unlock()
	if knownLoaded {
		return
	}
	knownProduct = make(map[string]bool)
	list := defaultKnownProducts
	if env := os.Getenv("PULSE_KNOWN_PRODUCTS"); env != "" {
		list = strings.Split(env, ",")
	}
	for _, p := range list {
		n := Normalize(p)
		if n != "" {
			knownProduct[n] = true
		}
	}
	knownLoaded = true
}

// Normalize aplica trim + lowercase ao product. Garante consistência com
// filtros SQL (WHERE product = 'nitro') independente da capitalização enviada
// pelo SDK.
func Normalize(product string) string {
	return strings.ToLower(strings.TrimSpace(product))
}

// IsKnown retorna true se product (já normalizado) está na allowlist
// configurada. Vazio retorna false.
func IsKnown(product string) bool {
	if product == "" {
		return false
	}
	loadKnown()
	knownMu.RLock()
	defer knownMu.RUnlock()
	return knownProduct[product]
}

// NormalizeAndCheck normaliza o product e emite log de alerta se ele não
// pertencer à allowlist. Não rejeita o evento — apenas registra para que
// times de produto investiguem SDKs enviando valores divergentes.
func NormalizeAndCheck(product string) string {
	n := Normalize(product)
	if n != "" && !IsKnown(n) {
		log.Printf("pulse.products: product desconhecido recebido: %q (original=%q)", n, product)
	}
	return n
}

// ResetForTest reinicializa o cache da allowlist. Uso restrito a testes que
// alteram PULSE_KNOWN_PRODUCTS em runtime.
func ResetForTest() {
	knownMu.Lock()
	knownLoaded = false
	knownProduct = nil
	knownMu.Unlock()
}
