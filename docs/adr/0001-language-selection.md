## ADR 0001: Auswahl der Implementierungssprache für Windows-Server und WSL2-FUSE-Client

Datum: 2025-09-25

Status: Vorschlag

### Kontext
Ziel ist ein Werkzeug, das Windows-Verzeichnisse per gRPC im Windows-Host bereitstellt und in WSL2/Linux per FUSE einbindet. Wichtige Anforderungen sind:
- Hohe Performance bei Dateioperationen (I/O, Verzeichnislisten) und geringer Latenz
- Zuverlässige Änderungsbenachrichtigungen (inotify-Äquivalent) über Mapping von Windows-Events
- Stabiler Betrieb (Dienste), geringe Laufzeit-Abhängigkeiten, einfache Verteilung
- Gute Ökosysteme für gRPC, FUSE, File-Watching und System-Integration (Windows-Dienst, systemd-User-Service)

Zur Diskussion stehen insbesondere Go (Golang) und .NET (C#/.NET 8+).

### Optionen
1) Go
   - Stärken:
     - Einfache, statisch gelinkte Binaries, sehr gute Cross-Plattform-Story
     - Reife gRPC-Implementierung und Tools; gute Performance
     - Solide FUSE-Bibliotheken (z. B. `hanwen/go-fuse/v2`) mit produktivem Einsatz
     - Gute Windows-Interop über `syscall`/`x/sys/windows`; Zugriff auf `ReadDirectoryChangesW`, USN Journal, Overlapped I/O
     - Geringe Start-/Speicher-Overheads; gut für CLI/Daemon
   - Schwächen:
     - Garbage Collector kann bei sehr latenzkritischen Workloads gelegentlich Jitter erzeugen (praktisch meist unkritisch)
     - Teilweise weniger ausgereifte High-Level-Windows-APIs als in .NET (muss über `x/sys/windows` erfolgen)

2) .NET (C#)
   - Stärken:
     - Erstklassige Windows-Integration, reiches API-Set (FileSystemWatcher, IO Pipelines)
     - Sehr gute gRPC-Unterstützung; Top-Developer-Experience
     - Hohe Produktivität, moderne Spracheigenschaften
   - Schwächen:
     - FUSE-Clientseite in Linux/WSL2 hat weniger verbreitete, reife Bibliotheken in .NET; häufig ist native/Interop nötig
     - Größere Runtime-Abhängigkeit (selbstenthaltende Deployments sind möglich, aber größer)
     - Cross-Compiling und sehr kleine, selbstständige Binaries sind machbar, aber nicht so leichtgewichtig wie Go

### Bewertung nach Komponenten
- Windows gRPC Server:
  - Beide Sprachen geeignet. .NET bietet sehr bequeme Windows-APIs, Go bietet einfache Deployments und gute Performance.
  - Eventing (ReadDirectoryChangesW/USN): In Go direkt via `x/sys/windows`; in .NET komfortabel über Wrapper, aber FileSystemWatcher basiert nicht auf USN und kann Limitierungen haben. Für hohe Last ist ein eigener USN/IO-Layer nötig – in beiden Sprachen machbar.

- WSL2/Linux FUSE Client:
  - Go hat mit `go-fuse` eine reife, gut dokumentierte Bibliothek. In .NET wären P/Invoke/Native-Bindings nötig oder Drittbibliotheken, die weniger verbreitet sind.
  - Latenz und Ressourcenbedarf fallen zugunsten von Go aus.

### Entscheidung
Wir wählen Go für beide Komponenten (Windows-Server und WSL2-FUSE-Client).

Begründung:
- Reife FUSE-Unterstützung in Go und einfache, schlanke Deployments für den Client
- Sehr gute gRPC-Ökosystem- und Performance-Eigenschaften
- Direkter Zugriff auf Windows-APIs ohne große Runtime; gute Kontrolle über I/O und Concurrency
- Einheitlicher Technologie-Stack vereinfacht Build, CI/CD und gemeinsame Codebasis

### Konsequenzen
- Implementierung auf Go-Basis: `go-fuse` auf Client-Seite, `x/sys/windows` und Win32-APIs auf Server-Seite
- Nutzung von Protobuf/gRPC für Schnittstellen und Streaming-Events
- Bereitstellung als zwei Binaries: `fsdriver-serve` (Windows), `fsdriver-mount` (WSL2)
- Option: Später hybride Variante evaluieren (z. B. .NET nur für Windows-Server), falls Windows-spezifische Features einfacher werden. Aktuell kein Bedarf.

### Alternativen und warum verworfen
- .NET für beide Komponenten: FUSE-Story ist schwächer; größere Runtimes; mehr Interop notwendig
- .NET Server + Go Client: Mischtechnologie erhöht Komplexität (Build/CI, Diagnose) ohne klaren Vorteil im MVP


