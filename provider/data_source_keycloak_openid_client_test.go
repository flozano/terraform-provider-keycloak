package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakDataSourceOpenidClient_basic(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.keycloak_openid_client.test"
	resourceName := "keycloak_openid_client.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakOpenidClientConfig(clientId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "client_id", resourceName, "client_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "realm_id", resourceName, "realm_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enabled", resourceName, "enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "access_type", resourceName, "access_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "standard_flow_enabled", resourceName, "standard_flow_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "implicit_flow_enabled", resourceName, "implicit_flow_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "direct_access_grants_enabled", resourceName, "direct_access_grants_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "service_account_user_id", resourceName, "service_account_user_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "service_accounts_enabled", resourceName, "service_accounts_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "resource_server_id", resourceName, "resource_server_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "full_scope_allowed", resourceName, "full_scope_allowed"),
					resource.TestCheckResourceAttrPair(dataSourceName, "consent_required", resourceName, "consent_required"),
					resource.TestCheckResourceAttrPair(dataSourceName, "consent_screen_text", resourceName, "consent_screen_text"),
					resource.TestCheckResourceAttrPair(dataSourceName, "display_on_consent_screen", resourceName, "display_on_consent_screen"),
				),
			},
		},
	})
}

func TestAccKeycloakDataSourceOpenidClient_adminPermissions(t *testing.T) {
	skipIfVersionIsLessThan(testCtx, t, keycloakClient, keycloak.Version_26_2)
	skipIfFGAPv2NotEnabled(testCtx, t, keycloakClient)

	dataSourceName := "data.keycloak_openid_client.admin_permissions"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakOpenidClientConfig_adminPermissions(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeycloakDataSourceOpenidClientIsAdminPermissionsClient(dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "client_id", "admin-permissions"),
					resource.TestCheckResourceAttr(dataSourceName, "client_secret", ""),
				),
			},
		},
	})
}

func testAccCheckKeycloakDataSourceOpenidClientIsAdminPermissionsClient(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found: %s", dataSourceName)
		}

		adminPermissionsClientId, err := keycloakClient.GetAdminPermissionsClientId(testCtx, rs.Primary.Attributes["realm_id"])
		if err != nil {
			return fmt.Errorf("error getting admin-permissions client id: %s", err)
		}
		if rs.Primary.ID != adminPermissionsClientId {
			return fmt.Errorf("data source id %s was not equal to %s (the id of the admin-permissions client)", rs.Primary.ID, adminPermissionsClientId)
		}

		return nil
	}
}

func TestAccKeycloakDataSourceOpenidClient_extraConfig(t *testing.T) {
	t.Parallel()
	clientId := acctest.RandomWithPrefix("tf-acc-test-extra-config")
	dataSourceName := "data.keycloak_openid_client.test_extra_config"
	resourceName := "keycloak_openid_client.test_extra_config"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeycloakOpenidClientConfig_extraConfig(clientId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "extra_config.key1", resourceName, "extra_config.key1"),
				),
			},
		},
	})
}

func testAccKeycloakOpenidClientConfig(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "test" {
	name                     	= "%s"
	client_id                	= "%s"
	realm_id                 	= data.keycloak_realm.realm.id
	description              	= "a test openid client"
	standard_flow_enabled    	= true
	access_type              	= "CONFIDENTIAL"
	service_accounts_enabled 	= true
	client_secret            	= "secret"
	valid_redirect_uris      	= [
		"http://localhost:5555/callback",
	]
	authorization {
		policy_enforcement_mode = "ENFORCING"
	}
	web_origins              	= [
		"http://localhost"
	]
	full_scope_allowed       	= false
	consent_required         	= true
	display_on_consent_screen	= true
	consent_screen_text      	= "some consent screen text"
}

data "keycloak_openid_client" "test" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.test.client_id

	depends_on = [
		keycloak_openid_client.test,
	]
}
`, testAccRealm.Realm, clientId, clientId)
}

func testAccKeycloakOpenidClientConfig_extraConfig(clientId string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "test_extra_config" {
	name                     = "%s"
	client_id                = "%s"
	realm_id                 = data.keycloak_realm.realm.id
	description              = "a test openid client with extra_conf"
	access_type              = "CONFIDENTIAL"
	extra_config             = {
		"key1"				 = "value1"
	}
}

data "keycloak_openid_client" "test_extra_config" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = keycloak_openid_client.test_extra_config.client_id

	depends_on = [
		keycloak_openid_client.test_extra_config,
	]
}
`, testAccRealm.Realm, clientId, clientId)
}

func testAccKeycloakOpenidClientConfig_adminPermissions() string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

data "keycloak_openid_client" "admin_permissions" {
	realm_id  = data.keycloak_realm.realm.id
	client_id = "admin-permissions"
}
`, testAccRealmFGAPv2.Realm)
}
