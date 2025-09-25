## T05: Event-gestützte Cache-Invalidierung (Client)

Ziel:
Events vom Server in FUSE-Invalidierungen umsetzen (Dentry/Attr), sodass `ls`/`stat` zeitnah aktualisierte Inhalte zeigen.

Umfang:
- Empfang des `Watch`-Streams im Client
- Mapping von Eventtypen auf FUSE-Invalidate-Calls
- Rate-Limit/Debounce pro Pfad

Akzeptanzkriterien:
- Manuelle Dateiänderung in Windows wird ohne Remount im Mount sichtbar
- Keine endlosen Invalidierungsloops


