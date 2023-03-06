from scipy import signal
from scipy.signal import butter, filtfilt
import os
import  allantools
import matplotlib.pyplot as plt
import time
import pandas as pd
import numpy as np

#Filter the phase data with 0.1Hz low-pass filter
filename="./16Dec22_SC_ptp4l.csv"
if os.path.exists(filename):
    df=pd.read_csv(filename, dtype={"tstamp": "float64", "phase": "float64", "state": "float64", "freq": "float64", "delay": "float64", "event": "string"})
    print ("reading file")
    print (df.dtypes)
    df.describe()
else:
    print ("no Such File")
    sys.exit()

starttime = df['tstamp'].values[0]


s2_count = (df['state'].values == 2).sum()
print ("number of s2 :",s2_count)

firstS2 = (df['state'].values == 2).argmax()
prevS2 = firstS2
cumS2delta = 0
x=0
S2count=0
# calculate update rate from ptp4l stats frequency of reception of new samples from ptp4l process
for x in range (firstS2+1, 1024): #use just about 1k samples (minus ~ init S0/S1's and any events)\n",
    if df.loc[x].state == 2:
    	cumS2delta = cumS2delta+(df.loc[x].tstamp-df.loc[prevS2].tstamp)
    	prevS2=x
    	S2count=S2count+1

update_rate = round((1/(cumS2delta/S2count)))
print ("Update rate estimate from S2 deltas: ",update_rate, "updates/s")

#initial transient sync period is fixed to 5'
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


# G.8273.2 7.1 Max. Absolute Time Error, 0.1Hz Low-Pass Filtered; max|TE| <= 30ns
phase_min    = df.loc[end_initial_syncperiod:len(df)].phase.min()
phase_mean   = df.loc[end_initial_syncperiod:len(df)].phase.mean()
phase_stddev = df.loc[end_initial_syncperiod:len(df)].phase.std()
phase_max    = df.loc[end_initial_syncperiod:len(df)].phase.max()
phase_pktpk  = phase_max-phase_min

#if phase_pktpk<1001:
#    phase_bins=round(phase_pktpk)
#else:
#	phase_bins=1000
max_abs_te = max(phase_max, abs(phase_min))
print ("pk-pk phase:",'{:.3f}'.format(phase_pktpk), "ns")
print ("Mean phase:",'{:.3f}'.format(phase_mean), "ns")
print ("Min phase:",'{:.3f}'.format(phase_min), "ns")
print ("Max phase:",'{:.3f}'.format(phase_max), "ns")
print ("Phase stddev:",'{:.3f}'.format(phase_stddev), "ns")
print ("Max |TE|:",'{:.3f}'.format(max_abs_te), "ns")

# G.8273.2 7.1 Max. Constant Time Error; cTE <= [-10ns,+10ns]
# calculate and print cTE stats\n",

if s2_count > 2000 * update_rate:
	df['MovAvg'] = df.phase.rolling(1000*update_rate,min_periods=1000*update_rate).mean()
	df.loc[end_initial_syncperiod+1000*update_rate:len(df)].plot(kind='line',linewidth=0.25,x='tstamp', y='MovAvg', color='blue')
	plt.title("Constant Time Error (cTE), 1000s Moving Average", fontsize=12, weight='bold')
	plt.xlabel("Time (s)")
	plt.ylabel("cTE (ns)")
	plt.show()
	cte_min    = df.loc[(end_initial_syncperiod+1000*update_rate):len(df)].MovAvg.min()
	cte_mean   = df.loc[(end_initial_syncperiod+1000*update_rate):len(df)].MovAvg.mean()
	cte_stddev = df.loc[(end_initial_syncperiod+1000*update_rate):len(df)].MovAvg.std()
	cte_max = df.loc[(end_initial_syncperiod+1000*update_rate):len(df)].MovAvg.max()
	cte_var = max(cte_max, abs(cte_mean-cte_min))
	cte_pktpk = cte_max-cte_min
	print ("Mean cTE:",'{:.3f}'.format(cte_mean), "ns")
	print ("Min cTE:",'{:.3f}'.format(cte_min), "ns")
	print ("Max cTE:",'{:.3f}'.format(cte_max), "ns")
	print ("pk-pk cTE:",'{:.3f}'.format(cte_pktpk), "ns")
	print ("cTE stddev:",'{:.3f}'.format(cte_stddev), "ns")
	print ("cTE Var: +/-",'{:.3f}'.format(cte_var), "ns")
else:
	print("Insufficient data for cTE computation & plot, need at least 2000s")
    
# G.8273.2 7.1.2 Max. Dynamic Time Error, 0.1Hz Low-Pass Filtered; MTIE < 10ns
mtie_taus, mtie_devs, mtie_errs, ns = allantools.mtie(lpf_signal, rate=update_rate,
                                                     data_type='phase', taus=taus_list)
plt.plot(mtie_taus, mtie_devs, color='blue'),
plt.title("MTIE, G.8273.2 Class-C mask; Constant Temperature", fontsize=12, weight='bold'),
plt.xlabel("Tau (s)")
plt.ylabel("MTIE (ns)")
plt.show()
mtie_min = mtie_devs.min()
mtie_max = mtie_devs.max()
mtie_pktpk = mtie_max-mtie_min
print ("Min MTIE:",'{:.3f}'.format(mtie_min), "ns")
print ("Max MTIE:",'{:.3f}'.format(mtie_max), "ns")
print ("Max-Min MTIE:",'{:.3f}'.format(mtie_pktpk), "ns")


# G.8273.2 7.1.2 Max. Dynamic Time Error, 0.1Hz Low-Pass Filtered; TDEV < 2ns
tdev_taus, tdev_devs, tdev_errs, ns = allantools.tdev(lpf_signal, rate=update_rate,
                                                     data_type='phase', taus=taus_list)

plt.plot(tdev_taus, tdev_devs, color='blue'),
plt.title("TDEV, G.8273.2 Class-C mask; Constant Temperature", fontsize=12, weight='bold'),
plt.xlabel("Tau (s)")
plt.ylabel("TDEV (ns)")
plt.show()
tdev_min = tdev_devs.min()
tdev_max = tdev_devs.max()
tdev_pktpk = tdev_max-tdev_min
print ("Min TDEV:",'{:.3f}'.format(tdev_min), "ns")
print ("Max TDEV:",'{:.3f}'.format(tdev_max), "ns")
print ("Max-Min TDEV:",'{:.3f}'.format(tdev_pktpk), "ns")

# G.8273.2 7.1.3 Max. dynamic Time Error, 0.1Hz High Pass Filtered; dTE 

# generate HPF
#btype='high'        	 							# band type is type of filter 
#b, a = butter(fiorder, w, btype, analog, output)
#hpf_signal = filtfilt(b, a, input_signal)
#plt.plot(hpf_signal, linewidth=0.25, color='blue')
#plt.title("Time Error (TE), 0.1Hz high-pass filtered\", fontsize=12, weight='bold')",
#plt.xlabel("Time (s))"
#plt.ylabel("TE (ns))"
#plt.show()

# G.8273.2 7.1.4.1 Relative Constant Time Error Noise Generation; cTE


# G.8273.2 7.1.4.2 Relative Dynamic Time Error Low-Pass Filtered Noise Generation (MTIE)
