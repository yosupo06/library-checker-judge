import Foundation
nonisolated(unsafe) var input: [Int] = readLine()!.split(separator: " ").compactMap{ Int($0) }
print(input[0] + input[1])
