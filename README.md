# MijnAurum Telegraf Input Plugin

The `mijnaurum` plugin gathers heating metrics from the mijnaurum smart heating system REST API.

### Installation

- Clone the repository
```
git clone git@github.com:peterpeerdeman/mijnaurum-telegraf-plugin.git
```

- Build the plugin for your own architecture and for raspberry pi arm64
```
go build -o mijnaurum cmd/main.go
env GOOS=linux GOARCH=arm64 GOARM=7 go build -o mijnaurum cmd/main.go
```

- Call it from telegraf
```toml
[[inputs.execd]]
  command = ["/path/to/mijnaurum", "-config", "/path/to/mijnaurum.conf"]
  signal = "none"
```

### Configuration

```toml
 # Gathers Metrics MijnAurum V2 REST API
[[inputs.mijnaurum]]
  ## Credentials
  username = "user1"
  password = "pass123"

  ## Metrics to collect from mijnaurum
  # collectors = ["heat"]
```


### Metrics

- mijnaurum
  - tags:
    - source
    - source_type
    - rate_unit
    - unit
    - meter_id
    - location_id
  - fields:
    - day_value
    - day_cost
    - week_value
    - week_cost
    - month_value
    - month_cost
    - year_value
    - year_cost

### Example output

```
mijnaurum,location_id=z4O1ho3dGhmH-w2zu2d4YOgUsP77jfQadSw0lCL3SnqONvtExoxS-tgjiAmyxdmK,meter_id=D0zkFgSTtzvSSmqvAmshBMJp_qVkgpjqEuibvND3l9n-VEhaVEyFF97vQNfzVY5j,rate_unit=J/h,source=dZGWYt_pp20TnzlFzHxBKzsOR6X-cXA-xLZLSTNMuJaVzodeWGHa1SJS03mDjykT,source_type=heat,unit=GJ week_value=0.011000000000000001,year_cost=35.71992,year_value=1.6560000000000001,day_cost=0.23727000000000004,day_value=0.011000000000000001,month_cost=8.951550000000001,month_value=0.41500000000000004,week_cost=0.23727000000000004 1653329096960710000
```
