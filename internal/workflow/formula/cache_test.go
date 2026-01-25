package formula

import (
	"sync"
	"testing"
)

// TestExpressionCacheBasic tests basic cache operations
func TestExpressionCacheBasic(t *testing.T) {
	cache := NewExpressionCache(10)

	t.Run("cache miss returns not found", func(t *testing.T) {
		_, found := cache.Get("test * 2")
		if found {
			t.Error("Expected cache miss, got hit")
		}
	})

	t.Run("cache hit returns stored value", func(t *testing.T) {
		expr := "value * 2"
		expected := &CachedExpression{
			Program: nil, // Would be a real compiled program
			Hash:    "test-hash",
		}

		cache.Put(expr, expected)
		result, found := cache.Get(expr)

		if !found {
			t.Error("Expected cache hit, got miss")
			return
		}
		if result == nil {
			t.Error("Expected cached expression, got nil")
			return
		}
		if result.Hash != expected.Hash {
			t.Errorf("Hash mismatch: got %s, want %s", result.Hash, expected.Hash)
		}
	})

	t.Run("cache overwrite updates value", func(t *testing.T) {
		expr := "value * 3"
		first := &CachedExpression{Hash: "first"}
		second := &CachedExpression{Hash: "second"}

		cache.Put(expr, first)
		cache.Put(expr, second)

		result, found := cache.Get(expr)
		if !found {
			t.Error("Expected cache hit")
		}
		if result.Hash != "second" {
			t.Errorf("Expected second value, got %s", result.Hash)
		}
	})
}

// TestExpressionCacheLRU tests LRU eviction behavior
func TestExpressionCacheLRU(t *testing.T) {
	cache := NewExpressionCache(3) // Small cache for testing eviction

	t.Run("cache evicts least recently used", func(t *testing.T) {
		// Fill cache
		cache.Put("expr1", &CachedExpression{Hash: "1"})
		cache.Put("expr2", &CachedExpression{Hash: "2"})
		cache.Put("expr3", &CachedExpression{Hash: "3"})

		// Add one more, should evict expr1
		cache.Put("expr4", &CachedExpression{Hash: "4"})

		// expr1 should be evicted
		_, found := cache.Get("expr1")
		if found {
			t.Error("Expected expr1 to be evicted")
		}

		// Others should still be present
		if _, found := cache.Get("expr2"); !found {
			t.Error("Expected expr2 to be present")
		}
		if _, found := cache.Get("expr3"); !found {
			t.Error("Expected expr3 to be present")
		}
		if _, found := cache.Get("expr4"); !found {
			t.Error("Expected expr4 to be present")
		}
	})

	t.Run("accessing entry updates LRU order", func(t *testing.T) {
		cache := NewExpressionCache(2)
		cache.Put("a", &CachedExpression{Hash: "a"})
		cache.Put("b", &CachedExpression{Hash: "b"})

		// Access 'a' to make it most recently used
		cache.Get("a")

		// Add 'c', should evict 'b' (least recently used)
		cache.Put("c", &CachedExpression{Hash: "c"})

		if _, found := cache.Get("a"); !found {
			t.Error("Expected 'a' to still be present")
		}
		if _, found := cache.Get("b"); found {
			t.Error("Expected 'b' to be evicted")
		}
		if _, found := cache.Get("c"); !found {
			t.Error("Expected 'c' to be present")
		}
	})
}

// TestExpressionCacheConcurrency tests thread safety
func TestExpressionCacheConcurrency(t *testing.T) {
	cache := NewExpressionCache(100)
	iterations := 1000

	t.Run("concurrent writes are safe", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					expr := "test" + string(rune(id))
					cache.Put(expr, &CachedExpression{Hash: expr})
				}
			}(i)
		}

		wg.Wait()
		// If we get here without deadlock/race, test passes
	})

	t.Run("concurrent reads and writes are safe", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(20)

		// Writers
		for i := 0; i < 10; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					expr := "test" + string(rune(id))
					cache.Put(expr, &CachedExpression{Hash: expr})
				}
			}(i)
		}

		// Readers
		for i := 0; i < 10; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					expr := "test" + string(rune(id))
					cache.Get(expr)
				}
			}(i)
		}

		wg.Wait()
	})
}

// TestExpressionCacheSize tests cache size limits
func TestExpressionCacheSize(t *testing.T) {
	t.Run("cache respects size limit", func(t *testing.T) {
		maxSize := 5
		cache := NewExpressionCache(maxSize)

		// Add more than max size
		for i := 0; i < maxSize*2; i++ {
			cache.Put("expr"+string(rune(i)), &CachedExpression{Hash: string(rune(i))})
		}

		// Count items we can retrieve
		foundCount := 0
		for i := 0; i < maxSize*2; i++ {
			if _, found := cache.Get("expr" + string(rune(i))); found {
				foundCount++
			}
		}

		if foundCount > maxSize {
			t.Errorf("Cache exceeded max size: found %d items, max %d", foundCount, maxSize)
		}
	})
}

