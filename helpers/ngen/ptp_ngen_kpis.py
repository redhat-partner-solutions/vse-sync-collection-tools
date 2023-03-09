import argparse
import os
import pandas as pd
import numpy as np
import preprocess
import process



"""
1) Read input parameters
2) Preprocess samples and calculate required interim data 
3) process: Calculate NGEN KPIs
"""

parser = argparse.ArgumentParser(description='Process ptp4l samples to calculate NGEN KPIs.')
required = parser.add_argument_group('required arguments')
required.add_argument('-i', '--input', type = str,
                    help='Input sample data', required=True)
parser.add_argument('-c', '--clockclass', type = str,
                    help='clock class-[C,D] requirement to satisfy, defaults to Class-C', default = 'C')
parser.add_argument('-t', '--transient', type = int,
                    help='transient period, defaults to 300sec', default = 300)
parser.add_argument('-s', '--steady', type = int,
                    help='minimum steady state period to enable calculations, defaults to 2000sec', default = 2000)
parser.add_argument('-p','--plot', type = bool,
                    help='add plots to the results', default = "False")
parser.add_argument('-o','--output', 
                    help='Output file name, defaults to stdout', default="stdout")

args = parser.parse_args()


if os.path.exists(args.input):
    df=pd.read_csv(args.input, dtype={"tstamp": "float64", "phase": "float64", "state": "float64", "freq": "float64", "delay": "float64", "event": "string"})
else:
    print ("no Such File named ", args.input)

lpf_signal, update_rate, s2_count = preprocess.run(df, args.transient)
process.run(df, args.transient, args.clockclass, args.steady, args.plot, s2_count, update_rate, lpf_signal)