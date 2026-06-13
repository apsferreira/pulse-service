-- fix(#66): versionamento + revalidação periódica do churn model.
-- model_version       — incrementado a cada Train (drift detection)
-- last_validated_at   — timestamp da última revalidação semanal
-- last_validation_accuracy — acurácia do modelo na revalidação semanal (0.0-1.0)

ALTER TABLE project_churn_models
    ADD COLUMN IF NOT EXISTS model_version           INT NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS last_validated_at       TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_validation_accuracy FLOAT;
