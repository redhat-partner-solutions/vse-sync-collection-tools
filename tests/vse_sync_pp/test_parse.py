"""Test cases for vse_sync_pp.parse"""

import json
from io import StringIO
from decimal import Decimal

from unittest import TestCase

from vse_sync_pp.parse import (
    parse, Parsed,
    DecimalEncoder,
)

class TestParse(TestCase):
    """Test cases for vse_sync_pp.parse.parse"""
    line = (
        'ts2phc[681011.839]: [ts2phc.0.config] '
        'ens7f1 master offset          0 s2 freq      -0'
    )
    def test_parse(self):
        """Test vse_sync_pp.parse.parse parses test line"""
        self.assertEqual(
            tuple(parse(StringIO(self.line))),
            (Parsed(Decimal('681011.839'), 'ens7f1', 0, 's2'),),
        )

class TestDecimalEncoder(TestCase):
    """Test cases for vse_sync_pp.parse.DecimalEncoder"""
    def test_encode(self):
        """Test vse_sync_pp.parse.DecimalEncoder encodes Decimal"""
        self.assertEqual(
            json.dumps(Decimal('123.456'), cls=DecimalEncoder),
            '123.456',
        )
