package llm

// LLMBackend defines the interface for different LLM implementations
type LLMBackend interface {
	LoadModel() error
	UnloadModel() error
	IsLoaded() bool
	GenerateResponse(prompt string, config GenerationConfig) (string, error)
	GetModelInfo() ModelInfo
}

// GenerationConfig holds parameters for text generation
type GenerationConfig struct {
	MaxTokens   int      `json:"max_tokens"`
	Temperature float32  `json:"temperature"`
	TopP        float32  `json:"top_p"`
	TopK        int      `json:"top_k"`
	StopTokens  []string `json:"stop_tokens"`
}

// PromptTemplate defines how to format prompts for smart home control
type PromptTemplate struct {
	SystemPrompt string
	UserTemplate string
}

// SmartHomePromptTemplate is optimized for device control
var SmartHomePromptTemplate = PromptTemplate{
	SystemPrompt: `You are a smart home assistant. Your job is to understand natural language requests and respond with device control actions and helpful responses.

Available device types: light, switch, sensor, climate, cover, fan, media_player

For device control, respond with JSON actions in this format:
{
  "response": "I'll turn on the lights for you.",
  "actions": [
    {
      "action": "turn_on",
      "device_type": "light", 
      "parameters": {"entity_id": "light.living_room"}
    }
  ]
}

Common actions:
- turn_on, turn_off (lights, switches)
- set_brightness (lights, value 0-255)
- set_temperature (climate, value in celsius)
- set_volume (media_player, value 0-1)

Be conversational and helpful. If you can't determine the specific device, ask for clarification.`,

	UserTemplate: `User request: {{.Message}}

{{if .Context.ReferencedDevices}}
Recently mentioned devices: {{.Context.ReferencedDevices}}
{{end}}

{{if .Context.LastAction}}
Last action performed: {{.Context.LastAction.Action}}
{{end}}

Please respond with device control actions and a helpful message.`,
}
