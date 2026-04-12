-- BKL-420: Pulse — Session Replay (gravação de sessão do usuário)
-- Armazena eventos de sessão anonimizados para análise de UX.
-- LGPD: dados anonimizados, sem PII direta. user_agent truncado em 200 chars.
-- Opt-out: verificado no SDK antes de enviar qualquer dado.

CREATE TABLE IF NOT EXISTS session_replays (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id  TEXT        NOT NULL,
    project_id  UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id     TEXT        NOT NULL DEFAULT 'anonymous',
    page_url    TEXT,
    user_agent  TEXT,
    event_count INT         NOT NULL DEFAULT 0,
    duration_ms INT         NOT NULL DEFAULT 0,
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_event_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(session_id)
);

CREATE TABLE IF NOT EXISTS session_replay_events (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id  TEXT        NOT NULL REFERENCES session_replays(session_id) ON DELETE CASCADE,
    project_id  UUID        NOT NULL,
    event_type  VARCHAR(30) NOT NULL CHECK (event_type IN ('dom_mutation', 'click', 'scroll', 'viewport', 'url_change')),
    timestamp_ms BIGINT     NOT NULL,
    data        JSONB       NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_session_replays_project_id   ON session_replays(project_id);
CREATE INDEX IF NOT EXISTS idx_session_replays_started_at   ON session_replays(started_at);
CREATE INDEX IF NOT EXISTS idx_session_replay_events_session ON session_replay_events(session_id);
CREATE INDEX IF NOT EXISTS idx_session_replay_events_type   ON session_replay_events(event_type);

-- Cleanup automático após 30 dias (LGPD — retenção mínima)
-- Trigger agendado via pg_cron ou job Go:
-- DELETE FROM session_replays WHERE started_at < NOW() - INTERVAL '30 days';

-- DOWN:
-- DROP INDEX IF EXISTS idx_session_replay_events_type;
-- DROP INDEX IF EXISTS idx_session_replay_events_session;
-- DROP INDEX IF EXISTS idx_session_replays_started_at;
-- DROP INDEX IF EXISTS idx_session_replays_project_id;
-- DROP TABLE IF EXISTS session_replay_events;
-- DROP TABLE IF EXISTS session_replays;
