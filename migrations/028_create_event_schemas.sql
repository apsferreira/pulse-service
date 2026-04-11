-- BKL-1080: Schema registry de eventos.
-- Define o schema esperado de cada event_type (nome + propriedades + tipos).
-- O backend valida os eventos ingeridos contra o schema (se existir).
-- Eventos sem schema passam sem validação (comportamento compatível com versão anterior).
-- DOWN:
--   DROP TABLE IF EXISTS event_schema_properties;
--   DROP TABLE IF EXISTS event_schemas;
--   DROP TYPE IF EXISTS event_property_type;

CREATE TYPE event_property_type AS ENUM ('string', 'number', 'boolean', 'object', 'array');

-- Schema de um event_type dentro de um projeto
CREATE TABLE IF NOT EXISTS event_schemas (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID        NOT NULL,
    event_name      TEXT        NOT NULL,                -- ex: "checkout_completed"
    description     TEXT        NOT NULL DEFAULT '',
    strict          BOOLEAN     NOT NULL DEFAULT false,   -- true = rejeita props não declaradas
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, event_name)
);

-- Propriedades esperadas de cada schema
CREATE TABLE IF NOT EXISTS event_schema_properties (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schema_id    UUID        NOT NULL REFERENCES event_schemas(id) ON DELETE CASCADE,
    name         TEXT        NOT NULL,                    -- nome da propriedade
    type         event_property_type NOT NULL,
    required     BOOLEAN     NOT NULL DEFAULT false,
    description  TEXT        NOT NULL DEFAULT '',
    UNIQUE(schema_id, name)
);

CREATE INDEX IF NOT EXISTS idx_event_schemas_project_id  ON event_schemas(project_id);
CREATE INDEX IF NOT EXISTS idx_event_schemas_event_name  ON event_schemas(project_id, event_name);

COMMENT ON TABLE  event_schemas            IS 'BKL-1080: Schema registry — define estrutura esperada de cada event_type';
COMMENT ON COLUMN event_schemas.strict     IS 'Se true, o backend rejeita eventos com propriedades não declaradas no schema';
