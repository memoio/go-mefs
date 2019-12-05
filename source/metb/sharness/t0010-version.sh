#!/bin/sh

test_description="metb --version and --help tests"

. lib/test-lib.sh

test_expect_success "metb binary is here" '
	test -f ../bin/metb
'

test_expect_success "'metb --version' works" '
	metb --version >actual
'

test_expect_success "'metb --version' output looks good" '
	egrep "^metb version [0-9]+.[0-9]+.[0-9]+$" actual
'

test_expect_success "'metb --help' works" '
	metb --help >actual
'

test_expect_success "'metb --help' output looks good" '
	grep "COMMANDS" actual &&
	grep "USAGE" actual
'

test_done
