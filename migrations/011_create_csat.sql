-- UP
CREATE TABLE IF NOT EXISTS csat_surveys (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product     VARCHAR(100) NOT NULL,
    question    TEXT NOT NULL,
    scale       SMALLINT NOT NULL DEFAULT 10,
    type        VARCHAR(20) NOT NULL DEFAULT 'nps',
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS csat_responses (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_id   UUID NOT NULL REFERENCES csat_surveys(id) ON DELETE CASCADE,
    score       SMALLINT NOT NULL,
    comment     TEXT,
    respondent  VARCHAR(255),
    metadata    JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_csat_responses_survey  ON csat_responses (survey_id);
CREATE INDEX IF NOT EXISTS idx_csat_responses_created ON csat_responses (created_at);
CREATE INDEX IF NOT EXISTS idx_csat_surveys_product   ON csat_surveys (product);

-- DOWN
-- DROP INDEX IF EXISTS idx_csat_surveys_product;
-- DROP INDEX IF EXISTS idx_csat_responses_created;
-- DROP INDEX IF EXISTS idx_csat_responses_survey;
-- DROP TABLE IF EXISTS csat_responses;
-- DROP TABLE IF EXISTS csat_surveys;
