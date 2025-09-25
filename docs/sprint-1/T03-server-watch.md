## T03: Windows Watch/Events – Stream

Ziel:
Bereitstellung eines Event-Streams, der Änderungen unterhalb der Share-Root meldet.

Umfang:
- `ReadDirectoryChangesW`-basierter Watcher (rekursiv)
- Event-Kategorien: Create/Delete/Rename/Modify/Attrib
- Debounce/Coalescing einfacher Heuristik nach Pfad/Typ

Akzeptanzkriterien:
- Manuelle Änderung von Dateien/Ordnern erzeugt erwartete Events
- Backpressure und Reconnect-Handhabung grundlegend vorhanden


