# WSL2 Networking Configuration

## Problem
When running the fsdriver client from WSL2, you might encounter connection issues when trying to connect to the Windows server. This is due to WSL2's default networking mode which creates a virtual network interface.

## Solution: Mirrored Networking Mode

Add the following configuration to your `.wslconfig` file (usually located at `%USERPROFILE%\.wslconfig`):

```ini
[wsl2]
networkingMode=mirrored
```

After making this change, restart WSL2:
```cmd
wsl --shutdown
```

## Benefits
- Allows using `127.0.0.1` from WSL2 to connect to Windows services
- Simplifies networking configuration
- Eliminates the need to find the Windows host IP address

## Alternative: Manual IP Configuration
If you prefer not to use mirrored networking, you can:
1. Find the Windows host IP from WSL2: `ip route | grep default`
2. Start the server with: `--addr 0.0.0.0:PORT`
3. Connect the client to the Windows host IP

## References
- [WSL2 Networking Documentation](https://docs.microsoft.com/en-us/windows/wsl/networking)
- [WSL2 Configuration Options](https://docs.microsoft.com/en-us/windows/wsl/wsl-config)
