## T02: Windows gRPC Server – Read-only Basis

Ziel:
gRPC-Server, der ein konfiguriertes Windows-Verzeichnis bereitstellt und Read-only Operationen bedient.

Umfang:
- Implementiere `Stat`, `ReadDir`, `Open`, `Read`, `Close`
- Pfadnormalisierung und Sicherheit: Zugriff strikt innerhalb der Share-Root
- Chunked Read, konfigurierbare Blockgröße
- Fehler-Mapping (z. B. `ERROR_FILE_NOT_FOUND` → `ENOENT`)

Akzeptanzkriterien:
- Manuelle Tests per gRPC-Client gegen Testordner erfolgreich
- Logging bei Fehlern, grundlegende Metriken (optional)


