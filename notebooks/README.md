# Notebook-Bestand: Ordnerstruktur und Arbeitsablauf

Dieses Verzeichnis ist die **kuratierte Projektansicht**, nicht der Speicherort für den ungeprüften 78-Notebook-Originalbestand. Das Manifest ist ein leeres Erfassungsgerüst mit den IDs `NB-001` bis `NB-078`; Titel, Formate, Inhalte, Eigentümer und Klassifikationen wurden nicht aus Screenshots abgeleitet.

## Verzeichnisbaum

```text
notebooks/
├── README.md
├── manifest-template.csv                 # 78 leere Erfassungsdatensätze
├── schemas/
│   └── manifest-entry.schema.json        # Vertrag für einen Manifestdatensatz
├── curated/                              # nur geprüfte, freigegebene Ableitungen
│   ├── governance/
│   ├── security/
│   ├── architecture/
│   ├── research/
│   └── operations/
├── inbox/                                # lokal/geschützt, niemals committen
└── derived/                              # temporäre Exporte/Renderings, nicht committen
```

Die Originale verbleiben im kontrollierten Evidence Vault. `inbox/` und `derived/` sind optionale lokale Arbeitsbereiche und werden durch `.gitignore` vom Repository ausgeschlossen. Sie sollten nicht zur dauerhaften Ablage von Originalen verwendet werden.

## Ablage eines freigegebenen Eintrags

Nach Sichtung erhält ein kuratierter Eintrag eine eigene Ablage:

```text
notebooks/curated/<bereich>/NB-001-<kurzer-slug>/
├── README.md              # Quelle, Zweck, Freigabe und Verknüpfungen
└── <kuratierte-datei>     # freigegebene .ipynb, .md oder andere passende Form
```

Der Bereich muss einer der fünf kuratierten Ordner sein. Das Format wird erst nach Prüfung des Originals festgelegt; „Notebook“ wird nicht automatisch als Jupyter-Datei interpretiert. Die Originaldatei wird nicht überschrieben oder durch eine Konvertierung ersetzt.

## Manifest-Felder

- `notebook_id`: stabile interne ID `NB-001` … `NB-078`; nicht aus Dateinamen ableiten.
- `title`, `source_filename`, `source_format`, `source_location`, `source_system`: erst nach Inventur befüllen. `source_location` soll auf einen geschützten Speicherverweis zeigen, keine öffentlich erreichbare URL.
- `project_area`: zunächst `unknown`; nach Triage z. B. `governance`, `security`, `architecture`, `research` oder `operations`.
- `confidentiality`: zunächst `unclassified`; zulässige Einstufungen sind `public`, `internal`, `confidential`, `restricted` und `unknown`.
- `source_sha256`: Prüfsumme des unveränderten Originals; erst nach dem Abruf selbst berechnen.
- `curated_sha256`: getrennte Prüfsumme für die freigegebene Ableitung, falls eine solche entsteht.
- `review_status`: startet als `untriaged`; danach `in_review`, `curated`, `restricted`, `duplicate` oder `excluded`.
- `curated_path`: Git-Pfad nur für freigegebene, bereinigte Ableitungen. Für Originale leer lassen.
- `owner`, Zeitstempel, `reviewed_by`, `reviewed_at`, `license_rights`, `related_artifact_ids` und `notes`: nur mit verifizierten Angaben befüllen.

Leere Zellen bedeuten **nicht erhoben**. Keine unbekannten Metadaten schätzen. Niemals Geheimnisse, API-Schlüssel, personenbezogene Daten oder ungeprüfte Inhalte in das öffentliche Repository übernehmen.

## Aufnahmeprozess

1. Originale unverändert im zugriffsbeschränkten Vault belassen; dort Quelle, Zugriffsrechte und Aufbewahrung festlegen.
2. Inventar erstellen: ID, tatsächlicher Dateiname und Format, Dateigröße, Quelle und SHA-256. Duplikate erkennen, nicht löschen.
3. Inhalt und Rechte prüfen; Geheimnisse/PII kennzeichnen. Einstufung und Triage-Status aktualisieren.
4. Nur notwendige, bereinigte und freigegebene Ableitungen nach `curated/<bereich>/` kopieren.
5. Transformation dokumentieren; Quell- und Ableitungsprüfsummen getrennt halten.
6. Freigegebene Erkenntnisse in `core/`, `docs/`, Tests oder Beispielen neu implementieren. Das Notebook bleibt Nachweis/Arbeitsmaterial, nicht automatisch Produktionscode.
7. Änderungen reviewen und committen; Originale sowie sensible/temporäre Daten nie mit Git synchronisieren.

## Noch offen

Vor einer echten Massenaufnahme müssen die 78 Dateien mit Dateiliste oder Export inventarisiert werden. Format, Eigentümer, Lizenz, Vertraulichkeitsstufe und Storage-Anbieter sind derzeit unbekannt. Das Template erzeugt keine Aussage über den Inhalt der Dateien.
