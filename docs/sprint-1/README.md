## Sprint 1 – Ziel und Inkrement

Ziel: End-to-end Vertical Slice (Read-only). Aus WSL2 heraus ein Windows-Verzeichnis read-only mounten, Verzeichnisinhalte lesen (Stat/ReadDir, Open/Read), und grundlegende Change-Events empfangen (Mapping Windows → Inotify), um Directory- und Attribute-Caches invalidieren zu können.

Definition of Done:
- CLI kann einen lokalen Windows-Share in WSL2 read-only mounten
- `ls`, `cat`, `stat` funktionieren zuverlässig auf dem Mount
- Änderungen (Create/Delete/Rename/Modify) im Share invalidieren Directory/Attr-Caches sichtbar (z. B. `ls` aktualisiert sich ohne remount)
- Basis-Dokumentation vorhanden; manuelle Smoke-Tests beschrieben

Tickets liegen als einzelne Markdown-Dateien in diesem Ordner.

### Aktueller Stand (nach T01-T04)
- T01: `proto/fsdriver.proto` erstellt, Go-Stubs generiert
- T02: Windows gRPC-Server (read-only) implementiert (`server/*`), CLI `--share`, `--addr`
- T03: Watch-Stream mit fsnotify implementiert (rekursiv, Event-Mapping)
- T04: FUSE-Client (read-only) implementiert (`client/*`), Linux-only Build-Constraints
- Build/Run-Hinweise siehe `docs/usage.md`


