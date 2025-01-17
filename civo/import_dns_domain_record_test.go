package civo

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCivoDNSDomainRecord_importBasic(t *testing.T) {
	resourceName := "civo_dns_domain_record.www"
	var domainName = acctest.RandomWithPrefix("tf-test-record") + ".example"
	var recordName = acctest.RandomWithPrefix("record")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCivoDNSDomainNameRecordDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckCivoDNSDomainNameRecordConfigBasic(domainName, recordName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccDNSDomainNameRecordImportID(resourceName),
			},
		},
	})
}

func testAccDNSDomainNameRecordImportID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		domainID := rs.Primary.Attributes["domain_id"]
		id := rs.Primary.ID

		return fmt.Sprintf("%s:%s", domainID, id), nil
	}
}
