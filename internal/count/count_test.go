package count

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClassesOnFilesystem(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "class-count-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	err = os.WriteFile(filepath.Join(tempDir, "Test.class"), []byte{}, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.MkdirAll(filepath.Join(tempDir, "com", "example"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(tempDir, "com", "example", "Another.class"), []byte{}, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Test class counting
	count, err := JarClasses(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	expected := 2
	if count != expected {
		t.Errorf("Expected %d classes, got %d", expected, count)
	}
}

func TestClassesWithNonClassFiles(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "class-count-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	err = os.WriteFile(filepath.Join(tempDir, "Test.class"), []byte{}, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(tempDir, "readme.txt"), []byte{}, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(tempDir, "config.properties"), []byte{}, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Test class counting - should only count .class files
	count, err := JarClasses(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	expected := 1
	if count != expected {
		t.Errorf("Expected %d classes, got %d", expected, count)
	}
}

func TestJarClassesFromMissingFiles(t *testing.T) {
	paths := []string{"/nonexistent/file1.jar", "/nonexistent/file2.jar"}

	classCount, skippedCount, err := JarClassesFrom(paths...)
	if err != nil {
		t.Fatal(err)
	}

	if classCount != 0 {
		t.Errorf("Expected 0 classes, got %d", classCount)
	}

	if skippedCount != 2 {
		t.Errorf("Expected 2 skipped files, got %d", skippedCount)
	}
}

func TestEstimateModuleClasses(t *testing.T) {
	// Create a temporary file to simulate modules file
	tempFile, err := os.CreateTemp("", "modules")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	// Write some data to simulate a modules file
	data := make([]byte, 10000) // 10KB file
	_, err = tempFile.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	tempFile.Close()

	count, err := estimateModuleClasses(tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Should be at least 1000 (minimum) and roughly 10000/100 = 100
	if count < 1000 {
		t.Errorf("Expected at least 1000 classes, got %d", count)
	}
}

func TestClassesWithModulesFile(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "modules-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create lib directory
	libDir := filepath.Join(tempDir, "lib")
	err = os.MkdirAll(libDir, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	// Create modules file
	modulesFile := filepath.Join(libDir, "modules")
	data := make([]byte, 50000) // 50KB modules file
	err = os.WriteFile(modulesFile, data, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Test Classes function with modules file present
	count, err := Classes(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	if count < 1000 {
		t.Errorf("Expected at least 1000 classes for modules, got %d", count)
	}
}

func TestClassesWithoutModulesFile(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "no-modules-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a class file
	err = os.WriteFile(filepath.Join(tempDir, "Test.class"), []byte{}, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Test Classes function without modules file - should fall back to JarClasses
	count, err := Classes(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	expected := 1
	if count != expected {
		t.Errorf("Expected %d classes, got %d", expected, count)
	}
}

func TestJarClassesWithVariousExtensions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "extensions-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Test all supported class extensions
	extensions := []string{".class", ".classdata", ".clj", ".groovy", ".kts"}

	for i, ext := range extensions {
		filename := filepath.Join(tempDir, "Test"+string(rune('0'+i))+ext)
		err = os.WriteFile(filename, []byte{}, 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}

	count, err := JarClasses(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	expected := len(extensions)
	if count != expected {
		t.Errorf("Expected %d classes with various extensions, got %d", expected, count)
	}
}

func TestJarClassesWithEmptyDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "empty-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	count, err := JarClasses(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Errorf("Expected 0 classes in empty directory, got %d", count)
	}
}

func TestJarClassesWithDeepNesting(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "deep-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create deeply nested structure
	deepPath := filepath.Join(tempDir, "a", "b", "c", "d", "e", "f", "g")
	err = os.MkdirAll(deepPath, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	// Add class files at various levels
	paths := []string{
		filepath.Join(tempDir, "Root.class"),
		filepath.Join(tempDir, "a", "Level1.class"),
		filepath.Join(tempDir, "a", "b", "c", "Level3.class"),
		filepath.Join(deepPath, "Deep.class"),
	}

	for _, path := range paths {
		err = os.WriteFile(path, []byte{}, 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}

	count, err := JarClasses(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	expected := len(paths)
	if count != expected {
		t.Errorf("Expected %d classes in nested structure, got %d", expected, count)
	}
}

func TestJarClassesWithActualJarFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "jar-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a real JAR file with class entries
	jarPath := filepath.Join(tempDir, "test.jar")
	jarFile, err := os.Create(jarPath)
	if err != nil {
		t.Fatal(err)
	}

	zipWriter := zip.NewWriter(jarFile)

	// Add class files to the JAR
	classFiles := []string{
		"com/example/Test.class",
		"com/example/util/Helper.class",
		"META-INF/MANIFEST.MF",   // Non-class file
		"application.properties", // Non-class file
	}

	for _, fileName := range classFiles {
		fileWriter, writeErr := zipWriter.Create(fileName)
		if writeErr != nil {
			t.Fatal(writeErr)
		}
		_, writeErr = fileWriter.Write([]byte("fake class content"))
		if writeErr != nil {
			t.Fatal(writeErr)
		}
	}

	zipWriter.Close()
	jarFile.Close()

	count, err := JarClasses(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should count only .class files (2 out of 4 files)
	expected := 2
	if count != expected {
		t.Errorf("Expected %d classes in JAR file, got %d", expected, count)
	}
}

func TestJarClassesWithZeroByteNoneJar(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "none-jar-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create zero-byte JAR with 'none' in name
	noneJarPath := filepath.Join(tempDir, "svm-none.jar")
	err = os.WriteFile(noneJarPath, []byte{}, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Should not cause error and should skip the file
	count, err := JarClasses(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Errorf("Expected 0 classes from zero-byte none JAR, got %d", count)
	}
}

func TestJarClassesWithInvalidJar(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "invalid-jar-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a file with .jar extension but invalid ZIP content
	invalidJarPath := filepath.Join(tempDir, "invalid.jar")
	err = os.WriteFile(invalidJarPath, []byte("this is not a zip file"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Should handle invalid JAR gracefully and continue
	count, err := JarClasses(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Errorf("Expected 0 classes from invalid JAR, got %d", count)
	}
}

func TestJarClassesFromMixedPaths(t *testing.T) {
	tempDir1, err := os.MkdirTemp("", "mixed-test1")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir1)

	tempDir2, err := os.MkdirTemp("", "mixed-test2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir2)

	// Create class files in first directory
	err = os.WriteFile(filepath.Join(tempDir1, "Test1.class"), []byte{}, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Create class files in second directory
	err = os.WriteFile(filepath.Join(tempDir2, "Test2.class"), []byte{}, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	paths := []string{tempDir1, tempDir2, "/nonexistent"}
	classCount, skippedCount, err := JarClassesFrom(paths...)
	if err != nil {
		t.Fatal(err)
	}

	if classCount != 2 {
		t.Errorf("Expected 2 classes from mixed paths, got %d", classCount)
	}

	if skippedCount != 1 {
		t.Errorf("Expected 1 skipped path, got %d", skippedCount)
	}
}

func TestJarContents(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected int
	}{
		{"Class file", "Test.class", 1},
		{"Classdata file", "Test.classdata", 1},
		{"Clojure file", "core.clj", 1},
		{"Groovy file", "Script.groovy", 1},
		{"Kotlin script", "build.kts", 1},
		{"Non-class file", "README.txt", 0},
		{"Properties file", "config.properties", 0},
		{"Empty name", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock zip.File entry
			file := &zip.File{
				FileHeader: zip.FileHeader{
					Name: tt.filename,
				},
			}

			result := jarContents(file)
			if result != tt.expected {
				t.Errorf("jarContents(%s) = %d, expected %d", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("Permission denied directory", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("Skipping permission test when running as root")
		}

		tempDir, err := os.MkdirTemp("", "permission-test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)

		// Create subdirectory with no permissions
		restrictedDir := filepath.Join(tempDir, "restricted")
		err = os.MkdirAll(restrictedDir, 0o000)
		if err != nil {
			t.Fatal(err)
		}

		// This should handle the permission error gracefully
		_, err = JarClasses(tempDir)
		if err == nil {
			t.Log("Expected permission error, but got none - this is OK if filesystem doesn't enforce permissions")
		} else if !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("Expected permission denied error, got: %v", err)
		}
	})

	t.Run("Very small modules file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "tiny-modules")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tempFile.Name())

		// Write minimal data (less than 100 bytes)
		err = os.WriteFile(tempFile.Name(), []byte("tiny"), 0o644)
		if err != nil {
			t.Fatal(err)
		}

		count, err := estimateModuleClasses(tempFile.Name())
		if err != nil {
			t.Fatal(err)
		}

		// Should still return minimum of 1000
		if count != 1000 {
			t.Errorf("Expected 1000 classes for tiny modules file, got %d", count)
		}
	})

	t.Run("Modules file stat error", func(t *testing.T) {
		_, err := estimateModuleClasses("/nonexistent/modules")
		if err == nil {
			t.Error("Expected error for nonexistent modules file")
		}
	})

	t.Run("Classes with lib dir stat error", func(t *testing.T) {
		// Try to access a directory that doesn't exist
		_, err := Classes("/nonexistent/directory")
		if err == nil {
			t.Error("Expected error for nonexistent directory")
		}
	})
}
