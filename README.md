# velib-exporter

Want to export your stats from https://www.velib-metropole.fr/ ?

from this ugly interface:

![](https://user-images.githubusercontent.com/62110608/154846420-f89e2604-7d30-40a7-8344-97f55622e355.png)

to:

_TODO: insert Grafana screenshot here_

## Features

- Your total Velib trip number
- Your total distance in Velib (in km)
- Your total distance in electrical Velib (in km)
- Your total distance in mechanical Velib (in km)
- Your highest trip distance (your personal record! yay!) (in km)
- Your average Velib trip duration (in minutes)
- Total CO2 saved according to Velib (in grams)

```prometheus
$> curl -s 127.0.0.1:5050/metrics | grep -B2 -E ^velib_
# HELP velib_co2_total_saved Total CO2 saved by using Velib in grams
# TYPE velib_co2_total_saved gauge
velib_co2_total_saved 7328.200000000001
# HELP velib_distance_electrical Distance total in electrical Velib in meters
# TYPE velib_distance_electrical gauge
velib_distance_electrical 33737
# HELP velib_distance_mechanical Distance total in mechanical Velib in meters
# TYPE velib_distance_mechanical gauge
velib_distance_mechanical 32283
# HELP velib_distance_total Distance total in Velib in meters
# TYPE velib_distance_total gauge
velib_distance_total 66020
# HELP velib_trip_average_duration Velib trip average duration in minutes
# TYPE velib_trip_average_duration gauge
velib_trip_average_duration 12
# HELP velib_trip_highest_distance Velib trip highest distance in meters
# TYPE velib_trip_highest_distance gauge
velib_trip_highest_distance 6554
# HELP velib_trip_number Number of Velib trips
# TYPE velib_trip_number gauge
velib_trip_number 25
```

ℹ️ Stats are refreshed every hours.

## Installation

```console
$ git clone https://github.com/qselle/velib-exporter
$ cd velib-exporter
$ go build
```

## Usage

```console
Usage of velib-exporter:
  -address string
        Exporter listening address (default "127.0.0.1")
  -debug
        Debug mode
  -port string
        Exporter listening port (default "5050")
  -token string
        Velib API token
```

_TODO: add tutorial on how to get the token_

## Grafana

_TODO: insert grafana dashboard ID_

---

Feel free to contribute.