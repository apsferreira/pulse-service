# Pulse CS — Playbook: Win-back Campaign para Churned Tenants

> BKL-1087 | Versão 1.0 | Abril 2026
> Owner: Customer Success / @growth

---

## Definição de Churn

**Tenant churned:** conta que ficou 30 dias consecutivos sem nenhum login de qualquer usuário do tenant.

**Critério técnico de identificação:**
```sql
SELECT t.slug, t.email_admin, t.plan, t.created_at,
       MAX(s.created_at) AS last_seen
FROM tenants t
LEFT JOIN sessions s ON s.tenant_id = t.id
GROUP BY t.id
HAVING MAX(s.created_at) < NOW() - INTERVAL '30 days'
   OR MAX(s.created_at) IS NULL
ORDER BY last_seen DESC;
```

**Excluir da campanha:** tenants em plano `internal` (IIT), tenants com cancelamento formal registrado (já receberam off-boarding), tenants que solicitaram remoção de dados (LGPD).

**Frequência de execução:** verificação semanal, toda segunda-feira.

---

## Sequência de 5 Emails

### Email 1 — D+0 (dia do churn detectado)

**Assunto:** "[NOME], a gente notou sua ausência"

**Remetente:** antonio@institutoitinerante.com.br (humanizado, não noreply)

---

Oi, [NOME].

Sou o Antonio, fundador do Pulse.

Notei que a equipe do [NOME_EMPRESA] não entrou no Pulse há um tempão. Não quero mandar email automático sem sentido — então aqui vai um direto:

Você parou de usar por algum problema específico?

Pode ter sido um bug que travou alguma coisa. Pode ter sido a interface que não encaixou no jeito de trabalhar do time. Pode ter sido falta de tempo mesmo.

Me conta. Responde esse email e eu leio pessoalmente.

Se tiver alguma coisa quebrada, vou resolver. Se a ferramenta não serviu pro que você precisava, quero saber — isso me ajuda a fazer o produto melhor.

Antonio Pedro
Fundador, Instituto Itinerante de Tecnologia

---

**Configurações de envio:**
- Sem rastreio de abertura neste primeiro email (cria confiança)
- Monitorar respostas manualmente — responder em até 24h
- Se resposta recebida → sair da sequência automática e tratar manualmente

---

### Email 2 — D+7 (1 semana após churn)

**Assunto:** "O que mudou no Pulse desde a última vez que você entrou"

---

Oi, [NOME].

Semana passada eu mandei um email mas não quero parecer insistente sem motivo. Então aqui vai algo concreto.

Desde quando você usou o Pulse pela última vez, a gente lançou:

- **Dashboard de série temporal**: agora você vê evolução de KPIs ao longo do tempo, não só snapshot atual
- **Alertas configuráveis**: recebe email quando um evento passa de um threshold que você definiu
- **Export de dados em CSV**: pede e recebe os dados brutos do Pulse sem precisar chamar a API

Pode ser que o Pulse de hoje seja diferente do que você testou.

Se quiser explorar, é só entrar: [LINK_LOGIN]

Qualquer dúvida, responde aqui.

Antonio

P.S.: Se você não usa mais porque mudou de necessidade (parou o produto, mudou de empresa), sem problema — me avisa e eu encerro a conta direitinho.

---

**Configurações de envio:**
- Personalizar as features listadas com base nos eventos reais do tenant (o que ele usava antes)
- Se o tenant estava no plano free: destacar features que só existem nos planos pagos

---

### Email 3 — D+30 (1 mês após churn)

**Assunto:** "Proposta: 2 meses com 50% de desconto para voltar"

---

Oi, [NOME].

Faz um mês que a conta do [NOME_EMPRESA] no Pulse está inativa.

Sem enrolação: quero que você dê uma segunda chance pro Pulse, e por isso estou fazendo uma proposta concreta.

**Se você reativar nos próximos 7 dias: os próximos 2 meses saem com 50% de desconto.**

No plano [PLANO_ATUAL], isso significa:
- De R$ [VALOR_ATUAL] → R$ [VALOR_DESCONTO] por mês, nos dois primeiros meses
- Sem fidelidade, sem multa de saída

Para ativar, é só responder esse email com "quero voltar" e eu configuro manualmente para você.

Ou entra na plataforma e aplica o cupom: **VOLTEI50**

