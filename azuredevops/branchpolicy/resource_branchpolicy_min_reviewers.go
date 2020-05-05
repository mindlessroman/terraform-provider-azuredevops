package branchpolicy

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/microsoft/azure-devops-go-api/azuredevops/policy"
)

const (
	schemaReviewerCount    = "reviewer_count"
	schemaSubmitterCanVote = "submitter_can_vote"
)

type minReviewerPolicySettings struct {
	ApprovalCount    int  `json:"minimumApproverCount"`
	SubmitterCanVote bool `json:"creatorVoteCounts"`
}

func ResourceBranchPolicyMinReviewers() *schema.Resource {
	resource := genBasePolicyResource(&policyCrudArgs{
		flattenFunc,
		expandFunc,
		minReviewerCount,
	})

	settingsSchema := resource.Schema[schemaSettings].Elem.(*schema.Resource).Schema
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

func flattenFunc(d *schema.ResourceData, policy *policy.PolicyConfiguration, projectID *string) error {
	err := baseFlattenFunc(d, policy, projectID)
	if err != nil {
		return err
	}
	policySettings := minReviewerPolicySettings{}
	json.Unmarshal([]byte(fmt.Sprintf("%v", policy.Settings)), &policySettings)

	settingsList := d.Get(schemaSettings).([]interface{})
	settings := settingsList[0].(map[string]interface{})

	settings[schemaReviewerCount] = policySettings.ApprovalCount
	settings[schemaSubmitterCanVote] = policySettings.SubmitterCanVote

	d.Set(schemaSettings, settings)
	return nil
}

func expandFunc(d *schema.ResourceData, typeID uuid.UUID) (*policy.PolicyConfiguration, *string, error) {
	policyConfig, projectID, err := baseExpandFunc(d, typeID)
	if err != nil {
		return nil, nil, err
	}

	settingsList := d.Get(schemaSettings).([]interface{})
	settings := settingsList[0].(map[string]interface{})

	policySettings := policyConfig.Settings.(map[string]interface{})
	policySettings["minimumApproverCount"] = settings[schemaReviewerCount].(int)
	policySettings["creatorVoteCounts"] = settings[schemaSubmitterCanVote].(bool)

	return policyConfig, projectID, nil
}
