# Projektstruktur

Die Ordner bilden die in der README beschriebene Zielarchitektur ab. Noch nicht implementierte Bereiche sind zunächst leer und durch `.gitkeep` im Repository sichtbar; die Ordneranlage bedeutet nicht, dass dort bereits Komponenten vorhanden sind.

## Bereiche

- `architecture/` – Architekturübersichten und Modulbeschreibungen
- `core/` – Runtime-Governance, Mesh-Sicherheit, Audit-Ledger und Reasoning
- `products/` – Produktmodule
- `agents/` – Agenten-Tiers und späteres Registry
- `infrastructure/` – Kubernetes-, Docker- und Terraform-Konfiguration
- `frontend/` – Web-App und Dashboard
- `tests/` – Unit-, Integrations- und Sicherheitstests
- `docs/` – API-, Compliance- und Deployment-Dokumentation
- `examples/` – Compliance-, Sicherheits- und Agentenbeispiele
- `assets/` – Bilder, Logos und Branding
- `.github/workflows/` – GitHub-Actions-Workflows

Bestehende Dateien im Repository-Stamm bleiben an Ort und Stelle. Insbesondere bleiben `index.html` und `index.json` unverändert, damit die vorhandene Website-Konfiguration nicht durch das Gerüst beeinträchtigt wird.
