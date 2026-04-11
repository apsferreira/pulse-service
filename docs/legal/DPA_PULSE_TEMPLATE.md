# Data Processing Agreement (DPA) — Pulse
## Acordo de Processamento de Dados Pessoais

**BKL-1033** | Versão 1.0 | Vigente a partir de: _______

---

## Partes

**Controlador:** [NOME DO CLIENTE], CNPJ [CNPJ DO CLIENTE], ("**Controlador**")

**Operador:** Instituto Itinerante de Tecnologia LTDA, CNPJ a definir, Av. [endereço completo], Salvador/BA ("**IIT**" ou "**Operador**")

O presente Acordo de Processamento de Dados Pessoais ("DPA") é parte integrante dos Termos de Serviço do Pulse e regulamenta o processamento de dados pessoais realizado pela IIT em nome do Controlador, nos termos da Lei nº 13.709/2018 (LGPD) e do Regulamento (UE) 2016/679 (GDPR), quando aplicável.

---

## 1. Objeto

O Controlador utiliza a plataforma Pulse para:
- Coleta e análise de eventos de usuários finais
- Gestão de feature flags com targeting por usuário/tenant
- Análise estatística de experimentos A/B
- Monitoramento de uptime de serviços

Para essas finalidades, o Controlador instrui a IIT a processar dados pessoais de seus usuários finais.

---

## 2. Categorias de Dados Pessoais

A IIT processa, por instrução do Controlador, as seguintes categorias de dados:

| Categoria | Exemplos | Base Legal |
|-----------|----------|------------|
| Identificadores | `user_id`, `session_id`, `tenant_id` | Art. 7º VI LGPD — execução contratual |
| Dados de comportamento | tipo de evento, propriedades do evento, timestamp | Art. 7º IX — legítimo interesse |
| Dados de dispositivo | `user_agent`, IP anonimizado | Art. 7º IX — segurança e fraude |
| Email do administrador | email do owner da conta Pulse | Art. 7º VI — execução contratual |

**Dados sensíveis:** A IIT não processa dados sensíveis (art. 5º II LGPD) por meio do Pulse. O Controlador é responsável por não enviar dados sensíveis nos campos de propriedades de eventos.

---

## 3. Finalidade e Instrução de Processamento

A IIT processa dados pessoais **exclusivamente** para as seguintes finalidades, conforme instrução do Controlador:

1. Armazenamento e processamento de eventos de analytics
2. Avaliação de feature flags por usuário/segmento
3. Geração de relatórios e métricas para o dashboard do Controlador
4. Execução de experimentos A/B com análise estatística
5. Envio de notificações operacionais ao administrador da conta

A IIT não usa dados do Controlador para fins próprios, publicidade ou treinamento de modelos de IA.

---

## 4. Prazo de Retenção

| Dado | Retenção | Critério |
|------|----------|----------|
| Eventos de analytics | 90 dias (Free), 180 dias (Starter/Growth), 1 ano (Pro) | Plano contratado |
| Flag evaluations | 1 ano | Compliance LGPD Art. 37 |
| Logs de sistema | 30 dias | Segurança operacional |
| Email do administrador | Duração do contrato + 5 anos | Obrigação legal fiscal |

Após expiração, os dados são deletados automaticamente ou anonimizados de forma irreversível.

---

## 5. Subprocessadores

A IIT utiliza os seguintes subprocessadores para entregar o serviço Pulse:

| Subprocessador | País | Finalidade | Garantia |
|----------------|------|------------|----------|
| Cloudflare Inc. | EUA | CDN, proxy de rede | DPA Cloudflare + SCC |
| Hetzner Online GmbH | Alemanha | Infraestrutura K3s (fallback) | GDPR Art. 28 |
| Homelab IIT (K3s) | Brasil | Processamento principal | Ambiente on-premise brasileiro |
| TrueNAS (NAS local) | Brasil | Backup de dados | On-premise, LGPD Art. 46 |
| Resend Inc. | EUA | Notificações por email | DPA Resend + SCC |
| Redis Labs | EUA | Cache de feature flags | DPA Redis + SCC |

