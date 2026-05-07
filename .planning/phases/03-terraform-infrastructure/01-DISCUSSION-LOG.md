# Phase 3: Terraform Infrastructure - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-06
**Phase:** 03-terraform-infrastructure
**Areas discussed:** Environment strategy, Node sizing, State bootstrap strategy, CosmosDB consistency

---

## Environment Strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Single 'dev' only | One environment, lowest cost, sufficient for PoC | |
| dev + prod workspaces | Two workspaces, mirrors landing-zones OUs | |
| dev + hom + prod | Three environments (user-provided) | ✓ |

**User's choice:** dev + hom + prod (via free-text — added homologation environment)
**Notes:** "hom" = homologation (staging/QA). Three full environments across both clouds.

---

### Environment variable storage

| Option | Description | Selected |
|--------|-------------|----------|
| tfvars files | One .tfvars per environment, clean diff | ✓ |
| Terraform workspaces | Single codebase, workspace variables | |
| You decide | Claude picks | |

**User's choice:** tfvars files

---

### State backend isolation

| Option | Description | Selected |
|--------|-------------|----------|
| Separate backends per env | Own S3 bucket + Azure container per env, full isolation | ✓ |
| Shared backend, prefixed keys | One backend, keys like dev/terraform.tfstate | |

**User's choice:** Separate backends per environment

---

## Node Sizing (EKS / AKS)

### Instance size

| Option | Description | Selected |
|--------|-------------|----------|
| Small — t3.medium / Standard_D2s_v3 | 2 vCPU, 4 GB RAM, ~$30-40/mo per node | ✓ |
| Minimal — t3.small / Standard_B2s | 2 vCPU, 2 GB RAM, cheapest | |
| Medium — m5.large / Standard_D4s_v3 | 4 vCPU, 8-16 GB RAM, ~$70-100/mo | |

**User's choice:** t3.medium / Standard_D2s_v3

---

### Node count

| Option | Description | Selected |
|--------|-------------|----------|
| 2 nodes minimum | Scheduling resilience, required for HA controllers | ✓ |
| 1 node (cheapest) | Single point of failure | |
| 3 nodes (multi-AZ) | One per AZ, most realistic | |

**User's choice:** 2 nodes minimum

---

## State Bootstrap Strategy

### Bootstrap mechanism

| Option | Description | Selected |
|--------|-------------|----------|
| Bootstrap module with local state | infra/bootstrap/ creates backends via local state, then migrate | ✓ |
| Manual pre-creation | AWS CLI + az CLI commands, documented | |
| CI handles it | Assumes backends exist, CI creates them | |

**User's choice:** Bootstrap module with local state

---

### Bootstrap scope

| Option | Description | Selected |
|--------|-------------|----------|
| Per-environment | Separate bootstrap per dev / hom / prod | ✓ |
| Shared single bootstrap | One bootstrap for all envs | |

**User's choice:** Per-environment

---

## CosmosDB Consistency

### Consistency level

| Option | Description | Selected |
|--------|-------------|----------|
| Session | Read-your-writes per session, default, no extra cost | ✓ |
| Bounded Staleness | Consistent within time window, more expensive | |
| Eventual | Lowest cost/latency, no ordering guarantee | |

**User's choice:** Session

---

### Geo-redundancy

| Option | Description | Selected |
|--------|-------------|----------|
| No — single region | Simpler, significantly cheaper for PoC | ✓ |
| Yes — geo-redundant reads | One write + one read replica, adds cost | |

**User's choice:** Single region, no geo-redundancy

---

## Claude's Discretion

- CIDR ranges for VPC/VNet/subnets — Claude picks non-overlapping RFC1918 ranges
- Kubernetes version — Claude pins to latest stable EKS/AKS-supported version
- Terraform and provider version pins — Claude pins to current stable releases

## Deferred Ideas

None mentioned.
