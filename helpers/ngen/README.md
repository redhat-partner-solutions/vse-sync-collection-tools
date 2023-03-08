# QuickStart


## Build 

1) Build the image to calculate Noise Generation KPIs according to G.8273.2:

```
# podman build -t quay.io/jnunez/ngen_kpis:0.1 -f ./ContainerFile
```

## Run

1) Pull the G.8273.2 Noise Generation KPIs container image:

```
# podman pull quay.io/jnunez/ngen_kpis:0.1
```

2) Data File derivated from ptp4l messages:

```
tstamp,phase,state,freq,delay,event
1184958.065,,,,,"port 1: UNCALIBRATED to SLAVE on MASTER_CLOCK_SELECTED"
1184958.128,-98,2,-15910,22273
1184958.190,-91,2,-15908,22273
1184958.253,-67,2,-15876,22272
```

where:

tstamp: Timestamp
phase: Phase offset in ns
state: Servo clock state
freq: Frequency offset
delay: Path delay
event: Event information   

3) Run the container to calculate G.8273.2 Noise Generation KPIs consuming post-processed data from ptp4l in `/tmp/data.csv`:

```
# podman run -it --rm -v /tmp:/tmp quay.io/jnunez/ngen_kpis:0.1 /tmp/data.csv
```


