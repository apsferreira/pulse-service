-- Migration 043: placeholder — número reservado, conteúdo movido para 044.
--
-- Histórico: durante o desenvolvimento de BKL-1078 (webhooks), o arquivo
-- desta migration foi criado com o número 044 por engano e o número 043
-- ficou ausente da sequência, gerando inconsistência entre o nome do arquivo
-- e o comentário de cabeçalho de 044 (que dizia "Migration 043").
--
-- As alterações originalmente atribuídas à 043
-- (webhook_type, success, duration_ms) estão aplicadas em
-- 044_webhook_types_and_delivery_history.sql.
--
-- Este arquivo existe apenas para fechar o gap na sequência e manter
-- coerência ao auditar `ls migrations/*.sql | sort`. Não executa DDL.

SELECT 1;
