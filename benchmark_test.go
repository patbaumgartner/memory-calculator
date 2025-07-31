package main

import (
	"testing"
)

func BenchmarkParseMemoryString(b *testing.B) {
	testCases := []string{
		"2G",
		"512M",
		"1024K",
		"2147483648",
		"1.5G",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			_, _ = parseMemoryString(tc)
		}
	}
}

func BenchmarkFormatMemory(b *testing.B) {
	testCases := []int64{
		1024,
		1024 * 1024,
		1024 * 1024 * 1024,
		2 * 1024 * 1024 * 1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			_ = formatMemory(tc)
		}
	}
}

func BenchmarkDetectContainerMemory(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detectContainerMemory()
	}
}

func BenchmarkExtractJVMFlag(b *testing.B) {
	javaToolOptions := "-XX:MaxDirectMemorySize=10M -Xmx324661K -XX:MaxMetaspaceSize=211914K -XX:ReservedCodeCacheSize=240M -Xss1M"
	flags := []string{"-Xmx", "-Xss", "-XX:MaxMetaspaceSize", "-XX:ReservedCodeCacheSize"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, flag := range flags {
			_ = extractJVMFlag(javaToolOptions, flag)
		}
	}
}

func BenchmarkReadCgroupsV1(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = readCgroupsV1()
	}
}

func BenchmarkReadCgroupsV2(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = readCgroupsV2()
	}
}
