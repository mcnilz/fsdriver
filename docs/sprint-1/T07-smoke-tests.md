## T07: Manuelle Smoke-Tests (E2E Read-only)

Ziel:
Schnelle, manuelle Validierung des Vertical Slice.

Checkliste:
- Server starten mit Testordner
- Client mountet read-only
- `ls` im Root-Verzeichnis zeigt Inhalte
- `cat` einer Datei liefert den erwarteten Inhalt
- Datei in Windows erstellen/umbenennen/löschen → nach kurzer Zeit sichtbar
- `stat` zeigt aktualisierte Timestamps/Größen
- Unmount/Graceful Shutdown ohne Hänger

Dokumentation:
- Kurze Schritt-für-Schritt-Anleitung in `docs/sprint-1/README.md` verlinken


