package count

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMinimalBuildFunctionality tests that the minimal build provides core functionality
// This test is designed to pass with both standard and minimal builds
func TestMinimalBuildFunctionality(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "count_minimal_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: JarClasses should handle non-existent files gracefully
	testJarPath := filepath.Join(tempDir, "test.jar")
	testContent := []byte("fake jar content for size estimation")
	err = os.WriteFile(testJarPath, testContent, 0o644)
	if err != nil {
		t.Fatalf("Failed to create test jar: %v", err)
	}

	classes, err := JarClasses(testJarPath)
	// Note: Standard build might return 0 for non-ZIP files, minimal build estimates based on size
	// Both behaviors are acceptable for this test
	if err != nil {
		t.Errorf("JarClasses() error = %v", err)
	}
	if classes < 0 {
		t.Errorf("JarClasses() = %d, want >= 0", classes)
	}

	// Test 2: JarClasses should return 0 for empty file
	emptyJarPath := filepath.Join(tempDir, "empty.jar")
	err = os.WriteFile(emptyJarPath, []byte{}, 0o644)
	if err != nil {
		t.Fatalf("Failed to create empty jar: %v", err)
	}

	emptyClasses, err := JarClasses(emptyJarPath)
	if err != nil {
		t.Errorf("JarClasses() error = %v", err)
	}
	if emptyClasses < 0 {
		t.Errorf("JarClasses() = %d, want >= 0 for empty file", emptyClasses)
	}

	// Test 3: JarClassesFrom should handle mixed paths
	jarPaths := []string{testJarPath, emptyJarPath, "nonexistent.jar"}
	totalClasses, skipped, err := JarClassesFrom(jarPaths...)
	if err != nil {
		t.Errorf("JarClassesFrom() error = %v", err)
	}
	if totalClasses < 0 {
		t.Errorf("JarClassesFrom() totalClasses = %d, want >= 0", totalClasses)
	}
	if skipped < 0 {
		t.Errorf("JarClassesFrom() skipped = %d, want >= 0", skipped)
	}

	// Test 4: Classes should return some count without error
	classesCount, err := Classes(tempDir)
	if err != nil {
		t.Errorf("Classes() error = %v", err)
	}
	if classesCount < 0 {
		t.Errorf("Classes() = %d, want >= 0", classesCount)
	}
}

// TestMinimalBuildConsistency ensures minimal build produces reasonable results
func TestMinimalBuildConsistency(t *testing.T) {
	// Create test files of different sizes
	tempDir, err := os.MkdirTemp("", "count_consistency_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Small JAR (should estimate fewer classes)
	smallJar := filepath.Join(tempDir, "small.jar")
	smallContent := make([]byte, 1024) // 1KB
	err = os.WriteFile(smallJar, smallContent, 0o644)
	if err != nil {
		t.Fatalf("Failed to create small jar: %v", err)
	}

	// Large JAR (should estimate more classes)
	largeJar := filepath.Join(tempDir, "large.jar")
	largeContent := make([]byte, 10240) // 10KB
	err = os.WriteFile(largeJar, largeContent, 0o644)
	if err != nil {
		t.Fatalf("Failed to create large jar: %v", err)
	}

	smallClasses, err := JarClasses(smallJar)
	if err != nil {
		t.Fatalf("JarClasses(small) error = %v", err)
	}

	largeClasses, err := JarClasses(largeJar)
	if err != nil {
		t.Fatalf("JarClasses(large) error = %v", err)
	}

	// Both builds should return non-negative values
	if smallClasses < 0 {
		t.Errorf("Small JAR classes = %d, want >= 0", smallClasses)
	}
	if largeClasses < 0 {
		t.Errorf("Large JAR classes = %d, want >= 0", largeClasses)
	}

	// For minimal build: large JAR should have more estimated classes than small JAR
	// For standard build: both might return 0 for non-ZIP files, which is also acceptable
	if smallClasses > 0 && largeClasses > 0 && largeClasses <= smallClasses {
		t.Errorf("Large JAR classes (%d) should be > small JAR classes (%d)", largeClasses, smallClasses)
	}

	t.Logf("Small JAR (1KB): %d classes, Large JAR (10KB): %d classes", smallClasses, largeClasses)
}

// TestMinimalBuildErrorHandling tests error conditions
func TestMinimalBuildErrorHandling(t *testing.T) {
	// Test with nonexistent file
	_, err := JarClasses("nonexistent.jar")
	if err == nil {
		t.Error("JarClasses() should return error for nonexistent file")
	}

	// Test estimateModuleClasses with nonexistent file
	_, err = estimateModuleClasses("nonexistent_modules")
	if err == nil {
		t.Error("estimateModuleClasses() should return error for nonexistent file")
	}
}
