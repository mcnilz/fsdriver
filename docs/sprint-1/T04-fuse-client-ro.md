## T04: FUSE-Client – Read-only Mount

Ziel:
FUSE-Dateisystem, das über gRPC mit dem Windows-Server spricht und read-only Operationen anbietet.

Umfang:
- Mount eines Shares an angegebenem Mountpoint
- Implementiere `Lookup/Stat`, `ReadDir`, `Open`, `Read`, `Release`
- Attribute- und Dentry-Cache mit kurzer TTL

Akzeptanzkriterien:
- `ls`, `cat`, `stat` funktionieren gegen den Mount
- Errors werden korrekt nach POSIX gemappt


