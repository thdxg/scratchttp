#!/usr/bin/env nu

let testdata = (open ./testdata.json)
mut pass = true

for test in $testdata {
  let res = ($test.req | nc -w 1 localhost $env.PORT)
  $pass = $pass and ($res == $test.res)
}

let result = (if $pass { 'PASSED' } else { 'FAILED' })

print $result
