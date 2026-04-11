-- BKL: Corrige URL do monitor do contract-service
-- O subdomínio correto é "contract" (sem 's'), não "contracts"
UPDATE monitors
SET url = 'https://contract.institutoitinerante.com.br/health'
WHERE name = 'Contracts Service'
  AND url = 'https://contracts.institutoitinerante.com.br/health';
