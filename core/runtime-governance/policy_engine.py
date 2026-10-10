"""Dependency-free deterministic policy evaluator for exact action names."""

from __future__ import annotations

import argparse
import json
import re
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any

POLICY_VERSION = "1.0.0"
EFFECTS = {"allow", "deny", "review"}
_RULE_ID = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")
_ROOT_KEYS = {"policy_version", "rules"}
_RULE_KEYS = {"id", "action", "effect", "description"}


class PolicyError(ValueError):
    """Raised when a policy file is malformed or violates the policy contract."""


@dataclass(frozen=True)
class Rule:
    id: str
    action: str
    effect: str
    description: str


@dataclass(frozen=True)
class Decision:
    action: str
    outcome: str
    reason: str
    matched_rule_ids: tuple[str, ...]

    def to_dict(self) -> dict[str, Any]:
        result = asdict(self)
        result["matched_rule_ids"] = list(self.matched_rule_ids)
        return result


def _reject_duplicate_keys(pairs: list[tuple[str, Any]]) -> dict[str, Any]:
    result: dict[str, Any] = {}
    for key, value in pairs:
        if key in result:
            raise PolicyError(f"duplicate JSON key: {key!r}")
        result[key] = value
    return result


def _reject_nonstandard_constant(value: str) -> None:
    raise PolicyError(f"non-standard JSON constant: {value}")


def load_rules(path: str | Path) -> tuple[Rule, ...]:
    """Load and strictly validate a policy file."""
    policy_path = Path(path)
    try:
        raw = policy_path.read_text(encoding="utf-8")
    except OSError as exc:
        raise PolicyError(f"cannot read policy {policy_path}: {exc}") from exc

    try:
        document = json.loads(
            raw,
            object_pairs_hook=_reject_duplicate_keys,
            parse_constant=_reject_nonstandard_constant,
        )
    except json.JSONDecodeError as exc:
        raise PolicyError(f"invalid JSON at line {exc.lineno}, column {exc.colno}: {exc.msg}") from exc

    if not isinstance(document, dict):
        raise PolicyError("policy root must be a JSON object")
    if set(document) != _ROOT_KEYS:
        missing = sorted(_ROOT_KEYS - set(document))
        extra = sorted(set(document) - _ROOT_KEYS)
        raise PolicyError(f"policy root keys mismatch (missing={missing}, extra={extra})")
    if document["policy_version"] != POLICY_VERSION:
        raise PolicyError(f"policy_version must be {POLICY_VERSION!r}")
    if not isinstance(document["rules"], list):
        raise PolicyError("'rules' must be an array")

    rules: list[Rule] = []
    seen_ids: set[str] = set()
    for index, item in enumerate(document["rules"]):
        location = f"rules[{index}]"
        if not isinstance(item, dict):
            raise PolicyError(f"{location} must be a JSON object")
        if set(item) != _RULE_KEYS:
            missing = sorted(_RULE_KEYS - set(item))
            extra = sorted(set(item) - _RULE_KEYS)
            raise PolicyError(f"{location} keys mismatch (missing={missing}, extra={extra})")
        for field in ("id", "action", "effect", "description"):
            value = item[field]
            if not isinstance(value, str) or not value.strip():
                raise PolicyError(f"{location}.{field} must be a non-empty string")
        if not _RULE_ID.fullmatch(item["id"]):
            raise PolicyError(f"{location}.id must be lowercase kebab-case")
        if item["id"] in seen_ids:
            raise PolicyError(f"duplicate rule id: {item['id']!r}")
        seen_ids.add(item["id"])
        if item["effect"] not in EFFECTS:
            raise PolicyError(f"{location}.effect must be one of {sorted(EFFECTS)}")
        rules.append(Rule(**item))
    return tuple(rules)


class PolicyEngine:
    """Evaluate actions using exact-match rules and fail-closed conflict handling.

    Resolution is deterministic: deny takes precedence over review, which takes
    precedence over allow. An action without a matching rule requires review.
    """

    def __init__(self, rules: tuple[Rule, ...] | list[Rule] = ()) -> None:
        self._rules = tuple(rules)

    def evaluate(self, action: str) -> Decision:
        if not isinstance(action, str) or not action.strip():
            raise ValueError("action must be a non-empty string")

        matches = tuple(rule for rule in self._rules if rule.action == action)
        ids = tuple(rule.id for rule in matches)
        if not matches:
            return Decision(action, "review", "no_matching_rule", ())

        effects = {rule.effect for rule in matches}
        if "deny" in effects:
            return Decision(action, "deny", "deny_rule_matched", ids)
        if "review" in effects:
            return Decision(action, "review", "review_rule_matched", ids)
        return Decision(action, "allow", "allow_rule_matched", ids)


def main() -> int:
    parser = argparse.ArgumentParser(description="Evaluate one action against the policy rules.")
    parser.add_argument("action", help="exact action name to evaluate")
    parser.add_argument(
        "--policy",
        type=Path,
        default=Path(__file__).with_name("policy_rules.json"),
        help="policy JSON file (defaults to policy_rules.json beside this module)",
    )
    args = parser.parse_args()
    try:
        engine = PolicyEngine(load_rules(args.policy))
        decision = engine.evaluate(args.action)
    except (PolicyError, ValueError) as exc:
        parser.exit(2, f"Policy evaluation failed: {exc}\n")
    print(json.dumps(decision.to_dict(), ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
