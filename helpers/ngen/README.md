# QuickStart


## Build 

1) Build the image to calculate Noise Generation KPIs according to G.8273.2:

```
# podman build -t quay.io/jnunez/ngen_kpis:0.1/ngen_kpis:0.1 -f ./ContainerFile
```

## Run

1) Pull the G.8273.2 Noise Generation KPIs container image:

```
# podman pull quay.io/jnunez/ngen_kpis:0.1
```

2) Run the container to calculate G.8273.2 Noise Generation KPIs consuming post-processed data from ptp4l in `/tmp/data.csv`:

```
# podman run -it --rm -v /tmp:/tmp localhost/ngen_kpis:0.1 /tmp/data.csv
```

