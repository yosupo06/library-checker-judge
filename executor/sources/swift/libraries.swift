import Algorithms
import Collections
import DequeModule
import HeapModule
import OrderedCollections

// Exercise linked implementations as well as generic/inlinable APIs.
var deque: Deque<Int> = [2, 3]
deque.prepend(1)
deque.append(4)
print(deque.popFirst()!, deque.popLast()!)

var heap = Heap([3, 1, 4, 2])
print(heap.popMin()!, heap.popMax()!)

let ordered = OrderedSet([3, 1, 3, 2])
print(ordered.map(String.init).joined(separator: " "))
print(Array((1...5).chunks(ofCount: 2)).map { $0.reduce(0, +) })
print(Array((1...4).combinations(ofCount: 2)).count)
let wide = Int128(Int64.max) + 1
print(wide)

// Default MainActor isolation allows ordinary helpers to use top-level state.
var total = 0
func add(_ value: Int) { total += value }
add(7)
add(8)
print(total)
