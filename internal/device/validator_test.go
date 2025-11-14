package device

import (
	"testing"

	"github.com/tienpdinh/gpt-home/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestValidatorBrightness(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name      string
		action    *models.DeviceAction
		wantValid bool
		wantError string
	}{
		{
			name: "valid brightness",
			action: &models.DeviceAction{
				Action: "set_brightness",
				Parameters: map[string]any{
					"brightness": 128,
				},
			},
			wantValid: true,
		},
		{
			name: "brightness as float",
			action: &models.DeviceAction{
				Action: "set_brightness",
				Parameters: map[string]any{
					"brightness": 128.5,
				},
			},
			wantValid: true,
		},
		{
			name: "brightness too high",
			action: &models.DeviceAction{
				Action: "set_brightness",
				Parameters: map[string]any{
					"brightness": 300,
				},
			},
			wantValid: false,
			wantError: "cannot exceed 255",
		},
		{
			name: "brightness negative",
			action: &models.DeviceAction{
				Action: "set_brightness",
				Parameters: map[string]any{
					"brightness": -10,
				},
			},
			wantValid: false,
			wantError: "cannot be negative",
		},
		{
			name: "missing brightness parameter",
			action: &models.DeviceAction{
				Action:     "set_brightness",
				Parameters: map[string]any{},
			},
			wantValid: false,
			wantError: "requires 'brightness' parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateAction(tt.action)
			assert.Equal(t, tt.wantValid, result.Valid)
			if !tt.wantValid && tt.wantError != "" {
				assert.Contains(t, result.Error, tt.wantError)
			}
		})
	}
}

func TestValidatorTemperature(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name      string
		action    *models.DeviceAction
		wantValid bool
		wantError string
	}{
		{
			name: "valid temperature",
			action: &models.DeviceAction{
				Action: "set_temperature",
				Parameters: map[string]any{
					"temperature": 22.5,
				},
			},
			wantValid: true,
		},
		{
			name: "comfortable temperature",
			action: &models.DeviceAction{
				Action: "set_temperature",
				Parameters: map[string]any{
					"temperature": 20,
				},
			},
			wantValid: true,
		},
		{
			name: "temperature too high",
			action: &models.DeviceAction{
				Action: "set_temperature",
				Parameters: map[string]any{
					"temperature": 50,
				},
			},
			wantValid: false,
			wantError: "outside safe range",
		},
		{
			name: "temperature too low",
			action: &models.DeviceAction{
				Action: "set_temperature",
				Parameters: map[string]any{
					"temperature": 5,
				},
			},
			wantValid: false,
			wantError: "outside safe range",
		},
		{
			name: "missing temperature parameter",
			action: &models.DeviceAction{
				Action:     "set_temperature",
				Parameters: map[string]any{},
			},
			wantValid: false,
			wantError: "requires 'temperature' parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateAction(tt.action)
			assert.Equal(t, tt.wantValid, result.Valid)
			if !tt.wantValid && tt.wantError != "" {
				assert.Contains(t, result.Error, tt.wantError)
			}
		})
	}
}

func TestValidatorColorTemp(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name      string
		action    *models.DeviceAction
		wantValid bool
		wantError string
	}{
		{
			name: "valid color temp warm",
			action: &models.DeviceAction{
				Action: "set_color_temp",
				Parameters: map[string]any{
					"color_temp": 2700,
				},
			},
			wantValid: true,
		},
		{
			name: "valid color temp cool",
			action: &models.DeviceAction{
				Action: "set_color_temp",
				Parameters: map[string]any{
					"color_temp": 6500,
				},
			},
			wantValid: true,
		},
		{
			name: "color temp too low",
			action: &models.DeviceAction{
				Action: "set_color_temp",
				Parameters: map[string]any{
					"color_temp": 2000,
				},
			},
			wantValid: false,
			wantError: "outside typical range",
		},
		{
			name: "color temp too high",
			action: &models.DeviceAction{
				Action: "set_color_temp",
				Parameters: map[string]any{
					"color_temp": 7000,
				},
			},
			wantValid: false,
			wantError: "outside typical range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateAction(tt.action)
			assert.Equal(t, tt.wantValid, result.Valid)
			if !tt.wantValid && tt.wantError != "" {
				assert.Contains(t, result.Error, tt.wantError)
			}
		})
	}
}

func TestValidatorOnOff(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name      string
		action    *models.DeviceAction
		wantValid bool
	}{
		{
			name: "turn on",
			action: &models.DeviceAction{
				Action:     "turn_on",
				Parameters: map[string]any{},
			},
			wantValid: true,
		},
		{
			name: "turn off",
			action: &models.DeviceAction{
				Action:     "turn_off",
				Parameters: map[string]any{},
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateAction(tt.action)
			assert.Equal(t, tt.wantValid, result.Valid)
		})
	}
}

func TestValidatorUnknownAction(t *testing.T) {
	validator := NewValidator()

	action := &models.DeviceAction{
		Action: "unknown_action",
	}

	result := validator.ValidateAction(action)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Error, "unknown action")
}