O Controlador autoriza o uso dos subprocessadores listados. Alterações na lista serão comunicadas com **30 dias de antecedência** por email ao administrador da conta.

---

## 6. Transferências Internacionais

Dados trafegam via Cloudflare CDN (EUA) para otimização de latência. Essa transferência é coberta por Cláusulas Contratuais Padrão (SCC) aprovadas pela Comissão Europeia, e cumpre o art. 33 da LGPD por garantia contratual adequada.

---

## 7. Medidas de Segurança (Art. 46 LGPD)

A IIT implementa:

- **Criptografia em trânsito:** TLS 1.3 em todas as conexões
- **Criptografia em repouso:** volumes K3s com LUKS (em homologação)
- **Controle de acesso:** RBAC K3s, API keys por projeto/ambiente, X-Service-Token
- **Isolamento multi-tenant:** `project_id` em todas as tabelas, queries parametrizadas
- **Monitoramento:** Logs centralizados, alertas de anomalia
- **Backup:** pg_basebackup semanal + WAL archiving (RPO ~5 minutos) em TrueNAS RAID-Z1
- **Resposta a incidentes:** SLA de notificação ao Controlador em até 72h (LGPD Art. 48)

---

## 8. Direitos dos Titulares (Art. 18 LGPD)

A IIT disponibiliza APIs para o Controlador exercer direitos dos titulares:

- **Portabilidade:** `GET /api/v1/users/:user_id/export` — exporta todos os dados do usuário em JSON
- **Exclusão:** `DELETE /api/v1/users/:user_id/data` — remove eventos e flag evaluations
- **Confirmação de processamento:** via dashboard Pulse → Configurações → Dados

O Controlador é responsável por receber e atender as solicitações dos titulares, usando as APIs da IIT como facilitador. Prazo de atendimento: 15 dias (LGPD Art. 18 §3°).

---

## 9. Notificação de Incidentes (Art. 48 LGPD)

Em caso de incidente de segurança que possa causar dano aos titulares, a IIT notificará o Controlador em até **72 horas** após a ciência do incidente, por email ao administrador da conta, contendo:

- Natureza do incidente e dados afetados
- Estimativa de titulares impactados
- Medidas tomadas e plano de mitigação
- Contato do responsável pela resposta ao incidente

---

## 10. Obrigações do Controlador

O Controlador declara e se obriga a:

1. Ter base legal válida para coletar e enviar dados pessoais ao Pulse
2. Informar usuários finais sobre o uso de analytics (Aviso de Privacidade)
3. Não enviar dados sensíveis (saúde, biometria, crenças) como propriedades de eventos
4. Usar as APIs de exclusão quando um usuário exercer o direito ao esquecimento
5. Manter credenciais (API keys) em segurança e não compartilhá-las publicamente

---

## 11. Auditoria

O Controlador pode solicitar relatório de auditoria de conformidade (SOC 2 Type II, quando disponível) ou questionário de segurança. Auditorias presenciais estão sujeitas a agendamento e custos.

---

## 12. Responsabilidade

A IIT é responsável pelos danos causados por falha no cumprimento deste DPA ou da LGPD, limitada ao valor pago pelo Controlador nos 12 meses anteriores ao evento danoso. A IIT não é responsável por danos decorrentes de instruções incorretas do Controlador ou uso indevido das APIs.

---

## 13. Vigência e Rescisão

Este DPA vigora enquanto o Controlador usar o Pulse. Ao encerrar a conta, a IIT deletará os dados processados em até **30 dias**, salvo obrigação legal de retenção.

---

## 14. Foro

Disputas decorrentes deste DPA serão resolvidas perante o Foro da Comarca de Salvador/BA, exceto quando a legislação de defesa do consumidor exigir outro foro.

---

## Assinaturas

| Controlador | IIT (Operador) |
|-------------|----------------|
| Nome: _________________ | Nome: Antonio Pedro Ferreira |
| Cargo: _________________ | Cargo: Responsável Técnico |
| Data: _________________ | Data: _________________ |
| Assinatura: _____________ | Assinatura: _____________ |

---

*Para assinar digitalmente, enviar para: juridico@institutoitinerante.com.br*
