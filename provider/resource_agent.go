package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAgent() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAgentCreate,
		ReadContext:   resourceAgentRead,
		UpdateContext: resourceAgentUpdate,
		DeleteContext: resourceAgentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"agent_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"conversation_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tts": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"voice_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"stability": {
										Type:     schema.TypeFloat,
										Optional: true,
									},
									"speed": {
										Type:     schema.TypeFloat,
										Optional: true,
									},
									"similarity_boost": {
										Type:     schema.TypeFloat,
										Optional: true,
									},
								},
							},
						},
						"agent": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"first_message": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"language": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"prompt": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"prompt": {
													Type:     schema.TypeString,
													Required: true,
												},
												"llm": {
													Type:     schema.TypeString,
													Optional: true,
													Default:  "gpt-4o-mini",
												},
												"tools": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"temperature": {
													Type:     schema.TypeFloat,
													Optional: true,
												},
												"max_tokens": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"knowledge_base": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"type": {
																Type:     schema.TypeString,
																Required: true,
															},
															"name": {
																Type:     schema.TypeString,
																Required: true,
															},
															"id": {
																Type:     schema.TypeString,
																Required: true,
															},
															"usage_mode": {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceAgentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	agent := &Agent{
		Name: d.Get("name").(string),
	}

	if v, ok := d.Get("tags").(*schema.Set); ok && v.Len() > 0 {
		tags := make([]string, v.Len())
		for i, tag := range v.List() {
			tags[i] = tag.(string)
		}
		agent.Tags = tags
	}

	if v, ok := d.Get("conversation_config").([]interface{}); ok && len(v) > 0 {
		configData := v[0].(map[string]interface{})
		convConfig := &ConversationConfig{}

		if ttsList, ok := configData["tts"].([]interface{}); ok && len(ttsList) > 0 {
			ttsData := ttsList[0].(map[string]interface{})
			ttsConfig := &TTSConfig{}
			if val, ok := ttsData["voice_id"].(string); ok && val != "" {
				ttsConfig.VoiceID = val
			}
			if val, ok := ttsData["stability"].(float64); ok {
				v := val
				ttsConfig.Stability = &v
			}
			if val, ok := ttsData["speed"].(float64); ok {
				v := val
				ttsConfig.Speed = &v
			}
			if val, ok := ttsData["similarity_boost"].(float64); ok {
				v := val
				ttsConfig.SimilarityBoost = &v
			}
			convConfig.TTS = ttsConfig
		}

		if agentList, ok := configData["agent"].([]interface{}); ok && len(agentList) > 0 {
			agentData := agentList[0].(map[string]interface{})
			agentConfig := &AgentConfig{
				FirstMessage: agentData["first_message"].(string),
				Language:     agentData["language"].(string),
			}

			if promptList, ok := agentData["prompt"].([]interface{}); ok && len(promptList) > 0 {
				promptData := promptList[0].(map[string]interface{})
				promptConfig := &PromptConfig{
					Prompt: promptData["prompt"].(string),
					LLM:    promptData["llm"].(string),
				}
				if val, ok := promptData["temperature"].(float64); ok {
					v := val
					promptConfig.Temperature = &v
				}
				if val, ok := promptData["max_tokens"].(int); ok {
					v := val
					promptConfig.MaxTokens = &v
				}
				if toolSet, ok := promptData["tools"].(*schema.Set); ok && toolSet.Len() > 0 {
					tools := make([]string, toolSet.Len())
					for i, tool := range toolSet.List() {
						tools[i] = tool.(string)
					}
					promptConfig.ToolIDs = tools
				}
				if kbList, ok := promptData["knowledge_base"].([]interface{}); ok && len(kbList) > 0 {
					kbs := make([]*KnowledgeBaseLocator, len(kbList))
					for i, item := range kbList {
						kbData := item.(map[string]interface{})
						kbs[i] = &KnowledgeBaseLocator{
							Type:      kbData["type"].(string),
							Name:      kbData["name"].(string),
							ID:        kbData["id"].(string),
							UsageMode: kbData["usage_mode"].(string),
						}
					}
					promptConfig.KnowledgeBase = kbs
				}
				agentConfig.Prompt = promptConfig
			}
			convConfig.Agent = agentConfig
		}
		agent.ConversationConfig = convConfig
	}

	createdAgent, err := client.CreateAgent(ctx, agent)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(createdAgent.AgentID)
	return resourceAgentRead(ctx, d, m)
}

func resourceAgentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	agentID := d.Id()

	agent, err := client.GetAgent(ctx, agentID)
	if err != nil {
		return diag.FromErr(err)
	}

	if agent == nil {
		d.SetId("")
		return nil
	}

	d.Set("agent_id", agent.AgentID)
	d.Set("name", agent.Name)
	if err := d.Set("tags", agent.Tags); err != nil {
		return diag.FromErr(err)
	}

	if agent.ConversationConfig != nil {
		convConfigMap := make(map[string]interface{})

		if agent.ConversationConfig.TTS != nil {
			ttsMap := make(map[string]interface{})
			ttsMap["voice_id"] = agent.ConversationConfig.TTS.VoiceID
			if agent.ConversationConfig.TTS.Stability != nil {
				ttsMap["stability"] = *agent.ConversationConfig.TTS.Stability
			}
			if agent.ConversationConfig.TTS.Speed != nil {
				ttsMap["speed"] = *agent.ConversationConfig.TTS.Speed
			}
			if agent.ConversationConfig.TTS.SimilarityBoost != nil {
				ttsMap["similarity_boost"] = *agent.ConversationConfig.TTS.SimilarityBoost
			}
			convConfigMap["tts"] = []interface{}{ttsMap}
		}

		if agent.ConversationConfig.Agent != nil {
			agentConfigMap := make(map[string]interface{})
			agentConfigMap["first_message"] = agent.ConversationConfig.Agent.FirstMessage
			agentConfigMap["language"] = agent.ConversationConfig.Agent.Language

			if agent.ConversationConfig.Agent.Prompt != nil {
				promptMap := make(map[string]interface{})
				promptMap["prompt"] = agent.ConversationConfig.Agent.Prompt.Prompt
				promptMap["llm"] = agent.ConversationConfig.Agent.Prompt.LLM
				promptMap["tools"] = agent.ConversationConfig.Agent.Prompt.ToolIDs
				if agent.ConversationConfig.Agent.Prompt.Temperature != nil {
					promptMap["temperature"] = *agent.ConversationConfig.Agent.Prompt.Temperature
				}
				if agent.ConversationConfig.Agent.Prompt.MaxTokens != nil {
					promptMap["max_tokens"] = *agent.ConversationConfig.Agent.Prompt.MaxTokens
				}
				if agent.ConversationConfig.Agent.Prompt.KnowledgeBase != nil {
					kbList := make([]interface{}, len(agent.ConversationConfig.Agent.Prompt.KnowledgeBase))
					for i, kb := range agent.ConversationConfig.Agent.Prompt.KnowledgeBase {
						kbMap := make(map[string]interface{})
						kbMap["type"] = kb.Type
						kbMap["name"] = kb.Name
						kbMap["id"] = kb.ID
						kbMap["usage_mode"] = kb.UsageMode
						kbList[i] = kbMap
					}
					promptMap["knowledge_base"] = kbList
				}
				agentConfigMap["prompt"] = []interface{}{promptMap}
			}
			convConfigMap["agent"] = []interface{}{agentConfigMap}
		}

		if err := d.Set("conversation_config", []interface{}{convConfigMap}); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceAgentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	agentID := d.Id()

	if d.HasChange("name") || d.HasChange("tags") || d.HasChange("conversation_config") {
		agent := &Agent{
			Name: d.Get("name").(string),
		}

		if v, ok := d.Get("tags").(*schema.Set); ok && v.Len() > 0 {
			tags := make([]string, v.Len())
			for i, tag := range v.List() {
				tags[i] = tag.(string)
			}
			agent.Tags = tags
		}

		if v, ok := d.Get("conversation_config").([]interface{}); ok && len(v) > 0 {
			configData := v[0].(map[string]interface{})
			convConfig := &ConversationConfig{}

			if ttsList, ok := configData["tts"].([]interface{}); ok && len(ttsList) > 0 {
				ttsData := ttsList[0].(map[string]interface{})
				ttsConfig := &TTSConfig{}
				if val, ok := ttsData["voice_id"].(string); ok && val != "" {
					ttsConfig.VoiceID = val
				}
				if val, ok := ttsData["stability"].(float64); ok {
					v := val
					ttsConfig.Stability = &v
				}
				if val, ok := ttsData["speed"].(float64); ok {
					v := val
					ttsConfig.Speed = &v
				}
				if val, ok := ttsData["similarity_boost"].(float64); ok {
					v := val
					ttsConfig.SimilarityBoost = &v
				}
				convConfig.TTS = ttsConfig
			}

			if agentList, ok := configData["agent"].([]interface{}); ok && len(agentList) > 0 {
				agentData := agentList[0].(map[string]interface{})
				agentConfig := &AgentConfig{
					FirstMessage: agentData["first_message"].(string),
					Language:     agentData["language"].(string),
				}

				if promptList, ok := agentData["prompt"].([]interface{}); ok && len(promptList) > 0 {
					promptData := promptList[0].(map[string]interface{})
					promptConfig := &PromptConfig{
						Prompt: promptData["prompt"].(string),
						LLM:    promptData["llm"].(string),
					}
					if val, ok := promptData["temperature"].(float64); ok {
						v := val
						promptConfig.Temperature = &v
					}
					if val, ok := promptData["max_tokens"].(int); ok {
						v := val
						promptConfig.MaxTokens = &v
					}
					if toolSet, ok := promptData["tools"].(*schema.Set); ok && toolSet.Len() > 0 {
						tools := make([]string, toolSet.Len())
						for i, tool := range toolSet.List() {
							tools[i] = tool.(string)
						}
						promptConfig.ToolIDs = tools
					}
					if kbList, ok := promptData["knowledge_base"].([]interface{}); ok && len(kbList) > 0 {
						kbs := make([]*KnowledgeBaseLocator, len(kbList))
						for i, item := range kbList {
							kbData := item.(map[string]interface{})
							kbs[i] = &KnowledgeBaseLocator{
								Type:      kbData["type"].(string),
								Name:      kbData["name"].(string),
								ID:        kbData["id"].(string),
								UsageMode: kbData["usage_mode"].(string),
							}
						}
						promptConfig.KnowledgeBase = kbs
					}
					agentConfig.Prompt = promptConfig
				}
				convConfig.Agent = agentConfig
			}
			agent.ConversationConfig = convConfig
		}

		err := client.UpdateAgent(ctx, agentID, agent)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceAgentRead(ctx, d, m)
}

func resourceAgentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	agentID := d.Id()

	err := client.DeleteAgent(ctx, agentID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
