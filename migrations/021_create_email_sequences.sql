-- BKL-1082: Email sequence de onboarding automático.
-- Controla o envio de emails via Resend para novos tenants do Pulse.
-- Tabelas: email_sequence_templates (o que enviar) + email_sequence_events (o que já foi enviado).

CREATE TABLE IF NOT EXISTS email_sequence_templates (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    trigger_day INTEGER NOT NULL, -- dias após signup: 0, 1, 7, 30, 60
    subject     TEXT NOT NULL,
    body_md     TEXT NOT NULL,    -- corpo em Markdown (renderizado pelo Resend)
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Templates padrão de onboarding
INSERT INTO email_sequence_templates (name, trigger_day, subject, body_md) VALUES
('welcome', 0,
 'Bem-vindo ao Pulse 🚀',
 '# Bem-vindo ao Pulse!\n\nOlá {{name}},\n\nSeu projeto **{{project_name}}** está pronto.\n\n**API Key:** `{{api_key}}`\n\n**Próximo passo:** instale o SDK em 2 minutos:\n\n```bash\nnpm install @pulse/js\n```\n\nPrecisa de ajuda? Veja o [Guia de início rápido](https://pulse.institutoitinerante.com.br/docs/guides/01-por-onde-comecar).\n\nAté logo,\nTime Pulse'),

('sdk_reminder', 1,
 'Você ainda não integrou o SDK — posso ajudar?',
 '# SDK ainda não integrado\n\nOlá {{name}},\n\nNotamos que o projeto **{{project_name}}** ainda não rastreou nenhum evento.\n\nLeva menos de 5 minutos para integrar:\n\n```typescript\nimport { PulseSDK } from "@pulse/js"\nPulseSDK.init({ projectKey: "{{api_key}}" })\nPulseSDK.track("app_opened")\n```\n\nVeja o [guia completo](https://pulse.institutoitinerante.com.br/docs/guides/02-instalar-sdk-js).\n\nResponda este email se tiver dúvidas — respondemos em 24h.'),

('first_flag_milestone', 7,
 '🎉 Primeira semana com o Pulse',
 '# Parabéns pela primeira semana!\n\nOlá {{name}},\n\n{{#if has_flags}}Você criou {{flag_count}} flags. Que tal adicionar rollout percentual para liberar features gradualmente?{{else}}Ainda não criou nenhuma flag. Que tal criar sua primeira agora?\n\nVá para o dashboard → Nova Flag → ative para 10% dos usuários.{{/if}}\n\n[Abrir dashboard](https://pulse.institutoitinerante.com.br)'),

('health_check_30d', 30,
 'Como está o Pulse no seu produto?',
 '# Checkin do mês\n\nOlá {{name}},\n\nJá faz 30 dias! Aqui está um resumo do projeto **{{project_name}}**:\n\n- **Eventos rastreados:** {{event_count}}\n- **Flags ativas:** {{active_flag_count}}\n- **Health score:** {{health_score}}/100\n\n{{#if low_health}}Seu health score está abaixo de 50. Quer uma call para otimizar o uso do Pulse?{{/if}}\n\nPrecisa de algo? Responda este email.'),

('winback_60d', 60,
 'Ainda usando o Pulse? Temos novidades',
 '# Tem novidades no Pulse!\n\nOlá {{name}},\n\nNo último mês lançamos:\n- ✅ Audit log de flags\n- ✅ Ambientes por projeto (prod/staging/dev)\n- ✅ A/B testing com chi-squared\n\n{{#if is_free_plan}}Está no plano Free. Que tal fazer upgrade para o Starter (R$399/mês) e ter 500k eventos/mês + A/B testing?{{/if}}\n\n[Ver novidades](https://pulse.institutoitinerante.com.br/changelog)')
ON CONFLICT (name) DO NOTHING;

-- Tabela de controle: quais emails já foram enviados por projeto
CREATE TABLE IF NOT EXISTS email_sequence_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    template_id UUID NOT NULL REFERENCES email_sequence_templates(id),
    email       VARCHAR(255) NOT NULL,
    sent_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status      VARCHAR(20) NOT NULL DEFAULT 'sent' CHECK (status IN ('sent', 'failed', 'skipped')),
    error       TEXT,
    UNIQUE(project_id, template_id) -- cada template enviado apenas 1x por projeto
);

CREATE INDEX IF NOT EXISTS idx_email_seq_events_project ON email_sequence_events(project_id);
CREATE INDEX IF NOT EXISTS idx_email_seq_events_sent    ON email_sequence_events(sent_at DESC);

-- Controle de unsubscribe (compliance LGPD)
CREATE TABLE IF NOT EXISTS email_unsubscribes (
    email      VARCHAR(255) PRIMARY KEY,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    reason     VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE email_sequence_templates IS 'Templates de email sequence de onboarding Pulse — BKL-1082.';
COMMENT ON TABLE email_sequence_events IS 'Log de emails enviados por projeto (evita duplicatas) — BKL-1082.';
COMMENT ON TABLE email_unsubscribes IS 'Lista de unsubscribe LGPD para email sequences Pulse — BKL-1082.';
