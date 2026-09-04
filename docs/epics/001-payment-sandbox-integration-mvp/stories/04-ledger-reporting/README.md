---
id: 001-story-04-ledger-reporting
aliases: []
tags:
  - payments
  - ledger
  - reporting
epic: 001-payment-sandbox-integration-mvp
---

# Story: Ledger and Reporting

## Intent

Add a simple ledger and reporting layer so Rails can reconcile local records against sandbox truth.

## Scope

### In Scope

- [ ] Double-entry ledger entries
- [x] Transaction report endpoint [Evidence](./TRANSACTION_REPORT_CONTRACT.md)
- [x] Balance projection [Evidence](./BALANCE_PROJECTION_CONTRACT.md)
- [x] Daily settlement projection [Evidence](./SETTLEMENT_PROJECTION_CONTRACT.md)
- [ ] Reconciliation snapshot export
- [ ] Immutable financial movements
- [ ] Reportable fees and refunds

### Out of Scope

- [ ] Bank settlement files
- [ ] Full accounting back office
- [ ] External accounting integrations
- [ ] Tax or invoice generation

## System Design

```mermaid
flowchart LR
    API[Go Sandbox API] --> LED[Ledger Engine]
    LED --> EN[Ledger Entries]
    EN --> REP[Transaction Report]
    REP --> RJ[Rails Reconciliation Job]
    LED --> BAL[Balance Projection]
```

## Explanation

1. Payment actions write immutable ledger entries.
2. The ledger engine validates debits and credits remain balanced.
3. The report endpoint exposes payment and balance state for reconciliation.
4. Rails compares the report with its own local projections.
5. Balances are derived from the ledger, not updated directly.
6. Fees, refunds, and reversals must remain traceable to their originating transaction.

## Contract Notes

- Ledger entries must be append-only.
- A balance is a projection of entries, not a primary source of truth.
- The transaction report must include enough data for reconciliation.
- Every report line must be traceable back to one or more ledger entries.
- Reconciliation snapshots should be exportable and comparable from Rails.

## Acceptance Criteria

- [ ] Every financial movement is reflected in the ledger.
- [x] The report can be consumed by Rails for reconciliation.
- [ ] Balances can be derived from entries.
- [ ] Ledger entries remain immutable.
- [ ] Fees and refunds are represented in the report.

## Dependencies

- [ ] Sandbox API base
- [ ] Rails adapter

## Related Docs

- [Transaction Report Contract](./TRANSACTION_REPORT_CONTRACT.md)
- [Balance Projection Contract](./BALANCE_PROJECTION_CONTRACT.md)
- [Settlement Projection Contract](./SETTLEMENT_PROJECTION_CONTRACT.md)
