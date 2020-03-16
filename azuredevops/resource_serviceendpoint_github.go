package azuredevops

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/microsoft/azure-devops-go-api/azuredevops/serviceendpoint"
	crud "github.com/microsoft/terraform-provider-azuredevops/azuredevops/crud/serviceendpoint"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/converter"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/tfhelper"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/validate"
)

const (
	personalAccessToken = "personal_access_token"
)

func resourceServiceEndpointGitHub() *schema.Resource {
	r := crud.GenBaseServiceEndpointResource(flattenServiceEndpointGitHub, expandServiceEndpointGitHub, parseImportedProjectIDAndServiceEndpointID)
	authPersonal := &schema.Resource{
		Schema: map[string]*schema.Schema{
			personalAccessToken: {
				Type:             schema.TypeString,
				Required:         true,
				DefaultFunc:      schema.EnvDefaultFunc("AZDO_GITHUB_SERVICE_CONNECTION_PAT", nil),
				Description:      "The GitHub personal access token which should be used.",
				Sensitive:        true,
				ValidateFunc:     validate.NoEmptyStrings,
				DiffSuppressFunc: tfhelper.DiffFuncSuppressSecretChanged,
			},
		},
	}
	patHashKey, patHashSchema := tfhelper.GenerateSecreteMemoSchema(personalAccessToken)
	authPersonal.Schema[patHashKey] = patHashSchema
	r.Schema["auth_personal"] = &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MinItems:      1,
		MaxItems:      1,
		Elem:          authPersonal,
		ConflictsWith: []string{"auth_oath"},
	}

	r.Schema["auth_oath"] = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"oauth_configuration_id": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Configuration ID for OAuth",
				},
			},
		},
		ConflictsWith: []string{"auth_personal"},
	}

	return r
}

func expandAuthOauth(d map[string]interface{}) map[string]string {
	return map[string]string{
		"ConfigurationId": d["oauth_configuration_id"].(string),
	}
}
func expandAuthOauthList(d []interface{}) []map[string]string {
	vs := make([]map[string]string, 0, len(d))
	for _, v := range d {
		val, ok := v.(map[string]interface{})
		if ok {
			vs = append(vs, expandAuthOauth(val))
		}
	}
	return vs
}
func expandAuthOauthSet(authOauth []interface{}) map[string]string {
	d2 := expandAuthOauthList(authOauth)
	if len(d2) != 1 {
		return nil
	}
	return d2[0]
}

func expandAuthPersonal(d map[string]interface{}) map[string]string {
	return map[string]string{
		"accessToken": d[personalAccessToken].(string),
	}
}
func expandAuthPersonalList(d []interface{}) []map[string]string {
	vs := make([]map[string]string, 0, len(d))
	for _, v := range d {
		val, ok := v.(map[string]interface{})
		if ok {
			vs = append(vs, expandAuthPersonal(val))
		}
	}
	return vs
}
func expandAuthPersonalSet(authPersonal []interface{}) map[string]string {
	d2 := expandAuthPersonalList(authPersonal)
	if len(d2) != 1 {
		return nil
	}
	return d2[0]
}

// Convert internal Terraform data structure to an AzDO data structure
func expandServiceEndpointGitHub(d *schema.ResourceData) (*serviceendpoint.ServiceEndpoint, *string) {
	serviceEndpoint, projectID := crud.DoBaseExpansion(d)
	scheme := "InstallationToken"

	parameters := &map[string]string{}
	authPersonal := expandAuthPersonalSet(d.Get("auth_personal").([]interface{}))
	authGrant := expandAuthOauthSet(d.Get("auth_oath").([]interface{}))

	if authPersonal != nil {
		scheme = "PersonalAccessToken"
		parameters = &authPersonal
	}

	if authGrant != nil {
		scheme = "OAuth"
		parameters = &authGrant
	}

	serviceEndpoint.Authorization = &serviceendpoint.EndpointAuthorization{
		Parameters: parameters,
		Scheme:     &scheme,
	}
	serviceEndpoint.Type = converter.String("github")
	serviceEndpoint.Url = converter.String("http://github.com")

	return serviceEndpoint, projectID
}

// Convert AzDO data structure to internal Terraform data structure
func flattenServiceEndpointGitHub(d *schema.ResourceData, serviceEndpoint *serviceendpoint.ServiceEndpoint, projectID *string) error {
	err := crud.DoBaseFlattening(d, serviceEndpoint, projectID)
	if err != nil {
		return err
	}
	if *serviceEndpoint.Authorization.Scheme == "OAuth" {
		err = d.Set("auth_oath", []map[string]interface{}{
			{
				"oauth_configuration_id": (*serviceEndpoint.Authorization.Parameters)["ConfigurationId"],
			},
		})
		if err != nil {
			return err
		}
	}
	if *serviceEndpoint.Authorization.Scheme == "PersonalAccessToken" {
		authPersonalList := d.Get("auth_personal").([]interface{})
		if len(authPersonalList) == 1 {
			params := authPersonalList[0].(map[string]interface{})
			hash, key := tfhelper.HelpFlattenSecretNested(d, "auth_personal", params, personalAccessToken)
			params[key] = hash

			d.Set("auth_personal", authPersonalList)
		}
	}
	return nil
}
