import sequtils, strutils
let v = stdin.readLine.splitWhitespace.map(parseInt)
let A = v[0]
let B = v[1]
echo(A+B)
