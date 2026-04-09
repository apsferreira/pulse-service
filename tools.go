//go:build tools

// BKL-226: ferramentas de geração de código.
// Para instalar: go install go.uber.org/mock/mockgen@latest
// Para gerar mocks: go generate ./...
package tools

import (
	_ "go.uber.org/mock/mockgen"
)
