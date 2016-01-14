# Purpose

To connect to a given list of Redis Sentinels, and set the
down-after-milliseconds value to the value you specify, and report any
others. By default will only change a pod from the default value, not
just any value.


# Usage

Arguments:
  -s sentinel IP. Repeat for each sentinel to connect to.
  -c commit. By default it will only report them. Pass this flag to
		actually make the change
  -a make the change on ALL pods regardless of current value.
