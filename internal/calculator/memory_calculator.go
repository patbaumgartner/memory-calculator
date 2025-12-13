/*
 * Copyright 2018-2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Original file copied from https://github.com/paketo-buildpacks/libjvm/blob/main/helper/memory_calculator.go

// Package calculator calculates JVM memory settings based on total memory and other constraints.
package calculator

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/patbaumgartner/memory-calculator/internal/calc"
	"github.com/patbaumgartner/memory-calculator/internal/count"
	"github.com/patbaumgartner/memory-calculator/internal/logger"
	"github.com/patbaumgartner/memory-calculator/internal/parser"
)

const (
	// ClassLoadFactor is the percentage of classes loaded (35%).
	ClassLoadFactor = 0.35
	// DefaultHeadroom is the default percentage of memory to leave for the OS.
	DefaultHeadroom = 0
	// DefaultMemoryLimitPathV1 is the path to the cgroup v1 memory limit file.
	DefaultMemoryLimitPathV1 = "/sys/fs/cgroup/memory/memory.limit_in_bytes"
	// DefaultMemoryLimitPathV2 is the path to the cgroup v2 memory limit file.
	DefaultMemoryLimitPathV2 = "/sys/fs/cgroup/memory.max"
	// DefaultMemoryInfoPath is the path to /proc/meminfo.
	DefaultMemoryInfoPath = "/proc/meminfo"
	// DefaultThreadCount is the default thread count (250).
	DefaultThreadCount = 250
	// MaxJVMSize is the maximum size of the JVM.
	MaxJVMSize = 64 * calc.Tebi
	// UnsetTotalMemory is the default value for unset total memory.
	UnsetTotalMemory = int64(9_223_372_036_854_771_712)
)

// MemoryCalculator calculates JVM memory configuration.
type MemoryCalculator struct {
	Logger            *logger.Logger
	MemoryLimitPathV1 string
	MemoryLimitPathV2 string
	MemoryInfoPath    string
}

// Create creates a new MemoryCalculator.
func Create(quiet bool) *MemoryCalculator {
	return &MemoryCalculator{
		Logger:            logger.Create(quiet),
		MemoryLimitPathV1: DefaultMemoryLimitPathV1,
		MemoryLimitPathV2: DefaultMemoryLimitPathV2,
		MemoryInfoPath:    DefaultMemoryInfoPath,
	}
}

// Execute performs the memory calculation and returns environment variables.
func (m MemoryCalculator) Execute() (map[string]string, error) {
	c := calc.Calculator{
		HeadRoom:    DefaultHeadroom,
		ThreadCount: DefaultThreadCount,
	}

	// Parse configuration from environment variables
	if err := m.parseHeadroomConfig(&c); err != nil {
		return nil, err
	}

	if err := m.parseThreadCountConfig(&c); err != nil {
		return nil, err
	}

	var values []string
	opts, ok := os.LookupEnv("JAVA_TOOL_OPTIONS")
	if ok {
		values = append(values, opts)
	}

	// Parse class count configuration
	if err := m.parseClassCountConfig(&c, opts); err != nil {
		return nil, err
	}

	// Determine total memory
	totalMemory, err := m.determineTotalMemory()
	if err != nil {
		return nil, err
	}

	c.TotalMemory = totalMemory

	r, err := c.Calculate(opts)
	if err != nil {
		return nil, fmt.Errorf("unable to calculate memory configuration\n%w", err)
	}

	// Build calculated values
	calculated := m.buildCalculatedValues(r)
	values = append(values, calculated...)

	m.Logger.Infof(
		"Calculated JVM Memory Configuration: %s (Total Memory: %s, Thread Count: %d, "+
			"Loaded Class Count: %d, Headroom: %d%%)",
		strings.Join(calculated, " "), c.TotalMemory, c.ThreadCount, c.LoadedClassCount, c.HeadRoom)

	return map[string]string{"JAVA_TOOL_OPTIONS": strings.Join(values, " ")}, nil
}

func (m MemoryCalculator) getMemoryLimitFromPath(memoryLimitPath string) int64 {
	// #nosec G304 - memoryLimitPath comes from predefined cgroup paths, not user input
	if b, readErr := os.ReadFile(memoryLimitPath); readErr != nil && !os.IsNotExist(readErr) {
		m.Logger.Infof("WARNING: Unable to read %s: %s", memoryLimitPath, readErr)
	} else if readErr == nil {
		limit := strings.TrimSpace(string(b))
		if size, parseErr := calc.ParseSize(limit); parseErr != nil {
			if limit == "max" {
				return UnsetTotalMemory
			}
			m.Logger.Infof(
				"WARNING: Unable to convert memory limit %q from path %q as int: %s",
				limit, memoryLimitPath, parseErr)
		} else {
			return size.Value
		}
	}
	return UnsetTotalMemory
}

func (m MemoryCalculator) parseMemoryString(memStr string) (int64, error) {
	size, err := calc.ParseSize(memStr)
	if err != nil {
		// Try parsing with common suffixes
		memStr = strings.TrimSpace(strings.ToUpper(memStr))

		// Handle decimal values like 2G, 512M, etc.
		if strings.HasSuffix(memStr, "G") || strings.HasSuffix(memStr, "GB") {
			numStr := strings.TrimSuffix(strings.TrimSuffix(memStr, "GB"), "G")
			if num, parseErr := strconv.ParseFloat(numStr, 64); parseErr == nil {
				return int64(num * float64(calc.Gibi)), nil
			}
		} else if strings.HasSuffix(memStr, "M") || strings.HasSuffix(memStr, "MB") {
			numStr := strings.TrimSuffix(strings.TrimSuffix(memStr, "MB"), "M")
			if num, parseErr := strconv.ParseFloat(numStr, 64); parseErr == nil {
				return int64(num * float64(calc.Mebi)), nil
			}
		} else if strings.HasSuffix(memStr, "K") || strings.HasSuffix(memStr, "KB") {
			numStr := strings.TrimSuffix(strings.TrimSuffix(memStr, "KB"), "K")
			if num, parseErr := strconv.ParseFloat(numStr, 64); parseErr == nil {
				return int64(num * float64(calc.Kibi)), nil
			}
		} else {
			// Try as plain bytes
			if num, parseErr := strconv.ParseInt(memStr, 10, 64); parseErr == nil {
				return num, nil
			}
		}
		return 0, err
	}
	return size.Value, nil
}

func parseMemInfo(s string) (int64, error) {
	// Look for MemAvailable line without regex
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "MemAvailable:") {
			// Remove "MemAvailable:" prefix and parse
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}

			num, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("unable to convert available memory %s to integer\n%w", parts[1], err)
			}

			// Handle unit (usually kB)
			var unit int64 = 1
			if len(parts) > 2 {
				switch strings.ToLower(parts[2]) {
				case "kb":
					unit = calc.Kibi
				case "mb":
					unit = calc.Mebi
				case "gb":
					unit = calc.Gibi
				}
			}
			return num * unit, nil
		}
	}
	return 0, fmt.Errorf("failed to find MemAvailable in meminfo")
}

// CountAgentClasses counts classes in agent JARs.
func (m MemoryCalculator) CountAgentClasses(opts string) (int, error) {
	var agentClassCount, skippedAgents int
	p, err := parser.ParseFlags(opts)
	if err != nil {
		return 0, fmt.Errorf("unable to parse $JAVA_TOOL_OPTIONS\n%w", err)
	}

	var agentPaths []string
	for _, s := range p {
		if strings.HasPrefix(s, "-javaagent:") {
			agentPaths = append(agentPaths, strings.Split(s, ":")[1])
		}
	}
	if len(agentPaths) > 0 {
		agentClassCount, skippedAgents, err = count.JarClassesFrom(agentPaths...)
		if err != nil {
			return 0, fmt.Errorf("error counting agent jar classes \n%w", err)
		} else if skippedAgents > 0 {
			m.Logger.Infof(
				`WARNING: could not count classes from all agent jars (skipped %d), `+
					`class count and metaspace may not be sized correctly`, skippedAgents)
		}
	}
	return agentClassCount, nil
}

// parseHeadroomConfig parses headroom configuration from environment variables
func (m MemoryCalculator) parseHeadroomConfig(c *calc.Calculator) error {
	var deprecatedHeadroom bool

	if s, ok := os.LookupEnv("BPL_JVM_HEADROOM"); ok {
		headroom, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("unable to convert $BPL_JVM_HEADROOM=%s to integer\n%w", s, err)
		}
		c.HeadRoom = headroom
		deprecatedHeadroom = true
		m.Logger.Info("WARNING: BPL_JVM_HEADROOM is deprecated and will be removed, please switch to BPL_JVM_HEAD_ROOM")
	}

	if s, ok := os.LookupEnv("BPL_JVM_HEAD_ROOM"); ok {
		headroom, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("unable to convert $BPL_JVM_HEAD_ROOM=%s to integer\n%w", s, err)
		}
		c.HeadRoom = headroom
		if deprecatedHeadroom {
			m.Logger.Info(
				"WARNING: You have set both BPL_JVM_HEAD_ROOM and BPL_JVM_HEADROOM. " +
					"BPL_JVM_HEADROOM has been deprecated, so it will be ignored.")
		}
	}

	return nil
}

// parseThreadCountConfig parses thread count configuration from environment variables
func (m MemoryCalculator) parseThreadCountConfig(c *calc.Calculator) error {
	if threadCount, ok := os.LookupEnv("BPL_JVM_THREAD_COUNT"); ok {
		count, err := strconv.Atoi(threadCount)
		if err != nil {
			return fmt.Errorf("unable to convert $BPL_JVM_THREAD_COUNT=%s to integer\n%w", threadCount, err)
		}
		c.ThreadCount = count
	}
	return nil
}

// parseClassCountConfig parses class count configuration from environment variables
func (m MemoryCalculator) parseClassCountConfig(c *calc.Calculator, opts string) error {
	if s, ok := os.LookupEnv("BPL_JVM_LOADED_CLASS_COUNT"); ok {
		count, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("unable to convert $BPL_JVM_LOADED_CLASS_COUNT=%s to integer\n%w", s, err)
		}
		c.LoadedClassCount = count
		return nil
	}

	// Calculate class count dynamically
	appPath := "/app" // Default application path
	if path, ok := os.LookupEnv("BPI_APPLICATION_PATH"); ok {
		appPath = path
	}

	jvmClassCount := 1000 // Default JVM class count
	if jvmCountStr, ok := os.LookupEnv("BPI_JVM_CLASS_COUNT"); ok {
		count, err := strconv.Atoi(jvmCountStr)
		if err != nil {
			return fmt.Errorf("unable to convert $BPI_JVM_CLASS_COUNT=%s to integer\n%w", jvmCountStr, err)
		}
		jvmClassCount = count
	}

	adjustmentFactor := 100
	if adjustmentStr, ok := os.LookupEnv("BPI_CLASS_ADJUSTMENT_FACTOR"); ok {
		factor, err := strconv.Atoi(adjustmentStr)
		if err != nil {
			return fmt.Errorf("unable to convert $BPI_CLASS_ADJUSTMENT_FACTOR=%s to integer\n%w", adjustmentStr, err)
		}
		adjustmentFactor = factor
	}

	staticAdjustment := 0
	if staticStr, ok := os.LookupEnv("BPI_CLASS_STATIC_ADJUSTMENT"); ok {
		adjustment, err := strconv.Atoi(staticStr)
		if err != nil {
			return fmt.Errorf("unable to convert $BPI_CLASS_STATIC_ADJUSTMENT=%s to integer\n%w", staticStr, err)
		}
		staticAdjustment = adjustment
	}

	agentClassCount, err := m.CountAgentClasses(opts)
	if err != nil {
		return fmt.Errorf("unable to determine agent class count\n%w", err)
	}

	appClassCount, err := count.Classes(appPath)
	if err != nil {
		return fmt.Errorf("unable to determine class count\n%w", err)
	}

	totalClasses := float64(jvmClassCount+appClassCount+agentClassCount+staticAdjustment) *
		(float64(adjustmentFactor) / 100.0)

	m.Logger.Debugf(
		"Memory Calculation: (%d%% * (%d + %d + %d + %d)) * %0.2f",
		adjustmentFactor, jvmClassCount, appClassCount, agentClassCount, staticAdjustment, ClassLoadFactor)

	c.LoadedClassCount = int(totalClasses * ClassLoadFactor)
	return nil
}

// determineTotalMemory determines the total memory available to the JVM
func (m MemoryCalculator) determineTotalMemory() (calc.Size, error) {
	totalMemory := UnsetTotalMemory

	// Check if total memory was specified via command line argument
	if totalMemStr, ok := os.LookupEnv("BPL_JVM_TOTAL_MEMORY"); ok {
		if size, parseErr := m.parseMemoryString(totalMemStr); parseErr != nil {
			m.Logger.Infof("WARNING: Unable to parse total memory %q: %s, falling back to detection", totalMemStr, parseErr)
		} else {
			totalMemory = size
			m.Logger.Infof("Using specified memory: %s", calc.Size{Value: totalMemory}.String())
		}
	}

	// Fall back to cgroup detection if not specified
	if totalMemory == UnsetTotalMemory {
		totalMemory = m.getMemoryLimitFromPath(m.MemoryLimitPathV1)
		if totalMemory == UnsetTotalMemory {
			totalMemory = m.getMemoryLimitFromPath(m.MemoryLimitPathV2)
		}
	}

	// Finally try /proc/meminfo
	if totalMemory == UnsetTotalMemory {
		if b, readErr := os.ReadFile(m.MemoryInfoPath); readErr != nil && !os.IsNotExist(readErr) {
			m.Logger.Infof(`WARNING: failed to read %q: %s`, m.MemoryInfoPath, readErr)
		} else if readErr == nil {
			if mem, parseErr := parseMemInfo(string(b)); parseErr != nil {
				m.Logger.Infof(`WARNING: failed to parse available memory from path %q: %s`, m.MemoryInfoPath, parseErr)
			} else {
				m.Logger.Infof("Calculating JVM memory based on %s available memory", calc.Size{Value: mem}.String())
				m.Logger.Info(
					"For more information on this calculation, see " +
						"https://paketo.io/docs/reference/java-reference/#memory-calculator")
				totalMemory = mem
			}
		}
	}

	// Set default or validate bounds
	if totalMemory == UnsetTotalMemory {
		m.Logger.Info("WARNING: Unable to determine memory limit. Configuring JVM for 1G container.")
		return calc.Size{Value: calc.Gibi}, nil
	} else if totalMemory > MaxJVMSize {
		m.Logger.Info("WARNING: Container memory limit too large. Configuring JVM for 64T container.")
		return calc.Size{Value: MaxJVMSize}, nil
	}

	return calc.Size{Value: totalMemory}, nil
}

// buildCalculatedValues builds the list of calculated JVM memory options
func (m MemoryCalculator) buildCalculatedValues(r calc.MemoryRegions) []string {
	var calculated []string
	if r.DirectMemory.Provenance != calc.UserConfigured {
		calculated = append(calculated, r.DirectMemory.String())
	}
	if r.Heap.Provenance != calc.UserConfigured {
		calculated = append(calculated, r.Heap.String())
	}
	if r.Metaspace.Provenance != calc.UserConfigured {
		calculated = append(calculated, r.Metaspace.String())
	}
	if r.ReservedCodeCache.Provenance != calc.UserConfigured {
		calculated = append(calculated, r.ReservedCodeCache.String())
	}
	if r.Stack.Provenance != calc.UserConfigured {
		calculated = append(calculated, r.Stack.String())
	}
	return calculated
}
