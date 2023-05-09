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

import argparse
import os
import pandas as pd
import process

"""
1) Read input parameters
2) Preprocess samples and calculate required interim data
3) process: Calculate DPLL KPIs
"""

parser = argparse.ArgumentParser(
    description="Process dpll samples to calculate NGEN KPIs."
)
required = parser.add_argument_group("required arguments")
required.add_argument(
    "-i", "--input", type=str, help="Input sample data", required=True
)

args = parser.parse_args()


if os.path.exists(args.input):
    df = pd.read_csv(
        args.input,
        dtype={
            "tstamp": "float64",
            "eecst": "int",
            "phasest": "int",
            "offset": "float64",
        },
    )
else:
    raise FileNotFoundError


process.tgm_terror(df)
process.tgm_cte(df)
process.tgm_mtie(df)
process.tgm_tdev(df)
