// swift-tools-version:6.4
import PackageDescription

let package = Package(
    name: "JudgeLibraries",
    products: [.library(name: "JudgeLibraries", type: .static, targets: ["JudgeLibraries"])],
    dependencies: [
        .package(url: "https://github.com/apple/swift-collections.git", exact: "1.7.6"),
        .package(url: "https://github.com/apple/swift-algorithms.git", exact: "1.2.1"),
        // Algorithms depends on RealModule; pin it for reproducible image builds.
        .package(url: "https://github.com/apple/swift-numerics.git", exact: "1.1.1"),
    ],
    targets: [
        .target(name: "JudgeLibraries", dependencies: [
            .product(name: "Collections", package: "swift-collections"),
            .product(name: "Algorithms", package: "swift-algorithms"),
            .product(name: "RealModule", package: "swift-numerics"),
        ]),
    ]
)
