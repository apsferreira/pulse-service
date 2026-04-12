-- BKL-1002: Adiciona coluna prompt_variant à tabela feature_flags.
-- Permite que flags carregem configuração de LLM para A/B test de prompts.
-- Estrutura esperada: {"system_prompt": "...", "model": "gemini-2.5-flash-preview-04-17"}

ALTER TABLE feature_flags
    ADD COLUMN IF NOT EXISTS prompt_variant JSONB;

-- DOWN:
-- ALTER TABLE feature_flags DROP COLUMN IF EXISTS prompt_variant;
