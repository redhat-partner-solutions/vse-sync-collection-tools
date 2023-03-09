import sys
import os
from scipy import signal
from scipy.signal import butter, filtfilt
import  allantools
import time
import pandas as pd
import numpy as np

"""
Calculate TE

Input: samples, sample number identifiying end of transient period, mask, taus
Output:  max|TE|
"""
def calculate_abs_te(df, end_initial_syncperiod, mask_te, taus_list):
    print ("G.8273.2 7.1 Max. Absolute Time Error, unfiltered measured; max|TE| <= ", mask_te,"ns")
    phase_min    = df.loc[end_initial_syncperiod:len(df)].phase.min()
    phase_mean   = df.loc[end_initial_syncperiod:len(df)].phase.mean()
    phase_stddev = df.loc[end_initial_syncperiod:len(df)].phase.std()
    phase_max    = df.loc[end_initial_syncperiod:len(df)].phase.max()
    phase_pktpk  = phase_max-phase_min

    max_abs_te = max(phase_max, abs(phase_min))
    print ("Max |TE|:",'{:.3f}'.format(max_abs_te), "ns")
    if max_abs_te > mask_te:
        print ("Max Absolute Time Error unfiltered above 30ns")
        sys.exit(1)
    else:
        print ("pk-pk phase:",'{:.3f}'.format(phase_pktpk), "ns")
        print ("Mean phase:",'{:.3f}'.format(phase_mean), "ns")
        print ("Min phase:",'{:.3f}'.format(phase_min), "ns")
        print ("Max phase:",'{:.3f}'.format(phase_max), "ns")
        print ("Phase stddev:",'{:.3f}'.format(phase_stddev), "ns")

"""
Calculate cTE

Input: samples, amount_of_useful_samples, sample_rate, sample number identifiying end of transient period, mask, taus
Output:  max|cTE|
"""
def calculate_const_te(df, s2_count, update_rate, end_initial_syncperiod, mask_cte, taus_list):
    print ("G.8273.2 7.1.1 Max. Constant Time Error averaged over 1000sec cTE <= ", mask_cte, "ns")
    if s2_count > 2000 * update_rate:
        df['MovAvg'] = df.phase.rolling(1000*update_rate,min_periods=1000*update_rate).mean()
        cte_min    = df.loc[(end_initial_syncperiod+1000*update_rate):len(df)].MovAvg.min()
        cte_mean   = df.loc[(end_initial_syncperiod+1000*update_rate):len(df)].MovAvg.mean()
        cte_stddev = df.loc[(end_initial_syncperiod+1000*update_rate):len(df)].MovAvg.std()
        cte_max = df.loc[(end_initial_syncperiod+1000*update_rate):len(df)].MovAvg.max()
        cte_var = max(cte_max, abs(cte_mean-cte_min))
        cte_pktpk = cte_max-cte_min
        max_abs_cte = max(cte_max, abs(cte_min))
        print ("Max |cTE|:",'{:.3f}'.format(max_abs_cte), "ns")
        if max_abs_cte > mask_cte:
            print ("Max constant Time Error averaged over 1000s above 10ns")
        else:
            print ("Mean cTE:",'{:.3f}'.format(cte_mean), "ns")
            print ("Min cTE:",'{:.3f}'.format(cte_min), "ns")
            print ("Max cTE:",'{:.3f}'.format(cte_max), "ns")
            print ("pk-pk cTE:",'{:.3f}'.format(cte_pktpk), "ns")
            print ("cTE stddev:",'{:.3f}'.format(cte_stddev), "ns")
            print ("cTE Var: +/-",'{:.3f}'.format(cte_var), "ns")
    else:
        print ("Insufficient data for cTE computation, at least 2000s are needed")

"""
Calculate mtie

Input: samples, low_pass_filter, sample_rate,  mask, taus
Output:  max|MTIE|
"""
def calculate_mtie(df, lpf_signal, update_rate, mask_mtie, taus_list):
    print("G.8273.2 7.1.2 Max. Dynamic Time Error, 0.1Hz Low-Pass Filtered; MTIE <= ", mask_mtie, "ns")
    mtie_taus, mtie_devs, mtie_errs, ns = allantools.mtie(lpf_signal, rate=update_rate,
                                                     data_type='phase', taus=taus_list)
    mtie_min = mtie_devs.min()
    mtie_max = mtie_devs.max()
    mtie_pktpk = mtie_max-mtie_min
    max_abs_mtie = max(mtie_max, abs(mtie_min))
    print ("Max |MTIE|:",'{:.3f}'.format(max_abs_mtie), "ns")
    if max_abs_mtie > mask_mtie:
        print ("Max Dynamic Time Error, dTE (MTIE) above 10ns")
    else:
        print ("Min MTIE:",'{:.3f}'.format(mtie_min), "ns")
        print ("Max MTIE:",'{:.3f}'.format(mtie_max), "ns")
        print ("Max-Min MTIE:",'{:.3f}'.format(mtie_pktpk), "ns")

"""
Calculate tdev

Input: samples, low_pass_filter, sample_rate,  mask, taus
Output:  max|TDEV|
"""
def calculate_tdev(df, lpf_signal, update_rate, mask_tdev, taus_list):
    print("G.8273.2 7.1.2 Max. Dynamic Time Error, 0.1Hz Low-Pass Filtered; TDEV <= ", mask_tdev, "ns")
    tdev_taus, tdev_devs, tdev_errs, ns = allantools.tdev(lpf_signal, rate=update_rate, 
                                                     data_type='phase', taus=taus_list)
    tdev_min = tdev_devs.min()
    tdev_max = tdev_devs.max()
    tdev_pktpk = tdev_max-tdev_min
    max_abs_tdev = max(tdev_max, abs(tdev_min))
    print ("Max |TDEV|:",'{:.3f}'.format(max_abs_tdev), "ns")
    if max_abs_tdev > mask_tdev:
        print ("Max Dynamic Time Error (TDEV) above 2ns")
    else:
        print ("Min TDEV:",'{:.3f}'.format(tdev_min), "ns")
        print ("Max TDEV:",'{:.3f}'.format(tdev_max), "ns")
        print ("Max-Min TDEV:",'{:.3f}'.format(tdev_pktpk), "ns")

"""
Calculate NGEN KPIs

Input: samples, number_of_samples, sample identifiying end of transient period, update rate in seconds
Output: update_rate in 
"""
def run(df, transient_period, clk_class, s2_count,  update_rate, lpf_signal):
    
    if clk_class == 'C':
        mask_te=30
        mask_cte=10
        mask_mtie=10
        mask_tdev=2
    else: # Class-D
        mask_te=15
        mask_cte=4
        mask_mtie=3
        mask_tdev=1

    #initial transient sync period is fixed to 5 minutes
    end_initial_syncperiod=transient_period*update_rate

    taus_list = np.array([1/update_rate,
                     1, 2, 3, 4,5,6,7,8,9,
                     10,20,30,40,50,60,70,80,90,
                     100,200,300,400,500,600,700,800,900,
                     1000,2000,3000,4000,5000,6000,7000,8000,9000,10000])

    calculate_abs_te(df, end_initial_syncperiod, mask_te, taus_list)
    
    calculate_const_te(df, s2_count, update_rate, end_initial_syncperiod, mask_cte, taus_list) 

    calculate_mtie(df, lpf_signal, update_rate, mask_mtie, taus_list)

    calculate_tdev(df, lpf_signal, update_rate, mask_tdev, taus_list)
