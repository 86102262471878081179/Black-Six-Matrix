# Runtime Governance: Policy Engine

Dependency-free Python implementation of a small deterministic policy evaluator. Rules match **exact action names**; there are no wildcards or implicit grants.

## Decision behavior

- No matching rule → `review` (`no_matching_rule`).
- Matching rules resolve in this order: `deny` > `review` > `allow`.
- Rule IDs must be unique; invalid or ambiguous policy files fail to load.
- The checked-in `policy_rules.json` is intentionally empty, so every action currently requires review. No real-world action is approved by default.

## Try it

```bash
python3 core/runtime-governance/policy_engine.py "send_email"
python3 core/runtime-governance/policy_engine.py "send_email" --policy path/to/policy.json
python3 -m unittest discover -s tests/unit -p 'test_*.py' -v
```

Policy files use version `1.0.0`. Each rule requires `id`, `action`, `effect` (`allow`, `deny`, or `review`) and a non-empty `description`. Action matching is case-sensitive and exact. This component is a foundation only: it does not claim to contain the 125+ rules mentioned in the project README and it does not execute actions, authenticate callers, or write an audit ledger.
