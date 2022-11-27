# Admin panel for WireGuard VPN

## Installation

### Migrations
Just run `migrate -path migrations -database "mysql://root:root@tcp(localhost:3306)/<your-database>" up`

### Config
Create in `configs/` file `panel.toml` with following variables:
```
    bind_addr = ":80"
    log_level = "debug"
    database_url = "root:root@tcp(127.0.0.1:3306)/<your-database>"
    session_key = "<your-session-key>"
    commands_path = "/path/to/wg-admin/internal/app/services/commands/"
```

### Start panel
Run `make` and then run `./panel`
