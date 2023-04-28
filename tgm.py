"""A Python implementation of tgm FSM"""

# pylint: disable=invalid-name

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

class Callbacks(object):
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

class Fsm(object):
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

