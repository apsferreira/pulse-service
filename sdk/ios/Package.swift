// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "Pulse",
    platforms: [
        .iOS(.v14),
        .macOS(.v11)
    ],
    products: [
        .library(
            name: "Pulse",
            targets: ["Pulse"]
        )
    ],
    targets: [
        .target(
            name: "Pulse",
            dependencies: [],
            path: "Sources/Pulse"
        ),
        .testTarget(
            name: "PulseTests",
            dependencies: ["Pulse"],
            path: "Tests/PulseTests"
        )
    ]
)