// TestExpressionCacheEmpty tests empty cache operations
func TestExpressionCacheEmpty(t *testing.T) {
	cache := NewExpressionCache(10)

	t.Run("get from empty cache returns false", func(t *testing.T) {
		_, found := cache.Get("nonexistent")
		if found {
			t.Error("Expected false for empty cache")
		}
	})

	t.Run("put then get works", func(t *testing.T) {
		cache.Put("test", &CachedExpression{Hash: "test"})
		_, found := cache.Get("test")
		if !found {
			t.Error("Expected to find just-added item")
		}
	})
}

// TestExpressionCacheStats tests cache statistics
func TestExpressionCacheStats(t *testing.T) {
	cache := NewExpressionCache(10)

	t.Run("stats track hits and misses", func(t *testing.T) {
		// Initial stats
		stats := cache.Stats()
		if stats.Hits != 0 || stats.Misses != 0 {
			t.Error("Expected zero stats initially")
		}

		// Add an item
		cache.Put("test", &CachedExpression{Hash: "test"})

		// Hit
		cache.Get("test")
		stats = cache.Stats()
		if stats.Hits != 1 {
			t.Errorf("Expected 1 hit, got %d", stats.Hits)
		}

		// Miss
		cache.Get("nonexistent")
		stats = cache.Stats()
		if stats.Misses != 1 {
			t.Errorf("Expected 1 miss, got %d", stats.Misses)
		}
	})

	t.Run("hit rate calculation", func(t *testing.T) {
		cache := NewExpressionCache(10)
		cache.Put("test", &CachedExpression{Hash: "test"})

		// 3 hits, 1 miss -> 75% hit rate
		cache.Get("test")
		cache.Get("test")
		cache.Get("test")
		cache.Get("nonexistent")

		stats := cache.Stats()
		expectedRate := 0.75
		if stats.HitRate() != expectedRate {
			t.Errorf("Expected hit rate %.2f, got %.2f", expectedRate, stats.HitRate())
		}
	})
}

// TestCachedEvaluatorBasic tests the cached evaluator
func TestCachedEvaluatorBasic(t *testing.T) {
	evaluator := NewCachedEvaluator(100)

	t.Run("first evaluation caches result", func(t *testing.T) {
		expr := "value * 2"
		context := map[string]interface{}{"value": 10}

		result1, err := evaluator.Evaluate(expr, context)
		if err != nil {
			t.Fatalf("First evaluation failed: %v", err)
		}
		if result1 != 20 {
			t.Errorf("Expected 20, got %v", result1)
		}

		// Second evaluation should use cache
		result2, err := evaluator.Evaluate(expr, context)
		if err != nil {
			t.Fatalf("Second evaluation failed: %v", err)
		}
		if result2 != 20 {
			t.Errorf("Expected 20, got %v", result2)
		}

		// Check cache was used
		stats := evaluator.CacheStats()
		if stats.Hits == 0 {
			t.Error("Expected cache hit on second evaluation")
		}
	})

	t.Run("different expressions are cached separately", func(t *testing.T) {
		evaluator := NewCachedEvaluator(100)
		context := map[string]interface{}{"value": 10}

		result1, _ := evaluator.Evaluate("value * 2", context)
		result2, _ := evaluator.Evaluate("value * 3", context)

		if result1 == result2 {
			t.Error("Different expressions should produce different results")
		}

		stats := evaluator.CacheStats()
		if stats.Misses != 2 {
			t.Errorf("Expected 2 cache misses for different expressions, got %d", stats.Misses)
		}
	})

	t.Run("same expression with different context uses cache", func(t *testing.T) {
		evaluator := NewCachedEvaluator(100)
		expr := "value * 2"

		result1, _ := evaluator.Evaluate(expr, map[string]interface{}{"value": 10})
		result2, _ := evaluator.Evaluate(expr, map[string]interface{}{"value": 20})

		if result1 != 20 || result2 != 40 {
			t.Errorf("Expected 20 and 40, got %v and %v", result1, result2)
		}

		// Should have 1 hit (same compiled expression reused)
		stats := evaluator.CacheStats()
		if stats.Hits == 0 {
			t.Error("Expected cache hit for same expression with different context")
		}
	})
}

// TestCachedEvaluatorConcurrency tests thread safety of cached evaluator
func TestCachedEvaluatorConcurrency(t *testing.T) {
	evaluator := NewCachedEvaluator(100)
	iterations := 100

	t.Run("concurrent evaluations are safe", func(t *testing.T) {
		var wg sync.WaitGroup
		expressions := []string{
			"value * 2",
			"value + 10",
			"value / 2",
			"round(value * 1.5)",
		}

		for _, expr := range expressions {
			wg.Add(1)
			go func(e string) {
				defer wg.Done()
				for i := 0; i < iterations; i++ {
					context := map[string]interface{}{"value": i}
					_, _ = evaluator.Evaluate(e, context)
				}
			}(expr)
		}

		wg.Wait()

		// Check that cache was effective
		stats := evaluator.CacheStats()
		if stats.Hits == 0 {
			t.Error("Expected some cache hits with concurrent access")
		}
	})
}
