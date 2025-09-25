## Nutzung (MVP Read-only)

### Voraussetzungen
- Windows 10/11 mit WSL2
- Go >= 1.22 installiert
- protoc installiert und im PATH (`protoc --version`)

### Build
```bash
go build ./...
```

### Server starten (Windows)
```bash
fsdriver\server.exe --share C:\\path\\to\\dir --addr 127.0.0.1:50051
```

Parameter:
- `--share`: Root-Verzeichnis, das freigegeben wird (muss existieren)
- `--addr`: Listen-Adresse (Default: 127.0.0.1:50051)

### Proto neu generieren (bei Änderungen)
```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/fsdriver.proto
```

### Hinweise
- Aktuell nur Read-only Operationen (Stat, ReadDir, Open/Read, Close)
- Pfad-Zugriffe sind strikt auf `--share` begrenzt
- Logs sind strukturiert (einfaches Key-Value über stdout)


