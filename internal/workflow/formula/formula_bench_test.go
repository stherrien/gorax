package formula

import (
	"testing"
	"time"
)

// BenchmarkSimpleExpression benchmarks evaluation of simple expressions
func BenchmarkSimpleExpression(b *testing.B) {
	evaluator := NewEvaluator()
	context := map[string]interface{}{
		"value": 42,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = evaluator.Evaluate("value * 2", context)
	}
}

// BenchmarkComplexExpression benchmarks evaluation of complex expressions
func BenchmarkComplexExpression(b *testing.B) {
	evaluator := NewEvaluator()
	context := map[string]interface{}{
		"value":  42,
		"factor": 2.5,
		"offset": 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = evaluator.Evaluate("(value * factor + offset) / 2", context)
	}
}

// BenchmarkStringOperations benchmarks string manipulation functions
func BenchmarkStringOperations(b *testing.B) {
	evaluator := NewEvaluator()
	context := map[string]interface{}{
		"text": "Hello World",
	}

	b.Run("upper", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("upper(text)", context)
		}
	})

	b.Run("lower", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("lower(text)", context)
		}
	})

	b.Run("trim", func(b *testing.B) {
		ctx := map[string]interface{}{"text": "  hello  "}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("trim(text)", ctx)
		}
	})

	b.Run("concat", func(b *testing.B) {
		ctx := map[string]interface{}{
			"a": "hello",
			"b": "world",
		}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate(`concat(a, " ", b)`, ctx)
		}
	})

	b.Run("substr", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("substr(text, 0, 5)", context)
		}
	})
}

// BenchmarkMathOperations benchmarks mathematical functions
func BenchmarkMathOperations(b *testing.B) {
	evaluator := NewEvaluator()

	b.Run("round", func(b *testing.B) {
		context := map[string]interface{}{"value": 4.7}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("round(value)", context)
		}
	})

	b.Run("ceil", func(b *testing.B) {
		context := map[string]interface{}{"value": 4.2}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("ceil(value)", context)
		}
	})

	b.Run("floor", func(b *testing.B) {
		context := map[string]interface{}{"value": 4.9}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("floor(value)", context)
		}
	})

	b.Run("abs", func(b *testing.B) {
		context := map[string]interface{}{"value": -42.5}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("abs(value)", context)
		}
	})

	b.Run("min", func(b *testing.B) {
		context := map[string]interface{}{
			"a": 5,
			"b": 3,
			"c": 7,
		}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("min(a, b, c)", context)
		}
	})

	b.Run("max", func(b *testing.B) {
		context := map[string]interface{}{
			"a": 5,
			"b": 3,
			"c": 7,
		}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("max(a, b, c)", context)
		}
	})
}

// BenchmarkDateOperations benchmarks date/time functions
func BenchmarkDateOperations(b *testing.B) {
	evaluator := NewEvaluator()

	b.Run("now", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("now()", map[string]interface{}{})
		}
	})

	b.Run("dateFormat", func(b *testing.B) {
		context := map[string]interface{}{"date": time.Now()}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate(`dateFormat(date, "2006-01-02")`, context)
		}
	})

	b.Run("dateParse", func(b *testing.B) {
		context := map[string]interface{}{"dateStr": "2025-12-20"}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate(`dateParse(dateStr, "2006-01-02")`, context)
		}
	})

	b.Run("addDays", func(b *testing.B) {
		context := map[string]interface{}{"date": time.Now()}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("addDays(date, 7)", context)
		}
	})
}

// BenchmarkArrayOperations benchmarks array operations
func BenchmarkArrayOperations(b *testing.B) {
	evaluator := NewEvaluator()

	b.Run("len_array", func(b *testing.B) {
		context := map[string]interface{}{
			"items": []interface{}{1, 2, 3, 4, 5},
		}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("len(items)", context)
		}
	})

	b.Run("len_string", func(b *testing.B) {
		context := map[string]interface{}{
			"text": "hello world",
		}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("len(text)", context)
		}
	})
}

// BenchmarkConditionalExpressions benchmarks conditional expressions
func BenchmarkConditionalExpressions(b *testing.B) {
	evaluator := NewEvaluator()

	b.Run("simple_condition", func(b *testing.B) {
		context := map[string]interface{}{"value": 42}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("value > 50 ? 'high' : 'low'", context)
		}
	})

	b.Run("complex_condition", func(b *testing.B) {
		context := map[string]interface{}{
			"value":  42,
			"status": "active",
		}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate(
				"value > 50 && status == 'active' ? 'proceed' : 'wait'",
				context,
			)
		}
	})

	b.Run("nested_condition", func(b *testing.B) {
		context := map[string]interface{}{"value": 42}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate(
				"value > 100 ? 'high' : (value > 50 ? 'medium' : 'low')",
				context,
			)
		}
	})
}

