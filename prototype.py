#!/usr/bin/env python3

"""Prototype test script for https://issues.redhat.com/browse/TTINTEL-239"""

from argparse import ArgumentParser
import sys
from contextlib import nullcontext

import re
from collections import namedtuple
from decimal import Decimal

import numpy as np
import matplotlib.pyplot as plt

### <generated>
# pylint: disable=invalid-name,unused-argument

STATE_TESTING = 0
STATE_TESTING_INITIALIZING = 1
STATE_TESTING_LOCKED = 2
STATE_TESTING_LOCKED_STABILIZING = 3
STATE_TESTING_LOCKED_STABLE = 4
STATE_TESTING_LOCKED_WOBBLE = 5
STATE_TESTING_LOCKING = 6
STATE_TESTING_UNKNOWN = 7

def initial_transition(fsm, arg):
    """Transition into the initial state"""
    fsm.state = STATE_TESTING
    fsm.state = STATE_TESTING_UNKNOWN

def handle_S0_in_TESTING(fsm, arg):
    """Handle event S0 in state /TESTING"""
    fsm.state = STATE_TESTING_INITIALIZING

def handle_S0_in_TESTING_INITIALIZING(fsm, arg):
    """Handle event S0 in state /TESTING/INITIALIZING"""
    fsm.state = STATE_TESTING_INITIALIZING
    fsm.state = STATE_TESTING_INITIALIZING

def handle_S0_in_TESTING_LOCKED(fsm, arg):
    """Handle event S0 in state /TESTING/LOCKED"""
    fsm.callbacks.action_unstable(fsm, arg)
    fsm.state = STATE_TESTING_LOCKED
    fsm.state = STATE_TESTING_INITIALIZING

def handle_S0_in_TESTING_LOCKED_STABILIZING(fsm, arg):
    """Handle event S0 in state /TESTING/LOCKED/STABILIZING"""
    fsm.state = STATE_TESTING_LOCKED_STABILIZING
    fsm.callbacks.action_unstable(fsm, arg)
    fsm.state = STATE_TESTING_LOCKED
    fsm.state = STATE_TESTING_INITIALIZING

def handle_S0_in_TESTING_LOCKED_STABLE(fsm, arg):
    """Handle event S0 in state /TESTING/LOCKED/STABLE"""
    fsm.state = STATE_TESTING_LOCKED_STABLE
    fsm.callbacks.action_unstable(fsm, arg)
    fsm.state = STATE_TESTING_LOCKED
    fsm.state = STATE_TESTING_INITIALIZING

def handle_S0_in_TESTING_LOCKED_WOBBLE(fsm, arg):
    """Handle event S0 in state /TESTING/LOCKED/WOBBLE"""
    fsm.state = STATE_TESTING_LOCKED_WOBBLE
    fsm.callbacks.action_unstable(fsm, arg)
    fsm.state = STATE_TESTING_LOCKED
    fsm.state = STATE_TESTING_INITIALIZING

def handle_S0_in_TESTING_LOCKING(fsm, arg):
    """Handle event S0 in state /TESTING/LOCKING"""
    fsm.state = STATE_TESTING_LOCKING
    fsm.state = STATE_TESTING_INITIALIZING

def handle_S0_in_TESTING_UNKNOWN(fsm, arg):
    """Handle event S0 in state /TESTING/UNKNOWN"""
    fsm.state = STATE_TESTING_UNKNOWN
    fsm.state = STATE_TESTING_INITIALIZING

TRANSITION_ON_EVENT_S0 = {
    STATE_TESTING: handle_S0_in_TESTING,
    STATE_TESTING_INITIALIZING: handle_S0_in_TESTING_INITIALIZING,
    STATE_TESTING_LOCKED: handle_S0_in_TESTING_LOCKED,
    STATE_TESTING_LOCKED_STABILIZING: handle_S0_in_TESTING_LOCKED_STABILIZING,
    STATE_TESTING_LOCKED_STABLE: handle_S0_in_TESTING_LOCKED_STABLE,
    STATE_TESTING_LOCKED_WOBBLE: handle_S0_in_TESTING_LOCKED_WOBBLE,
    STATE_TESTING_LOCKING: handle_S0_in_TESTING_LOCKING,
    STATE_TESTING_UNKNOWN: handle_S0_in_TESTING_UNKNOWN,
}

def handle_S1_in_TESTING(fsm, arg):
    """Handle event S1 in state /TESTING"""
    fsm.state = STATE_TESTING_LOCKING

def handle_S1_in_TESTING_INITIALIZING(fsm, arg):
    """Handle event S1 in state /TESTING/INITIALIZING"""
    fsm.state = STATE_TESTING_INITIALIZING
    fsm.state = STATE_TESTING_LOCKING

