## T06: CLI und minimale Konfiguration

Ziel:
Kommandozeile für beide Seiten: `serve` (Windows) und `mount` (WSL2) mit minimalen Flags.

Umfang:
- `fsdriver serve --share C:\\path\\to\\dir --addr localhost:50051`
- `fsdriver mount --share myshare --mountpoint /mnt/fsdriver/myshare --addr <win-host>:50051 --ro`
- Einfache YAML/TOML-Config optional, Flags haben Vorrang

Akzeptanzkriterien:
- Starten/Stoppen der Prozesse per CLI, valide Fehlerausgaben
- Hilfe/Usage vorhanden


