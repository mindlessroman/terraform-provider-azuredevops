package testhelper

import (
	"fmt"
	"strings"
)

// AccTestHCLAzureGitRepoResource HCL describing an AzDO GIT repository resource
func AccTestHCLAzureGitRepoResource(projectName string, gitRepoName string, initType string) string {
	azureGitRepoResource := fmt.Sprintf(`
resource "azuredevops_git_repository" "gitrepo" {
	project_id      = azuredevops_project.project.id
	name            = "%s"
	initialization {
		init_type = "%s"
	}
}`, gitRepoName, initType)

	projectResource := AccTestHCLProjectResource(projectName)
	return fmt.Sprintf("%s\n%s", projectResource, azureGitRepoResource)
}

// AccTestHCLGroupDataSource HCL describing an AzDO Group Data Source
func AccTestHCLGroupDataSource(projectName string, groupName string) string {
	dataSource := fmt.Sprintf(`
data "azuredevops_group" "group" {
	project_id = azuredevops_project.project.id
	name       = "%s"
}`, groupName)

	projectResource := AccTestHCLProjectResource(projectName)
	return fmt.Sprintf("%s\n%s", projectResource, dataSource)
}

// AccTestHCLProjectResource HCL describing an AzDO project
func AccTestHCLProjectResource(projectName string) string {
	if projectName == "" {
		return ""
	}
	return fmt.Sprintf(`
resource "azuredevops_project" "project" {
	project_name       = "%s"
	description        = "%s-description"
	visibility         = "private"
	version_control    = "Git"
	work_item_template = "Agile"
}`, projectName, projectName)
}

// AccTestHCLUserEntitlementResource HCL describing an AzDO UserEntitlement
func AccTestHCLUserEntitlementResource(principalName string) string {
	return fmt.Sprintf(`
resource "azuredevops_user_entitlement" "user" {
	principal_name     = "%s"
	account_license_type = "express"
}`, principalName)
}

// AccTestHCLServiceEndpointGitHubResource HCL describing an AzDO service endpoint
func AccTestHCLServiceEndpointGitHubResource(projectName string, serviceEndpointName string) string {
	serviceEndpointResource := fmt.Sprintf(`
resource "azuredevops_serviceendpoint_github" "serviceendpoint" {
	project_id             = azuredevops_project.project.id
	service_endpoint_name  = "%s"
	auth_personal {
	}
}`, serviceEndpointName)

	projectResource := AccTestHCLProjectResource(projectName)
	return fmt.Sprintf("%s\n%s", projectResource, serviceEndpointResource)
}

// AccTestHCLServiceEndpointDockerHubResource HCL describing an AzDO service endpoint
func AccTestHCLServiceEndpointDockerHubResource(projectName string, serviceEndpointName string) string {
	serviceEndpointResource := fmt.Sprintf(`
resource "azuredevops_serviceendpoint_dockerhub" "serviceendpoint" {
	project_id             = azuredevops_project.project.id
	service_endpoint_name  = "%s"
}`, serviceEndpointName)

	projectResource := AccTestHCLProjectResource(projectName)
	return fmt.Sprintf("%s\n%s", projectResource, serviceEndpointResource)
}

// AccTestHCLServiceEndpointAzureRMResource HCL describing an AzDO service endpoint
func AccTestHCLServiceEndpointAzureRMResource(projectName string, serviceEndpointName string) string {
	serviceEndpointResource := fmt.Sprintf(`
resource "azuredevops_serviceendpoint_azurerm" "serviceendpointrm" {
	project_id             = azuredevops_project.project.id
	service_endpoint_name  = "%s"
	azurerm_spn_clientid 	="e318e66b-ec4b-4dff-9124-41129b9d7150"
	azurerm_spn_tenantid      = "9c59cbe5-2ca1-4516-b303-8968a070edd2"
    azurerm_subscription_id   = "3b0fee91-c36d-4d70-b1e9-fc4b9d608c3d"
    azurerm_subscription_name = "Microsoft Azure DEMO"
    azurerm_scope             = "/subscriptions/3b0fee91-c36d-4d70-b1e9-fc4b9d608c3d"
	azurerm_spn_clientsecret ="d9d210dd-f9f0-4176-afb8-a4df60e1ae72"

}`, serviceEndpointName)

	projectResource := AccTestHCLProjectResource(projectName)
	return fmt.Sprintf("%s\n%s", projectResource, serviceEndpointResource)
}

