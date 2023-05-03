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


def plot_phase_dpll(df):
    plt.figure()
    df.loc[1 : len(df)].plot(kind="line", linewidth=0.25, y="offset", color="blue")
    plt.title("DPLL Phase Offset", fontsize=12, weight="bold")
    plt.xlabel("Time (s)")
    plt.ylabel("Phase Offset (ns)")
    plt.savefig("dpll.png")


def run(df):
    """
    Calculate  KPIs

    Input: samples
    Output: plot phase dpll
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
