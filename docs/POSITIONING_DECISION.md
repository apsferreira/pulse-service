# ADR-001: Pulse Positioning — SaaS vs Self-Hosted vs Open Source

**Status:** Proposed
**Date:** 2026-04-11
**Author:** Growth Agent (IIT)
**BKL:** BKL-556

---

## Context

Pulse is IIT's product analytics and feature flag platform. As it approaches production readiness, a fundamental go-to-market question must be answered: what deployment model will drive the most sustainable growth for the IIT team?

The competitive landscape includes three clear archetypes:

| Competitor | Model | Key insight |
|---|---|---|
| PostHog | Open source + optional cloud | Community-driven adoption; self-host = free; cloud = revenue |
| LaunchDarkly | Enterprise SaaS only | Deep integrations, compliance focus; high ACV |
| GrowthBook | Open source + cloud | Smaller team than PostHog; hybrid GTM |
| Mixpanel | Cloud SaaS | Analytics-only; no feature flags |
| Amplitude | Cloud SaaS | Analytics-only; enterprise positioning |

PostHog's strategy is worth studying closely: open source generates community trust and organic inbound, while the cloud tier converts users who hit operational limits. However, PostHog has a large engineering team. Executing both tracks simultaneously with a small team leads to split focus and delayed shipping.

IIT's current constraints:
- Small team (1 engineering manager + contract engineers)
- No dedicated DevOps headcount for supporting self-hosted installations
- Revenue is needed to fund product development
- Time-to-market matters more than marginal adoption via open source

---

## Decision

**Pulse will launch as SaaS-first, with a self-hosted / open-core tier planned for a future Enterprise phase.**

The initial product will be cloud-only. There is no open source release at launch.

---

## Rationale

### 1. Market velocity

A SaaS product ships faster. There is no need to build and maintain Docker Compose stacks, Helm charts, upgrade migration guides or a support pipeline for self-hosted deployments. Every engineering hour goes into product features that paying customers use directly.

### 2. IIT is a small team

Supporting self-hosted users is a non-trivial operational commitment. Self-hosters hit database edge cases, misconfigure environment variables, run outdated versions and require assistance that is indistinguishable from paid support — but without revenue. A two-person core team cannot sustain that.

### 3. Self-hosted cannibalizes SaaS revenue at early stage

For a platform at Pulse's stage (pre-PMF), self-hosted is a revenue drain, not a funnel. Users with the technical ability to self-host are also the most price-sensitive and least likely to convert to a paid cloud plan. The conversion from open source to cloud is a long-term play that requires years of ecosystem building. IIT needs revenue now.

### 4. Compliance customers require self-hosted — but that is an Enterprise conversation

Enterprise customers in regulated industries (healthcare, financial services, government) often require data residency and cannot use third-party cloud services. This is a real need — but it is not the IIT target customer at launch. When IIT is ready to pursue Enterprise, a self-hosted option can be offered as a paid tier with an SLA and dedicated support, not as a free community download.

---

## Consequences

### Immediate

- Pulse roadmap focuses exclusively on cloud product quality, reliability and onboarding.
- No open source repository is created or maintained in 2026.
- All infrastructure investments go toward IIT-hosted cloud (K3s on Proxmox / AWS).

### Medium term (6–12 months)

- A public REST API with clear versioning will be designed from the start. This investment pays off in two ways: (a) it accelerates enterprise adoption when self-hosted is offered, and (b) it enables integrations with third-party tools that drive inbound.
- SDK quality (JavaScript, Go) becomes a key differentiator — better than GrowthBook, approaching PostHog.

### Long term (12+ months)

- **Enterprise tier:** Self-hosted or on-premise deployment, offered as a paid SKU with SLA, onboarding and support. Price: contact sales (minimum R$3.000/month).
- **Open core consideration:** If Pulse gains sufficient community traction, specific modules (e.g., the SDK client libraries) may be open-sourced to drive adoption without giving away the analytics backend.

### Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Developer audience prefers open source competitors | Medium | Invest in documentation, DX and SDK quality — compete on experience, not price |
| Enterprise customer blocked by data residency requirement | Low at launch | Defer to Enterprise tier roadmap; flag as known limitation on pricing page |
| PostHog free tier undercuts Pulse on price | High | Compete on simplicity, Brazilian market fit, and IIT ecosystem integration, not feature parity |

---

## Alternatives Considered

### Open source from day one (rejected)

Would generate goodwill and inbound SEO, but would require maintaining a public repository, responding to issues, building upgrade guides and handling community support — all before achieving product-market fit. The overhead is disproportionate to the team size.

### Hybrid: SaaS + free self-hosted simultaneously (rejected)

Splits engineering focus. The SaaS product suffers as effort goes to Helm charts and Docker images. Revenue pressure from early customers demands full attention on the cloud product.

### Enterprise SaaS only (partially considered)

Pulse could target only large customers from day one. Rejected because IIT's current distribution strength is with SMBs and indie developers. The sales cycle for enterprise is long and requires a dedicated sales function that IIT does not have.

---

## Review Trigger

This decision should be revisited when any of the following occurs:

- Pulse reaches 50 paying tenants on the cloud plan
- An inbound enterprise opportunity requires self-hosted as a condition
- A strategic partnership makes open source distribution attractive (e.g., CNCF, GDG ecosystem)

---

**Decision owner:** Antonio Pedro Ferreira (CEO, IIT)
**Next review date:** 2026-10-01
