package azuredevops

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/microsoft/azure-devops-go-api/azuredevops/policy"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/crud/branchpolicy"
)

const (
	schemaReviewerCount    = "reviewer_count"
	schemaSubmitterCanVote = "submitter_can_vote"
)

type minReviewerPolicySettings struct {
	ApprovalCount    int  `json:"minimumApproverCount"`
	SubmitterCanVote bool `json:"creatorVoteCounts"`
}

func resourceBranchPolicyMinReviewers() *schema.Resource {
	resource := branchpolicy.GenBasePolicyResource(&branchpolicy.PolicyCrudArgs{
		flattenFunc,
		expandFunc,
		branchpolicy.MinReviewerCount,
	})

	settingsSchema := resource.Schema[branchpolicy.SchemaSettings].Elem.(*schema.Resource).Schema
	settingsSchema[schemaReviewerCount] = &schema.Schema{
		Type:         schema.TypeInt,
		Required:     true,
		ValidateFunc: validation.IntAtLeast(1),
	}
	settingsSchema[schemaSubmitterCanVote] = &schema.Schema{
		Type:     schema.TypeBool,
		Default:  false,
		Optional: true,
	}
	return resource
}

func flattenFunc(d *schema.ResourceData, policy *policy.PolicyConfiguration, projectID *string) {
	branchpolicy.BaseFlattenFunc(d, policy, projectID)
	policySettings := minReviewerPolicySettings{}
	json.Unmarshal([]byte(fmt.Sprintf("%v", policy.Settings)), &policySettings)

	settingsList := d.Get(branchpolicy.SchemaSettings).([]interface{})
	settings := settingsList[0].(map[string]interface{})

	settings[schemaReviewerCount] = policySettings.ApprovalCount
	settings[schemaSubmitterCanVote] = policySettings.SubmitterCanVote

	d.Set(branchpolicy.SchemaSettings, settings)
}

func expandFunc(d *schema.ResourceData, typeID uuid.UUID) (*policy.PolicyConfiguration, *string, error) {
	policyConfig, projectID, err := branchpolicy.BaseExpandFunc(d, typeID)
	if err != nil {
		return nil, nil, err
	}

	settingsList := d.Get(branchpolicy.SchemaSettings).([]interface{})
	settings := settingsList[0].(map[string]interface{})

	policySettings := policyConfig.Settings.(map[string]interface{})
	policySettings["minimumApproverCount"] = settings[schemaReviewerCount].(int)
	policySettings["creatorVoteCounts"] = settings[schemaSubmitterCanVote].(bool)

	return policyConfig, projectID, nil
}
