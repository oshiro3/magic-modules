package vertexai_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-google/google/acctest"
	"github.com/hashicorp/terraform-provider-google/google/tpgresource"
	transport_tpg "github.com/hashicorp/terraform-provider-google/google/transport"
)

func TestAccVertexAIEndpoint_vertexAiEndpointNetwork(t *testing.T) {
	t.Parallel()

	context := map[string]interface{}{
		"endpoint_name": fmt.Sprint(acctest.RandInt(t) % 9999999999),
		"kms_key_name":  acctest.BootstrapKMSKeyInLocation(t, "us-central1").CryptoKey.Name,
		"network_name":  acctest.BootstrapSharedServiceNetworkingConnection(t, "vertex-ai-endpoint-update-1"),
		"random_suffix": acctest.RandString(t, 10),
	}

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckVertexAIEndpointDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccVertexAIEndpoint_vertexAiEndpointNetwork(context),
			},
			{
				ResourceName:            "google_vertex_ai_endpoint.endpoint",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag", "location", "region", "labels", "terraform_labels"},
			},
			{
				Config: testAccVertexAIEndpoint_vertexAiEndpointNetworkUpdate(context),
			},
			{
				ResourceName:            "google_vertex_ai_endpoint.endpoint",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag", "location", "region", "labels", "terraform_labels"},
			},
		},
	})
}

func testAccVertexAIEndpoint_vertexAiEndpointNetwork(context map[string]interface{}) string {
	return acctest.Nprintf(`
resource "google_vertex_ai_endpoint" "endpoint" {
  name         = "%{endpoint_name}"
  display_name = "sample-endpoint"
  description  = "A sample vertex endpoint"
  location     = "us-central1"
  region       = "us-central1"
  labels       = {
    label-one = "value-one"
  }
  network      = "projects/${data.google_project.project.number}/global/networks/${data.google_compute_network.vertex_network.name}"
  encryption_spec {
    kms_key_name = "%{kms_key_name}"
  }
}

data "google_compute_network" "vertex_network" {
  name       = "%{network_name}"
}

resource "google_kms_crypto_key_iam_member" "crypto_key" {
  crypto_key_id = "%{kms_key_name}"
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:service-${data.google_project.project.number}@gcp-sa-aiplatform.iam.gserviceaccount.com"
}

data "google_project" "project" {}
`, context)
}

func testAccVertexAIEndpoint_vertexAiEndpointNetworkUpdate(context map[string]interface{}) string {
	return acctest.Nprintf(`
resource "google_vertex_ai_endpoint" "endpoint" {
  name         = "%{endpoint_name}"
  display_name = "new-sample-endpoint"
  description  = "An updated sample vertex endpoint"
  location     = "us-central1"
  region       = "us-central1"
  labels       = {
    label-two = "value-two"
  }
  network      = "projects/${data.google_project.project.number}/global/networks/${data.google_compute_network.vertex_network.name}"
  encryption_spec {
    kms_key_name = "%{kms_key_name}"
  }
}

data "google_compute_network" "vertex_network" {
  name       = "%{network_name}"
}

resource "google_kms_crypto_key_iam_member" "crypto_key" {
  crypto_key_id = "%{kms_key_name}"
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:service-${data.google_project.project.number}@gcp-sa-aiplatform.iam.gserviceaccount.com"
}

data "google_project" "project" {}
`, context)
}

func testAccCheckVertexAIEndpointDestroyProducer(t *testing.T) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for name, rs := range s.RootModule().Resources {
			if rs.Type != "google_vertex_ai_endpoint" {
				continue
			}
			if strings.HasPrefix(name, "data.") {
				continue
			}

			config := acctest.GoogleProviderConfig(t)

			url, err := tpgresource.ReplaceVarsForTest(config, rs, "{{VertexAIBasePath}}projects/{{project}}/locations/{{location}}/endpoints/{{name}}")
			if err != nil {
				return err
			}

			billingProject := ""

			if config.BillingProject != "" {
				billingProject = config.BillingProject
			}

			_, err = transport_tpg.SendRequest(transport_tpg.SendRequestOptions{
				Config:    config,
				Method:    "GET",
				Project:   billingProject,
				RawURL:    url,
				UserAgent: config.UserAgent,
			})
			if err == nil {
				return fmt.Errorf("VertexAIEndpoint still exists at %s", url)
			}
		}

		return nil
	}
}
