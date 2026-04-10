-- BKL-811: Seed da survey NPS para o SocialMake.
-- Usada pelo NPS30DayScheduler para coletar satisfação dos tenants com 30 dias de uso.
-- A survey é criada via csat_surveys (infraestrutura já existente do pulse-service).

INSERT INTO csat_surveys (product, question, scale, type, active)
VALUES (
    'socialmake',
    'Em uma escala de 0 a 10, o quanto você recomendaria o SocialMake para um colega ou amigo?',
    10,
    'nps',
    true
)
ON CONFLICT DO NOTHING;
