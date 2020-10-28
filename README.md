# velux

A go library and cli for interacting with Velux Active devices using the HomeKit protocol. It can be used to read data from and control windows.

This sits on top of [mctofu/homekit](https://github.com/mctofu/homekit) to provide an easier interface to deal specifically with windows.

## CLI Usage

### Installing

```shell
go get github.com/mctofu/velux/cmd/velux
```

### Initialize
For now you first need to create and pair a controller using the homekit cli (https://github.com/mctofu/homekit). Once you've paired with the Velux Active gateway you need to import the pairing.
```shell
$ velux import
Imported controller and 1 pairings
```

### Read status
The status command retrieves environment measurements and window positions
```shell
$ velux status
Temperature: 75.2F
Relative Humidity: 36.0
CO2 PPM: 623
1234567890abcdef (Roof Window): 65 (65)
2234567890abcdef (Roof Window): 20 (20)
3234567890abcdef (Blinds): 0 (0)
```

### Setup friendly identifiers (Optional)
Assign a short code and descriptor to a window for easier usage
```shell
velux setup --serial 1234567890abcdef --code fwin --desc "front window"
```
```shell
$ velux status
Temperature: 75.2F
Relative Humidity: 36.0
CO2 PPM: 626
front window (fwin): 65 (65)
2234567890abcdef (Roof Window): 20 (20)
3234567890abcdef (Blinds): 0 (0)
```

### Set window positions
Set all window positions or filter by serial, type or code
```
$ velux set -c fwin -p 25
Updated windows:
front window (fwin): 65 (25)
```
