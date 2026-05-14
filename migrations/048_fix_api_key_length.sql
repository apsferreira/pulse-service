-- Migration 048: corrigir tipo da coluna api_key em projects
-- VARCHAR(64) não comporta o formato "pulse_<hex64>" (70 chars). Expandir para TEXT.

ALTER TABLE projects ALTER COLUMN api_key TYPE TEXT;
