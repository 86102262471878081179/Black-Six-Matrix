# Evidence Vault: Zielarchitektur und Repository-Grenze

## Grundsatz

GitHub enthält Quellcode, Dokumentation und **freigegebene, kleine Metadaten**. Es ist nicht der Ablageort des vollständigen Evidence Vaults. Der gemeldete Bestand enthält mindestens tausende TXT-/PDF-Objekte sowie weitere Medien und Archive; die sichtbaren Zählungen ergeben nicht zwingend einen vollständigen oder disjunkten Gesamtbestand. Die Rohdateien werden daher nicht in dieses Repository kopiert.

## Logische Bereiche

```text
Evidence Vault (separater, zugriffsbeschränkter Speicher)
├── originals/       # unveränderte Originale, unveränderlicher/versionsgebundener Bestand
├── derivatives/     # geprüfte Exporte, OCR oder bereinigte Fassungen
├── manifests/       # Inventare und Zuordnung Quelle ↔ Prüfsumme
└── audit/           # Zugriffe, Importe, Verifikation und Aufbewahrungsereignisse

Black-Six-Matrix GitHub
├── notebooks/       # Template + freigegebene, kuratierte Inhalte
├── docs/evidence-vault/
│   └── ARCHITECTURE.md
├── core/cosmic-memory/  # spätere Audit-/Prüfsummen-Integration, keine Rohdokumente
└── tools/evidence-ingest/  # künftiger kontrollierter Import/Verifikator
```

Die Baumdarstellung ist ein **logisches Modell**, kein bereits bereitgestellter Storage-Dienst. Der Anbieter, Zugriffskontrollen, Verschlüsselung, Aufbewahrungsfristen und Wiederherstellung müssen vor produktiver Nutzung festgelegt werden.

## Datenfluss

1. **Inventarisieren:** Dateiliste aus der Quelle exportieren, stabile ID `NB-001` bis `NB-078` bzw. eigene Evidence-ID zuweisen. Dateiendung nicht als Inhaltstyp missverstehen.
2. **Original sichern:** Original unverändert in den geschützten Vault aufnehmen. Quelle, Zeit, Größe und SHA-256 protokollieren. Identische Prüfsummen als mögliche Duplikate markieren, nicht automatisch entfernen.
3. **Prüfen:** Dateiformat erkennen, Lesbarkeit prüfen, Malware-/Makro-/Archivrisiken berücksichtigen, Zugangsdaten und personenbezogene Daten kennzeichnen. Keine automatische OCR/Extraktion als wahrheitsgetreue Quelle behandeln.
4. **Klassifizieren:** `public`, `internal`, `confidential`, `restricted` oder `unknown`. Nicht klassifizierte Objekte bleiben gesperrt für öffentliche GitHub-Übernahme.
5. **Ableiten:** Nur notwendige, rechtlich und inhaltlich freigegebene Fassungen erzeugen. Quelle und Ableitung erhalten getrennte IDs/Prüfsummen sowie eine protokollierte Transformationsbeschreibung.
6. **Kuratieren:** Relevante Inhalte werden in `notebooks/curated/<bereich>/` oder `docs/` übernommen. Produktionsverhalten wird in Code und Tests unter `core/` implementiert; die Quelldatei wird nicht ungeprüft als Code ausgeführt.
7. **Verifizieren und auditieren:** Prüfsummen gegen das Original validieren; Ereignisse mit Actor, Zeitpunkt, Objekt-ID und Aktion erfassen. Keine Geheimnisse oder privaten Speicher-URLs in öffentliche Logs schreiben.

## GitHub-Grenzen

- Keine Roh-PDF/TXT/PPTX/DOCX/ZIP/HEIC-Dateien, vollständigen Datenbank-Backups, `.pyc`, geheimen Schlüssel oder persönlichen Daten committen.
- Nur Inhalte mit bestätigtem Nutzungsrecht, Einstufung und Review dürfen in `notebooks/curated/` eingecheckt werden.
- Große Binärdateien und vertrauliche Bestände bleiben im Vault; Git-LFS ist kein Ersatz für Zugriffskontrolle oder Records-Management.
- Prüfsummen allein belegen Integrität nur gegenüber einem bekannten Referenzwert. Für Nachweisbarkeit müssen Referenzwert, Signatur/Vertrauensanker und Prüfprotokoll geschützt werden.
- `core/cosmic-memory/` ist derzeit ein geplantes Projektverzeichnis, kein bereits implementierter oder zertifizierter WORM-Speicher.

## Entscheidungen vor produktivem Betrieb

1. Welcher Storage-Dienst und welche Identitäts-/Zugriffssteuerung werden verwendet?
2. Welche Personen dürfen Originale lesen, exportieren, freigeben und löschen?
3. Welche rechtlichen Grundlagen, Lizenzbedingungen und Aufbewahrungsfristen gelten?
4. Sind die Kategorien aus dem Screenshot disjunkt? Die sichtbaren Zahlen summieren sich auf 12.025; weitere Pages-/ZIP-Zahlen sind abgeschnitten.
5. Wie wird der Bestand gesichert und wiederhergestellt, und wie werden Lösch-/Sperrvermerke nachverfolgt?

Dieses Dokument definiert Grenzen und Fluss, aber weder eine Speicheranbindung noch eine Compliance-Zertifizierung.
