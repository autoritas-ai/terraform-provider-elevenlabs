package provider

import (
	"context"
	"encoding/json"
	"fmt"

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
		Schema: map[string]*schema.Schema{
			"tool_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "webhook",
			},
			"response_timeout_secs": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"disable_interruptions": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"force_pre_tool_speech": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"api_schema": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"method": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "GET",
						},
						"path_params_schema": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"description": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"query_params_schema": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"properties": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"type": {
													Type:     schema.TypeString,
													Required: true,
												},
												"description": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"required": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"request_body_schema": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The schema for the request body, specified as a JSON string. It is recommended to use the `jsonencode` function to construct this value.",
							ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
								s, ok := v.(string)
								if !ok {
									es = append(es, fmt.Errorf("expected type of %s to be string", k))
									return
								}
								if s == "" {
									return
								}
								var js interface{}
								if err := json.Unmarshal([]byte(s), &js); err != nil {
									es = append(es, fmt.Errorf("%q contains an invalid JSON: %s", k, err))
								}
								return
							},
						},
						"request_headers": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func resourceToolCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	var apiSchema *APISchema
	if apiSchemaList, ok := d.Get("api_schema").([]interface{}); ok && len(apiSchemaList) > 0 {
		apiSchemaData := apiSchemaList[0].(map[string]interface{})
		apiSchema = &APISchema{
			URL:    apiSchemaData["url"].(string),
			Method: apiSchemaData["method"].(string),
		}

		if v, ok := apiSchemaData["path_params_schema"].([]interface{}); ok && len(v) > 0 {
			apiSchema.PathParamsSchema = make(map[string]LiteralJsonSchemaProperty)
			for _, item := range v {
				param := item.(map[string]interface{})
				name := param["name"].(string)
				apiSchema.PathParamsSchema[name] = LiteralJsonSchemaProperty{
					Type:        param["type"].(string),
					Description: param["description"].(string),
				}
			}
		}

		if v, ok := apiSchemaData["query_params_schema"].([]interface{}); ok && len(v) > 0 {
			queryData := v[0].(map[string]interface{})
			queryParamsSchema := &QueryParamsJsonSchema{
				Properties: make(map[string]LiteralJsonSchemaProperty),
			}
			if props, ok := queryData["properties"].([]interface{}); ok {
				for _, item := range props {
					prop := item.(map[string]interface{})
					name := prop["name"].(string)
					queryParamsSchema.Properties[name] = LiteralJsonSchemaProperty{
						Type:        prop["type"].(string),
						Description: prop["description"].(string),
					}
				}
			}
			if req, ok := queryData["required"].([]interface{}); ok {
				for _, r := range req {
					queryParamsSchema.Required = append(queryParamsSchema.Required, r.(string))
				}
			}
			apiSchema.QueryParamsSchema = queryParamsSchema
		}

		if v, ok := apiSchemaData["request_headers"].(map[string]interface{}); ok && len(v) > 0 {
			apiSchema.RequestHeaders = make(map[string]string)
			for key, val := range v {
				apiSchema.RequestHeaders[key] = val.(string)
			}
		}

		if v, ok := apiSchemaData["request_body_schema"].(string); ok && v != "" {
			apiSchema.RequestBodySchema = json.RawMessage(v)
		}
	}

	tool := &Tool{
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		Type:                 d.Get("type").(string),
		ResponseTimeoutSecs:  d.Get("response_timeout_secs").(int),
		DisableInterruptions: d.Get("disable_interruptions").(bool),
		ForcePreToolSpeech:   d.Get("force_pre_tool_speech").(bool),
		APISchema:            apiSchema,
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
	d.Set("type", tool.ToolConfig.Type)
	d.Set("response_timeout_secs", tool.ToolConfig.ResponseTimeoutSecs)
	d.Set("disable_interruptions", tool.ToolConfig.DisableInterruptions)
	d.Set("force_pre_tool_speech", tool.ToolConfig.ForcePreToolSpeech)

	if tool.ToolConfig.APISchema != nil {
		apiSchema := make(map[string]interface{})
		apiSchema["url"] = tool.ToolConfig.APISchema.URL
		apiSchema["method"] = tool.ToolConfig.APISchema.Method

		if tool.ToolConfig.APISchema.PathParamsSchema != nil {
			pathParamsList := make([]interface{}, 0, len(tool.ToolConfig.APISchema.PathParamsSchema))
			for k, v := range tool.ToolConfig.APISchema.PathParamsSchema {
				param := make(map[string]interface{})
				param["name"] = k
				param["type"] = v.Type
				param["description"] = v.Description
				pathParamsList = append(pathParamsList, param)
			}
			apiSchema["path_params_schema"] = pathParamsList
		}

		if tool.ToolConfig.APISchema.QueryParamsSchema != nil {
			queryParams := make(map[string]interface{})
			propsList := make([]interface{}, 0, len(tool.ToolConfig.APISchema.QueryParamsSchema.Properties))
			for k, v := range tool.ToolConfig.APISchema.QueryParamsSchema.Properties {
				prop := make(map[string]interface{})
				prop["name"] = k
				prop["type"] = v.Type
				prop["description"] = v.Description
				propsList = append(propsList, prop)
			}
			queryParams["properties"] = propsList
			queryParams["required"] = tool.ToolConfig.APISchema.QueryParamsSchema.Required
			apiSchema["query_params_schema"] = []interface{}{queryParams}
		}

		if tool.ToolConfig.APISchema.RequestHeaders != nil {
			apiSchema["request_headers"] = tool.ToolConfig.APISchema.RequestHeaders
		}

		if tool.ToolConfig.APISchema.RequestBodySchema != nil {
			requestBodySchema, err := json.Marshal(tool.ToolConfig.APISchema.RequestBodySchema)
			if err != nil {
				return diag.FromErr(err)
			}
			// Unmarshal and then re-marshal to get a compact, canonical JSON string
			var v interface{}
			if err := json.Unmarshal(requestBodySchema, &v); err != nil {
				return diag.FromErr(err)
			}
			compactBody, err := json.Marshal(v)
			if err != nil {
				return diag.FromErr(err)
			}
			apiSchema["request_body_schema"] = string(compactBody)
		}

		if err := d.Set("api_schema", []interface{}{apiSchema}); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceToolUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	toolID := d.Id()

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("type") || d.HasChange("response_timeout_secs") || d.HasChange("disable_interruptions") || d.HasChange("force_pre_tool_speech") || d.HasChange("api_schema") {
		var apiSchema *APISchema
		if apiSchemaList, ok := d.Get("api_schema").([]interface{}); ok && len(apiSchemaList) > 0 {
			apiSchemaData := apiSchemaList[0].(map[string]interface{})
			apiSchema = &APISchema{
				URL:    apiSchemaData["url"].(string),
				Method: apiSchemaData["method"].(string),
			}

			if v, ok := apiSchemaData["path_params_schema"].([]interface{}); ok && len(v) > 0 {
				apiSchema.PathParamsSchema = make(map[string]LiteralJsonSchemaProperty)
				for _, item := range v {
					param := item.(map[string]interface{})
					name := param["name"].(string)
					apiSchema.PathParamsSchema[name] = LiteralJsonSchemaProperty{
						Type:        param["type"].(string),
						Description: param["description"].(string),
					}
				}
			}

			if v, ok := apiSchemaData["query_params_schema"].([]interface{}); ok && len(v) > 0 {
				queryData := v[0].(map[string]interface{})
				queryParamsSchema := &QueryParamsJsonSchema{
					Properties: make(map[string]LiteralJsonSchemaProperty),
				}
				if props, ok := queryData["properties"].([]interface{}); ok {
					for _, item := range props {
						prop := item.(map[string]interface{})
						name := prop["name"].(string)
						queryParamsSchema.Properties[name] = LiteralJsonSchemaProperty{
							Type:        prop["type"].(string),
							Description: prop["description"].(string),
						}
					}
				}
				if req, ok := queryData["required"].([]interface{}); ok {
					for _, r := range req {
						queryParamsSchema.Required = append(queryParamsSchema.Required, r.(string))
					}
				}
				apiSchema.QueryParamsSchema = queryParamsSchema
			}

			if v, ok := apiSchemaData["request_headers"].(map[string]interface{}); ok && len(v) > 0 {
				apiSchema.RequestHeaders = make(map[string]string)
				for key, val := range v {
					apiSchema.RequestHeaders[key] = val.(string)
				}
			}

		if v, ok := apiSchemaData["request_body_schema"].(string); ok && v != "" {
			apiSchema.RequestBodySchema = json.RawMessage(v)
		}
		}

		tool := &Tool{
			Name:                 d.Get("name").(string),
			Description:          d.Get("description").(string),
			Type:                 d.Get("type").(string),
			ResponseTimeoutSecs:  d.Get("response_timeout_secs").(int),
			DisableInterruptions: d.Get("disable_interruptions").(bool),
			ForcePreToolSpeech:   d.Get("force_pre_tool_speech").(bool),
			APISchema:            apiSchema,
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