def handle_S1_in_TESTING_LOCKED(fsm, arg):
    """Handle event S1 in state /TESTING/LOCKED"""
    fsm.callbacks.action_unstable(fsm, arg)
    fsm.state = STATE_TESTING_LOCKED
    fsm.state = STATE_TESTING_LOCKING

def handle_S1_in_TESTING_LOCKED_STABILIZING(fsm, arg):
    """Handle event S1 in state /TESTING/LOCKED/STABILIZING"""
    fsm.state = STATE_TESTING_LOCKED_STABILIZING
    fsm.callbacks.action_unstable(fsm, arg)
    fsm.state = STATE_TESTING_LOCKED
    fsm.state = STATE_TESTING_LOCKING

def handle_S1_in_TESTING_LOCKED_STABLE(fsm, arg):
    """Handle event S1 in state /TESTING/LOCKED/STABLE"""
    fsm.state = STATE_TESTING_LOCKED_STABLE
    fsm.callbacks.action_unstable(fsm, arg)
    fsm.state = STATE_TESTING_LOCKED
    fsm.state = STATE_TESTING_LOCKING

def handle_S1_in_TESTING_LOCKED_WOBBLE(fsm, arg):
    """Handle event S1 in state /TESTING/LOCKED/WOBBLE"""
    fsm.state = STATE_TESTING_LOCKED_WOBBLE
    fsm.callbacks.action_unstable(fsm, arg)
    fsm.state = STATE_TESTING_LOCKED
    fsm.state = STATE_TESTING_LOCKING

def handle_S1_in_TESTING_LOCKING(fsm, arg):
    """Handle event S1 in state /TESTING/LOCKING"""
    fsm.state = STATE_TESTING_LOCKING
    fsm.state = STATE_TESTING_LOCKING

def handle_S1_in_TESTING_UNKNOWN(fsm, arg):
    """Handle event S1 in state /TESTING/UNKNOWN"""
    fsm.state = STATE_TESTING_UNKNOWN
    fsm.state = STATE_TESTING_LOCKING

TRANSITION_ON_EVENT_S1 = {
    STATE_TESTING: handle_S1_in_TESTING,
    STATE_TESTING_INITIALIZING: handle_S1_in_TESTING_INITIALIZING,
    STATE_TESTING_LOCKED: handle_S1_in_TESTING_LOCKED,
    STATE_TESTING_LOCKED_STABILIZING: handle_S1_in_TESTING_LOCKED_STABILIZING,
    STATE_TESTING_LOCKED_STABLE: handle_S1_in_TESTING_LOCKED_STABLE,
    STATE_TESTING_LOCKED_WOBBLE: handle_S1_in_TESTING_LOCKED_WOBBLE,
    STATE_TESTING_LOCKING: handle_S1_in_TESTING_LOCKING,
    STATE_TESTING_UNKNOWN: handle_S1_in_TESTING_UNKNOWN,
}

def handle_S2_in_TESTING(fsm, arg):
    """Handle event S2 in state /TESTING"""
    if fsm.callbacks.condition_tight(fsm, arg):
        fsm.state = STATE_TESTING_LOCKED
        fsm.state = STATE_TESTING_LOCKED_STABLE
        fsm.callbacks.action_stable(fsm, arg)
        return
    fsm.state = STATE_TESTING_LOCKED
    fsm.state = STATE_TESTING_LOCKED_STABILIZING

def handle_S2_in_TESTING_INITIALIZING(fsm, arg):
    """Handle event S2 in state /TESTING/INITIALIZING"""
    if fsm.callbacks.condition_tight(fsm, arg):
        fsm.state = STATE_TESTING_INITIALIZING
        fsm.state = STATE_TESTING_LOCKED
        fsm.state = STATE_TESTING_LOCKED_STABLE
        fsm.callbacks.action_stable(fsm, arg)
        return
    fsm.state = STATE_TESTING_INITIALIZING
    fsm.state = STATE_TESTING_LOCKED
    fsm.state = STATE_TESTING_LOCKED_STABILIZING

def handle_S2_in_TESTING_LOCKED(fsm, arg):
    """Handle event S2 in state /TESTING/LOCKED"""
    if fsm.callbacks.condition_tight(fsm, arg):
        fsm.state = STATE_TESTING_LOCKED_STABLE
        fsm.callbacks.action_stable(fsm, arg)
        return

def handle_S2_in_TESTING_LOCKED_STABILIZING(fsm, arg):
    """Handle event S2 in state /TESTING/LOCKED/STABILIZING"""
    if fsm.callbacks.condition_tight(fsm, arg):
        fsm.state = STATE_TESTING_LOCKED_STABILIZING
        fsm.state = STATE_TESTING_LOCKED_STABLE
        fsm.callbacks.action_stable(fsm, arg)
        return

