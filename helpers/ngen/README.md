# QuickStart


## G.8272.2 Noise Generation KPIs


```console
# python3 ptp_ngen_kpis.py -h
usage: ptp_ngen_kpis.py [-h] -i INPUT [-c CLOCKCLASS] [-t TRANSIENT] [--plot PLOT] [-o OUTPUT]

optional arguments:
  -h, --help            show this help message and exit
  -c CLOCKCLASS, --clockclass CLOCKCLASS
                        clock class-[C,D] requirement to satisfy, defaults to
                        Class-C
  -t TRANSIENT, --transient TRANSIENT
                        transient period, defaults to 300sec
  -s STEADY, --steady STEADY
                        minimum steady state period to enable calculations,
                        defaults to 2000sec
  -p PLOT, --plot PLOT  add plots to the results
  -o OUTPUT, --output OUTPUT
                        Output file name, defaults to stdout

required arguments:
  -i INPUT, --input INPUT
                        Input sample data
```

## Build

1) Build the image to calculate Noise Generation KPIs according to G.8273.2:

```
# podman build -t quay.io/redhat-partner-solutions/ngen_kpis:0.1 -f ./ContainerFile
```

## Run

1) Pull the G.8273.2 Noise Generation KPIs container image:

```
# podman pull quay.io/redhat-partner-solutions/ngen_kpis:0.1
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
# podman run -it --rm -v /tmp:/tmp quay.io/jnunez/ngen_kpis:0.1 -i /tmp/data.csv
```
