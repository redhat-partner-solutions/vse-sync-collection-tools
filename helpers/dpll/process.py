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

import matplotlib.pyplot as plt
import numpy as np
import allantools


def plot_phase_dpll(df):
    """
    plot Phase DPLL
    """
    plt.figure()
    df.loc[1 : len(df)].plot(kind="line", linewidth=0.25, y="offset", color="blue")
    plt.title("DPLL Phase Offset", fontsize=12, weight="bold")
    plt.xlabel("Time (s)")
    plt.ylabel("Phase Offset (ns)")
    plt.savefig("dpll.png")


def tgm_terror(df):
    """
    Calculate  KPIs

    Input: samples
    Output:
        plot phase dpll
    """

    phase_min = df.loc[1 : len(df)].offset.min()
    phase_mean = df.loc[1 : len(df)].offset.mean()
    phase_stddev = df.loc[1 : len(df)].offset.std()
    phase_max = df.loc[1 : len(df)].offset.max()
    phase_pktpk = phase_max - phase_min
    max_abs_phase = max(phase_max, abs(phase_min))
    print(
        f"""
           Mean DPLL Phase Offset: {phase_mean:.3f} ns
           Min DPLL Phase Offset: {phase_min:.3f} ns
           Max DPLL Phase Offset: {phase_max:.3f} ns
           Max |DPLL Phase Offset|: {max_abs_phase:.3f} ns
           pk-pk DPLL Phase Offset: {phase_pktpk:.3f} ns
           DPLL Phase Offset stdev: {phase_stddev:.3f} ns
           """
    )
    plot_phase_dpll(df)


def plot_cte_phase_dpll(df, window_size):
    """
    plot cTE
    """
    plt.figure()
    df.loc[0 : (len(df) - (window_size + 1))].plot(
        kind="line", linewidth=0.25, x="tstamp", y="MovAvg", color="blue"
    )
    plt.title(
        "DPLL Phase Constant Time Error, averaged over 100 samples",
        fontsize=12,
        weight="bold",
    )
    plt.xlabel("Time (s)")
    plt.ylabel("cTE Phase(ns)")
    plt.savefig("dpll_constant_time_error.png")


def moving_average_low_pass_filter(samples, window_size):
    """
    Input:
      `samples`: the input data samples
      `window_size`: number of samples from which the average is calculated
    Output:
      moving average lows pass filter for all data points in `samples`
    """

    n = len(samples)
    mov_avg = samples.copy()
    for i in range(0, (n - window_size)):
        avg = np.mean(samples[i : i + window_size])
        mov_avg[i] = avg
    return mov_avg


def tgm_cte(df):
    """
    Calculate constant Time Error measured through a moving-average low-pass filter of at least 100 consecutive time error samples. This filter is applied by the test equipment to remove errors caused by timestamp quantization, or any quantization of packet position in the test equipment, before calculating the maximum time error
    Input: samples
    Output:
        plot cTE from phase samples in phase DPLL
    """

    df["MovAvg"] = moving_average_low_pass_filter(df.offset, 100)
    plot_cte_phase_dpll(df, 100)


def plot_mtie(mtie_taus, mtie_devs):
    """
    Plot MTIE
    """
    plt.figure()
    plt.axes(xscale="log", yscale="log")
    plt.plot(mtie_taus, mtie_devs, color="blue")
    mask_B = tgm_wander_generation_mtie(mtie_taus, "PRTC-B")
    mask_A = tgm_wander_generation_mtie(mtie_taus, "PRTC-A")
    plt.plot(mtie_taus, mask_B, color="red", linestyle="dashed", label="PRTC-B")
    plt.plot(mtie_taus, mask_A, color="green", linestyle="dashed", label="PRTC-A")
    plt.title("MTIE, DPLL Phase", fontsize=12, weight="bold")
    plt.xlabel("Tau (s)")
    plt.ylabel("MTIE (ns)")
    plt.legend()
    plt.savefig("dpll_phase_mtie.png")