def handle_S2_in_TESTING_LOCKED_STABLE(fsm, arg):
    """Handle event S2 in state /TESTING/LOCKED/STABLE"""
    if not fsm.callbacks.condition_tight(fsm, arg):
        fsm.state = STATE_TESTING_LOCKED_STABLE
        fsm.state = STATE_TESTING_LOCKED_WOBBLE
        fsm.callbacks.action_wobble(fsm, arg)
        return

def handle_S2_in_TESTING_LOCKED_WOBBLE(fsm, arg):
    """Handle event S2 in state /TESTING/LOCKED/WOBBLE"""
    if fsm.callbacks.condition_tight(fsm, arg):
        fsm.state = STATE_TESTING_LOCKED_WOBBLE
        fsm.state = STATE_TESTING_LOCKED_STABLE
        fsm.callbacks.action_stable(fsm, arg)
        return

def handle_S2_in_TESTING_LOCKING(fsm, arg):
    """Handle event S2 in state /TESTING/LOCKING"""
    if fsm.callbacks.condition_tight(fsm, arg):
        fsm.state = STATE_TESTING_LOCKING
        fsm.state = STATE_TESTING_LOCKED
        fsm.state = STATE_TESTING_LOCKED_STABLE
        fsm.callbacks.action_stable(fsm, arg)
        return
    fsm.state = STATE_TESTING_LOCKING
    fsm.state = STATE_TESTING_LOCKED
    fsm.state = STATE_TESTING_LOCKED_STABILIZING

def handle_S2_in_TESTING_UNKNOWN(fsm, arg):
    """Handle event S2 in state /TESTING/UNKNOWN"""
    if fsm.callbacks.condition_tight(fsm, arg):
        fsm.state = STATE_TESTING_UNKNOWN
        fsm.state = STATE_TESTING_LOCKED
        fsm.state = STATE_TESTING_LOCKED_STABLE
        fsm.callbacks.action_stable(fsm, arg)
        return
    fsm.state = STATE_TESTING_UNKNOWN
    fsm.state = STATE_TESTING_LOCKED
    fsm.state = STATE_TESTING_LOCKED_STABILIZING

TRANSITION_ON_EVENT_S2 = {
    STATE_TESTING: handle_S2_in_TESTING,
    STATE_TESTING_INITIALIZING: handle_S2_in_TESTING_INITIALIZING,
    STATE_TESTING_LOCKED: handle_S2_in_TESTING_LOCKED,
    STATE_TESTING_LOCKED_STABILIZING: handle_S2_in_TESTING_LOCKED_STABILIZING,
    STATE_TESTING_LOCKED_STABLE: handle_S2_in_TESTING_LOCKED_STABLE,
    STATE_TESTING_LOCKED_WOBBLE: handle_S2_in_TESTING_LOCKED_WOBBLE,
    STATE_TESTING_LOCKING: handle_S2_in_TESTING_LOCKING,
    STATE_TESTING_UNKNOWN: handle_S2_in_TESTING_UNKNOWN,
}

class Callbacks():
    """Interface for tgm FSM condition and action callbacks"""
    @staticmethod
    def condition_tight(fsm, arg):
        """Callback for tgm FSM condition tight"""
        raise NotImplementedError
    @staticmethod
    def action_stable(fsm, arg):
        """Callback for tgm FSM action stable"""
        raise NotImplementedError
    @staticmethod
    def action_unstable(fsm, arg):
        """Callback for tgm FSM action unstable"""
        raise NotImplementedError
    @staticmethod
    def action_wobble(fsm, arg):
        """Callback for tgm FSM action wobble"""
        raise NotImplementedError

class Fsm():
    """A class for tgm FSM instances"""
    def __init__(self, callbacks=None, data=None, arg=None):
        self.state = None
        self.callbacks = self if callbacks is None else callbacks
        self.data = self if data is None else data
        initial_transition(self, arg)
    def inject_S0(self, arg=None):
        """Inject event S0 with event `arg`"""
        try:
            TRANSITION_ON_EVENT_S0[self.state](self, arg)
        except KeyError:
            pass
    def inject_S1(self, arg=None):
        """Inject event S1 with event `arg`"""
        try:
            TRANSITION_ON_EVENT_S1[self.state](self, arg)
        except KeyError:
            pass
    def inject_S2(self, arg=None):
        """Inject event S2 with event `arg`"""
        try:
            TRANSITION_ON_EVENT_S2[self.state](self, arg)
        except KeyError:
            pass

