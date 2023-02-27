package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func Float64PtrToFloat64(n *float64) float64 {
	if n == nil {
		return 0
	}
	return *n
}

func Float64ToFloat64Ptr(n float64) *float64 {
	return &n
}

// https://go.dev/play/p/ocsLE_AeICK
// TODO: add tests to prove this
func TrimFloatToRight(v float64, numToTrim int) float64 {
	str := fmt.Sprintf("%f", v)
	if len(str) < numToTrim+1 {
		return 0
	}

	str = strings.Trim(str, "0")
	newStr := str[:len(str)-numToTrim]
	newVal, _ := strconv.ParseFloat(newStr, 64)

	return newVal
}

func ConvertPercentageToDecimal(v float64) float64 {
	return v / 100
}
