#!/usr/bin/env nu

let testdata = (open ./testdata.json)
mut pass = true

for test in $testdata {
  let res = ($test.req | nc -w 1 localhost $env.PORT)
  if $res != $test.res {
    $pass = false
    print $"request: ($test.req | to json)"
    let dir = (mktemp -d)
    $test.res | save $"($dir)/expected"
    $res | save $"($dir)/actual"
    do -i { diff -u $"($dir)/expected" $"($dir)/actual" }
    rm -r $dir
    break
  }
}

let result = (if $pass { 'PASSED' } else { 'FAILED' })

print $result
