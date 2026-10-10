# Agent Registry

`agent_registry.json` ist das zentrale Verzeichnis der Agenten. Jeder Eintrag enthält eine stabile ID, einen Namen, ein Tier, einen Lebenszyklusstatus und eine Beschreibung. Der erste Eintrag ist nur ein geplanter Platzhalter; er führt keinen Agenten aus.

## Validieren

Mit Python 3 (keine Drittanbieterpakete erforderlich):

```bash
python3 -m agents.registry
python3 -m unittest discover -s tests/unit -p 'test_*.py' -v
```

Für eine andere Registerdatei kann der Pfad als Argument angegeben werden:

```bash
python3 -m agents.registry path/to/agent_registry.json
```

Das JSON-Schema für Editor- und CI-Integration liegt in `agent_registry.schema.json`. Zulässige Tiers: `tier-1-governance`, `tier-2-security`, `tier-3-analysis`. Zulässige Status: `planned`, `active`, `deprecated`.
