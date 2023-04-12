# QuickStart


## Build

```
# podman build -t gnss-utils:3.25_dev -f ./ContainerFile
```

## Run in OpenShift

1) Deploy the manifest `pod_gpsd.yaml`

```
# oc apply -f pod_gpsd.yaml
```

2) Check logs

```
# oc logs -f gnss-utils
GNSS Device : gnss0
PTY device: /dev/pts/0
gpsd:INFO: launching (Version 3.25.1~dev, revision release-3.25-26-g8ab2872d3)
gpsd:INFO: starting uid 0, gid 0
gpsd:INFO: Command line: /usr/local/sbin/gpsd -p -n -S 2947 -G -N -D 5 /dev/gnss0
....
psd:PROG: CORE: switch_driver(u-blox) called...
gpsd:PROG: CORE: selecting u-blox driver...
gpsd:INFO: CORE: /dev/gnss0 identified as type u-blox, 0 sec
gpsd:PROG: => GPS: UBX class: 0a, id: 04, len: 0, crc: 0e34
gpsd:IO: SER: => GPS: b5620a0400000e34
gpsd:PROG: UBX-NAV-PVT
gpsd:PROG: NAV-PVT: flags=01 time= 1681327668.999550643 lat=42.55 lon=-71.21 altHAE=25.98 track=0.00 speed=0.00 mode=3 status=1 used=0
gpsd:PROG:   headVeh 0.00000 magDec nan magAcc nan
gpsd:PROG: UBX: new PROTVER 14 was 0
gpsd:IO: UBX: len 28
gpsd:PROG: UBX-NAV-POSECEF: iTOW=329287000 ECEF x=1515414.24 y=-4455226.90 z=4290845.27 pAcc=2.61
gpsd:IO: UBX: len 28
gpsd:PROG: UBX-NAV-VELECEF
gpsd:PROG: UBX-NAV-VELECEF: iTOW=329287000 ECEF vx=0.00 vy=0.00 vz=0.00 vAcc=0.02
```

## Run in RHEL

1) Pull the container:

```
# podman pull quay.io/jnunez/gnss-utils:3.25_dev
```

2) Run the container locally:

```
# podman run -it --rm --privileged -v /dev/gnss0:/dev/gnss0 -e "BUS=gnss0" quay.io/jnunez/gnss-utils:3.25_dev
```

3) Check gpsd is running:

```
# podman  logs -f your_gnss_utils_container_name
GNSS Device : gnss0
gpsd:INFO: launching (Version 3.25.1~dev, revision release-3.25-13-g5d8d5b0dc)
gpsd:INFO: starting uid 0, gid 0
gpsd:INFO: Command line: /usr/local/sbin/gpsd -n -S 2947 -G -N -D 5 /dev/ttyGNSS_5100
....
```

4) With GPSD running, launch `ubxtool` command to read gnss subsystem:

```
# podman exec -it CONTAINER_ID bash
#
# ubxtool -w 1 -v 1 -p MON-VER -P 29.20
UBX-NAV-PVT:
  iTOW 390009000 time 2023/2/23 12:19:51 valid x37
  tAcc 23 nano 277947 fixType 3 flags x1 flags2 xea
  numSV 9 lon -969945065 lat 329430676 height 141289
  hMSL 166434 hAcc 991 vAcc 1889
  velN 3 velE 1 velD -6 gSpeed 3 headMot 21707050
  sAcc 14 headAcc 9905136 pDOP 230 reserved1 0 46364 10063
  headVeh 2576309 magDec 0 magAcc 0
 ....
 UBX-MON-VER:
  swVersion EXT CORE 1.00 (3fda8e)
  hwVersion 00190000
  extension ROM BASE 0x118B2060
  extension FWVER=TIM 2.20
  extension PROTVER=29.20
  extension MOD=ZED-F9T
  extension GPS;GLO;GAL;BDS
  extension SBAS;QZSS
  extension NAVIC
```
