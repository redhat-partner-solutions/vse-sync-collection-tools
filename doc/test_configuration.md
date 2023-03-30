# Test Configuration File

Following is a reference configuration template for Test Configuration file:

```
suite_configs:
  - ptp_tests_config:
      namespace: <namespace-where-ptp-lives>
      pod_name: <ptp-pod-name>
      container: <ptp-container-name>
      interface_name: <interface-of-ptpconfiguration>
      tty_timeout: <tty_timeout_in_seconds>
      dpll_reads: <number_of_dpll_reads>
```

* `namespace` (string, required): namespace where ptp operator lives (usually `openshift-ptp`)
* `pod_name` (string, required): name of the ptp daemon pod (usually `linuxptp-daemon-something`)
* `container` (string, required): name of the container (usually linuxtptp-daemon-container)
* `interface_name` (string, optional): interface name of ptp port under test (e.g., ens7f0)
* `tty_timeout` (int, required): timeout for reading GPS signal in segonds (eg., 20)
* `dpll_reads` (int, required): number of times 1PPS Digital Phase Locked Loop (DPLL) is read (e.g., 4)
