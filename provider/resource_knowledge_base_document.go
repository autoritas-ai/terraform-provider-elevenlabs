package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceKnowledgeBaseDocument() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKnowledgeBaseDocumentCreate,
		ReadContext:   resourceKnowledgeBaseDocumentRead,
		DeleteContext: resourceKnowledgeBaseDocumentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"document_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the document.",
			},
			"file_path": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				ConflictsWith: []string{"url", "text_content"},
				Description: "The path to the file to upload.",
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				ConflictsWith: []string{"file_path", "text_content"},
				Description: "The URL of the document to import.",
			},
			"text_content": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				ConflictsWith: []string{"file_path", "url"},
				Description: "The text content of the document.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of the document.",
			},
			"size_bytes": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The size of the document in bytes.",
			},
			"created_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp of when the document was created.",
			},
			"last_updated_at": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The timestamp of when the document was last updated.",
			},
		},
	}
}

func resourceKnowledgeBaseDocumentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	name := d.Get("name").(string)

	var resp *AddKnowledgeBaseResponse
	var err error

	if filePath, ok := d.GetOk("file_path"); ok {
		resp, err = client.CreateKnowledgeBaseDocumentFromFile(ctx, name, filePath.(string))
	} else if url, ok := d.GetOk("url"); ok {
		resp, err = client.CreateKnowledgeBaseDocumentFromURL(ctx, name, url.(string))
	} else if textContent, ok := d.GetOk("text_content"); ok {
		resp, err = client.CreateKnowledgeBaseDocumentFromText(ctx, name, textContent.(string))
	} else {
		return diag.Errorf("one of file_path, url, or text_content must be specified")
	}

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)
	return resourceKnowledgeBaseDocumentRead(ctx, d, m)
}

func resourceKnowledgeBaseDocumentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	documentID := d.Id()

	doc, err := client.GetKnowledgeBaseDocument(ctx, documentID)
	if err != nil {
		return diag.FromErr(err)
	}

	if doc == nil {
		d.SetId("")
		return nil
	}

	d.Set("document_id", doc.ID)
	d.Set("name", doc.Name)
	d.Set("type", doc.Type)
	if doc.Metadata != nil {
		d.Set("size_bytes", doc.Metadata.SizeBytes)
		d.Set("created_at", doc.Metadata.CreatedAt)
		d.Set("last_updated_at", doc.Metadata.LastUpdatedAt)
	}
	if doc.Type == "url" {
		d.Set("url", doc.URL)
	}

	return nil
}

func resourceKnowledgeBaseDocumentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	documentID := d.Id()

	err := client.DeleteKnowledgeBaseDocument(ctx, documentID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}