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
		Description: "Provides a resource to manage an agent on ElevenLabs.",
		Schema: map[string]*schema.Schema{
			"agent_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the agent.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the agent.",
			},
			"conversation_config": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Defines the conversation configuration for the agent.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"agent": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "Configures the agent's behavior.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"first_message": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The initial message the agent sends.",
									},
									"prompt": {
										Type:        schema.TypeList,
										Required:    true,
										MaxItems:    1,
										Description: "Defines the agent's prompt.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"prompt": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The system prompt for the agent.",
												},
												"llm": {
													Type:        schema.TypeString,
													Optional:    true,
													Default:     "gpt-4o-mini",
													Description: "The language model to use for the agent. Defaults to `gpt-4o-mini`.",
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

	conversationConfigList := d.Get("conversation_config").([]interface{})
	conversationConfigData := conversationConfigList[0].(map[string]interface{})
	agentConfigList := conversationConfigData["agent"].([]interface{})
	agentConfigData := agentConfigList[0].(map[string]interface{})
	promptConfigList := agentConfigData["prompt"].([]interface{})
	promptConfigData := promptConfigList[0].(map[string]interface{})

	agent := &Agent{
		Name: d.Get("name").(string),
		ConversationConfig: ConversationConfig{
			Agent: AgentConfig{
				FirstMessage: agentConfigData["first_message"].(string),
				Prompt: PromptConfig{
					Prompt: promptConfigData["prompt"].(string),
					LLM:    promptConfigData["llm"].(string),
				},
			},
		},
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

	conversationConfig := make([]map[string]interface{}, 1)
	agentConfig := make([]map[string]interface{}, 1)
	promptConfig := make([]map[string]interface{}, 1)

	promptConfig[0] = map[string]interface{}{
		"prompt": agent.ConversationConfig.Agent.Prompt.Prompt,
		"llm":    agent.ConversationConfig.Agent.Prompt.LLM,
	}

	agentConfig[0] = map[string]interface{}{
		"first_message": agent.ConversationConfig.Agent.FirstMessage,
		"prompt":        promptConfig,
	}

	conversationConfig[0] = map[string]interface{}{
		"agent": agentConfig,
	}

	if err := d.Set("conversation_config", conversationConfig); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAgentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	agentID := d.Id()

	if d.HasChange("name") || d.HasChange("conversation_config") {
		conversationConfigList := d.Get("conversation_config").([]interface{})
		conversationConfigData := conversationConfigList[0].(map[string]interface{})
		agentConfigList := conversationConfigData["agent"].([]interface{})
		agentConfigData := agentConfigList[0].(map[string]interface{})
		promptConfigList := agentConfigData["prompt"].([]interface{})
		promptConfigData := promptConfigList[0].(map[string]interface{})

		agent := &Agent{
			Name: d.Get("name").(string),
			ConversationConfig: ConversationConfig{
				Agent: AgentConfig{
					FirstMessage: agentConfigData["first_message"].(string),
					Prompt: PromptConfig{
						Prompt: promptConfigData["prompt"].(string),
						LLM:    promptConfigData["llm"].(string),
					},
				},
			},
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
