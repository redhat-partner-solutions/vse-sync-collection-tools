#!/bin/sh


echo "GNSS Device : "${BUS}

SOCAT_OUTPUT=/tmp/out-socat
socat -d -d pty,raw,echo=0 pty,raw,echo=0 > /dev/null 2> $SOCAT_OUTPUT &
sleep 2
DEVICE_PTY=$(head -n1 $SOCAT_OUTPUT | grep -o '/dev/.\+')
echo "PTY device: "${DEVICE_PTY}
/usr/local/sbin/gpsd -p -n -S 2947 -G -N -D 5 /dev/${BUS}
