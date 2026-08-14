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
	"strings"

	"github.com/patbaumgartner/memory-calculator/internal/calc"
	"github.com/patbaumgartner/memory-calculator/internal/cgroups"
	"github.com/patbaumgartner/memory-calculator/internal/count"
	"github.com/patbaumgartner/memory-calculator/internal/logger"
	"github.com/patbaumgartner/memory-calculator/internal/parser"
)

const (
	// ClassLoadFactor is the percentage of classes loaded (35%).
	ClassLoadFactor = 0.35
	// DefaultHeadroom is the default percentage of memory to leave for the OS.
	DefaultHeadroom = 0
	// DefaultThreadCount is the default thread count (250).
	DefaultThreadCount = 250
	// DefaultTotalMemory is used when no memory limit can be determined.
	DefaultTotalMemory = calc.Gibi
	// MaxJVMSize is the maximum size of the JVM.
	MaxJVMSize = 64 * calc.Tebi
)

// MemoryCalculator calculates JVM memory configuration.
type MemoryCalculator struct {
	Logger   *logger.Logger
	Detector *cgroups.Detector
}

// Create creates a new MemoryCalculator.
func Create(quiet bool) *MemoryCalculator {
	return &MemoryCalculator{
		Logger:   logger.Create(quiet),
		Detector: cgroups.Create(),
	}
}

// Input is the typed contract for a memory calculation. Environment variables and command-line
// flags are external adapters; the calculator itself has no dependency on process-global state.
type Input struct {
	TotalMemory      *int64
	ThreadCount      int
	LoadedClassCount *int
	HeadRoom         int
	ApplicationPath  string
	JavaToolOptions  string
	JVMClassCount    int
	AdjustmentFactor int
	StaticAdjustment int
}

// Result is the outcome of a memory calculation.
type Result struct {
	// JavaToolOptions is the complete JAVA_TOOL_OPTIONS value, including any options the caller
	// already had set.
	JavaToolOptions string
	// TotalMemory is the memory budget the calculation was based on.
	TotalMemory calc.Size
	// Regions holds the individual memory regions that were calculated.
	Regions calc.MemoryRegions
	// ThreadCount is the thread count used to size stack memory.
	ThreadCount int
	// LoadedClassCount is the class count used to size metaspace.
	LoadedClassCount int
	// HeadRoom is the percentage of total memory that was reserved.
	HeadRoom int
}

// Environment renders the result as environment variables to export.
func (r Result) Environment() map[string]string {
	return map[string]string{"JAVA_TOOL_OPTIONS": r.JavaToolOptions}
}

// Execute performs the memory calculation from typed input.
func (m MemoryCalculator) Execute(input Input) (Result, error) {
	c := calc.Calculator{
		HeadRoom:    input.HeadRoom,
		ThreadCount: input.ThreadCount,
	}

	if input.LoadedClassCount != nil {
		c.LoadedClassCount = *input.LoadedClassCount
	} else if err := m.calculateClassCount(&c, input); err != nil {
		return Result{}, err
	}

	totalMemory := m.determineTotalMemory(input.TotalMemory)
	c.TotalMemory = totalMemory

	regions, err := c.Calculate(input.JavaToolOptions)
	if err != nil {
		return Result{}, fmt.Errorf("unable to calculate memory configuration\n%w", err)
	}

	values := make([]string, 0, 6)
	if input.JavaToolOptions != "" {
		values = append(values, input.JavaToolOptions)
	}
	calculated := m.buildCalculatedValues(regions)
	values = append(values, calculated...)

	m.Logger.Infof(
		"Calculated JVM Memory Configuration: %s (Total Memory: %s, Thread Count: %d, "+
			"Loaded Class Count: %d, Headroom: %d%%)",
		strings.Join(calculated, " "), c.TotalMemory, c.ThreadCount, c.LoadedClassCount, c.HeadRoom)

	return Result{
		JavaToolOptions:  strings.Join(values, " "),
		TotalMemory:      c.TotalMemory,
		Regions:          regions,
		ThreadCount:      c.ThreadCount,
		LoadedClassCount: c.LoadedClassCount,
		HeadRoom:         c.HeadRoom,
	}, nil
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
			agent := strings.TrimPrefix(s, "-javaagent:")
			path, _, _ := strings.Cut(agent, "=")
			agentPaths = append(agentPaths, path)
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

// calculateClassCount determines the effective loaded class count from application and agent JARs.
func (m MemoryCalculator) calculateClassCount(c *calc.Calculator, input Input) error {
	agentClassCount, err := m.CountAgentClasses(input.JavaToolOptions)
	if err != nil {
		return fmt.Errorf("unable to determine agent class count\n%w", err)
	}

	appClassCount, err := count.Classes(input.ApplicationPath)
	if err != nil {
		return fmt.Errorf("unable to determine class count\n%w", err)
	}

	totalClasses := float64(input.JVMClassCount+appClassCount+agentClassCount+input.StaticAdjustment) *
		(float64(input.AdjustmentFactor) / 100.0)

	m.Logger.Debugf(
		"Memory Calculation: (%d%% * (%d + %d + %d + %d)) * %0.2f",
		input.AdjustmentFactor, input.JVMClassCount, appClassCount, agentClassCount,
		input.StaticAdjustment, ClassLoadFactor)

	c.LoadedClassCount = int(totalClasses * ClassLoadFactor)
	return nil
}

// determineTotalMemory uses an explicit memory budget or detects one from cgroups and host memory.
func (m MemoryCalculator) determineTotalMemory(explicit *int64) calc.Size {
	if explicit != nil {
		m.Logger.Infof("Using specified memory: %s", calc.Size{Value: *explicit})
		return m.clamp(*explicit)
	}

	detection := m.Detector.Detect()
	if !detection.Found() {
		m.Logger.Infof("WARNING: Unable to determine memory limit. Configuring JVM for %s container.",
			calc.Size{Value: DefaultTotalMemory})
		return calc.Size{Value: DefaultTotalMemory}
	}

	m.Logger.Infof("Calculating JVM memory based on %s available memory (source: %s)",
		calc.Size{Value: detection.Limit}, detection.Source)
	m.Logger.Info(
		"For more information on this calculation, see " +
			"https://paketo.io/docs/reference/java-reference/#memory-calculator")

	return m.clamp(detection.Limit)
}

// clamp caps the total at the largest heap the JVM can address.
func (m MemoryCalculator) clamp(totalMemory int64) calc.Size {
	if totalMemory > MaxJVMSize {
		m.Logger.Infof("WARNING: Memory limit %s is too large. Configuring JVM for %s.",
			calc.Size{Value: totalMemory}, calc.Size{Value: MaxJVMSize})
		return calc.Size{Value: MaxJVMSize}
	}

	return calc.Size{Value: totalMemory}
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
