package elevenlabs

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"hash/fnv"
	"strconv"
)

func dataSourceVoices() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVoicesRead,
		Schema: map[string]*schema.Schema{
			"search": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sort": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sort_direction": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"voice_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"category": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"voices": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"voice_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"labels": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourceVoicesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	params := &SearchVoicesParams{
		Search:        d.Get("search").(string),
		Sort:          d.Get("sort").(string),
		SortDirection: d.Get("sort_direction").(string),
		VoiceType:     d.Get("voice_type").(string),
		Category:      d.Get("category").(string),
	}

	resp, err := client.SearchVoices(ctx, params)
	if err != nil {
		return diag.FromErr(err)
	}

	voices := make([]map[string]interface{}, len(resp.Voices))
	for i, v := range resp.Voices {
		voice := make(map[string]interface{})
		voice["voice_id"] = v.VoiceID
		voice["name"] = v.Name
		voice["description"] = v.Description
		voice["labels"] = v.Labels
		voices[i] = voice
	}

	if err := d.Set("voices", voices); err != nil {
		return diag.FromErr(err)
	}

	// Create a unique ID for the data source based on the search parameters
	h := fnv.New32a()
	h.Write([]byte(fmt.Sprintf("%v", params)))
	d.SetId(strconv.FormatUint(uint64(h.Sum32()), 10))

	return nil
}