// AccTestHCLVariableGroupResource HCL describing an AzDO variable group
func AccTestHCLVariableGroupResource(projectName string, variableGroupName string, allowAccess bool) string {
	variableGroupResource := fmt.Sprintf(`
resource "azuredevops_variable_group" "vg" {
	project_id  = azuredevops_project.project.id
	name        = "%s"
	description = "A sample variable group."
	allow_access = %t
	variable {
		name      = "key1"
		value     = "value1"
		is_secret = true
	}

	variable {
		name  = "key2"
		value = "value2"
	}

	variable {
		name = "key3"
	}
}`, variableGroupName, allowAccess)

	projectResource := AccTestHCLProjectResource(projectName)
	return fmt.Sprintf("%s\n%s", projectResource, variableGroupResource)
}

// AccTestHCLAgentPoolResource HCL describing an AzDO Agent Pool
func AccTestHCLAgentPoolResource(poolName string) string {
	return fmt.Sprintf(`
resource "azuredevops_agent_pool" "pool" {
	name           = "%s"
	auto_provision = false
	pool_type      = "automation"
	}`, poolName)
}

// AccTestHCLBuildDefinitionResource HCL describing an AzDO build definition
func AccTestHCLBuildDefinitionResource(projectName string, buildDefinitionName string, buildPath string) string {
	buildDefinitionResource := fmt.Sprintf(`
resource "azuredevops_build_definition" "build" {
	project_id      = azuredevops_project.project.id
	name            = "%s"
	agent_pool_name = "Hosted Ubuntu 1604"
	path			= "%s"

	repository {
	  repo_type             = "GitHub"
	  repo_name             = "repoOrg/repoName"
	  branch_name           = "branch"
	  yml_path              = "path/to/yaml"
	}
}`, buildDefinitionName, strings.ReplaceAll(buildPath, `\`, `\\`))

	projectResource := AccTestHCLProjectResource(projectName)
	return fmt.Sprintf("%s\n%s", projectResource, buildDefinitionResource)
}

// AccTestHCLGroupMembershipResource full terraform stanza to standup a group membership
func AccTestHCLGroupMembershipResource(projectName, groupName, userPrincipalName string) string {
	membershipDependenciesStanza := AccTestHCLGroupMembershipDependencies(projectName, groupName, userPrincipalName)
	membershipStanza := `
resource "azuredevops_group_membership" "membership" {
	group = data.azuredevops_group.group.descriptor
	members = [azuredevops_user_entitlement.user.descriptor]
}`

	return membershipDependenciesStanza + "\n" + membershipStanza
}

// AccTestHCLGroupMembershipDependencies all the dependencies needed to configure a group membership
func AccTestHCLGroupMembershipDependencies(projectName, groupName, userPrincipalName string) string {
	return fmt.Sprintf(`
resource "azuredevops_project" "project" {
	project_name = "%s"
}
data "azuredevops_group" "group" {
	project_id = azuredevops_project.project.id
	name       = "%s"
}
resource "azuredevops_user_entitlement" "user" {
	principal_name       = "%s"
	account_license_type = "express"
}

output "group_descriptor" {
	value = data.azuredevops_group.group.descriptor
}
output "user_descriptor" {
	value = azuredevops_user_entitlement.user.descriptor
}
`, projectName, groupName, userPrincipalName)
}

// AccTestHCLGroupResource HCL describing an AzDO group, if the projectName is empty, only a azuredevops_group instance is returned
func AccTestHCLGroupResource(groupResourceName, projectName, groupName string) string {
	return fmt.Sprintf(`
%s

resource "azuredevops_group" "%s" {
	scope        = azuredevops_project.project.id
	display_name = "%s"
}

output "group_id_%s" {
	value = azuredevops_group.%s.id
}
`, AccTestHCLProjectResource(projectName), groupResourceName, groupName, groupResourceName, groupResourceName)
}
