package azuredevops

// The tests in this file use the mock clients in mock_client.go to mock out
// the Azure DevOps client operations.

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/testhelper"
)

/**
 * Begin acceptance tests
 */

// Verifies that the following sequence of events occurrs without error:
//	(1) Branch policies can be created with no errors
//	(2) Branch policies can be updated with no errors
//	(3) Branch policies can be deleted with no errors
func TestAccAzureDevOpsBranchPolicy_CreateAndUpdate(t *testing.T) {
	projName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	repoName := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	opts1 := hclOptions{
		projectName: projName,
		repoName:    repoName,
		minReviewerOptions: policyOptions{
			true,
			true,
		},
	}

	opts2 := hclOptions{
		projectName: projName,
		repoName:    repoName,
		minReviewerOptions: policyOptions{
			false,
			false,
		},
	}

	minReviewerTfNode := "azuredevops_branch_policy_min_reviewers.p"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testhelper.TestAccPreCheck(t, nil) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: getHCL(opts1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(minReviewerTfNode, "id"),
					resource.TestCheckResourceAttr(minReviewerTfNode, "blocking", fmt.Sprintf("%t", opts1.minReviewerOptions.blocking)),
					resource.TestCheckResourceAttr(minReviewerTfNode, "enabled", fmt.Sprintf("%t", opts1.minReviewerOptions.enabled)),
				),
			}, {
				Config: getHCL(opts2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(minReviewerTfNode, "id"),
					resource.TestCheckResourceAttr(minReviewerTfNode, "blocking", fmt.Sprintf("%t", opts2.minReviewerOptions.blocking)),
					resource.TestCheckResourceAttr(minReviewerTfNode, "enabled", fmt.Sprintf("%t", opts2.minReviewerOptions.enabled)),
				),
			}, {
				ResourceName:      minReviewerTfNode,
				ImportStateIdFunc: testAccImportStateIDFunc(minReviewerTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

type policyOptions struct {
	enabled  bool
	blocking bool
}

type hclOptions struct {
	projectName        string
	repoName           string
	minReviewerOptions policyOptions
}

func getHCL(opts hclOptions) string {
	projectAndRepo := testhelper.TestAccAzureGitRepoResource(opts.projectName, opts.repoName, "Clean")
	minReviewCountPolicyFmt := `
	resource "azuredevops_branch_policy_min_reviewers" "p" {
		project_id = azuredevops_project.project.id
		enabled  = %t
		blocking = %t
		settings {
			reviewer_count     = 1
			submitter_can_vote = false
			scope {
				repository_id  = azuredevops_git_repository.gitrepo.id
				repository_ref = azuredevops_git_repository.gitrepo.default_branch
				match_type     = "exact"
			}
		}
	}`

	minReviewCountPolicy := fmt.Sprintf(
		minReviewCountPolicyFmt,
		opts.minReviewerOptions.enabled,
		opts.minReviewerOptions.blocking)

	return fmt.Sprintf("%s\n%s", projectAndRepo, minReviewCountPolicy)
}
