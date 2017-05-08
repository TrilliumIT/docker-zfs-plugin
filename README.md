# docker-zfs-plugin
Docker volume plugin for creating persistent volumes as a dedicated zfs dataset.

# Installation

Download the latest binary from github releases and place in `/usr/bin/`.

If using a systemd based distribution, copy
[docker-zfs-plugin.service](docker-zfs-plugin.service) to `/etc/systemd/system`.
Then enable and start the service with `systemctl daemon-reload && systemctl
enable docker-arp-ipam.service`.
