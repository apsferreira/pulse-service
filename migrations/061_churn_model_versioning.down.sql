ALTER TABLE project_churn_models
    DROP COLUMN IF EXISTS last_validation_accuracy,
    DROP COLUMN IF EXISTS last_validated_at,
    DROP COLUMN IF EXISTS model_version;
