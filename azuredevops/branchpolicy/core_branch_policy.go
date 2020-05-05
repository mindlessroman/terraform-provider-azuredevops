package branchpolicy

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/microsoft/azure-devops-go-api/azuredevops/policy"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/config"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/converter"
)

// Policy type IDs. These are global and can be listed using the following endpoint:
//	https://docs.microsoft.com/en-us/rest/api/azure/devops/policy/types/list?view=azure-devops-rest-5.1
var (
	noActiveComments = uuid.MustParse("c6a1889d-b943-4856-b76f-9e46bb6b0df2")
	minReviewerCount = uuid.MustParse("fa4e907d-c16b-4a4c-9dfa-4906e5d171dd")
	successfulBuild  = uuid.MustParse("0609b952-1397-4640-95ec-e00a01b2c241")
)

// Keys for schema elements
const (
	schemaProjectID     = "project_id"
	schemaEnabled       = "enabled"
	schemaBlocking      = "blocking"
	schemaSettings      = "settings"
	schemaScope         = "scope"
	schemaRepositoryID  = "repository_id"
	schemaRepositoryRef = "repository_ref"
	schemaMatchType     = "match_type"
)

// The type of repository branch name matching strategy used by the policy
const (
	matchTypeExact  string = "Exact"
	matchTypePrefix string = "Prefix"
)

// policyCrudArgs arguments for genBasePolicyResource
type policyCrudArgs struct {
	flattenFunc func(d *schema.ResourceData, policy *policy.PolicyConfiguration, projectID *string) error
	expandFunc  func(d *schema.ResourceData, typeID uuid.UUID) (*policy.PolicyConfiguration, *string, error)
	policyType  uuid.UUID
}

// genBasePolicyResource creates a Resource with the common elements of a build policy
func genBasePolicyResource(crudArgs *policyCrudArgs) *schema.Resource {
	return &schema.Resource{
		Create:   genPolicyCreateFunc(crudArgs),
		Read:     genPolicyReadFunc(crudArgs),
		Update:   genPolicyUpdateFunc(crudArgs),
		Delete:   genPolicyDeleteFunc(crudArgs),
		Importer: genPolicyImporter(),
		Schema:   genBaseSchema(),
	}
}

type commonPolicySettings struct {
	Scopes []struct {
		RepositoryID      string `json:"repositoryId"`
		RepositoryRefName string `json:"refName"`
		MatchType         string `json:"matchKind"`
	} `json:"scope"`
}

// baseFlattenFunc flattens each of the base elements of the schema
func baseFlattenFunc(d *schema.ResourceData, policyConfig *policy.PolicyConfiguration, projectID *string) error {
	if policyConfig.Id == nil {
		d.SetId("")
		return nil
	}
	d.SetId(strconv.Itoa(*policyConfig.Id))
	d.Set(schemaProjectID, converter.ToString(projectID, ""))
	d.Set(schemaEnabled, converter.ToBool(policyConfig.IsEnabled, true))
	d.Set(schemaBlocking, converter.ToBool(policyConfig.IsBlocking, true))
	settings, err := flattenSettings(d, policyConfig)
	if err != nil {
		return err
	}
	err = d.Set(schemaSettings, settings)
	if err != nil {
		return fmt.Errorf("Unable to persist policy settings configuration: %+v", err)
	}
	return nil
}

func flattenSettings(d *schema.ResourceData, policyConfig *policy.PolicyConfiguration) ([]interface{}, error) {
	policySettings := commonPolicySettings{}
	policyAsJSON, err := json.Marshal(policyConfig.Settings)
	if err != nil {
		return nil, fmt.Errorf("Unable to marshal policy settings into JSON: %+v", err)
	}

	json.Unmarshal(policyAsJSON, &policySettings)
	scopes := make([]interface{}, len(policySettings.Scopes))
	for index, scope := range policySettings.Scopes {
		scopes[index] = map[string]interface{}{
			schemaRepositoryID:  scope.RepositoryID,
			schemaRepositoryRef: scope.RepositoryRefName,
			schemaMatchType:     scope.MatchType,
		}
	}
	settings := []interface{}{
		map[string]interface{}{
			schemaScope: scopes,
		},
	}
	return settings, nil
}

// baseExpandFunc expands each of the base elements of the schema
func baseExpandFunc(d *schema.ResourceData, typeID uuid.UUID) (*policy.PolicyConfiguration, *string, error) {
	projectID := d.Get(schemaProjectID).(string)

	policyConfig := policy.PolicyConfiguration{
		IsEnabled:  converter.Bool(d.Get(schemaEnabled).(bool)),
		IsBlocking: converter.Bool(d.Get(schemaBlocking).(bool)),
		Type: &policy.PolicyTypeRef{
			Id: &typeID,
		},
		Settings: expandSettings(d),
	}

	if d.Id() != "" {
		policyID, err := strconv.Atoi(d.Id())
		if err != nil {
			return nil, nil, fmt.Errorf("Error parsing policy configuration ID: (%+v)", err)
		}
		policyConfig.Id = &policyID
	}

	return &policyConfig, &projectID, nil
}

