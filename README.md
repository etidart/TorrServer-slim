## TorrServer-slim

TorrServer-slim is a streamlined fork of [TorrServer](https://github.com/YouROK/TorrServer), designed to enhance performance by removing unnecessary components. While the original project supports a wide range of devices and operating systems, this fork intended for use in GNU/Linux and Windows PC/server environments only. However, it can also run on any device with sufficient resources.

The primary use case for TorrServer-slim is as a drop-in replacement for TorrServer, particularly when paired with Lampa.

## How to use

- Just launch the executable and get working service on `localhost:8090` (default).
- On Linux systems with systemd you can install and enable `torrserver.service` which will run program automatically on each startup.

## Removed features

- web frontend
- ssl support
- ip blacklists
- msx
- html auth support
- rutor search
- ffprobe
- telegram bot
- dlna server
- disk as a piece storage

## License

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version. See the [LICENSE](LICENSE) file for details.
