package actions

import (
	"context"
	"encoding/json"
	"fmt"
)

// TransformAction implements the Action interface for data transformation
type TransformAction struct{}

// TransformActionConfig represents the configuration for a transform action
type TransformActionConfig struct {
	// Expression is a JSONPath expression to extract a value
	Expression string `json:"expression,omitempty"`
	// Mapping defines field mappings from source paths to target keys
	Mapping map[string]string `json:"mapping,omitempty"`
	// Default value to use if extraction fails
	Default interface{} `json:"default,omitempty"`
}

// Execute implements the Action interface
func (a *TransformAction) Execute(ctx context.Context, input *ActionInput) (*ActionOutput, error) {
	// Parse config
	configBytes, err := json.Marshal(input.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var config TransformActionConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to parse transform action config: %w", err)
	}

	// Execute transformation
	result, err := a.executeTransform(ctx, config, input.Context)
	if err != nil {
		// If default is specified and transformation fails, use default
		if config.Default != nil {
			return NewActionOutput(config.Default), nil
		}
		return nil, err
	}

	return NewActionOutput(result), nil
}

// executeTransform executes the transformation
func (a *TransformAction) executeTransform(ctx context.Context, config TransformActionConfig, execContext map[string]interface{}) (interface{}, error) {
	// If mapping is provided, create output from mapping
	if len(config.Mapping) > 0 {
		return a.executeMapping(config.Mapping, execContext)
	}

	// If expression is provided, evaluate it
	if config.Expression != "" {
		return a.executeExpression(config.Expression, execContext)
	}

	// Return input context if no transformation specified
	return execContext, nil
}

// executeMapping creates a new object based on the mapping configuration
func (a *TransformAction) executeMapping(mapping map[string]string, context map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for targetKey, sourcePath := range mapping {
		value, err := GetValueByPath(context, sourcePath)
		if err != nil {
			// Use nil for missing values instead of failing
			result[targetKey] = nil
			continue
		}
		result[targetKey] = value
	}

	return result, nil
}

// executeExpression evaluates a JSONPath expression
func (a *TransformAction) executeExpression(expression string, context map[string]interface{}) (interface{}, error) {
	value, err := GetValueByPath(context, expression)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression '%s': %w", expression, err)
	}

	return value, nil
}

// Legacy function for backward compatibility
func ExecuteTransform(ctx context.Context, config TransformActionConfig, executionContext map[string]interface{}) (interface{}, error) {
	action := &TransformAction{}
	input := NewActionInput(config, executionContext)
	output, err := action.Execute(ctx, input)
	if err != nil {
		return nil, err
	}
	return output.Data, nil
}
