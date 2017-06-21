# docker-zfs-plugin
Docker volume plugin for creating persistent volumes as a dedicated zfs dataset.

# Installation

Download the latest binary from github releases and place in `/usr/bin/`.

If using a systemd based distribution, copy
[docker-zfs-plugin.service](docker-zfs-plugin.service) to `/etc/systemd/system`.
Then enable and start the service with `systemctl daemon-reload && systemctl
enable docker-zfs-plugin.service && systemctl start docker-zfs-plugin.service`.

* Usage

After the plugin is running, you can interact with it through normal `docker volume` commands.

Recently, support was added for passing in ZFS attributes from the `docker volume create` command:

`docker volume create -d zfs -o compression=lz4 -o dedup=on --name=data`
