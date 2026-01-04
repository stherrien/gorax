package messaging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid SQS config",
			config: Config{
				Type:   QueueTypeSQS,
				Region: "us-east-1",
			},
			wantErr: false,
		},
		{
			name: "valid Kafka config",
			config: Config{
				Type:    QueueTypeKafka,
				Brokers: []string{"localhost:9092"},
			},
			wantErr: false,
		},
		{
			name: "valid RabbitMQ config",
			config: Config{
				Type: QueueTypeRabbitMQ,
				URL:  "amqp://guest:guest@localhost:5672/",
			},
			wantErr: false,
		},
		{
			name:    "missing queue type",
			config:  Config{},
			wantErr: true,
			errMsg:  "queue type is required",
		},
		{
			name: "SQS missing region",
			config: Config{
				Type: QueueTypeSQS,
			},
			wantErr: true,
			errMsg:  "region is required for SQS",
		},
		{
			name: "Kafka missing brokers",
			config: Config{
				Type: QueueTypeKafka,
			},
			wantErr: true,
			errMsg:  "brokers are required for Kafka",
		},
		{
			name: "RabbitMQ missing URL",
			config: Config{
				Type: QueueTypeRabbitMQ,
			},
			wantErr: true,
			errMsg:  "URL is required for RabbitMQ",
		},
		{
			name: "unsupported queue type",
			config: Config{
				Type: "invalid",
			},
			wantErr: true,
			errMsg:  "unsupported queue type: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
