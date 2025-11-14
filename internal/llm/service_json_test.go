package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStructuredResponseValidJSON(t *testing.T) {
	service := NewService("http://localhost:11434", "test-model")

	jsonResponse := `{
		"understanding": "Turn on bedroom lights",
		"response": "I'll turn on the bedroom lights for you",
		"actions": [
			{
				"action": "turn_on",
				"parameters": {}
			}
		],
		"confidence": 0.95
	}`

	result := service.parseStructuredResponse(jsonResponse)
	require.NotNil(t, result)

	assert.Equal(t, "Turn on bedroom lights", result.Understanding)
	assert.Equal(t, "I'll turn on the bedroom lights for you", result.Response)
	assert.Equal(t, 1, len(result.Actions))
	assert.Equal(t, "turn_on", result.Actions[0].Action)
	assert.Equal(t, float32(0.95), result.Confidence)
}

func TestParseStructuredResponseMarkdownJSON(t *testing.T) {
	service := NewService("http://localhost:11434", "test-model")

	// Some models wrap JSON in markdown code blocks
	jsonResponse := "```json\n{\n\t\"understanding\": \"Set brightness to 50%\",\n\t\"response\": \"Setting brightness\",\n\t\"actions\": [\n\t\t{\n\t\t\t\"action\": \"set_brightness\",\n\t\t\t\"parameters\": {\"brightness\": 128}\n\t\t}\n\t],\n\t\"confidence\": 0.9\n}\n```"

	result := service.parseStructuredResponse(jsonResponse)
	require.NotNil(t, result)

	assert.Equal(t, "Set brightness to 50%", result.Understanding)
	assert.Equal(t, 1, len(result.Actions))
	assert.Equal(t, "set_brightness", result.Actions[0].Action)
	assert.Equal(t, float64(128), result.Actions[0].Parameters["brightness"])
}

func TestParseStructuredResponseWithoutMarkdownJSON(t *testing.T) {
	service := NewService("http://localhost:11434", "test-model")

	jsonResponse := "```\n{\n\t\"understanding\": \"Test\",\n\t\"response\": \"Testing\",\n\t\"actions\": [],\n\t\"confidence\": 0.85\n}\n```"

	result := service.parseStructuredResponse(jsonResponse)
	require.NotNil(t, result)

	assert.Equal(t, "Test", result.Understanding)
	assert.Equal(t, 0, len(result.Actions))
}

func TestParseStructuredResponseInvalidJSON(t *testing.T) {
	service := NewService("http://localhost:11434", "test-model")

	invalidJSON := "This is not JSON at all"

	result := service.parseStructuredResponse(invalidJSON)
	assert.Nil(t, result)
}

func TestParseStructuredResponseWithActions(t *testing.T) {
	service := NewService("http://localhost:11434", "test-model")

	jsonResponse := `{
		"understanding": "Adjust temperature and brightness",
		"response": "Setting temperature to 22°C and brightness to 75%",
		"actions": [
			{
				"action": "set_temperature",
				"parameters": {"temperature": 22}
			},
			{
				"action": "set_brightness",
				"parameters": {"brightness": 191}
			}
		],
		"confidence": 0.92
	}`

	result := service.parseStructuredResponse(jsonResponse)
	require.NotNil(t, result)

	assert.Equal(t, 2, len(result.Actions))
	assert.Equal(t, "set_temperature", result.Actions[0].Action)
	assert.Equal(t, "set_brightness", result.Actions[1].Action)
	assert.Equal(t, float64(22), result.Actions[0].Parameters["temperature"])
	assert.Equal(t, float64(191), result.Actions[1].Parameters["brightness"])
}

func TestParseStructuredResponseEmptyActions(t *testing.T) {
	service := NewService("http://localhost:11434", "test-model")

	jsonResponse := `{
		"understanding": "User asking for status",
		"response": "The lights are currently on in the living room",
		"actions": [],
		"confidence": 0.88
	}`

	result := service.parseStructuredResponse(jsonResponse)
	require.NotNil(t, result)

	assert.Equal(t, 0, len(result.Actions))
	assert.Equal(t, "User asking for status", result.Understanding)
}
