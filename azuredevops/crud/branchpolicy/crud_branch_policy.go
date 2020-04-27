package crudpolicy

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/microsoft/azure-devops-go-api/azuredevops/policy"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/config"
)

// PolicyCrudArgs arguments for GenBasePolicyResource
type PolicyCrudArgs struct {
	flattenFunc func(d *schema.ResourceData, policy *policy.PolicyConfiguration, projectID *string)
	expandFunc  func(d *schema.ResourceData) (*policy.PolicyConfiguration, *string)
	importFunc  func(clients *config.AggregatedClient, id string) (string, int, error)
	typeID      string
}

// GenBasePolicyResource creates a Resource with the common elements of a build policy
func GenBasePolicyResource(crudArgs *PolicyCrudArgs) *schema.Resource {
	return &schema.Resource{
		Create:   genPolicyCreateFunc(crudArgs),
		Read:     genPolicyReadFunc(crudArgs),
		Update:   genPolicyUpdateFunc(crudArgs),
		Delete:   genPolicyDeleteFunc(crudArgs),
		Importer: genPolicyImporter(crudArgs),
		Schema:   genBaseSchema(),
	}
}

// MatchType match types for branch policies
type MatchType string
type matchTypeValues struct {
	Exact  MatchType
	Prefix MatchType
}

// MatchTypeValues enum of MatchType
var MatchTypeValues = matchTypeValues{
	Exact:  "Exact",
	Prefix: "Prefix",
}

func genBaseSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"enabled": {
			Type:    schema.TypeBool,
			Default: true,
		},
		"blocking": {
			Type:    schema.TypeBool,
			Default: true,
		},
		"settings": {
			Type: schema.TypeSet,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"scope": {
						Type: schema.TypeSet,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"repository_id": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"repository_ref": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"match_type": {
									Type:     schema.TypeString,
									Optional: true,
									Default:  string(MatchTypeValues.Exact),
									ValidateFunc: validation.StringInSlice([]string{
										string(MatchTypeValues.Exact),
										string(MatchTypeValues.Prefix),
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

func genPolicyCreateFunc(crudArgs *PolicyCrudArgs) schema.CreateFunc {
	return func(d *schema.ResourceData, m interface{}) error {
		clients := m.(*config.AggregatedClient)
		policyConfig, projectID := crudArgs.expandFunc(d)

		updatedPolicy, err := clients.PolicyClient.CreatePolicyConfiguration(clients.Ctx, policy.CreatePolicyConfigurationArgs{
			Configuration: policyConfig,
			Project:       projectID,
		})

		if err != nil {
			return fmt.Errorf("Error updating policy in Azure DevOps: %+v", err)
		}

		crudArgs.flattenFunc(d, updatedPolicy, projectID)
		return nil
	}
}

func genPolicyReadFunc(crudArgs *PolicyCrudArgs) schema.ReadFunc {
	return func(d *schema.ResourceData, m interface{}) error {
		clients := m.(*config.AggregatedClient)
		projectID := d.Get("project_id").(string)
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

func genPolicyUpdateFunc(crudArgs *PolicyCrudArgs) schema.UpdateFunc {
	return func(d *schema.ResourceData, m interface{}) error {
		clients := m.(*config.AggregatedClient)
		policyConfig, projectID := crudArgs.expandFunc(d)

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

func genPolicyDeleteFunc(crudArgs *PolicyCrudArgs) schema.DeleteFunc {
	return func(d *schema.ResourceData, m interface{}) error {
		clients := m.(*config.AggregatedClient)
		policyConfig, projectID := crudArgs.expandFunc(d)

		err := clients.PolicyClient.DeletePolicyConfiguration(clients.Ctx, policy.DeletePolicyConfigurationArgs{
			ConfigurationId: policyConfig.Id,
			Project:         projectID,
		})

		if err != nil {
			return fmt.Errorf("Error deleting policy in Azure DevOps: %+v", err)
		}

		return nil
	}
}

func genPolicyImporter(crudArgs *PolicyCrudArgs) *schema.ResourceImporter {
	return &schema.ResourceImporter{
		State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
			projectID, policyID, err := crudArgs.importFunc(meta.(*config.AggregatedClient), d.Id())
			if err != nil {
				return nil, fmt.Errorf("Error parsing policyID from the Terraform resource data:  %v", err)
			}
			d.Set("project_id", projectID)
			d.SetId(strconv.Itoa(policyID))
			return []*schema.ResourceData{d}, nil
		},
	}
}
