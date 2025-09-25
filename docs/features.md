## Ziel
Ein performanter, bidirektionaler Zugriff von WSL2/Linux auf Windows-Verzeichnisse mit vollständiger Dateisystemfunktionalität und zuverlässiger Änderungsbenachrichtigung (inotify-ähnlich), realisiert über einen gRPC-Server auf Windows und einen FUSE-Client in WSL2/Linux.

## Funktionsumfang (MVP)
- Basis-Mount
  - Mount eines einzelnen Windows-Verzeichnisses nach Linux mittels FUSE
  - Konfigurierbarer Mount-Punkt (z. B. `/mnt/fsdriver/<share>`)
- Dateisystem-Operationen
  - Lookup/Stat von Dateien/Verzeichnissen
  - ReadDir (Auflisten von Verzeichnissen) mit Paginierung bei großen Ordnern
  - Open/Close, Read/Write von Dateien (sequenziell, ohne Random-Write-Optimierung im MVP)
  - Create/Unlink/Rename von Dateien
  - Mkdir/Rmdir für Verzeichnisse
  - Truncate/Chmod/Chown (soweit sinnvoll auf NTFS abbildbar)
  - Symlinks/Hardlinks: Lesen/Erstellen soweit abbildbar; sonst sinnvoller Fallback
- Änderungsbenachrichtigungen (Inotify)
  - Abbildung von Windows-Dateiänderungen (USN Journal/ReadDirectoryChangesW) auf FUSE/notify
  - Events: Create/Delete/Rename/Modify/Attribute/Move
  - Event-Drosselung und Debounce, um Event-Stürme zu vermeiden
- gRPC-Schnittstelle
  - Definierte Services für Stat, ReadDir, Open/Read/Write, Create/Rename usw.
  - Bidirektionales Streaming für Change-Events
  - Fehler- und Statuscodes (Mapping zwischen Windows- und POSIX-Fehlern)
- Sicherheit & Isolation
  - Read-only-Modus optional
  - Freigaben whitelist-basiert; Server limitiert auf explizit konfigurierte Verzeichnisse
  - Transportverschlüsselung optional (localhost/AF_UNIX kann unverschlüsselt; Remote optional TLS)
- Performance (MVP)
  - Read-Ahead/Write-Behind grundlegende Pufferung
  - Chunked I/O über gRPC (konfigurierbare Blockgröße)
  - Parallelisierung von ReadDir/Stat-Operationen

## Erweiterungen (Post-MVP)
- Caching
  - Page-Cache im Client (größen- und zeitbasiert)
  - Directory-Entry-Cache mit Invalidierung über Change-Events
  - Attribute-Cache (TTL), invalidiert über Events
- Random-Access-Optimierung
  - Parallele Read/Write-Requests mit Request-Merging
  - Memory-mapped Read (Server-seitig) wo sinnvoll
- Langlebige Handles
  - Server-seitiges Handle-Management zur Reduktion von Open/Close-Kosten
  - Lease/Keep-Alive für Hot-Files
- Robustheit
  - Reconnect-Logik (Client ↔ Server), Replay/Resync von Events
  - Backpressure auf Streaming-Kanälen
  - Zeitouts, Retry-Policy mit Exponential Backoff
- Sicherheit
  - Mutual TLS zwischen WSL2 und Windows
  - Per-Share ACL/Read-only/Executable-Flags
- Beobachtbarkeit
  - Strukturierte Logs (JSON), Metriken (Prometheus/OpenTelemetry)
  - Tracing für teure Operationen (OpenTelemetry)
- Verwaltung/UX
  - CLI: `fsdriver serve` (Windows) und `fsdriver mount` (WSL2)
  - Konfigurationsdateien (YAML/TOML) und Env-Variablen
  - Systemd-User-Service (WSL2) und Windows-Dienst-Installation
- Kompatibilität
  - Case-(in)Sensitivity-Strategien (NTFS vs. Linux)
  - Pfadlängen-Handhabung (Windows MAX_PATH vs. Long Paths)
  - Zeitstempel- und Attribut-Mapping (NTFS ↔ POSIX)

## Nicht-Ziele (vorerst)
- Vollständige POSIX-Konformität in Spezialfällen (z. B. komplexe Unix-Permissions, Device-Files)
- Netzwerkbetrieb jenseits von lokaler Windows↔WSL2-Kommunikation
- Distributed Caching oder Multi-Writer-Kohärenz über mehrere Clients

## Technische Leitplanken
- Windows-Server
  - gRPC-Server, Zugriff auf NTFS über Win32-APIs
  - Änderungsüberwachung über ReadDirectoryChangesW/USN Journal
- Linux/WSL2-Client
  - FUSE-Dateisystem, Übersetzung von POSIX-Operationen in gRPC-Calls
  - Invalidation/Notify auf Basis der Event-Streams


