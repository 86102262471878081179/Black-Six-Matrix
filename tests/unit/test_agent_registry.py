import json
import tempfile
import unittest
from pathlib import Path

from agents.registry import RegistryError, load_registry


ROOT = Path(__file__).resolve().parents[2]
REGISTRY = ROOT / "agents" / "agent_registry.json"


class AgentRegistryTests(unittest.TestCase):
    def test_loads_checked_in_registry(self):
        agents = load_registry(REGISTRY)
        self.assertEqual(len(agents), 1)
        self.assertEqual(agents[0].id, "compliance-governance-001")
        self.assertEqual(agents[0].status, "planned")

    def test_rejects_duplicate_ids(self):
        record = {
            "id": "example-agent",
            "name": "Example",
            "tier": "tier-1-governance",
            "status": "planned",
            "description": "Test record",
        }
        self.assert_invalid({"registry_version": "1.0.0", "agents": [record, record]}, "duplicate agent id")

    def test_rejects_unknown_tier(self):
        record = {
            "id": "example-agent",
            "name": "Example",
            "tier": "tier-4-unknown",
            "status": "planned",
            "description": "Test record",
        }
        self.assert_invalid({"registry_version": "1.0.0", "agents": [record]}, "tier")

    def test_rejects_unknown_status(self):
        record = {
            "id": "example-agent",
            "name": "Example",
            "tier": "tier-1-governance",
            "status": "running",
            "description": "Test record",
        }
        self.assert_invalid({"registry_version": "1.0.0", "agents": [record]}, "status")

    def test_rejects_duplicate_json_keys(self):
        raw = '{"registry_version":"1.0.0","agents":[],"agents":[]}'
        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir) / "registry.json"
            path.write_text(raw, encoding="utf-8")
            with self.assertRaisesRegex(RegistryError, "duplicate JSON key"):
                load_registry(path)

    def assert_invalid(self, data, message):
        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir) / "registry.json"
            path.write_text(json.dumps(data), encoding="utf-8")
            with self.assertRaisesRegex(RegistryError, message):
                load_registry(path)


if __name__ == "__main__":
    unittest.main()
