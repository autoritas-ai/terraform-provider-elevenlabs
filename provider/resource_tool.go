package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceToolCreate,
		ReadContext:   resourceToolRead,
		UpdateContext: resourceToolUpdate,
		DeleteContext: resourceToolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Provides a resource to manage a tool on ElevenLabs.",
		Schema: map[string]*schema.Schema{
			"tool_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the tool.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the tool.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A description of the tool.",
			},
			"api_schema": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "The API schema for the tool.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The URL of the API endpoint.",
						},
						"method": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "GET",
							Description: "The HTTP method to use for the API endpoint. Defaults to `GET`.",
						},
					},
				},
			},
		},
	}
}

func resourceToolCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	apiSchemaList := d.Get("api_schema").([]interface{})
	apiSchemaData := apiSchemaList[0].(map[string]interface{})

	tool := &Tool{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		APISchema: APISchema{
			URL:    apiSchemaData["url"].(string),
			Method: apiSchemaData["method"].(string),
		},
	}

	createdTool, err := client.CreateTool(ctx, tool)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(createdTool.ID)
	return resourceToolRead(ctx, d, m)
}

func resourceToolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	toolID := d.Id()

	tool, err := client.GetTool(ctx, toolID)
	if err != nil {
		return diag.FromErr(err)
	}

	if tool == nil {
		d.SetId("")
		return nil
	}

	d.Set("tool_id", tool.ID)
	d.Set("name", tool.ToolConfig.Name)
	d.Set("description", tool.ToolConfig.Description)

	apiSchema := make([]map[string]interface{}, 1)
	apiSchema[0] = map[string]interface{}{
		"url":    tool.ToolConfig.APISchema.URL,
		"method": tool.ToolConfig.APISchema.Method,
	}

	if err := d.Set("api_schema", apiSchema); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceToolUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	toolID := d.Id()

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("api_schema") {
		apiSchemaList := d.Get("api_schema").([]interface{})
		apiSchemaData := apiSchemaList[0].(map[string]interface{})

		tool := &Tool{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			APISchema: APISchema{
				URL:    apiSchemaData["url"].(string),
				Method: apiSchemaData["method"].(string),
			},
		}

		err := client.UpdateTool(ctx, toolID, tool)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceToolRead(ctx, d, m)
}

func resourceToolDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	toolID := d.Id()

	err := client.DeleteTool(ctx, toolID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