# pylint: enable=invalid-name,unused-argument
### </generated>

# example log file line of interest
#ts2phc[681011.839]: [ts2phc.0.config] ens7f1 master offset          0 s2 freq      -0

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

Targ = namedtuple('Targ', ('tdelta', 'terror'))

class Tester(Callbacks):
    """Prototype tester."""
    def __init__(self, transient, stable):
        super().__init__()
        self._transient = transient
        self._stable = stable
        self._tzero = None
        self._targs = []
        self._events = []
        self._fsm = Fsm(callbacks=self)
        self._inject = {
            's0': self._fsm.inject_S0,
            's1': self._fsm.inject_S1,
            's2': self._fsm.inject_S2,
        }
    # pylint: disable=arguments-differ
    def condition_tight(self, fsm, targ):
        return abs(targ.terror) <= self._stable
    def action_stable(self, fsm, targ):
        self._events.append((targ, 'stable'))
    def action_wobble(self, fsm, targ):
        self._events.append((targ, 'wobble'))
    def action_unstable(self, fsm, targ):
        self._events.append((targ, 'unstable'))
    # pylint: enable=arguments-differ
    def parsed(self, parsed):
        """Feed `parsed` sample from log file into FSM."""
        if self._tzero is None:
            self._tzero = parsed.timestamp
        targ = Targ(parsed.timestamp - self._tzero, parsed.terror)
        self._inject[parsed.state](targ)
        self._targs.append(targ)
    def plot(self, filename):
        """Plot samples over time with events, also histogram, to `filename`."""
        colors = {'stable': 'green', 'wobble': 'orange', 'unstable': 'red'}
        fig, (ax1, ax2) = plt.subplots(2, constrained_layout=True)
        fig.set_size_inches(10, 8)
        ax1.axhline(0, color='black')
        for (targ, event) in self._events:
            ax1.axvline(targ.tdelta, color=colors[event])
        (tdeltas, terrors) = zip(*self._targs)
        if any((abs(t) > 10 for t in terrors)):
            ax1.set_yscale('symlog', linthresh=10)
        idx = 0
        for tdelta in tdeltas:
            if self._transient <= tdelta:
                break
            idx += 1
        ax1.plot(tdeltas[:idx], terrors[:idx], '.', color='#AAA')
        ax1.plot(tdeltas[idx:], terrors[idx:], '.')
        ax1.grid()
        ax1.set_title('Time error')
        counts, bins = np.histogram(np.array(terrors), bins='scott')
        ax2.hist(bins[:-1], bins, weights=counts)
        ax2.set_yscale('symlog', linthresh=10)
        ax2.set_title('Histogram')
        plt.savefig(filename)
    def result(self):
        """Return a 2-tuple (boolean test result, descriptive detail).

        Test failure is any event other than stable occurring outside the
        transient window: otherwise success. (This is too simple: we need to
        consider what state the clock is in _just after_ the the transient
        window closes. Likely not the only condition, so treating as ok for a
        prototype.)
        """
        result = True
        detail = [f"Time errors greater than {self._stable}ns are significant"]
        if self._transient:
            detail.append(f"First {self._transient}s of logs not significant")
        for (targ, event) in self._events:
            if self._transient <= targ.tdelta:
                if event != 'stable':
                    result = False
                    detail.append(f"Test failed: {event} after {targ.tdelta}s")
        return (result, '\n'.join(detail))

def main():
    """Prototype test script for https://issues.redhat.com/browse/TTINTEL-239"""
    aparser = ArgumentParser(description=main.__doc__)
    aparser.add_argument(
        '-i', '--interface',
        help="only process lines related to this interface",
    )
    aparser.add_argument(
        '-t', '--transient', type=int, default=0,
        help="ignore instability until `transient` secs after first log",
    )
    aparser.add_argument(
        '-s', '--stable', type=int, default=40,
        help="consider clock stable when offset is within |stable| nanoseconds",
    )
    aparser.add_argument(
        'input',
        help="the log file to parse, or '-' to read from stdin",
    )
    aparser.add_argument(
        'output', nargs='?',
        help="optional filename for image plot",
    )
    args = aparser.parse_args()
    tester = Tester(
        transient=0 if args.transient <= 0 else args.transient,
        stable=abs(args.stable),
    )
    with (  open(args.input, encoding='utf-8')
            if args.input != '-' else
            nullcontext(sys.stdin)
        ) as fid:
        for parsed in parse(fid, args.interface):
            tester.parsed(parsed)
    if args.output:
        tester.plot(args.output)
    (result, detail) = tester.result()
    print(detail)
    sys.exit(int(not result))

if __name__ == '__main__':
    main()
