#!/usr/bin/env nu

let testdata = (open ./testdata.json)
mut pass = true

for test in $testdata {
  let actual = ($test.req | nc -w 1 localhost $env.PORT | into binary)
  let expected = ($test.res | into binary)
  if $actual != $expected {
    $pass = false
    print $"request: ($test.req)"
    let dir = (mktemp -d)
    $actual | save $"($dir)/actual"
    $expected | save $"($dir)/expected"
    do -i { diff -u $"($dir)/expected" $"($dir)/actual" }
    rm -r $dir
    break
  }
}

let result = (if $pass { 'PASSED' } else { 'FAILED' })

print $result