// BenchmarkComplexWorkflowContext benchmarks realistic workflow context
func BenchmarkComplexWorkflowContext(b *testing.B) {
	evaluator := NewEvaluator()
	context := map[string]interface{}{
		"trigger": map[string]interface{}{
			"type": "webhook",
			"body": map[string]interface{}{
				"user_id": 12345,
				"action":  "purchase",
				"amount":  99.99,
				"items": []interface{}{
					map[string]interface{}{"id": 1, "name": "Item 1", "price": 49.99},
					map[string]interface{}{"id": 2, "name": "Item 2", "price": 50.00},
				},
			},
		},
		"step1": map[string]interface{}{
			"result":    "success",
			"processed": true,
		},
		"step2": map[string]interface{}{
			"validation": "passed",
			"score":      95,
		},
	}

	b.Run("access_nested_data", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("trigger.body.amount", context)
		}
	})

	b.Run("calculate_from_nested", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("trigger.body.amount * 1.1", context)
		}
	})

	b.Run("complex_calculation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate(
				"round((trigger.body.amount * 1.1) + (step2.score / 10))",
				context,
			)
		}
	})
}

// BenchmarkExpressionCompilation benchmarks expression compilation only
func BenchmarkExpressionCompilation(b *testing.B) {
	evaluator := NewEvaluator()

	b.Run("simple", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = evaluator.ValidateExpression("value * 2")
		}
	})

	b.Run("complex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = evaluator.ValidateExpression(
				"(value * factor + offset) / 2 > threshold ? 'high' : 'low'",
			)
		}
	})

	b.Run("with_functions", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = evaluator.ValidateExpression(
				"upper(concat(firstName, ' ', lastName))",
			)
		}
	})
}

// BenchmarkMultipleEvaluations benchmarks sequential evaluations
func BenchmarkMultipleEvaluations(b *testing.B) {
	evaluator := NewEvaluator()
	context := map[string]interface{}{
		"value": 42,
	}

	expressions := []string{
		"value * 2",
		"value + 10",
		"value / 2",
		"value - 5",
		"round(value * 1.5)",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, expr := range expressions {
			_, _ = evaluator.Evaluate(expr, context)
		}
	}
}

// BenchmarkContextSizes benchmarks with different context sizes
func BenchmarkContextSizes(b *testing.B) {
	evaluator := NewEvaluator()

	b.Run("small_context", func(b *testing.B) {
		context := map[string]interface{}{
			"value": 42,
		}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("value * 2", context)
		}
	})

	b.Run("medium_context", func(b *testing.B) {
		context := map[string]interface{}{
			"value1": 42,
			"value2": 10,
			"value3": 5.5,
			"text":   "hello",
			"flag":   true,
		}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("value1 * value2 + value3", context)
		}
	})

	b.Run("large_context", func(b *testing.B) {
		context := map[string]interface{}{
			"value1": 42,
			"value2": 10,
			"value3": 5.5,
			"text1":  "hello",
			"text2":  "world",
			"flag1":  true,
			"flag2":  false,
			"nested": map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			"array": []interface{}{1, 2, 3, 4, 5},
		}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("value1 * value2 + value3", context)
		}
	})
}

// BenchmarkErrorHandling benchmarks error scenarios
func BenchmarkErrorHandling(b *testing.B) {
	evaluator := NewEvaluator()

	b.Run("invalid_expression", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = evaluator.ValidateExpression("invalid syntax here")
		}
	})

	b.Run("missing_variable", func(b *testing.B) {
		context := map[string]interface{}{}
		for i := 0; i < b.N; i++ {
			_, _ = evaluator.Evaluate("nonexistent_var * 2", context)
		}
	})
}

// BenchmarkMemoryAllocation benchmarks memory usage
func BenchmarkMemoryAllocation(b *testing.B) {
	evaluator := NewEvaluator()
	context := map[string]interface{}{
		"value": 42,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = evaluator.Evaluate("value * 2 + 10", context)
	}
}

// BenchmarkConcurrentEvaluations benchmarks concurrent evaluations
func BenchmarkConcurrentEvaluations(b *testing.B) {
	evaluator := NewEvaluator()

	b.RunParallel(func(pb *testing.PB) {
		context := map[string]interface{}{
			"value": 42,
		}
		for pb.Next() {
			_, _ = evaluator.Evaluate("value * 2", context)
		}
	})
}

// BenchmarkFunctionLookup benchmarks function retrieval
func BenchmarkFunctionLookup(b *testing.B) {
	evaluator := NewEvaluator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = evaluator.GetAvailableFunctions()
	}
}

// BenchmarkBuildEnvironment benchmarks environment creation
func BenchmarkBuildEnvironment(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = buildEnvironment()
	}
}
