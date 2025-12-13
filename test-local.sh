#!/bin/bash
set -e

echo "========================================"
echo "🧪 JVM Memory Calculator - Local Test"
echo "========================================"

# Cleanup
rm -f memory-calculator memory-calculator-minimal

# 1. Build Standard
echo ""
echo "🏗️  Building Standard Variant..."
go build -o memory-calculator ./cmd/memory-calculator
ls -lh memory-calculator

# 2. Build Minimal
echo ""
echo "🏗️  Building Minimal Variant..."
go build -tags minimal -o memory-calculator-minimal ./cmd/memory-calculator
ls -lh memory-calculator-minimal

# 3. Test Standard
echo ""
echo "🧪 Testing Standard Variant..."
echo "  - Help:"
./memory-calculator --help > /dev/null && echo "    ✅ Pass" || echo "    ❌ Fail"
echo "  - Calculation (2G, 100 threads, 1000 classes):"
RESULT=$(./memory-calculator --total-memory 2G --thread-count 100 --loaded-class-count 1000 --quiet)
if [[ "$RESULT" == *"-Xmx"* ]]; then
  echo "    ✅ Pass: $RESULT"
else
  echo "    ❌ Fail: $RESULT"
fi

# 4. Test Minimal
echo ""
echo "🧪 Testing Minimal Variant..."
echo "  - Help:"
./memory-calculator-minimal --help > /dev/null && echo "    ✅ Pass" || echo "    ❌ Fail"
echo "  - Calculation (2G, 100 threads, 1000 classes):"
RESULT_MIN=$(./memory-calculator-minimal --total-memory 2G --thread-count 100 --loaded-class-count 1000 --quiet)
if [[ "$RESULT_MIN" == *"-Xmx"* ]]; then
  echo "    ✅ Pass: $RESULT_MIN"
else
  echo "    ❌ Fail: $RESULT_MIN"
fi

# 5. Comparison
echo ""
echo "📊 Comparison:"
if [ "$RESULT" == "$RESULT_MIN" ]; then
    echo "  ✅ Outputs match exactly"
else
    echo "  ❌ Outputs differ!"
    echo "     Standard: $RESULT"
    echo "     Minimal:  $RESULT_MIN"
    exit 1
fi

SIZE_STD=$(du -k memory-calculator | cut -f1)
SIZE_MIN=$(du -k memory-calculator-minimal | cut -f1)
echo "  Stats: Standard=${SIZE_STD}KB, Minimal=${SIZE_MIN}KB"

echo ""
echo "✅ Test Complete!"
