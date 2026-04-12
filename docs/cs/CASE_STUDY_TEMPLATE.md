# Pulse CS — Template de Case Study

> BKL-1088 | Versão 1.0 | Abril 2026
> Usar este template para documentar histórias reais de clientes e criar materiais de marketing/vendas

---

## Template

```markdown
# [NOME DA EMPRESA] + Pulse

**Setor:** [ex: SaaS B2B, E-commerce, Agência Digital]
**Tamanho:** [ex: 8 pessoas, 2 produtos ativos]
**Plano Pulse:** [Growth / Pro / Custom]
**Tempo de uso:** [ex: 4 meses]

---

## O Problema

[2-3 parágrafos descrevendo a situação antes do Pulse.
Ser específico: qual dado não tinha visibilidade? Qual decisão estava sendo tomada no escuro?
Evitar generalidades — "não tínhamos analytics" não convence. "Lançamos uma feature nova e não sabíamos se alguém estava usando" convence.]

## A Solução

[Como implementaram o Pulse. Quantos eventos configuraram? Quais features usaram primeiro?
Incluir detalhe técnico leve — mostra que é real.]

## Os Resultados

| Métrica | Antes | Depois |
|---------|-------|--------|
| [ex: Tempo para detectar bug em feature] | [ex: 3-5 dias (via suporte)] | [ex: < 4 horas (via alerta Pulse)] |
| [ex: % de usuários usando feature X] | [ex: desconhecido] | [ex: 23% — abaixo do esperado] |
| [ex: Decisões de produto baseadas em dado] | [ex: ~20%] | [ex: ~80%] |

[Parágrafo narrativo explicando os números — contexto importa mais que o número sozinho.]

## O que o Cliente Diz

> "[Quote direto — mínimo 2 frases. Deve soar humano, não press release.
> Incluir algo específico que o Pulse fez que foi surpreendente ou inesperado.]"
>
> **— [Nome], [Cargo], [Empresa]**

## Próximos Passos

[O que o cliente planeja fazer a seguir com o Pulse. Mostra que a história continua — não é um projeto acabado.]
```

---

## Exemplo 1: Agência Focal — SocialMake + Pulse

**Setor:** Agência de Marketing Digital
**Tamanho:** 11 pessoas, 23 clientes ativos
**Plano Pulse:** Growth
**Tempo de uso:** 3 meses

---

### O Problema

A Agência Focal cresceu rápido: em 18 meses, passou de 5 para 23 clientes ativos gerenciando redes sociais. O problema é que cresceram na ferramenta errada — usavam planilhas para controlar o que havia sido postado, quando, e com qual resultado.

Quando um cliente perguntava "quantos posts publicamos em março?" a resposta demorava um dia. Quando perguntava "qual post teve mais engajamento?", a resposta era manual, inconsistente entre analistas, e às vezes errada.

A parceira Carla, responsável pela gestão de contas, descreveu assim: "A gente passou a maior parte do tempo do report mensal coletando dados, não analisando. Era planilha no Excel, print de Instagram, print de LinkedIn. Levava dois dias por cliente."

### A Solução

A Focal implementou o SocialMake para agendamento e o Pulse para rastrear a cadência de publicação. Configuraram 4 eventos: `post_published`, `post_scheduled`, `account_connected` e `report_generated`. Em uma tarde de trabalho, o analista responsável pela integração conectou os dois sistemas e criou o primeiro dashboard.

A primeira versão rastreava apenas volume de publicações por cliente. Nas semanas seguintes, adicionaram métricas de aprovação (quantos posts o cliente aprovou sem revisão), e descobriram que um dos clientes — uma rede de academias — aprovava 98% dos posts sem comentário, enquanto outro cliente rejeitava 60% na primeira versão. Esse dado virou pauta na próxima reunião de operações.

### Os Resultados

| Métrica | Antes | Depois |
|---------|-------|--------|
| Tempo de report mensal por cliente | 2 dias | 2 horas |
| Erros de dado no report | ~3 por mês | 0 |
| Visibilidade de clientes com baixa taxa de aprovação | 0 | Identificados em tempo real |
| Tempo para responder "quantos posts publicamos?" | 1 dia | 30 segundos |

O dado mais valioso não foi o tempo poupado — foi a descoberta de padrões que antes eram invisíveis. A Focal identificou que dois clientes no plano básico consumiam 40% do tempo de operação da equipe por causa do processo de aprovação. Com esse dado, renegociaram os contratos.

### O que o Cliente Diz

