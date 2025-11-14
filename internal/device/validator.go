package device

import (
	"fmt"

	"github.com/tienpdinh/gpt-home/pkg/models"
)

// ValidationResult represents the result of action validation
type ValidationResult struct {
	Valid       bool
	Error       string
	Warning     string
	SafeAction  *models.DeviceAction
}

// Validator performs safety checks on device actions
type Validator struct{}

// NewValidator creates a new device action validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateAction validates a device action for safety
func (v *Validator) ValidateAction(action *models.DeviceAction) ValidationResult {
	if action == nil {
		return ValidationResult{
			Valid: false,
			Error: "action cannot be nil",
		}
	}

	// Validate action name
	switch action.Action {
	case "turn_on", "turn_off":
		return v.validateOnOff(action)
	case "set_brightness":
		return v.validateBrightness(action)
	case "set_temperature":
		return v.validateTemperature(action)
	case "set_color_temp":
		return v.validateColorTemp(action)
	case "set_humidity":
		return v.validateHumidity(action)
	case "open", "close":
		return v.validateCoverAction(action)
	default:
		return ValidationResult{
			Valid: false,
			Error: fmt.Sprintf("unknown action: %s", action.Action),
		}
	}
}

// validateOnOff validates turn_on/turn_off actions
func (v *Validator) validateOnOff(action *models.DeviceAction) ValidationResult {
	if action.Parameters == nil {
		action.Parameters = make(map[string]any)
	}

	return ValidationResult{
		Valid:      true,
		SafeAction: action,
	}
}

// validateBrightness validates brightness values (0-255)
func (v *Validator) validateBrightness(action *models.DeviceAction) ValidationResult {
	if action.Parameters == nil {
		return ValidationResult{
			Valid:  false,
			Error:  "brightness action requires parameters",
		}
	}

	brightness, ok := action.Parameters["brightness"]
	if !ok {
		return ValidationResult{
			Valid:  false,
			Error:  "brightness action requires 'brightness' parameter",
		}
	}

	// Convert to float64 if needed
	var brightness_value float64
	switch v := brightness.(type) {
	case float64:
		brightness_value = v
	case int:
		brightness_value = float64(v)
	default:
		return ValidationResult{
			Valid:  false,
			Error:  "brightness must be a number",
		}
	}

	// Clamp to valid range
	if brightness_value < 0 {
		return ValidationResult{
			Valid:    false,
			Error:    "brightness cannot be negative",
			Warning:  "requested brightness was negative, clamped to 0",
		}
	}

	if brightness_value > 255 {
		return ValidationResult{
			Valid:    false,
			Error:    "brightness cannot exceed 255",
			Warning:  "requested brightness exceeded 255, clamped to 255",
		}
	}

	// Create safe action
	safeAction := &models.DeviceAction{
		Action: action.Action,
		Parameters: map[string]any{
			"brightness": int(brightness_value),
		},
	}

	return ValidationResult{
		Valid:      true,
		SafeAction: safeAction,
	}
}

// validateTemperature validates temperature values (18-28°C recommended)
func (v *Validator) validateTemperature(action *models.DeviceAction) ValidationResult {
	if action.Parameters == nil {
		return ValidationResult{
			Valid:  false,
			Error:  "temperature action requires parameters",
		}
	}

	temperature, ok := action.Parameters["temperature"]
	if !ok {
		return ValidationResult{
			Valid:  false,
			Error:  "temperature action requires 'temperature' parameter",
		}
	}

	// Convert to float64 if needed
	var temp_value float64
	switch v := temperature.(type) {
	case float64:
		temp_value = v
	case int:
		temp_value = float64(v)
	default:
		return ValidationResult{
			Valid:  false,
			Error:  "temperature must be a number",
		}
	}

	// Check for dangerous values
	if temp_value < 10 || temp_value > 40 {
		return ValidationResult{
			Valid:    false,
			Error:    fmt.Sprintf("temperature %.1f°C is outside safe range (10-40°C)", temp_value),
			Warning:  "extremely high or low temperature requested",
		}
	}

	// Warn for uncomfortable values
	var warning string
	if temp_value < 16 {
		warning = "temperature is very cold - ensure this is intentional"
	} else if temp_value > 28 {
		warning = "temperature is very warm - ensure this is intentional"
	}

	// Create safe action
	safeAction := &models.DeviceAction{
		Action: action.Action,
		Parameters: map[string]any{
			"temperature": temp_value,
		},
	}

	result := ValidationResult{
		Valid:      true,
		SafeAction: safeAction,
	}
	if warning != "" {
		result.Warning = warning
	}

	return result
}

// validateColorTemp validates color temperature values (2700-6500K)
func (v *Validator) validateColorTemp(action *models.DeviceAction) ValidationResult {
	if action.Parameters == nil {
		return ValidationResult{
			Valid:  false,
			Error:  "color_temp action requires parameters",
		}
	}

	colorTemp, ok := action.Parameters["color_temp"]
	if !ok {
		return ValidationResult{
			Valid:  false,
			Error:  "color_temp action requires 'color_temp' parameter",
		}
	}

	// Convert to float64 if needed
	var kelvin_value float64
	switch v := colorTemp.(type) {
	case float64:
		kelvin_value = v
	case int:
		kelvin_value = float64(v)
	default:
		return ValidationResult{
			Valid:  false,
			Error:  "color_temp must be a number in kelvin",
		}
	}

	// Valid range for typical smart bulbs
	if kelvin_value < 2700 || kelvin_value > 6500 {
		return ValidationResult{
			Valid:    false,
			Error:    fmt.Sprintf("color temperature %.0fK is outside typical range (2700-6500K)", kelvin_value),
		}
	}

	// Create safe action
	safeAction := &models.DeviceAction{
		Action: action.Action,
		Parameters: map[string]any{
			"color_temp": kelvin_value,
		},
	}

	return ValidationResult{
		Valid:      true,
		SafeAction: safeAction,
	}
}

// validateHumidity validates humidity values (30-70% recommended)
func (v *Validator) validateHumidity(action *models.DeviceAction) ValidationResult {
	if action.Parameters == nil {
		return ValidationResult{
			Valid:  false,
			Error:  "humidity action requires parameters",
		}
	}

	humidity, ok := action.Parameters["humidity"]
	if !ok {
		return ValidationResult{
			Valid:  false,
			Error:  "humidity action requires 'humidity' parameter",
		}
	}

	// Convert to float64 if needed
	var humidity_value float64
	switch v := humidity.(type) {
	case float64:
		humidity_value = v
	case int:
		humidity_value = float64(v)
	default:
		return ValidationResult{
			Valid:  false,
			Error:  "humidity must be a number (0-100)",
		}
	}

	if humidity_value < 0 || humidity_value > 100 {
		return ValidationResult{
			Valid:  false,
			Error:  "humidity must be between 0 and 100",
		}
	}

	// Create safe action
	safeAction := &models.DeviceAction{
		Action: action.Action,
		Parameters: map[string]any{
			"humidity": humidity_value,
		},
	}

	return ValidationResult{
		Valid:      true,
		SafeAction: safeAction,
	}
}

// validateCoverAction validates cover open/close actions
func (v *Validator) validateCoverAction(action *models.DeviceAction) ValidationResult {
	if action.Parameters == nil {
		action.Parameters = make(map[string]any)
	}

	if action.Action != "open" && action.Action != "close" {
		return ValidationResult{
			Valid:  false,
			Error:  "cover action must be 'open' or 'close'",
		}
	}

	return ValidationResult{
		Valid:      true,
		SafeAction: action,
	}
}
