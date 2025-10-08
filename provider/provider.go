package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ELEVENLABS_API_KEY", nil),
				Description: "The API key for ElevenLabs.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"elevenlabs_agent": resourceAgent(),
			"elevenlabs_tool":  resourceTool(),
		},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiKey := d.Get("api_key").(string)
	var diags diag.Diagnostics
	if apiKey == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "API key not found",
			Detail:   "API key for ElevenLabs is required.",
		})
		return nil, diags
	}
	return NewClient(apiKey), nil
}
