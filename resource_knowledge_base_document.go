package elevenlabs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"io"
	"os"
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_url": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"source_text", "source_file_path"},
			},
			"source_text": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"source_url", "source_file_path"},
			},
			"source_file_path": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"source_url", "source_text"},
			},
			"mime_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"source_hash": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceKnowledgeBaseDocumentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	name := d.Get("name").(string)
	mimeType := d.Get("mime_type").(string)

	var resp *KnowledgeBaseDocumentCreateResponse
	var err error
	var sourceHash string

	if sourceURL, ok := d.GetOk("source_url"); ok {
		doc := &KnowledgeBaseDocumentFromURLRequest{
			URL:      sourceURL.(string),
			Name:     name,
			MimeType: mimeType,
		}
		resp, err = client.CreateKnowledgeBaseDocumentFromURL(ctx, doc)
	} else if sourceText, ok := d.GetOk("source_text"); ok {
		text := sourceText.(string)
		hash := sha256.Sum256([]byte(text))
		sourceHash = hex.EncodeToString(hash[:])
		doc := &KnowledgeBaseDocumentFromTextRequest{
			Text:     text,
			Name:     name,
			MimeType: mimeType,
		}
		resp, err = client.CreateKnowledgeBaseDocumentFromText(ctx, doc)
	} else if sourceFilePath, ok := d.GetOk("source_file_path"); ok {
		path := sourceFilePath.(string)
		file, err_open := os.Open(path)
		if err_open != nil {
			return diag.FromErr(fmt.Errorf("error opening source file: %w", err_open))
		}
		defer file.Close()

		hash := sha256.New()
		if _, err_copy := io.Copy(hash, file); err_copy != nil {
			return diag.FromErr(fmt.Errorf("error hashing source file: %w", err_copy))
		}
		sourceHash = hex.EncodeToString(hash.Sum(nil))

		resp, err = client.CreateKnowledgeBaseDocumentFromFile(ctx, name, mimeType, path)
	} else {
		return diag.Errorf("one of source_url, source_text, or source_file_path must be specified")
	}

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.DocumentID)
	if sourceHash != "" {
		d.Set("source_hash", sourceHash)
	}

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
	d.Set("mime_type", doc.MimeType)

	// We don't get these back from the API, so we rely on the state
	// The hash will ensure that if the local content changes, the resource is recreated.
	if doc.SourceType == "url" {
		d.Set("source_url", doc.SourceURL)
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