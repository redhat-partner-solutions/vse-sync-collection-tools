# QuickStart

Pull the container:

```
# podman pull quay.io/jnunez/gnss-utils
```

Run the container locally:

```
# podman run -it --rm --privileged -v /dev/ttyGNSS_5100:/dev/ttyGNSS_5100 -e "BUS=5100" quay.io/jnunez/gnss-utils:3.25_dev
```

```
# podman  logs -f your_gnss_utils_container_name
GNSS Device : ttyGNSS_5100
gpsd:INFO: launching (Version 3.25.1~dev, revision release-3.25-13-g5d8d5b0dc)
gpsd:INFO: starting uid 0, gid 0
gpsd:INFO: Command line: /usr/local/sbin/gpsd -n -S 2947 -G -N -D 5 /dev/ttyGNSS_5100 
```

Launch ubxtool commands leveraging gpsd:

```
# podman exec -it CONTAINER_ID bash
# ubxtool -v 1 -P MON-VER
```