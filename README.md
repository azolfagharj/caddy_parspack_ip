# ParsPack IP Source Module for Caddy <a href="https://azolfagharj.github.io/donate/"><img src="https://img.shields.io/badge/Donate-Support%20Development-orange?style=for-the-badge" alt="Donate"></a>

This Caddy module retrieves <a href="https://parspack.com/">ParsPack</a> CDN IP ranges from their <a href="https://parspack.com/cdnips.txt">official source</a> and makes them available for use in Caddy's trusted proxies and IP matchers.

## Features

- Automatically fetches IPv4 IP ranges from ParsPack endpoint
- Periodically refreshes the IP list to stay up-to-date
- Configurable refresh interval and timeout
- Supports Caddyfile configuration

## Installation

Build Caddy with this module using xcaddy:

```bash
xcaddy build --with github.com/azolfagharj/caddy_parspack_ip
```


## Configuration

### Simple Usage

The simplest way to use this module is without any options. It will automatically fetch IP ranges from ParsPack and refresh them every hour:

```caddyfile
trusted_proxies parspack
```

### With Options

You can optionally configure the refresh interval and timeout:

```caddyfile
trusted_proxies parspack {
    interval 12h
    timeout 15s
}
```

## Configuration Options

| Name | Description | Type | Default |
|------|-------------|------|---------|
| interval | How often ParsPack IP lists are retrieved | duration | 1h |
| timeout | Maximum time to wait for a response from ParsPack | duration | no timeout |

Both options are optional. If not specified, the module uses the default values shown above.

## Requirements

- Caddy v2.6.3 or later

This module is compatible with all Caddy v2.6.3+ versions including v2.7.x, v2.8.x, v2.9.x, v2.10.x, and future versions. The `http.ip_sources` namespace was introduced in Caddy v2.6.3, which is the minimum required version.

## License

Apache-2.0


---
## Support this Project



 🤝 **Enjoying this free project?** <a href="https://azolfagharj.github.io/donate/">Consider supporting</a> its development

<a href="https://azolfagharj.github.io/donate/"><img src="https://img.shields.io/badge/Donate-Support%20Development-orange?style=for-the-badge" alt="Donate"></a>

---
