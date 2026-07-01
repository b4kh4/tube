# Tube VPN

Tube is a lightweight, high-performance Layer 3 P2P Mesh VPN written in Go. It establishes direct, end-to-end encrypted tunnels between computers, bypassing NAT restrictions without relying on centralized traffic relays.

## Core Capabilities

Because Tube operates at the OS network layer (L3) rather than the application layer, it routes all IP traffic natively. This enables:

1. **LAN Gaming over WAN:** Play games like Minecraft, Terraria, or classic RTS games over the internet exactly as if you were on the same local router.
2. **Secure Remote Access (RDP/SSH):** Connect to a remote desktop or terminal securely without opening ports on your home router.
3. **Private File Sharing:** Run local SMB, FTP, or SFTP servers for high-speed, direct peer-to-peer file transfers.
4. **Self-Hosted Web Services:** Expose local development servers or internal web applications to authorized peers effortlessly.

## How It Works

1. **Virtual Interface (TUN/TAP):** Tube creates a virtual network adapter on the host OS. Applications send standard IP packets to this adapter.
2. **Interception & Encryption:** The application intercepts these raw L3 packets and encrypts them using the ChaCha20-Poly1305 AEAD algorithm.
3. **Signaling & NAT Traversal:** Peers exchange public endpoint data via a lightweight signaling broker using a 16-character room code.
4. **UDP Hole Punching:** Both peers initiate UDP transmission simultaneously, penetrating domestic NATs to establish a direct P2P link. No traffic is routed through third-party servers.