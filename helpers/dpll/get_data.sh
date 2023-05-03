while true;
   do
    cat /proc/uptime | awk '{printf $1","}';
    cat /sys/class/net/ens7f0/device/dpll_0_state | awk '{printf $1","}';
    cat /sys/class/net/ens7f0/device/dpll_1_state | awk '{printf $1","}';
    cat /sys/class/net/ens7f0/device/dpll_1_offset | awk '{print $1/100}'
    sleep 1;
done