[LINK_LOGIN]

Se a oferta não fizer sentido pro seu momento agora, sem problema. Mas se tiver qualquer dúvida sobre o que o Pulse pode fazer pelo seu produto, estou disponível para uma call de 20 minutos: [LINK_CALENDLY]

Antonio

---

**Configurações de envio:**
- Cupom `VOLTEI50` válido por 7 dias a partir do envio
- Configurar cupom no billing-service antes de disparar
- Rastrear abertura e clique neste email (conversão esperada: 10%)

---

### Email 4 — D+60 (2 meses após churn)

**Assunto:** "Uma pergunta rápida — 30 segundos, sério"

---

Oi, [NOME].

Prometo que não vou mandar mais emails depois desse se você não quiser.

Uma pergunta rápida: o que fez você parar de usar o Pulse?

[ ] O produto não resolveu o que eu precisava
[ ] Ficou caro
[ ] A equipe mudou e a gente perdeu o contexto
[ ] Escolhemos outra ferramenta
[ ] Simplesmente não tivemos tempo de usar

Pode responder clicando em uma das opções acima, ou simplesmente responde esse email com sua resposta. Leva 30 segundos.

Sua resposta me ajuda a entender o que posso melhorar. Não vou usar ela para te convencer a voltar — só para fazer o produto melhor.

Obrigado.

Antonio

---

**Configurações de envio:**
- Links de tracking para cada opção — registrar resposta no CRM
- Respostas abrem ticket de CS para tratamento manual
- Este email é de pesquisa, não de venda — tom deliberadamente neutro

---

### Email 5 — D+90 (3 meses após churn)

**Assunto:** "Encerrando a conta do [NOME_EMPRESA] — mas a porta fica aberta"

---

Oi, [NOME].

Faz 3 meses que a conta do [NOME_EMPRESA] está inativa.

Vou arquivar a conta para liberar recursos — mas antes quero deixar claro o que acontece:

- **Seus dados ficam preservados por mais 60 dias.** Se mudar de ideia, é só entrar em contato e eu reativo tudo.
- **Após 60 dias**, os dados são deletados permanentemente (conforme nossa Política de Privacidade).
- Se quiser uma cópia dos seus dados agora, é só pedir — exporto e mando por email.

Se um dia o Pulse fizer sentido de novo, a porta está aberta. Não tem nenhum ressentimento aqui.

Para exportar dados antes do encerramento: [LINK_EXPORT]
Para reativar a conta: responde esse email

Obrigado por ter testado o Pulse.

Antonio

---

**Configurações de envio:**
- Disparar após confirmação manual de que o tenant não respondeu nenhum email anterior
- Processar LGPD: registrar data de notificação e data prevista de deleção
- Marcar tenant como `status: churned_notified` no banco
- Após 60 dias: executar script de deleção e marcar como `status: deleted`

---

## Métricas de Sucesso

| KPI | Meta |
|-----|------|
| Taxa de abertura (Email 1) | > 40% |
| Taxa de resposta (Email 1) | > 8% |
| Taxa de conversão (Email 3 — cupom) | > 10% |
| Win-back rate geral (30d de campanha) | > 10% |
| Tempo médio para win-back | < 45 dias |
| Churn prevenido (saiu da sequência por reativação) | rastrear |

**Win-back rate** = tenants que reativaram / total de tenants que entraram na campanha no período.

---

## Automação Recomendada

Sequência ideal via n8n (self-hosted em 192.168.x.x):

```
Trigger: Cron toda segunda 09h
→ Query SQL identifica novos churned tenants
→ Adiciona à sequência no RD Station (se disponível) ou fila manual
→ Email 1 disparado via notification-service (resend.com)
→ Aguardar 7 dias → Email 2 (se sem resposta)
→ Aguardar 30 dias → Email 3 + criar cupom via billing-service API
→ Aguardar 60 dias → Email 4
→ Aguardar 90 dias → Email 5 + agendar deleção
```

Enquanto n8n não está deployado: executar sequência manualmente via planilha de controle.

---

## Planilha de Controle Manual

Colunas: `tenant_slug | email_admin | plano | last_seen | email1_data | email2_data | email3_data | email4_data | email5_data | resposta | status`

Status possíveis: `em_campanha | respondeu | reativou | saiu | arquivado | deletado`
