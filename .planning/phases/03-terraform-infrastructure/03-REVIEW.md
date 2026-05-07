---
phase: 3
slug: terraform-infrastructure
status: clean
files_reviewed: 6
critical: 0
warning: 0
info: 2
total: 2
created: 2026-05-07
---

# Phase3 — Code Review (Wave 2)

## Summary

- **Files reviewed:** 6 (aws-eks: main.tf, variables.tf, outputs.tf; azure-aks: main.tf, variables.tf, outputs.tf)
- **Depth:** standard
- **Result:** Clean (0 critical, 0 warning, 2 info)

---

## Findings

### INFO-01 — Node group max_size fixed offset (Info)

**File:** `infra/modules/aws-eks/main.tf`  
**Line:** ~line 70 (scaling_config block)

**Note:** `max_size = var.node_count + 2` uses a fixed offset. Acceptable for PoC; for production consider using a variable or dynamic scaling policy.

---

### INFO-02 — AKS zones hard-coded (Info)

**File:** `infra/modules/azure-aks/main.tf`  
**Line:** ~line 32 (`zones = ["1","2","3"]`)

**Note:** Hard-coded zones may not be available in all Azure regions. Consider making this a variable with region-specific defaults.

---

## Secure Behaviors Verified

| Threat ID | Behavior | Status |
|-----------|----------|--------|
| T-03-03-01 | EKS IRSA trust policy (StringEquals, no wildcards) | ✅ Verified |
| T-03-03-02 | EKS private API endpoint | ✅ Verified |
| T-03-03-03 | EKS audit logging enabled | ✅ Verified |
| T-03-03-04 | OIDC thumbprint dynamic via tls_certificate | ✅ Verified |
| T-03-03-05 | cluster_ca output marked sensitive | ✅ Verified |
| T-03-04-01 | AKS Workload Identity + federated credential | ✅ Verified |
| T-03-04-02 | AKS private node pool (vnet_subnet_id) | ✅ Verified |

---

## Next Steps

- No critical or warning findings to fix.
- Informational notes are for awareness only; no immediate action required.
- Full report saved to `03-REVIEW.md`.

---

**Approval:** approved 2026-05-07
