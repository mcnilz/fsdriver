## Nutzung (MVP Read-only)

### Voraussetzungen
- Windows 10/11 mit WSL2
- Go >= 1.22 installiert
- protoc installiert und im PATH (`protoc --version`)
- **WSL2 Networking**: Siehe [WSL2 Networking Konfiguration](wsl-networking.md) für die richtige Einrichtung

### Build
```bash
go build ./...
```

### Server starten (Windows)
```bash
fsdriver\server.exe --share C:\\path\\to\\dir --addr 0.0.0.0:50052
```

Parameter:
- `--share`: Root-Verzeichnis, das freigegeben wird (muss existieren)
- `--addr`: Listen-Adresse (Default: 127.0.0.1:50051, empfohlen: 0.0.0.0:50052)

### Client mounten (WSL2)
```bash
# Mit mirrored networking (empfohlen)
sudo ./client --share test --mountpoint /mnt/fsdriver/test --addr 127.0.0.1:50052

# Mit manueller IP-Konfiguration
sudo ./client --share test --mountpoint /mnt/fsdriver/test --addr 172.20.16.1:50052
```

### Verbindung testen
```bash
# Test-Server-Erreichbarkeit von Windows
./test_client.exe 127.0.0.1:50052

# Test von WSL2 (wenn Networking korrekt konfiguriert)
./test_client.exe 127.0.0.1:50052
```

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


