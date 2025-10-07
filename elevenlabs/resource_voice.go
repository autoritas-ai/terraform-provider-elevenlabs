package elevenlabs

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVoice() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVoiceCreate,
		ReadContext:   resourceVoiceRead,
		DeleteContext: resourceVoiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"voice_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"files": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceVoiceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	labelsData := d.Get("labels").(map[string]interface{})
	filesData := d.Get("files").(*schema.Set).List()

	labels := make(map[string]string)
	for k, v := range labelsData {
		labels[k] = v.(string)
	}

	filePaths := make([]string, len(filesData))
	for i, v := range filesData {
		filePaths[i] = v.(string)
	}

	resp, err := client.CreateVoice(ctx, name, description, labels, filePaths)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.VoiceID)
	return resourceVoiceRead(ctx, d, m)
}

func resourceVoiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	voiceID := d.Id()

	voice, err := client.GetVoice(ctx, voiceID)
	if err != nil {
		return diag.FromErr(err)
	}

	if voice == nil {
		d.SetId("")
		return nil
	}

	d.Set("voice_id", voice.VoiceID)
	d.Set("name", voice.Name)
	d.Set("description", voice.Description)
	d.Set("labels", voice.Labels)

	return nil
}

func resourceVoiceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	voiceID := d.Id()

	err := client.DeleteVoice(ctx, voiceID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}