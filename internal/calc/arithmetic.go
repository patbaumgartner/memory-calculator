package calc

import (
	"fmt"
	"math"
)

func checkedAdd(label string, values ...int64) (int64, error) {
	var total int64
	for _, value := range values {
		if value < 0 {
			return 0, fmt.Errorf("%s contains a negative value: %d", label, value)
		}
		if total > math.MaxInt64-value {
			return 0, fmt.Errorf("%s overflows the maximum supported size", label)
		}
		total += value
	}
	return total, nil
}

func checkedMultiply(label string, left, right int64) (int64, error) {
	if left < 0 || right < 0 {
		return 0, fmt.Errorf("%s contains a negative value: %d * %d", label, left, right)
	}
	if left != 0 && right > math.MaxInt64/left {
		return 0, fmt.Errorf("%s overflows the maximum supported size", label)
	}
	return left * right, nil
}
