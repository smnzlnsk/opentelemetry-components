package formulae

import (
	"regexp"
	"strconv"

	"github.com/smnzlnsk/opentelemetry-components/internal/shared/calculation"
)

// Regex for extracting metrics from formula
const metricRegexString = `\[([^\[\]{}()]+)(?:\(([^()]+)\))?(?:{([^{}]+)})?\]`

var metricRegex = regexp.MustCompile(metricRegexString)

// Extract all necessary metrics from formula as a map to filter in the future
// The regex is designed to capture metrics with optional state and age parameters
// The regex is supposed to match against the following patterns:
// [metric(age){state}]
// [metric(age)]
// [metric{state}]
// [metric]
// This results in 3 groups:
// 1. The metric name
// 2. The age (optional)
// 3. The state (optional)
func FilterMetricsFromFormula(formula string) map[string]bool {
	matches := metricRegex.FindAllStringSubmatch(formula, -1)
	res := make(map[string]bool, len(matches))

	for _, match := range matches {
		if len(match) > 1 {
			// Extract just the metric name (without state or age)
			metric := match[1]
			res[metric] = true
		}
	}
	return res
}

// Sanitize formula by adding default states where missing
// Returns a formula where every variable has the format [metric(age){state}]
func Sanitize(formula string, defaultState string) string {
	return metricRegex.ReplaceAllStringFunc(formula, func(match string) string {
		submatches := metricRegex.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		metric := submatches[1]
		hasAge := len(submatches) > 2 && submatches[2] != ""
		hasState := len(submatches) > 3 && submatches[3] != ""

		ageStr := ""
		if hasAge {
			ageStr = "(" + submatches[2] + ")"
		} else {
			ageStr = "(0)" // Add default age of 0 if no age is specified
		}

		if !hasState {
			return "[" + metric + ageStr + "{" + defaultState + "}" + "]"
		}

		// If it has a state but no age, add the default age
		if hasState && !hasAge {
			return "[" + metric + ageStr + "{" + submatches[3] + "}" + "]"
		}

		return match
	})
}

func GetCalculationArguments(formula string) []calculation.Argument {
	matches := metricRegex.FindAllStringSubmatch(formula, -1)
	res := make([]calculation.Argument, 0, len(matches))

	for _, match := range matches {
		age, _ := strconv.Atoi(match[2])
		res = append(res, calculation.Argument{
			Metric: match[1],
			Age:    age,
			State:  match[3],
		})
	}
	return res
}
