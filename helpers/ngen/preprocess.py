import os
from scipy import signal
from scipy.signal import butter, filtfilt
import time
import pandas as pd
import numpy as np


"""
Calculate frequency of reception of new samples from ptp4l process

Input: samples, number_of_samples
Output: update_rate in seconds
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


"""
Preprocess data

Input: samples
Output: lpf, frequency rate and samples in locked state
"""
def run(df, transient_period):
    s2_count = (df['state'].values == 2).sum()
    print ("number of samples with servo s2 :",s2_count)

    update_rate = calculate_update_rate(df, s2_count)
    print ("Update rate estimate from S2 deltas: ",update_rate, "updates/s")

    #initial transient sync period is fixed to 5 minutes
    end_transient_period=transient_period*update_rate

    #input signal after transient sync period
    input_signal=df.phase[end_transient_period:len(df)]

    fc=0.1               # Cutoff Frequency 0.1Hz low-pass filter
    w=fc/(update_rate/2) # the critical frequency for digital filters w is normalized from 0 to 1 where 1 is Nyquist freqency.
    btype='low'          # band type is type of filter
    fiorder=5            # the order of the filter
    analog=False         # it is always a digital filter
    output='ba'          # type of output: b numerator coefficient vector and a is denominator coefficient vector

    b, a = butter(fiorder, w, btype, analog, output)

    lpf_signal = filtfilt(b, a, input_signal)

    return lpf_signal, update_rate, s2_count

