#!/bin/sh

test_description="metb stop tests"

. lib/test-lib.sh

METB_ROOT=.

test_expect_success "metb init works" '
	../bin/metb init -n 3
'

test_expect_success "metb start works" '
	../bin/metb start --args --debug
'

test_expect_success "metb stop works" '
	../bin/metb stop
'

for i in {0..2}; do
	test_expect_success "daemon '$i' was shut down gracefully" '
		cat testbed/'$i'/daemon.stderr | tail -1 | grep "Gracefully shut down daemon"
	'
done

test_done
