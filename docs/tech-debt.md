## Technische Schulden / Cleanups (Backlog)

- Architektur & Struktur
  - Konfiguration extrahieren (Datei + Env + Flags, Prioritäten definieren)
  - Logging-Backend austauschbar machen (zap/logrus), einheitliches JSON-Format
  - Einheitliche Fehler-Typen und Mapping (Win32→POSIX) zentralisieren, Codes katalogisieren
  - Konsistente Kontext-Nutzung (ctx-Weitergabe, Cancellation/Timeouts)
  - Graceful Shutdown: gRPC-Server, offene Handles, laufende Streams

- Windows-spezifisch
  - Langes Pfadpräfix (\\\\?\\) überall robust unterstützen
  - Case-Sensitivity-Strategie finalisieren (Lookup vs. Anzeige)
  - Symlink/Hardlink-Unterstützung (lesen/erstellen) definieren und testen
  - Präzisere Fehler-Mappings aus `GetLastError()` aufnehmen

- Performance & Robustheit
  - I/O-Chunkgrößen konfigurierbar; Read-Ahead/Write-Behind Heuristiken
  - Backpressure im Event-Stream; Retry/Backoff für Streams/RPCs
  - Handle-Leak-Detection und Limits; Idle-Handle-GC
  - Directory-Paginierung effizienter (Iteratoren statt vollständigem Readdir)

- Beobachtbarkeit
  - Metriken (Prometheus/OpenTelemetry), Tracing für teure Pfade
  - Strukturierte Fehler-Logs mit Korrelation (req_id)

- Sicherheit
  - TLS (mTLS optional) für gRPC (auch wenn lokal), Zertifikatsrotation
  - Share-Whitelist/ACLs, Read-only/Executable-Flags pro Share

- CLI & UX
  - Besseres CLI-Help/Examples, Validierung und klare Fehlermeldungen
  - Windows-Dienst-Installer (sc.exe/PowerShell) und Service-Logs
  - WSL2-Helper-Skripte (systemd user service) für Autostart

- Tests
  - Unit-Tests für Pfad-Normalisierung, Fehler-Mapping, Handle-Management
  - Integrationstests (E2E) mit Testverzeichnis; Golden-Tests für ReadDir/Stat
  - Lasttests für große Verzeichnisse und parallele Reads

- Watch/Events (für T03/T05)
  - ReadDirectoryChangesW/USN Journal evaluieren; Coalescing/De-Dup-Heuristiken
  - Reconnect/Resync-Protokoll definieren; Event-Verlust-Strategie

- Dokumentation
  - DESIGN-Notizen erweitern (Fehlermatrix, Pfadregeln, Eventmodell)
  - Betrieb/Deployment: Beispiele, Troubleshooting, bekannte Limitierungen