> "Antes eu achava que analytics era coisa para empresa grande. O Pulse me mostrou que analytics é para quem precisa tomar decisão com menos achismo — e isso somos nós, todo dia. O report que levava dois dias agora eu apresento ao vivo na reunião com o cliente, atualizando na tela em tempo real. Isso mudou como o cliente nos vê."
>
> **— Carla Mendonça, Sócia-Operacional, Agência Focal**

### Próximos Passos

A Focal está avaliando o plano Pro para adicionar alertas automáticos: quando um cliente ficar 5 dias sem post publicado, o gestor de conta recebe um aviso antes que o cliente precise ligar.

---

## Exemplo 2: Construtech Modular — Nitro + Pulse

**Setor:** Construtora de pequeno porte (gestão de projetos internos)
**Tamanho:** 18 pessoas, 4 obras simultâneas
**Plano Pulse:** Growth
**Tempo de uso:** 5 meses

---

### O Problema

A equipe de gestão da Construtech Modular adotou o Nitro para controlar as tarefas das obras. O problema que surgiu 60 dias depois: ninguém sabia se a equipe estava realmente usando. Os gerentes de obra criavam tarefas no início do projeto e depois sumiam do sistema.

O diretor de operações, Marcos, fez a pergunta certa numa reunião de resultado: "A gente está pagando por uma ferramenta que o time usa ou que o time ignora?" Não havia resposta.

A empresa não tinha como saber quantas sessões de trabalho eram feitas por dia, quais times eram ativos, e quais projetos estavam com engajamento zero na ferramenta. O resultado prático: decisões de processo eram feitas com base em impressão, não em dado.

### A Solução

O time de TI interno conectou o Nitro ao Pulse em uma tarde, instrumentando os eventos `session_started`, `task_created` e `task_completed`. Não fizeram dashboard sofisticado na primeira semana — só um painel simples: usuários ativos por semana, tarefas criadas vs. concluídas.

O primeiro dado que apareceu foi revelador: dos 18 usuários cadastrados, apenas 9 haviam logado nos últimos 30 dias. E dos 9, 3 concentravam 70% das tarefas concluídas.

### Os Resultados

| Métrica | Antes | Depois |
|---------|-------|--------|
| Visibilidade de adoção por usuário | Zero | 100% rastreado |
| Usuários ativos / cadastrados | Desconhecido | 50% → meta: 80% |
| Tarefas concluídas por semana | Desconhecido | +34% após onboarding refeito |
| Tempo para identificar equipe com baixa adoção | Não identificado | < 1 semana |

Com os dados em mãos, Marcos convocou onboarding individual para os 9 usuários inativos. Dos 9, 6 passaram a usar ativamente após o treinamento. Os outros 3 preferiram continuar com WhatsApp — e com esse dado, a gestão aceitou que a ferramenta tinha um teto de adoção interno, e parou de forçar o uso onde ele não fazia sentido.

"O Pulse não resolveu nosso problema de adoção. Ele nos mostrou qual era o problema real: não era a ferramenta, eram 3 pessoas específicas que tinham um processo de trabalho diferente. Isso mudou a conversa."

### O que o Cliente Diz

> "Eu não precisava de um Mixpanel cheio de funil e cohort. Precisava saber se meu time estava usando o que a gente comprou. O Pulse respondeu isso em dois dias, sem precisar de engenheiro. E agora eu consigo mostrar pro meu diretor financeiro, em número, que o investimento na ferramenta está gerando resultado. Antes eu falava 'acho que está funcionando'. Agora eu mostro o gráfico."
>
> **— Marcos Alves, Diretor de Operações, Construtech Modular**

### Próximos Passos

A Construtech está avaliando usar o Pulse para rastrear também o Brio (atendimento ao cliente dos moradores dos empreendimentos). O objetivo é entender quais tipos de solicitação chegam com mais frequência e usar esse dado para alimentar o treinamento do agente de IA.

---

## Orientações para Coletar Cases Reais

1. **Timing certo:** abordar o cliente entre o 2º e 4º mês de uso — novo o suficiente para lembrar o "antes", experiente o suficiente para ter resultados concretos

2. **Entrevista de 20 minutos:** perguntas-chave:
   - "O que você tentou resolver antes do Pulse?"
   - "Qual foi o primeiro dado que apareceu e te surpreendeu?"
   - "Qual decisão você tomou baseada num dado do Pulse?"
   - "Como você descreveria o Pulse para alguém que nunca ouviu falar?"

3. **Nunca inventar números:** se o cliente não tem o dado preciso, usar "estimativa do cliente" ou apenas narrativa qualitativa — credibilidade vale mais que número bonito

4. **Aprovação antes de publicar:** enviar o rascunho ao cliente, aguardar OK escrito (email basta)

5. **Canais de publicação:** blog IIT, LinkedIn do Antonio, página de cases no site do Pulse
