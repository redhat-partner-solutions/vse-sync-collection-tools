#!/bin/bash

# Use if mounted the whole /sys/class/net/*
# ETH=`grep 000e /sys/class/net/*/device/subsystem_device | awk -F"/" '{print$5}' | head -n 1`
# BUS=`grep PCI_SLOT_NAME /sys/class/net/${ETH}/device/uevent | cut -c 20- | head -c 5 | sed 's/://'
# BUS is in the bash environment

echo "GNSS Device : ttyGNSS_"${BUS}

/usr/local/sbin/gpsd -n -S 2947 -G -N -D 5 /dev/ttyGNSS_${BUS}
