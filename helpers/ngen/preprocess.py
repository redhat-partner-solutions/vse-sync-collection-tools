# Copyright 2023 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import os
from scipy import signal
from scipy.signal import butter, filtfilt
import time
import pandas as pd
import numpy as np


def calculate_transient_period(
    df, update_rate, first_locked_sample, max_allowed_transient_period
):
    """
    Calculate actual transient period and check if out of bounds

    Input: samples, update_rate, max_allowed_transient_period (in seconds)
    Output: update_rate in seconds
    """

    transient_period = first_locked_sample / update_rate
    if transient_period <= max_allowed_transient_period:
        return transient_period
    else:
        return -1


def calculate_update_rate(df, s2_count):
    """
    Calculate frequency of reception of new samples from ptp4l process

    Input: samples, number_of_samples
    Output: update_rate in seconds
    """
    firstS2 = (df["state"].values == 2).argmax()
    print("First Sample in Locked State is", firstS2)
    prevS2 = firstS2
    cumS2delta = 0
    x = 0
    S2count = 0
    # use just about 1k samples (minus ~ init S0/S1's and any events)
    for x in range(firstS2 + 1, 1024):
        if df.loc[x].state == 2:
            cumS2delta = cumS2delta + (df.loc[x].tstamp - df.loc[prevS2].tstamp)
            prevS2 = x
            S2count = S2count + 1
    return round((1 / (cumS2delta / S2count))), firstS2


def run(df, max_allowed_transient_period):
    """
    Preprocess data

    Input: samples
    Output: lpf, frequency rate and samples in locked state
    """
    s2_count = (df["state"].values == 2).sum()
    print("number of samples in Locked State :", s2_count)

    update_rate, first_locked_sample = calculate_update_rate(df, s2_count)
    print("Update rate estimate from S2 deltas:", update_rate, "updates/s")

    actual_transient_period = calculate_transient_period(
        df, update_rate, first_locked_sample, max_allowed_transient_period
    )
    if actual_transient_period > 0:
        print(
            "Transient period below max allowed transient period",
            actual_transient_period,
        )
    else:
        raise Exception(
            "Transient above max allowed transient period", actual_transient_period
        )

    # initial transient sync period is fixed to 5 minutes
    end_transient_period = max_allowed_transient_period * update_rate

    # input signal after transient sync period
    input_signal = df.phase[end_transient_period : len(df)]

    fc = 0.1  # Cutoff Frequency 0.1Hz low-pass filter
    w = fc / (
        update_rate / 2
    )  # the critical frequency for digital filters w is normalized from 0 to 1 where 1 is Nyquist freqency.
    btype = "low"  # band type is type of filter
    fiorder = 5  # the order of the filter
    analog = False  # it is always a digital filter
    output = "ba"  # type of output: b numerator coefficient vector and a is denominator coefficient vector

    b, a = butter(fiorder, w, btype, analog, output)

    lpf_signal = filtfilt(b, a, input_signal)

    return lpf_signal, update_rate, s2_count
