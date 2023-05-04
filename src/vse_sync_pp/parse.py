"""Parse log messages"""

from argparse import ArgumentParser
import sys
from contextlib import nullcontext

import re
from collections import namedtuple
from decimal import Decimal

import json

def build_regexp(interface=None):
    """Return a regular expression string for parsing log file lines.

    If `interface` then only parse lines for the specified interface.
    """
    return r'\s'.join((
        r'^ts2phc' +
        r'\[([1-9][0-9]*\.[0-9]{3})\]:', # timestamp
        r'\[ts2phc\.0\..*\]',
        fr'({interface})' if interface else r'(\S+)', # interface
        r'master offset\s*',
        r'(-?[0-9]+)', # time error
        r'(\S+)', # state
        r'.*$',
    ))

Parsed = namedtuple('Parsed', ('timestamp', 'interface', 'terror', 'state'))

def parse(fid, interface=None):
    """Generator yielding :class:`Parsed` tuples for `fid` lines of interest"""
    regexp = re.compile(build_regexp(interface))
    for line in fid:
        matched = regexp.match(line)
        if matched:
            yield Parsed(
                Decimal(matched.group(1)),
                matched.group(2),
                int(matched.group(3)),
                matched.group(4),
            )

class DecimalEncoder(json.JSONEncoder):
    """A JSON encoder accepting :class:`Decimal` values"""
    def default(self, o):
        """Return a commonly serializable value from `o`"""
        if isinstance(o, Decimal):
            return float(o)
        return super(o)

def main():
    """Parse log messages"""
    aparser = ArgumentParser(description=main.__doc__)
    aparser.add_argument(
        '-i', '--interface',
        help="only process lines related to this interface",
    )
    aparser.add_argument(
        'input',
        help="the log file to parse, or '-' to read from stdin",
    )
    args = aparser.parse_args()
    with (  open(args.input, encoding='utf-8')
            if args.input != '-' else
            nullcontext(sys.stdin)
        ) as fid:
        for parsed in parse(fid, args.interface):
            print(json.dumps(parsed, cls=DecimalEncoder))

if __name__ == '__main__':
    main()
