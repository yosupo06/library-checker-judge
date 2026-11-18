#!/bin/sh
set -eu

# Build dependencies once in the image, never while compiling a submission.
swift build --build-system native -c release --product JudgeLibraries --disable-automatic-resolution
bin_path=$(swift build --build-system native -c release --show-bin-path)
mkdir -p /opt/swift-libraries/modules /opt/swift-libraries/lib /opt/swift-libraries/include /opt/swift-libraries/licenses
cp -R "$bin_path/Modules/." /opt/swift-libraries/modules/
cp "$bin_path/libJudgeLibraries.a" /opt/swift-libraries/lib/
# Algorithms' RealModule dependency imports this Clang module, including from
# serialized @inlinable bodies. Preserve its relative header/module-map layout.
cp -R .build/checkouts/swift-numerics/Sources/_NumericsShims/include/. /opt/swift-libraries/include/
for package in swift-collections swift-algorithms swift-numerics; do
    mkdir -p "/opt/swift-libraries/licenses/$package"
    cp ".build/checkouts/$package/LICENSE.txt" "/opt/swift-libraries/licenses/$package/"
done
