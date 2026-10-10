import json
import sys
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "core" / "runtime-governance"))

from policy_engine import PolicyEngine, PolicyError, Rule, load_rules  # noqa: E402


class PolicyEngineTests(unittest.TestCase):
    def test_unknown_action_requires_review(self):
        decision = PolicyEngine().evaluate("delete_record")
        self.assertEqual(decision.outcome, "review")
        self.assertEqual(decision.reason, "no_matching_rule")
        self.assertEqual(decision.matched_rule_ids, ())

    def test_exact_allow_rule_allows_action(self):
        engine = PolicyEngine([Rule("allow-read", "read_report", "allow", "Read-only report")])
        decision = engine.evaluate("read_report")
        self.assertEqual(decision.outcome, "allow")
        self.assertEqual(decision.matched_rule_ids, ("allow-read",))

    def test_action_matching_is_exact_and_case_sensitive(self):
        engine = PolicyEngine([Rule("allow-read", "read_report", "allow", "Read-only report")])
        decision = engine.evaluate("READ_REPORT")
        self.assertEqual(decision.outcome, "review")
        self.assertEqual(decision.reason, "no_matching_rule")

    def test_deny_takes_precedence_over_allow(self):
        engine = PolicyEngine([
            Rule("allow-send", "send_email", "allow", "Example allow rule"),
            Rule("deny-send", "send_email", "deny", "Example deny rule"),
        ])
        decision = engine.evaluate("send_email")
        self.assertEqual(decision.outcome, "deny")
        self.assertEqual(decision.matched_rule_ids, ("allow-send", "deny-send"))

    def test_review_takes_precedence_over_allow(self):
        engine = PolicyEngine([
            Rule("allow-action", "export_data", "allow", "Example allow rule"),
            Rule("review-action", "export_data", "review", "Example review rule"),
        ])
        self.assertEqual(engine.evaluate("export_data").outcome, "review")

    def test_checked_in_empty_policy_reviews_unknown_actions(self):
        rules = load_rules(ROOT / "core" / "runtime-governance" / "policy_rules.json")
        self.assertEqual(rules, ())
        self.assertEqual(PolicyEngine(rules).evaluate("anything").outcome, "review")

    def test_loader_rejects_duplicate_rule_ids(self):
        rule = {
            "id": "rule-one",
            "action": "read_report",
            "effect": "allow",
            "description": "Test rule",
        }
        self.assert_invalid_policy({"policy_version": "1.0.0", "rules": [rule, rule]}, "duplicate rule id")

    def test_loader_rejects_invalid_effect(self):
        rule = {
            "id": "rule-one",
            "action": "read_report",
            "effect": "maybe",
            "description": "Test rule",
        }
        self.assert_invalid_policy({"policy_version": "1.0.0", "rules": [rule]}, "effect")

    def test_loader_rejects_duplicate_json_keys(self):
        raw = '{"policy_version":"1.0.0","rules":[],"rules":[]}'
        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir) / "policy.json"
            path.write_text(raw, encoding="utf-8")
            with self.assertRaisesRegex(PolicyError, "duplicate JSON key"):
                load_rules(path)

    def test_rejects_blank_action(self):
        with self.assertRaisesRegex(ValueError, "non-empty string"):
            PolicyEngine().evaluate("  ")

    def assert_invalid_policy(self, data, message):
        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir) / "policy.json"
            path.write_text(json.dumps(data), encoding="utf-8")
            with self.assertRaisesRegex(PolicyError, message):
                load_rules(path)


if __name__ == "__main__":
    unittest.main()
