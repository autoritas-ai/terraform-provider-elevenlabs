package elevenlabs

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVoiceSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVoiceSettingsCreateOrUpdate,
		ReadContext:   resourceVoiceSettingsRead,
		UpdateContext: resourceVoiceSettingsCreateOrUpdate,
		DeleteContext: resourceVoiceSettingsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"voice_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"stability": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"similarity_boost": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"style": {
				Type:     schema.TypeFloat,
				Optional: true,
				Default:  0.0,
			},
			"use_speaker_boost": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"speed": {
				Type:     schema.TypeFloat,
				Optional: true,
				Default:  1.0,
			},
		},
	}
}

func resourceVoiceSettingsCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	voiceID := d.Get("voice_id").(string)

	settings := &VoiceSettings{
		Stability:       d.Get("stability").(float64),
		SimilarityBoost: d.Get("similarity_boost").(float64),
		Style:           d.Get("style").(float64),
		UseSpeakerBoost: d.Get("use_speaker_boost").(bool),
		Speed:           d.Get("speed").(float64),
	}

	err := client.UpdateVoiceSettings(ctx, voiceID, settings)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(voiceID)
	return resourceVoiceSettingsRead(ctx, d, m)
}

func resourceVoiceSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	voiceID := d.Id()

	settings, err := client.GetVoiceSettings(ctx, voiceID)
	if err != nil {
		return diag.FromErr(err)
	}

	if settings == nil {
		d.SetId("")
		return nil
	}

	d.Set("voice_id", voiceID)
	d.Set("stability", settings.Stability)
	d.Set("similarity_boost", settings.SimilarityBoost)
	d.Set("style", settings.Style)
	d.Set("use_speaker_boost", settings.UseSpeakerBoost)
	d.Set("speed", settings.Speed)

	return nil
}

func resourceVoiceSettingsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	voiceID := d.Get("voice_id").(string)

	// Reset to default settings
	defaultSettings := &VoiceSettings{
		Stability:       0.5,
		SimilarityBoost: 0.75,
		Style:           0.0,
		UseSpeakerBoost: true,
		Speed:           1.0,
	}

	err := client.UpdateVoiceSettings(ctx, voiceID, defaultSettings)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}