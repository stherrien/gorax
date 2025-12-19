package credential

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
)

const (
	// DataKeySize is the size of AES-256 key in bytes
	DataKeySize = 32
	// DataKeyCacheTTL is the time-to-live for cached data keys
	DataKeyCacheTTL = 5 * time.Minute
)

// KMSClientInterface defines the interface for KMS operations
type KMSClientInterface interface {
	GenerateDataKey(ctx context.Context, keyID string, encryptionContext map[string]string) ([]byte, []byte, error)
	DecryptDataKey(ctx context.Context, encryptedKey []byte, encryptionContext map[string]string) ([]byte, error)
}

// KMSClient wraps AWS KMS client for envelope encryption
type KMSClient struct {
	client *kms.Client
	keyID  string
	cache  *dataKeyCache
}

// dataKeyCache caches data keys to reduce KMS calls
type dataKeyCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
}

type cacheEntry struct {
	plainKey     []byte
	encryptedKey []byte
	expiresAt    time.Time
}

// NewKMSClient creates a new KMS client wrapper with LocalStack support
func NewKMSClient(ctx context.Context, keyID string) (*KMSClient, error) {
	if keyID == "" {
		return nil, ErrInvalidKeyID
	}

	// Load AWS config with LocalStack support
	cfg, err := loadAWSConfig(ctx)
	if err != nil {
		return nil, &KMSError{
			Op:    "NewKMSClient",
			KeyID: keyID,
			Err:   fmt.Errorf("failed to load AWS config: %w", err),
		}
	}

	// Create KMS client
	kmsClient := kms.NewFromConfig(cfg)

	return &KMSClient{
		client: kmsClient,
		keyID:  keyID,
		cache: &dataKeyCache{
			entries: make(map[string]*cacheEntry),
		},
	}, nil
}

// loadAWSConfig loads AWS configuration with LocalStack support
func loadAWSConfig(ctx context.Context) (aws.Config, error) {
	// Check if running in LocalStack mode
	localStackEndpoint := os.Getenv("LOCALSTACK_ENDPOINT")

	if localStackEndpoint != "" {
		// LocalStack configuration
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion("us-east-1"),
		)
		if err != nil {
			return aws.Config{}, err
		}

		// Override endpoint for LocalStack
		cfg.BaseEndpoint = aws.String(localStackEndpoint)

		return cfg, nil
	}

	// Standard AWS configuration
	return config.LoadDefaultConfig(ctx)
}

// GenerateDataKey generates a new AES-256 data encryption key
func (c *KMSClient) GenerateDataKey(ctx context.Context, keyID string, encryptionContext map[string]string) ([]byte, []byte, error) {
	if keyID == "" {
		return nil, nil, ErrInvalidKeyID
	}

	// Check cache first
	if plainKey, encryptedKey := c.getCachedKey(keyID, encryptionContext); plainKey != nil {
		return plainKey, encryptedKey, nil
	}

	// Convert encryption context
	kmsContext := convertEncryptionContext(encryptionContext)

	// Generate data key via KMS
	input := &kms.GenerateDataKeyInput{
		KeyId:             aws.String(keyID),
		NumberOfBytes:     aws.Int32(DataKeySize),
		EncryptionContext: kmsContext,
	}

	result, err := c.client.GenerateDataKey(ctx, input)
	if err != nil {
		return nil, nil, &KMSError{
			Op:    "GenerateDataKey",
			KeyID: keyID,
			Err:   fmt.Errorf("KMS GenerateDataKey failed: %w", err),
		}
	}

	// Validate result
	if len(result.Plaintext) != DataKeySize {
		return nil, nil, &KMSError{
			Op:    "GenerateDataKey",
			KeyID: keyID,
			Err:   fmt.Errorf("invalid data key size: got %d, want %d", len(result.Plaintext), DataKeySize),
		}
	}

	// Cache the key
	c.cacheKey(keyID, encryptionContext, result.Plaintext, result.CiphertextBlob)

	// Return copies to prevent cache modification
	plainKey := make([]byte, len(result.Plaintext))
	copy(plainKey, result.Plaintext)

	encryptedKey := make([]byte, len(result.CiphertextBlob))
	copy(encryptedKey, result.CiphertextBlob)

	return plainKey, encryptedKey, nil
}

