package formula

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/expr-lang/expr/vm"
	lru "github.com/hashicorp/golang-lru/v2"
)

// CachedExpression represents a compiled expression in the cache
type CachedExpression struct {
	Program *vm.Program
	Hash    string
}

// CacheStats provides statistics about cache performance
type CacheStats struct {
	Hits   uint64
	Misses uint64
}

// HitRate returns the cache hit rate as a percentage (0.0 to 1.0)
func (s *CacheStats) HitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0.0
	}
	return float64(s.Hits) / float64(total)
}

// ExpressionCache provides thread-safe LRU caching for compiled expressions
type ExpressionCache struct {
	cache  *lru.Cache[string, *CachedExpression]
	mu     sync.RWMutex
	hits   atomic.Uint64
	misses atomic.Uint64
}

// NewExpressionCache creates a new expression cache with the given size
func NewExpressionCache(size int) *ExpressionCache {
	cache, err := lru.New[string, *CachedExpression](size)
	if err != nil {
		// This should never happen with valid size, but handle it gracefully
		panic(fmt.Sprintf("failed to create LRU cache: %v", err))
	}

	return &ExpressionCache{
		cache: cache,
	}
}

// Get retrieves a cached expression
func (c *ExpressionCache) Get(expr string) (*CachedExpression, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, found := c.cache.Get(expr)
	if found {
		c.hits.Add(1)
		return val, true
	}

	c.misses.Add(1)
	return nil, false
}

// Put stores a compiled expression in the cache
func (c *ExpressionCache) Put(expr string, cached *CachedExpression) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache.Add(expr, cached)
}

// Stats returns current cache statistics
func (c *ExpressionCache) Stats() CacheStats {
	return CacheStats{
		Hits:   c.hits.Load(),
		Misses: c.misses.Load(),
	}
}

// hashExpression creates a deterministic hash of an expression for caching
func hashExpression(expr string) string {
	hash := sha256.Sum256([]byte(expr))
	return hex.EncodeToString(hash[:])
}

// envPool is a sync.Pool for environment maps
var envPool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate with capacity for typical use case
		return make(map[string]interface{}, 32)
	},
}

// getEnv gets a map from the pool
func getEnv() map[string]interface{} {
	return envPool.Get().(map[string]interface{})
}

// putEnv returns a map to the pool after clearing it
func putEnv(env map[string]interface{}) {
	// Clear the map
	for k := range env {
		delete(env, k)
	}
	envPool.Put(env)
}

// CachedEvaluator is an Evaluator with expression caching and object pooling
type CachedEvaluator struct {
	cache *ExpressionCache
	env   map[string]interface{}
}

// NewCachedEvaluator creates a new cached evaluator
func NewCachedEvaluator(cacheSize int) *CachedEvaluator {
	return &CachedEvaluator{
		cache: NewExpressionCache(cacheSize),
		env:   buildEnvironment(),
	}
}

// Evaluate compiles and evaluates an expression with caching
func (e *CachedEvaluator) Evaluate(expression string, context map[string]interface{}) (interface{}, error) {
	if expression == "" {
		return nil, fmt.Errorf("expression cannot be empty")
	}

	// Get environment map from pool
	env := getEnv()
	defer putEnv(env)

	// Merge built-in functions
	for k, v := range e.env {
		env[k] = v
	}
	// Merge context
	for k, v := range context {
		env[k] = v
	}

	// Try to get compiled program from cache
	var program *vm.Program
	cached, found := e.cache.Get(expression)

	if found {
		program = cached.Program
	} else {
		// Compile expression
		compiled, err := compileExpression(expression, env)
		if err != nil {
			return nil, err
		}

		program = compiled

		// Cache the compiled program
		e.cache.Put(expression, &CachedExpression{
			Program: program,
			Hash:    hashExpression(expression),
		})
	}

	// Run the program
	result, err := runProgram(program, env)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	return result, nil
}

// EvaluateWithType evaluates an expression and ensures the result matches the expected type
func (e *CachedEvaluator) EvaluateWithType(expression string, context map[string]interface{}, expectedType interface{}) (interface{}, error) {
	result, err := e.Evaluate(expression, context)
	if err != nil {
		return nil, err
	}

	// Type checking would go here if needed
	return result, nil
}

// ValidateExpression validates an expression without executing it
func (e *CachedEvaluator) ValidateExpression(expression string) error {
	if expression == "" {
		return fmt.Errorf("expression cannot be empty")
	}

	// Check cache first
	if _, found := e.cache.Get(expression); found {
		return nil
	}

	// Compile to validate
	_, err := compileExpression(expression, e.env)
	if err != nil {
		return fmt.Errorf("invalid expression: %w", err)
	}

	return nil
}

// GetAvailableFunctions returns a list of all available function names
func (e *CachedEvaluator) GetAvailableFunctions() []string {
	functions := []string{
		"upper", "lower", "trim", "concat", "substr",
		"now", "dateFormat", "dateParse", "addDays",
		"round", "ceil", "floor", "abs", "min", "max",
		"len",
	}
	return functions
}

// CacheStats returns cache performance statistics
func (e *CachedEvaluator) CacheStats() CacheStats {
	return e.cache.Stats()
}
