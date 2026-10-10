"""Small, dependency-free loader and validator for the agent registry."""

from __future__ import annotations

import argparse
import json
import re
from dataclasses import dataclass
from pathlib import Path
from typing import Any

REGISTRY_VERSION = "1.0.0"
TIERS = {"tier-1-governance", "tier-2-security", "tier-3-analysis"}
STATUSES = {"planned", "active", "deprecated"}
_ID_PATTERN = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")
_REQUIRED_ROOT_KEYS = {"registry_version", "agents"}
_REQUIRED_AGENT_KEYS = {"id", "name", "tier", "status", "description"}


class RegistryError(ValueError):
    """Raised when registry JSON is malformed or violates the registry contract."""


@dataclass(frozen=True)
class Agent:
    id: str
    name: str
    tier: str
    status: str
    description: str


def _reject_duplicate_keys(pairs: list[tuple[str, Any]]) -> dict[str, Any]:
    result: dict[str, Any] = {}
    for key, value in pairs:
        if key in result:
            raise RegistryError(f"duplicate JSON key: {key!r}")
        result[key] = value
    return result


def _reject_nonstandard_constant(value: str) -> None:
    raise RegistryError(f"non-standard JSON constant: {value}")


def load_registry(path: str | Path) -> list[Agent]:
    """Read and validate a registry file, returning its agent records."""
    file_path = Path(path)
    try:
        raw = file_path.read_text(encoding="utf-8")
    except OSError as exc:
        raise RegistryError(f"cannot read registry {file_path}: {exc}") from exc

    try:
        document = json.loads(
            raw,
            object_pairs_hook=_reject_duplicate_keys,
            parse_constant=_reject_nonstandard_constant,
        )
    except json.JSONDecodeError as exc:
        raise RegistryError(f"invalid JSON at line {exc.lineno}, column {exc.colno}: {exc.msg}") from exc

    if not isinstance(document, dict):
        raise RegistryError("registry root must be a JSON object")
    if set(document) != _REQUIRED_ROOT_KEYS:
        missing = sorted(_REQUIRED_ROOT_KEYS - set(document))
        extra = sorted(set(document) - _REQUIRED_ROOT_KEYS)
        raise RegistryError(f"registry root keys mismatch (missing={missing}, extra={extra})")
    if document["registry_version"] != REGISTRY_VERSION:
        raise RegistryError(f"registry_version must be {REGISTRY_VERSION!r}")
    if not isinstance(document["agents"], list):
        raise RegistryError("'agents' must be an array")

    agents: list[Agent] = []
    seen_ids: set[str] = set()
    for index, item in enumerate(document["agents"]):
        location = f"agents[{index}]"
        if not isinstance(item, dict):
            raise RegistryError(f"{location} must be a JSON object")
        if set(item) != _REQUIRED_AGENT_KEYS:
            missing = sorted(_REQUIRED_AGENT_KEYS - set(item))
            extra = sorted(set(item) - _REQUIRED_AGENT_KEYS)
            raise RegistryError(f"{location} keys mismatch (missing={missing}, extra={extra})")

        for field in ("id", "name", "tier", "status", "description"):
            value = item[field]
            if not isinstance(value, str) or not value.strip():
                raise RegistryError(f"{location}.{field} must be a non-empty string")

        agent_id = item["id"]
        if not _ID_PATTERN.fullmatch(agent_id):
            raise RegistryError(f"{location}.id must be lowercase kebab-case")
        if agent_id in seen_ids:
            raise RegistryError(f"duplicate agent id: {agent_id!r}")
        seen_ids.add(agent_id)

        if item["tier"] not in TIERS:
            raise RegistryError(f"{location}.tier must be one of {sorted(TIERS)}")
        if item["status"] not in STATUSES:
            raise RegistryError(f"{location}.status must be one of {sorted(STATUSES)}")

        agents.append(Agent(**item))
    return agents


def main() -> int:
    parser = argparse.ArgumentParser(description="Validate the Black-Six-Matrix agent registry.")
    parser.add_argument(
        "registry",
        nargs="?",
        type=Path,
        default=Path(__file__).with_name("agent_registry.json"),
        help="path to registry JSON (defaults to agents/agent_registry.json)",
    )
    args = parser.parse_args()
    try:
        agents = load_registry(args.registry)
    except RegistryError as exc:
        parser.exit(1, f"Invalid registry: {exc}\n")
    print(f"Valid registry: {len(agents)} agent(s)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
