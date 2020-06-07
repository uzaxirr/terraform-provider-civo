package civo

import (
	"fmt"
	"log"

	"github.com/civo/civogo"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

// Constant all possible DNS Record
const (
	// DNSRecordTypeA represents an A record
	DNSRecordTypeA = "a"

	// DNSRecordTypeCName represents an CNAME record
	DNSRecordTypeCName = "cname"

	// DNSRecordTypeMX represents an MX record
	DNSRecordTypeMX = "mx"

	// DNSRecordTypeTXT represents an TXT record
	DNSRecordTypeTXT = "txt"
)

// DNS domain record resource with this we can create and manage DNS Domain
func resourceDNSDomainRecord() *schema.Resource {
	fmt.Print()
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"domain_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Id from domain name",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The choice of RR type from a, cname, mx or txt",
				ValidateFunc: validation.StringInSlice([]string{
					DNSRecordTypeA,
					DNSRecordTypeCName,
					DNSRecordTypeMX,
					DNSRecordTypeTXT,
				}, false),
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The portion before the domain name (e.g. www) or an @ for the apex/root domain (you cannot use an A record with an amex/root domain)",
			},
			"value": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The IP address (A or MX), hostname (CNAME or MX) or text value (TXT) to serve for this record",
				ValidateFunc: validation.NoZeroValues,
			},
			"priority": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.NoZeroValues,
				Description:  "Useful for MX records only, the priority mail should be attempted it (defaults to 10)",
			},
			"ttl": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
				Description:  "How long caching DNS servers should cache this record for, in seconds (the minimum is 600 and the default if unspecified is 600)",
			},
			// Computed resource
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Create: resourceDNSDomainRecordCreate,
		Read:   resourceDNSDomainRecordRead,
		Update: resourceDNSDomainRecordUpdate,
		Delete: resourceDNSDomainRecordDelete,
		//Exists: resourceExistsItem,
		Importer: &schema.ResourceImporter{
			State: resourceDNSDomainRecordImport,
		},
	}
}

// function to create a new record for the main domain
func resourceDNSDomainRecordCreate(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*civogo.Client)

	log.Printf("[INFO] configuring the domain record %s", d.Get("name").(string))
	config := &civogo.DNSRecordConfig{
		Name:  d.Get("name").(string),
		Value: d.Get("value").(string),
		TTL:   d.Get("ttl").(int),
	}

	if attr, ok := d.GetOk("priority"); ok {
		if d.Get("type").(string) != "mx" {
			return fmt.Errorf("[WARN] warning priority value is only allow in the MX records")
		}
		config.Priority = attr.(int)
	}

	if d.Get("type").(string) == "a" {
		config.Type = DNSRecordTypeA
	}

	if d.Get("type").(string) == "cname" {
		config.Type = DNSRecordTypeCName
	}

	if d.Get("type").(string) == "mx" {
		config.Type = DNSRecordTypeMX
	}

	if d.Get("type").(string) == "txt" {
		config.Type = DNSRecordTypeTXT
	}

	log.Printf("[INFO] Creating the domain record %s", d.Get("name").(string))
	dnsDomainRecord, err := apiClient.CreateDNSRecord(d.Get("domain_id").(string), config)
	if err != nil {
		return fmt.Errorf("[ERR] failed to create a new domain record: %s", err)
	}

	d.SetId(dnsDomainRecord.ID)

	return resourceDNSDomainRecordRead(d, m)
}

// function to read a dns domain record
func resourceDNSDomainRecordRead(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*civogo.Client)

	log.Printf("[INFO] retriving the domain record %s", d.Get("name").(string))
	resp, err := apiClient.GetDNSRecord(d.Get("domain_id").(string), d.Id())
	if err != nil {
		if resp == nil {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("[WARN] domain record (%s) not found", d.Id())
	}

	d.Set("name", resp.Name)
	d.Set("account_id", resp.AccountID)
	d.Set("domain_id", resp.DNSDomainID)
	d.Set("name", resp.Name)
	d.Set("value", resp.Value)
	d.Set("type", resp.Type)
	d.Set("priority", resp.Priority)
	d.Set("ttl", resp.TTL)
	d.Set("created_at", resp.CreatedAt.UTC().String())
	d.Set("updated_at", resp.UpdatedAt.UTC().String())

	return nil
}

// function to update a dns domain record
func resourceDNSDomainRecordUpdate(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*civogo.Client)

	resp, err := apiClient.GetDNSRecord(d.Get("domain_id").(string), d.Id())
	if err != nil {
		return fmt.Errorf("[WARN] domain record (%s) not found", d.Id())
	}

	config := &civogo.DNSRecordConfig{}

	if d.HasChange("name") || d.HasChange("value") || d.HasChange("priority") || d.HasChange("ttl") || d.HasChange("type") {
		config.Name = d.Get("name").(string)
		config.Value = d.Get("value").(string)
		config.Priority = d.Get("priority").(int)
		config.TTL = d.Get("ttl").(int)

		if d.Get("type").(string) == "a" {
			config.Type = DNSRecordTypeA
		}

		if d.Get("type").(string) == "cname" {
			config.Type = DNSRecordTypeCName
		}

		if d.Get("type").(string) == "mx" {
			config.Type = DNSRecordTypeMX
		}

		if d.Get("type").(string) == "txt" {
			config.Type = DNSRecordTypeTXT
		}
	}

	log.Printf("[INFO] Updating the domain record %s", d.Get("name").(string))
	_, err = apiClient.UpdateDNSRecord(resp, config)
	if err != nil {
		return fmt.Errorf("[ERR] an error occurred while renamed the domain record %s, %s", d.Id(), err)
	}

	return resourceDNSDomainRecordRead(d, m)
}

//function to delete a dns domain record
func resourceDNSDomainRecordDelete(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*civogo.Client)

	log.Printf("[INFO] Searching the domain record %s", d.Get("name").(string))
	resp, err := apiClient.GetDNSRecord(d.Get("domain_id").(string), d.Id())
	if err != nil {
		return fmt.Errorf("[WARN] domain record (%s) not found", d.Id())
	}

	log.Printf("[INFO] deleting the domain record %s", d.Get("name").(string))
	_, err = apiClient.DeleteDNSRecord(resp)
	if err != nil {
		return fmt.Errorf("[WARN] an error occurred while tring to delete the domain record %s", d.Id())
	}

	return nil
}

// custom import to able to add a main domain to the terraform
func resourceDNSDomainRecordImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	apiClient := m.(*civogo.Client)

	domainID, DomainRecordID, err := resourceCommonParseID(d.Id())
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] retriving the domain record %s", DomainRecordID)
	resp, err := apiClient.GetDNSRecord(domainID, DomainRecordID)
	if err != nil {
		if resp != nil {
			return nil, err
		}
	}

	d.SetId(resp.ID)
	d.Set("name", resp.Name)
	d.Set("account_id", resp.AccountID)
	d.Set("domain_id", resp.DNSDomainID)
	d.Set("name", resp.Name)
	d.Set("value", resp.Value)
	d.Set("type", resp.Type)
	d.Set("priority", resp.Priority)
	d.Set("ttl", resp.TTL)
	d.Set("created_at", resp.CreatedAt.UTC().String())
	d.Set("updated_at", resp.UpdatedAt.UTC().String())

	return []*schema.ResourceData{d}, nil
}
