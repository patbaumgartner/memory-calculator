//go:build minimal

package count

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Minimal version without ZIP support
func JarClassesFrom(jarPaths ...string) (int, int, error) {
	var classCount, skipped int

	for _, jarPath := range jarPaths {
		if !strings.HasSuffix(strings.ToLower(jarPath), ".jar") {
			skipped++
			continue
		}

		if _, err := os.Stat(jarPath); os.IsNotExist(err) {
			skipped++
			continue
		}

		// For minimal build, estimate based on file size
		if info, err := os.Stat(jarPath); err == nil {
			// Rough estimate: 1 class per 2KB
			estimatedClasses := int(info.Size() / 2048)
			if estimatedClasses < 10 {
				estimatedClasses = 10 // minimum estimate
			}
			classCount += estimatedClasses
		} else {
			skipped++
		}
	}

	return classCount, skipped, nil
}

// Minimal version that only counts .class files directly
func Classes(dirPath string) (int, error) {
	var classCount int

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors in minimal mode
		}

		if strings.HasSuffix(strings.ToLower(info.Name()), ".class") {
			classCount++
		}

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("unable to walk %s\n%w", dirPath, err)
	}

	return classCount, nil
}

// JarClasses estimates class count based on file size (minimal implementation)
func JarClasses(path string) (int, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	// Estimate classes based on file size
	// Rough estimate: 1 class per 2KB on average for typical JAR files
	size := fileInfo.Size()
	if size == 0 {
		return 0, nil
	}

	// Conservative estimate: divide by 2048 bytes per class
	estimatedClasses := int(size / 2048)
	if estimatedClasses == 0 {
		estimatedClasses = 1 // Assume at least 1 class for non-empty files
	}

	return estimatedClasses, nil
}

// estimateModuleClasses provides a simple estimate (not exported in minimal build)
func estimateModuleClasses(modulesFile string) (int, error) {
	// Simple size-based estimation for minimal build
	fileInfo, err := os.Stat(modulesFile)
	if err != nil {
		return 0, err
	}

	// Rough estimate: 10 classes per KB of modules file
	size := fileInfo.Size()
	if size < 1024 {
		return 100, nil // Minimum estimate
	}

	return int(size / 100), nil // 10 classes per 100 bytes
}

// jarContents is a minimal placeholder for tests (always returns 0)
func jarContents(interface{}) int {
	return 0 // Minimal implementation doesn't process ZIP contents
}