// DecryptDataKey decrypts an encrypted data key using KMS
func (c *KMSClient) DecryptDataKey(ctx context.Context, encryptedKey []byte, encryptionContext map[string]string) ([]byte, error) {
	if len(encryptedKey) == 0 {
		return nil, ErrInvalidCiphertext
	}

	// Convert encryption context
	kmsContext := convertEncryptionContext(encryptionContext)

	// Decrypt via KMS
	input := &kms.DecryptInput{
		CiphertextBlob:    encryptedKey,
		EncryptionContext: kmsContext,
	}

	result, err := c.client.Decrypt(ctx, input)
	if err != nil {
		return nil, &KMSError{
			Op:    "DecryptDataKey",
			KeyID: c.keyID,
			Err:   fmt.Errorf("KMS Decrypt failed: %w", err),
		}
	}

	// Validate result
	if len(result.Plaintext) != DataKeySize {
		return nil, &KMSError{
			Op:    "DecryptDataKey",
			KeyID: c.keyID,
			Err:   fmt.Errorf("invalid decrypted key size: got %d, want %d", len(result.Plaintext), DataKeySize),
		}
	}

	return result.Plaintext, nil
}

// getCachedKey retrieves a cached data key if available and not expired
func (c *KMSClient) getCachedKey(keyID string, encryptionContext map[string]string) ([]byte, []byte) {
	cacheKey := buildCacheKey(keyID, encryptionContext)

	c.cache.mu.RLock()
	defer c.cache.mu.RUnlock()

	entry, exists := c.cache.entries[cacheKey]
	if !exists || time.Now().After(entry.expiresAt) {
		return nil, nil
	}

	// Return copies to prevent cache modification
	plainKey := make([]byte, len(entry.plainKey))
	copy(plainKey, entry.plainKey)

	encryptedKey := make([]byte, len(entry.encryptedKey))
	copy(encryptedKey, entry.encryptedKey)

	return plainKey, encryptedKey
}

// cacheKey stores a data key in cache
func (c *KMSClient) cacheKey(keyID string, encryptionContext map[string]string, plainKey, encryptedKey []byte) {
	cacheKey := buildCacheKey(keyID, encryptionContext)

	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()

	// Store copies to prevent external modification
	plainKeyCopy := make([]byte, len(plainKey))
	copy(plainKeyCopy, plainKey)

	encryptedKeyCopy := make([]byte, len(encryptedKey))
	copy(encryptedKeyCopy, encryptedKey)

	c.cache.entries[cacheKey] = &cacheEntry{
		plainKey:     plainKeyCopy,
		encryptedKey: encryptedKeyCopy,
		expiresAt:    time.Now().Add(DataKeyCacheTTL),
	}
}

// buildCacheKey creates a cache key from key ID and encryption context
func buildCacheKey(keyID string, encryptionContext map[string]string) string {
	// Simple cache key - in production, might want to include sorted context
	key := keyID
	if len(encryptionContext) > 0 {
		for k, v := range encryptionContext {
			key += fmt.Sprintf(":%s=%s", k, v)
		}
	}
	return key
}

// convertEncryptionContext converts map[string]string to map[string]*string for KMS API
func convertEncryptionContext(ctx map[string]string) map[string]string {
	if ctx == nil {
		return nil
	}

	// AWS SDK v2 uses map[string]string directly
	return ctx
}

// ClearKey securely zeros out a key in memory
func ClearKey(key []byte) {
	for i := range key {
		key[i] = 0
	}
}

// ClearCache clears all cached keys and zeros them out
func (c *KMSClient) ClearCache() {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()

	for key, entry := range c.cache.entries {
		// Zero out sensitive data
		ClearKey(entry.plainKey)
		ClearKey(entry.encryptedKey)
		delete(c.cache.entries, key)
	}
}

// EvictExpiredKeys removes expired entries from cache
func (c *KMSClient) EvictExpiredKeys() {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()

	now := time.Now()
	for key, entry := range c.cache.entries {
		if now.After(entry.expiresAt) {
			// Zero out sensitive data
			ClearKey(entry.plainKey)
			ClearKey(entry.encryptedKey)
			delete(c.cache.entries, key)
		}
	}
}

// GetKeyID returns the default KMS key ID
func (c *KMSClient) GetKeyID() string {
	return c.keyID
}
