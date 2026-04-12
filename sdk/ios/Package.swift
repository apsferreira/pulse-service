// swift-tools-version: 5.9
// PulseSDK — Swift Package Manager configuration

import PackageDescription

let package = Package(
    name: "PulseSDK",
    platforms: [
        .iOS(.v15),
        .macOS(.v12)
    ],
    products: [
        .library(name: "PulseSDK", targets: ["PulseSDK"])
    ],
    targets: [
        .target(
            name: "PulseSDK",
            path: ".",
            sources: ["PulseSDK.swift"]
        )
    ]
)
