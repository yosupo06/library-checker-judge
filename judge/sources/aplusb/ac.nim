import sequtils, strutils
let v = stdin.readLine.splitWhiteSpace.map(parseInt)
let A = v[0]
let B = v[1]
echo(A+B)