def tgm_wander_generation_mtie(mtie_taus, prtc):
    """
    Calculate mtie limits

    Input:
        `tdev_taus` array with number of samples
        `prtc` the mtie mask in ns
    Output:  wander generation mtie limit in ns
    """
    n = len(mtie_taus)
    mask = mtie_taus.copy()
    for i in range(0, (n)):
        match prtc:
            case "PRTC-B":
                if mtie_taus[i] < 54.5:
                    mask[i] = (0.275 * mtie_taus[i]) + 25
                else:
                    mask[i] = 40
            case "PRTC-A":
                if mtie_taus[i] < 274:
                    mask[i] = (0.275 * mtie_taus[i]) + 25
                else:
                    mask[i] = 100
    return mask


def tgm_mtie(df):
    """
    Calculate mtie

    Input:
        `samples` array with number of samples
    Output:  max|MTIE|
    """
    mtie_taus, mtie_devs, mtie_errs, ns = allantools.mtie(
        df.offset, rate=1.0, data_type="phase", taus=None
    )
    mtie_min = mtie_devs.min()
    mtie_max = mtie_devs.max()
    mtie_pktpk = mtie_max - mtie_min
    max_abs_mtie = max(mtie_max, abs(mtie_min))
    print("Max |MTIE|:", "{:.3f}".format(max_abs_mtie), "ns")
    plot_mtie(mtie_taus, mtie_devs)


def tgm_wander_generation_tdev(tdev_taus, prtc):
    """
    Wander limit generation (TDEV) for PRTC-B, PRTC-A based on
    observation interval
    """
    n = len(tdev_taus)
    mask = tdev_taus.copy()
    for i in range(0, n):
        match prtc:
            case "PRTC-B":
                if tdev_taus[i] > 500:
                    mask[i] = 5
                elif tdev_taus[i] < 100:
                    mask[i] = 1
                else:
                    mask[i] = 0.01 * tdev_taus[i]
            case "PRTC-A":
                if tdev_taus[i] > 1000:
                    mask[i] = 30
                elif tdev_taus[i] < 100:
                    mask[i] = 3
                else:
                    mask[i] = 0.03 * tdev_taus[i]
    return mask


def plot_tdev(tdev_taus, tdev_devs):
    """
    Plot TDEV
    """
    plt.figure()
    plt.axes(xscale="log", yscale="log")
    plt.plot(tdev_taus, tdev_devs, color="blue")
    mask_B = tgm_wander_generation_tdev(tdev_taus, "PRTC-B")
    mask_A = tgm_wander_generation_tdev(tdev_taus, "PRTC-A")

    plt.plot(tdev_taus, mask_B, color="red", linestyle="dashed", label="PRTC-B")
    plt.plot(tdev_taus, mask_A, color="green", linestyle="dashed", label="PRTC-A")

    plt.title("TDEV, DPLL Phase; ", fontsize=12, weight="bold")
    plt.xlabel("Tau (s)")
    plt.ylabel("TDEV (ns)")
    plt.legend()
    plt.savefig("dpll_phase_tdev.png")


def tgm_tdev(df):
    """
    Calculate tdev

    Input:
        `samples` array with number of samples
        `mask_tdev` the tdev mask in ns
    Output:  max|TDEV|
    """
    tdev_taus, tdev_devs, tdev_errs, ns = allantools.tdev(
        df.offset, rate=1.0, data_type="phase", taus=None
    )
    tdev_min = tdev_devs.min()
    tdev_max = tdev_devs.max()
    tdev_pktpk = tdev_max - tdev_min
    max_abs_tdev = max(tdev_max, abs(tdev_min))
    print("Max |TDEV|:", "{:.3f}".format(max_abs_tdev), "ns")
    plot_tdev(tdev_taus, tdev_devs)
