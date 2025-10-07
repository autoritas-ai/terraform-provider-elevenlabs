# Terraform Provider for ElevenLabs

This Terraform provider allows you to manage resources on the ElevenLabs platform, including agents, tools, knowledge bases, and voices.

## Provider Configuration

To use this provider, you need to configure it with your ElevenLabs API key.

```hcl
terraform {
  required_providers {
    elevenlabs = {
      source  = "user/elevenlabs"
      version = "0.1.0"
    }
  }
}

provider "elevenlabs" {
  api_key = var.elevenlabs_api_key
}
```

You can provide the API key via the `ELEVENLABS_API_KEY` environment variable or directly in the provider configuration.

## Resources

### `elevenlabs_agent`

Manages an ElevenLabs agent.

#### Example Usage

```hcl
resource "elevenlabs_agent" "example" {
  name = "My Example Agent"

  conversation_config {
    agent {
      first_message = "Hello, how can I help you today?"
      prompt {
        prompt = "You are a helpful assistant."
        llm    = "gpt-4o-mini"
      }
    }
  }
}
```

#### Argument Reference

*   `name` - (Optional) The name of the agent.
*   `conversation_config` - (Required) A block that configures the agent's conversation behavior.
    *   `agent` - (Required) A block that configures the agent itself.
        *   `first_message` - (Optional) The first message the agent sends in a conversation.
        *   `prompt` - (Required) A block that defines the agent's system prompt.
            *   `prompt` - (Required) The system prompt text.
            *   `llm` - (Optional) The language model to use. Defaults to `gpt-4o-mini`.

#### Attribute Reference

*   `agent_id` - The unique ID of the agent.

---

### `elevenlabs_tool`

Manages an ElevenLabs tool for an agent.

#### Example Usage

```hcl
resource "elevenlabs_tool" "example" {
  name        = "Get Weather"
  description = "A tool to get the current weather for a location."

  api_schema {
    url    = "https://api.example.com/weather?location={location}"
    method = "GET"
  }
}
```

#### Argument Reference

*   `name` - (Required) The name of the tool.
*   `description` - (Optional) A description of the tool.
*   `api_schema` - (Required) A block that defines the API schema for the tool.
    *   `url` - (Required) The URL of the API endpoint.
    *   `method` - (Optional) The HTTP method to use. Defaults to `GET`.

#### Attribute Reference

*   `tool_id` - The unique ID of the tool.

---

### `elevenlabs_knowledge_base_document`

Manages a document in an ElevenLabs knowledge base.

#### Example Usage

**From a URL:**

```hcl
resource "elevenlabs_knowledge_base_document" "from_url" {
  name       = "ElevenLabs API Docs"
  source_url = "https://elevenlabs.io/docs/api-reference"
  mime_type  = "text/html"
}
```

**From text:**

```hcl
resource "elevenlabs_knowledge_base_document" "from_text" {
  name        = "My Custom Document"
  source_text = "This is the content of my custom document."
}
```

**From a file:**

```hcl
resource "elevenlabs_knowledge_base_document" "from_file" {
  name             = "My Local Document"
  source_file_path = "./my-document.txt"
}
```

#### Argument Reference

*   `name` - (Required) The name of the document.
*   `source_url` - (Optional) The URL to create the document from. Conflicts with `source_text` and `source_file_path`.
*   `source_text` - (Optional) The text content to create the document from. Conflicts with `source_url` and `source_file_path`.
*   `source_file_path` - (Optional) The local file path to create the document from. Conflicts with `source_url` and `source_text`.
*   `mime_type` - (Optional) The MIME type of the document.

#### Attribute Reference

*   `document_id` - The unique ID of the document.
*   `source_hash` - The SHA256 hash of the document content when created from `source_text` or `source_file_path`.

---

### `elevenlabs_voice`

Manages a custom voice in ElevenLabs.

#### Example Usage

```hcl
resource "elevenlabs_voice" "example" {
  name        = "My Custom Voice"
  description = "A voice created for my project."
  labels = {
    accent = "American"
    age    = "young"
  }
  files = [
    "./sample1.mp3",
    "./sample2.mp3",
  ]
}
```

#### Argument Reference

*   `name` - (Required) The name of the voice.
*   `description` - (Optional) A description for the voice.
*   `labels` - (Optional) A map of labels for the voice.
*   `files` - (Required) A set of file paths to the audio samples for cloning the voice.

#### Attribute Reference

*   `voice_id` - The unique ID of the voice.

---

### `elevenlabs_voice_settings`

Manages the settings for an ElevenLabs voice.

#### Example Usage

```hcl
resource "elevenlabs_voice_settings" "example" {
  voice_id         = elevenlabs_voice.example.voice_id
  stability        = 0.75
  similarity_boost = 0.8
}
```

#### Argument Reference

*   `voice_id` - (Required) The ID of the voice to manage the settings for.
*   `stability` - (Required) The stability of the voice. A value between 0 and 1.
*   `similarity_boost` - (Required) The similarity boost of the voice. A value between 0 and 1.
*   `style` - (Optional) The style exaggeration of the voice. Defaults to `0.0`.
*   `use_speaker_boost` - (Optional) Whether to use speaker boost. Defaults to `true`.
*   `speed` - (Optional) The speed of the voice. Defaults to `1.0`.