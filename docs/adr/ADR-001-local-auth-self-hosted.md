# ADR-001 — Autenticação local para modo self-hosted (PULSE_AUTH_MODE=local)

**Data:** 2026-07-19
**Status:** Aceito
**Autores:** @tl (auditoria sweep v96)
**STD relacionado:** STD-001

---

## Contexto

O pulse-service opera em dois modos:

- **`PULSE_AUTH_MODE=external`** (padrão IIT): valida JWT emitido pelo auth-service. Modo de produção no ecossistema IIT. **Conforme com STD-001.**
- **`PULSE_AUTH_MODE=local`** (self-hosted): autenticação local com `pulse_users` (email + password_hash + role). Sem conexão com o auth-service IIT.

O pulse-service é um produto de analytics/feature-flags oferecido como plataforma self-hosted para terceiros. Clientes externos que instalam o pulse em sua própria infra não têm acesso ao auth-service IIT.

## Decisão

Manter o modo `PULSE_AUTH_MODE=local` com a tabela `pulse_users` como exceção justificada ao STD-001, com as seguintes salvaguardas:

1. O modo local é **exclusivamente para instalações self-hosted** fora da infraestrutura IIT.
2. Na IIT production, `PULSE_AUTH_MODE` NÃO deve ser definido (default `external`).
3. `pulse_users` armazena apenas `email`, `name`, `password_hash` (bcrypt) e `role` — sem outros PII.
4. O modo local não aceita JWTs do auth-service; são dois sistemas mutuamente exclusivos, controlados por variável de ambiente.

## Conformidade IIT

| Aspecto | Estado |
|---------|--------|
| IIT production usa auth-service JWT | ✓ (PULSE_AUTH_MODE=external, default) |
| Tabela pulse_users em produção IIT | Existe no schema mas não usada |
| STD-001 em produção IIT | Conforme |
| Documentação ADR | Este documento |

## Consequências

- Qualquer deploy IIT deve verificar que `PULSE_AUTH_MODE` não está setado como `local` (ou deve ser `external` explicitamente).
- `pulse_users` deve ser excluída do backup de LGPD se o pulse for usado apenas no modo IIT (sem self-hosted).
- Novos endpoints administrativos devem ser protegidos apenas por `JWTAuth` (external) no modo IIT; handlers do `LocalAuthHandler` só devem ser registrados quando `PULSE_AUTH_MODE=local`.

## Alternativas consideradas

| Alternativa | Descartada por |
|------------|----------------|
| Remover modo local | Pulse é produto self-hosted — sem auth-service IIT disponível no ambiente do cliente |
| Usar auth-service IIT em self-hosted | Exige que cliente externo conecte à infra IIT — quebraria o modelo de produto isolado |
| Usar token estático (SERVICE_TOKEN) para admin | Já implementado como alternativa mais simples; pulse_users é para UI de dashboard com múltiplos usuários |
