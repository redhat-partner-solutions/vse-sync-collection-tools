### SPDX-License-Identifier: GPL-2.0-or-later

from .kernel import ppsdpll
#from .kernel import eecdpll

ANALYZERS = {
	cls.id_: cls for cls in (
        ppsdpll.TimeError,
        ppsdpll.constanTimeError,
        ppsdpll.TimeDeviaton,
        ppsdpll.MaxTimeIntervalError
    )
}

def main():
	"""Print Synchronization KPIs from parsed data to stdout"""
	
if __name__ == '__main__':
    main()