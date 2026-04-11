package domain

// BKL-1080: Schema registry de eventos — define estrutura esperada de cada event_type.

import "time"

// EventPropertyType define os tipos suportados de propriedades de eventos.
type EventPropertyType string

const (
	PropertyTypeString  EventPropertyType = "string"
	PropertyTypeNumber  EventPropertyType = "number"
	PropertyTypeBoolean EventPropertyType = "boolean"
	PropertyTypeObject  EventPropertyType = "object"
	PropertyTypeArray   EventPropertyType = "array"
)

// EventSchemaProperty define uma propriedade esperada em um schema de evento.
type EventSchemaProperty struct {
	ID          string            `json:"id"`
	SchemaID    string            `json:"schema_id"`
	Name        string            `json:"name"`
	Type        EventPropertyType `json:"type"`
	Required    bool              `json:"required"`
	Description string            `json:"description"`
}

// EventSchema define o schema esperado de um event_type dentro de um projeto.
type EventSchema struct {
	ID          string                `json:"id"`
	ProjectID   string                `json:"project_id"`
	EventName   string                `json:"event_name"`
	Description string                `json:"description"`
	Strict      bool                  `json:"strict"`       // true = rejeita props não declaradas
	Properties  []EventSchemaProperty `json:"properties"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

// CreateEventSchemaRequest é o payload para POST /api/v1/event-schemas.
type CreateEventSchemaRequest struct {
	EventName   string                         `json:"event_name"`
	Description string                         `json:"description"`
	Strict      bool                           `json:"strict"`
	Properties  []CreateEventSchemaPropertyReq `json:"properties"`
}

// CreateEventSchemaPropertyReq define uma propriedade ao criar um schema.
type CreateEventSchemaPropertyReq struct {
	Name        string            `json:"name"`
	Type        EventPropertyType `json:"type"`
	Required    bool              `json:"required"`
	Description string            `json:"description"`
}

// SchemaValidationError descreve um erro de validação de evento contra schema.
type SchemaValidationError struct {
	Property string `json:"property"`
	Message  string `json:"message"`
}

// ValidateAgainstSchema valida um mapa de propriedades contra um EventSchema.
// Retorna lista de erros (vazia se válido).
func ValidateAgainstSchema(schema EventSchema, properties map[string]any) []SchemaValidationError {
	var errs []SchemaValidationError

	// Verifica propriedades obrigatórias
	for _, prop := range schema.Properties {
		val, exists := properties[prop.Name]
		if prop.Required && !exists {
			errs = append(errs, SchemaValidationError{
				Property: prop.Name,
				Message:  "propriedade obrigatória ausente",
			})
			continue
		}
		if !exists {
			continue
		}
		// Valida tipo
		if typeErr := validatePropertyType(prop.Name, prop.Type, val); typeErr != nil {
			errs = append(errs, *typeErr)
		}
	}

	// Modo strict: rejeita propriedades não declaradas
	if schema.Strict {
		declared := make(map[string]bool, len(schema.Properties))
		for _, p := range schema.Properties {
			declared[p.Name] = true
		}
		for k := range properties {
			if !declared[k] {
				errs = append(errs, SchemaValidationError{
					Property: k,
					Message:  "propriedade não declarada no schema (modo strict)",
				})
			}
		}
	}

	return errs
}

// validatePropertyType verifica se um valor tem o tipo esperado.
func validatePropertyType(name string, expected EventPropertyType, val any) *SchemaValidationError {
	switch expected {
	case PropertyTypeString:
		if _, ok := val.(string); !ok {
			return &SchemaValidationError{Property: name, Message: "esperado string"}
		}
	case PropertyTypeNumber:
		switch val.(type) {
		case float64, float32, int, int64, int32:
			// ok
		default:
			return &SchemaValidationError{Property: name, Message: "esperado number"}
		}
	case PropertyTypeBoolean:
		if _, ok := val.(bool); !ok {
			return &SchemaValidationError{Property: name, Message: "esperado boolean"}
		}
	case PropertyTypeObject:
		if _, ok := val.(map[string]any); !ok {
			return &SchemaValidationError{Property: name, Message: "esperado object"}
		}
	case PropertyTypeArray:
		if _, ok := val.([]any); !ok {
			return &SchemaValidationError{Property: name, Message: "esperado array"}
		}
	}
	return nil
}