func expandSettings(d *schema.ResourceData) map[string]interface{} {
	settingsList := d.Get(schemaSettings).([]interface{})
	settings := settingsList[0].(map[string]interface{})
	settingsScopes := settings[schemaScope].([]interface{})

	scopes := make([]map[string]interface{}, len(settingsScopes))
	for index, scope := range settingsScopes {
		scopeMap := scope.(map[string]interface{})
		scopes[index] = map[string]interface{}{
			"repositoryId": scopeMap[schemaRepositoryID],
			"refName":      scopeMap[schemaRepositoryRef],
			"matchKind":    scopeMap[schemaMatchType],
		}
	}
	return map[string]interface{}{
		schemaScope: scopes,
	}
}

func genBaseSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		schemaProjectID: {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		schemaEnabled: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		schemaBlocking: {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		schemaSettings: {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					schemaScope: {
						Type: schema.TypeList,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								schemaRepositoryID: {
									Type:     schema.TypeString,
									Optional: true,
								},
								schemaRepositoryRef: {
									Type:     schema.TypeString,
									Optional: true,
								},
								schemaMatchType: {
									Type:     schema.TypeString,
									Optional: true,
									Default:  matchTypeExact,
									ValidateFunc: validation.StringInSlice([]string{
										matchTypeExact, matchTypePrefix,
									}, true),
								},
							},
						},
						Required: true,
						MinItems: 1,
					},
				},
			},
			Required: true,
			MinItems: 1,
			MaxItems: 1,
		},
	}
}

func genPolicyCreateFunc(crudArgs *policyCrudArgs) schema.CreateFunc {
	return func(d *schema.ResourceData, m interface{}) error {
		clients := m.(*config.AggregatedClient)
		policyConfig, projectID, err := crudArgs.expandFunc(d, crudArgs.policyType)
		if err != nil {
			return err
		}

		createdPolicy, err := clients.PolicyClient.CreatePolicyConfiguration(clients.Ctx, policy.CreatePolicyConfigurationArgs{
			Configuration: policyConfig,
			Project:       projectID,
		})

		if err != nil {
			return fmt.Errorf("Error creating policy in Azure DevOps: %+v", err)
		}

		crudArgs.flattenFunc(d, createdPolicy, projectID)
		return nil
	}
}

func genPolicyReadFunc(crudArgs *policyCrudArgs) schema.ReadFunc {
	return func(d *schema.ResourceData, m interface{}) error {
		clients := m.(*config.AggregatedClient)
		projectID := d.Get(schemaProjectID).(string)
		policyID, err := strconv.Atoi(d.Id())

		if err != nil {
			return fmt.Errorf("Error converting policy ID to an integer: (%+v)", err)
		}

		policyConfig, err := clients.PolicyClient.GetPolicyConfiguration(clients.Ctx, policy.GetPolicyConfigurationArgs{
			Project:         &projectID,
			ConfigurationId: &policyID,
		})

		if utils.ResponseWasNotFound(err) {
			d.SetId("")
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error looking up build policy configuration with ID (%v) and project ID (%v): %v", policyID, projectID, err)
		}

		crudArgs.flattenFunc(d, policyConfig, &projectID)
		return nil
	}
}

func genPolicyUpdateFunc(crudArgs *policyCrudArgs) schema.UpdateFunc {
	return func(d *schema.ResourceData, m interface{}) error {
		clients := m.(*config.AggregatedClient)
		policyConfig, projectID, err := crudArgs.expandFunc(d, crudArgs.policyType)
		if err != nil {
			return err
		}

		updatedPolicy, err := clients.PolicyClient.UpdatePolicyConfiguration(clients.Ctx, policy.UpdatePolicyConfigurationArgs{
			ConfigurationId: policyConfig.Id,
			Configuration:   policyConfig,
			Project:         projectID,
		})

		if err != nil {
			return fmt.Errorf("Error updating policy in Azure DevOps: %+v", err)
		}

		crudArgs.flattenFunc(d, updatedPolicy, projectID)
		return nil
	}
}

func genPolicyDeleteFunc(crudArgs *policyCrudArgs) schema.DeleteFunc {
	return func(d *schema.ResourceData, m interface{}) error {
		clients := m.(*config.AggregatedClient)
		policyConfig, projectID, err := crudArgs.expandFunc(d, crudArgs.policyType)
		if err != nil {
			return err
		}

		err = clients.PolicyClient.DeletePolicyConfiguration(clients.Ctx, policy.DeletePolicyConfigurationArgs{
			ConfigurationId: policyConfig.Id,
			Project:         projectID,
		})

		if err != nil {
			return fmt.Errorf("Error deleting policy in Azure DevOps: %+v", err)
		}

		return nil
	}
}

func genPolicyImporter() *schema.ResourceImporter {
	return &schema.ResourceImporter{
		State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
			id := d.Id()
			parts := strings.SplitN(id, "/", 2)
			if len(parts) != 2 || strings.EqualFold(parts[0], "") || strings.EqualFold(parts[1], "") {
				return nil, fmt.Errorf("unexpected format of ID (%s), expected projectid/resourceId", id)
			}

			_, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("Policy configuration ID (%s) isn't a valid Int", parts[1])
			}

			d.Set(schemaProjectID, parts[0])
			d.SetId(parts[1])
			return []*schema.ResourceData{d}, nil
		},
	}
}
