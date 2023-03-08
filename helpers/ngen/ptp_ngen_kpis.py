import sys
import os
from scipy import signal
from scipy.signal import butter, filtfilt
import  allantools
import time
import pandas as pd
import numpy as np

"""
Calculate frequency of reception of new samples from ptp4l process

Input: samples, number_of_samples
Output: update_rate in 
"""
def calculate_update_rate(df,s2_count):
    firstS2 = (df['state'].values == 2).argmax()
    prevS2 = firstS2
    cumS2delta = 0
    x=0
    S2count=0    
    for x in range (firstS2+1, 1024): #use just about 1k samples (minus ~ init S0/S1's and any events)\n",
        if df.loc[x].state == 2:
            cumS2delta = cumS2delta+(df.loc[x].tstamp-df.loc[prevS2].tstamp)
            prevS2=x
            S2count=S2count+1
    return round((1/(cumS2delta/S2count)))

if len(sys.argv) == 2:
    filename=sys.argv[1]
    if os.path.exists(filename):
        df=pd.read_csv(filename, dtype={"tstamp": "float64", "phase": "float64", "state": "float64", "freq": "float64", "delay": "float64", "event": "string"})
        df.describe()
    else:
        print ("no Such File named ", filename)
        sys.exit()
else:
    print ("no postprocessed input data received as input")


start_time = df['tstamp'].values[0]
s2_count = (df['state'].values == 2).sum()
print ("number of samples with servo s2 :",s2_count)

update_rate = calculate_update_rate(df, s2_count)
print ("Update rate estimate from S2 deltas: ",update_rate, "updates/s")

#initial transient sync period is fixed to 5 minutes
end_initial_syncperiod=300*update_rate


#input signal after transient sync period
input_signal=df.phase[end_initial_syncperiod:len(df)]

fc=0.1 		   		 # Cutoff Frequency 0.1Hz low-pass filter
w=fc/(update_rate/2) # the critical frequency for digital filters w is normalized from 0 to 1 where 1 is Nyquist freqency.
btype='low'        	 # band type is type of filter
fiorder=5 	       	 # the order of the filter
analog=False       	 # it is always a digital filter
output='ba'        	 # type of output: b numerator coefficient vector and a is denominator coefficient vector

b, a = butter(fiorder, w, btype, analog, output)

lpf_signal = filtfilt(b, a, input_signal)

taus_list = np.array([1/update_rate,
       				 1, 2, 3, 4,5,6,7,8,9,
                     10,20,30,40,50,60,70,80,90,
                     100,200,300,400,500,600,700,800,900,
                     1000,2000,3000,4000,5000,6000,7000,8000,9000,10000])


print ("G.8273.2 7.1 Max. Absolute Time Error, unfiltered measured; max|TE| <= 30ns")
phase_min    = df.loc[end_initial_syncperiod:len(df)].phase.min()
phase_mean   = df.loc[end_initial_syncperiod:len(df)].phase.mean()
phase_stddev = df.loc[end_initial_syncperiod:len(df)].phase.std()
phase_max    = df.loc[end_initial_syncperiod:len(df)].phase.max()
phase_pktpk  = phase_max-phase_min

max_abs_te = max(phase_max, abs(phase_min))
if max_abs_te > 30:
    print ("Max Absolute Time Error unfiltered above 30ns")
    print ("Max |TE|:",'{:.3f}'.format(max_abs_te), "ns")
    sys.exit(1)
else:
    print ("pk-pk phase:",'{:.3f}'.format(phase_pktpk), "ns")
    print ("Mean phase:",'{:.3f}'.format(phase_mean), "ns")
    print ("Min phase:",'{:.3f}'.format(phase_min), "ns")
    print ("Max phase:",'{:.3f}'.format(phase_max), "ns")
    print ("Phase stddev:",'{:.3f}'.format(phase_stddev), "ns")

print ("G.8273.2 7.1.1 Max. Constant Time Error averaged over 1000sec cTE <= 10ns")
if s2_count > 2000 * update_rate:
    df['MovAvg'] = df.phase.rolling(1000*update_rate,min_periods=1000*update_rate).mean()
    cte_min    = df.loc[(end_initial_syncperiod+1000*update_rate):len(df)].MovAvg.min()
    cte_mean   = df.loc[(end_initial_syncperiod+1000*update_rate):len(df)].MovAvg.mean()
    cte_stddev = df.loc[(end_initial_syncperiod+1000*update_rate):len(df)].MovAvg.std()
    cte_max = df.loc[(end_initial_syncperiod+1000*update_rate):len(df)].MovAvg.max()
    cte_var = max(cte_max, abs(cte_mean-cte_min))
    cte_pktpk = cte_max-cte_min
    max_abs_cte = max(cte_max, abs(cte_min))
    if max_abs_cte > 10:
        print ("Max constant Time Error averaged over 1000s above 10ns")
        print ("Max |cTE|:",'{:.3f}'.format(max_abs_cte), "ns")
    else:
        print ("Mean cTE:",'{:.3f}'.format(cte_mean), "ns")
        print ("Min cTE:",'{:.3f}'.format(cte_min), "ns")
        print ("Max cTE:",'{:.3f}'.format(cte_max), "ns")
        print ("pk-pk cTE:",'{:.3f}'.format(cte_pktpk), "ns")
        print ("cTE stddev:",'{:.3f}'.format(cte_stddev), "ns")
        print ("cTE Var: +/-",'{:.3f}'.format(cte_var), "ns")
else:
	print("Insufficient data for cTE computation, at least 2000s are needed")

print("G.8273.2 7.1.2 Max. Dynamic Time Error, 0.1Hz Low-Pass Filtered; MTIE <= 10ns")
mtie_taus, mtie_devs, mtie_errs, ns = allantools.mtie(lpf_signal, rate=update_rate,
                                                     data_type='phase', taus=taus_list)
mtie_min = mtie_devs.min()
mtie_max = mtie_devs.max()
mtie_pktpk = mtie_max-mtie_min
max_abs_mtie = max(mtie_max, abs(mtie_min))
if max_abs_mtie > 10:
    print ("Max Dynamic Time Error, dTE (MTIE) above 10ns")
    print ("Max |MTIE|:",'{:.3f}'.format(max_abs_mtie), "ns")
else:
    print ("Min MTIE:",'{:.3f}'.format(mtie_min), "ns")
    print ("Max MTIE:",'{:.3f}'.format(mtie_max), "ns")
    print ("Max-Min MTIE:",'{:.3f}'.format(mtie_pktpk), "ns")


print("G.8273.2 7.1.2 Max. Dynamic Time Error, 0.1Hz Low-Pass Filtered; TDEV <= 2ns")
tdev_taus, tdev_devs, tdev_errs, ns = allantools.tdev(lpf_signal, rate=update_rate,
                                                     data_type='phase', taus=taus_list)
tdev_min = tdev_devs.min()
tdev_max = tdev_devs.max()
tdev_pktpk = tdev_max-tdev_min
max_abs_tdev = max(tdev_max, abs(tdev_min))
if max_abs_tdev > 2:
    print ("Max Dynamic Time Error (TDEV) above 2ns")
    print ("Max |TDEV|:",'{:.3f}'.format(max_abs_tdev), "ns")
else:
    print ("Min TDEV:",'{:.3f}'.format(tdev_min), "ns")
    print ("Max TDEV:",'{:.3f}'.format(tdev_max), "ns")
    print ("Max-Min TDEV:",'{:.3f}'.format(tdev_pktpk), "ns")  
    print ("Max |TDEV| below 2ns:",'{:.3f}'.format(max_abs_tdev), "ns")