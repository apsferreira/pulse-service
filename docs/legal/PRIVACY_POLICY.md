# Política de Privacidade — Pulse

**Vigente a partir de:** 2026-06-09

Esta Política descreve como o pulse-service trata dados pessoais coletados ao
receber eventos de produtos integrados ao Pulse. Aplica-se em conjunto ao DPA
(`DPA_PULSE_TEMPLATE.md`) firmado com cada Controlador.

## Dados coletados em cada evento

| Campo | Conteúdo persistido | Base legal LGPD |
|-------|---------------------|-----------------|
| `ip_address` | **IP anonimizado** — IPv4 com o último octeto zerado (ex.: `200.100.50.42` → `200.100.50.0`); IPv6 com os últimos 80 bits zerados, preservando apenas o prefixo `/48` | Art. 7º IX — legítimo interesse (análise geográfica agregada e proteção contra abuso) |
| `user_agent` | String literal enviada pelo cliente | Art. 7º IX — legítimo interesse (compatibilidade de plataforma e detecção de fraude) |
| `user_id` / `session_id` | Identificadores opcionais escolhidos pelo Controlador | Art. 7º VI — execução contratual |
| Propriedades do evento | Definidas pelo Controlador no SDK | Art. 7º VI / IX, conforme finalidade declarada |

## IP — operação irreversível

A anonimização do IP é realizada **antes da persistência** pelo
`internal/pkg/privacy.AnonymizeIP`. Não existe sal, chave ou tabela de
reversão: o IP completo nunca é gravado em disco nem em índices secundários.
Logs de erro do serviço **não devem** registrar o IP bruto recebido na
requisição.

## Uso transitório do IP

O IP completo é mantido apenas em memória durante o ciclo de vida da
requisição HTTP, exclusivamente para:
- aplicação de rate limit anti-abuso (chaves Redis com TTL ≤ 60s);
- decisão de roteamento da própria requisição.

Não há persistência em banco de dados, fila ou backup do IP bruto.

## Solicitações de titulares

Solicitações dos titulares (acesso, correção, eliminação) devem ser
encaminhadas ao Controlador. A IIT, como Operador, atende solicitações
recebidas via Controlador conforme prazo definido no DPA.
