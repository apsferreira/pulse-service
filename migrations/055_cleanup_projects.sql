-- Correção de projetos duplicados e renomeação do projeto principal IIT

-- 1. Renomeia o projeto principal para o nome legal correto
UPDATE projects
  SET name = 'Instituto Itinerante de Tecnologia'
  WHERE slug = 'iit-site';

-- 2. Remove o Catraca Virtual duplicado (catraca-app, criado mais tarde, sem flags/eventos/ambientes)
--    O projeto canônico é catraca (b4000000-...) — ID seeded, mais antigo
DELETE FROM projects WHERE slug = 'catraca-app';
