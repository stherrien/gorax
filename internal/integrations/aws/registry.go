package aws

import (
	"github.com/gorax/gorax/internal/integrations"
)

// RegisterAWSActions registers all AWS integration actions with the registry
// Credentials should be injected from the credential vault at runtime
func RegisterAWSActions(registry *integrations.Registry, accessKey, secretKey, region string) error {
	// S3 actions
	if err := registry.Register(NewListBucketsAction(accessKey, secretKey, region)); err != nil {
		return err
	}
	if err := registry.Register(NewGetObjectAction(accessKey, secretKey, region)); err != nil {
		return err
	}
	if err := registry.Register(NewPutObjectAction(accessKey, secretKey, region)); err != nil {
		return err
	}
	if err := registry.Register(NewDeleteObjectAction(accessKey, secretKey, region)); err != nil {
		return err
	}

	// SNS actions
	if err := registry.Register(NewPublishMessageAction(accessKey, secretKey, region)); err != nil {
		return err
	}

	// SQS actions
	if err := registry.Register(NewSendMessageAction(accessKey, secretKey, region)); err != nil {
		return err
	}
	if err := registry.Register(NewReceiveMessageAction(accessKey, secretKey, region)); err != nil {
		return err
	}
	if err := registry.Register(NewDeleteMessageAction(accessKey, secretKey, region)); err != nil {
		return err
	}

	// Lambda actions
	if err := registry.Register(NewInvokeFunctionAction(accessKey, secretKey, region)); err != nil {
		return err
	}

	return nil
}

// GetAWSActionsList returns a list of all AWS action names
func GetAWSActionsList() []string {
	return []string{
		// S3
		"aws:s3:list_buckets",
		"aws:s3:get_object",
		"aws:s3:put_object",
		"aws:s3:delete_object",
		// SNS
		"aws:sns:publish",
		// SQS
		"aws:sqs:send_message",
		"aws:sqs:receive_message",
		"aws:sqs:delete_message",
		// Lambda
		"aws:lambda:invoke",
	}
}
