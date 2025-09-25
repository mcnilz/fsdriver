# Foundations: Pfade, Fehler, Events, I/O, Logging

## Pfadnormalisierung
- Server erzwingt Root-Relativität; verhindert Escape mittels `..`
- Unterstützt Windows Long Paths (`\\\\?\\` Prefix), intern normiert zu UTF-16/UTF-8
- Case-Handling: Default case-insensitive Lookup, Case-Preservation in Listings

## Fehler-Mapping (Windows → POSIX)
- `ERROR_FILE_NOT_FOUND` → `ENOENT`, `ERROR_ACCESS_DENIED` → `EACCES`, `ERROR_ALREADY_EXISTS` → `EEXIST`
- Unbekannte Fehler → `EIO`, inkl. Originalcode in Details/Logs

## Change-Events
- Quelle: ReadDirectoryChangesW (rekursiv) oder USN Journal (später)
- Eventtypen: Create/Delete/Rename/Modify/Attrib, Debounce/Coalescing per Pfad
- Versand über gRPC-Stream; Reconnect mit Resync-Marker

## I/O-Chunking & Caching
- Default Read-Chunk z. B. 128 KiB; Read-Ahead basierend auf sequentiellem Zugriff
- Client: kurze TTL für Attr/Dentry-Cache; Invalidierung über Events

## Logging & Metriken
- Strukturierte Logs (JSON) mit Korrelation (req_id)
- Basis-Metriken: Latenz/Fehler pro RPC, Eventrate, offene Handles

## Clean Code Leitlinien
- Benennungen: aussagekräftige, vollständige Wörter; Funktionen sind Verben, Variablen Substantive
- Funktionen klein halten (<50 Zeilen), Single Responsibility, frühe Rückgaben
- Fehlerbehandlung erstklassig: keine stummen Catches; eindeutige Fehlerpfade
- Keine tiefe Verschachtelung: Guard Clauses bevorzugen
- Trenne Domänenlogik von I/O (z. B. Pfad-Validierung separat von RPC-Handlern)
- Keine globalen Zustände außer bewusst verwalteten Singletons (z. B. Logger)
- Kommentare sparsam; erklären das „Warum“, nicht das „Wie“
- Einheitliches Format und Linting; keine toten Codes


