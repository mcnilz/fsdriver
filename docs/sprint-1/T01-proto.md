## T01: gRPC Proto für Read-only Slice

Ziel:
Minimaler Proto-Contract für Read-only: `Stat`, `ReadDir`, `Open`, `Read`, `Close`, sowie ein bidirektionaler `Watch`-Stream für Change-Events.

Akzeptanzkriterien:
- Proto definiert Nachrichten und Services inkl. Fehlercodes (Mapping Windows↔POSIX)
- Streaming-Events: Create/Delete/Rename/Modify, Pfade relativ zur Share-Root
- Datei-I/O: Chunked Reads mit Offset+Länge

Deliverables:
- Datei `proto/fsdriver.proto`
- Kurzer Kommentar zur Fehlermapping-Strategie und Stabilitätsgarantien


