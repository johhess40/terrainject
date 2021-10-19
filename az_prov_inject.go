/*
This package is intended to be used along with Terraform to inject provider values for Azure into the working directory
- Must pass in a JSON object, preferably a JSON file with correct formatting
*/

package terrainject

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/hcl2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"log"
	"os"
)

// AzProv is populated using JSON unmarshalling to creat a struct that can be used to populate provider fields for Microsoft Azure
// Optionally other methods other than a file can be used in order to populate the struct but function ReadFile is also available for this purpose
type AzProv struct {
	Azure struct {
		Features struct {
			APIMngmt struct {
				PurgeOnDestroy bool `json:"PurgeOnDestroy"`
			} `json:"ApiMngmt"`
			CogAccount struct {
				PurgeOnDestroy bool `json:"PurgeOnDestroy"`
			} `json:"CogAccount"`
			KeyVault struct {
				PurgeOnDestroy    bool `json:"PurgeOnDestroy"`
				RecoverSoftDelete bool `json:"RecoverSoftDelete"`
			} `json:"KeyVault"`
			LogAnalyticsWrkSpc struct {
				PermDeleteOnDestroy bool `json:"PermDeleteOnDestroy"`
			} `json:"LogAnalyticsWrkSpc"`
			ResourceGroup struct {
				PrevDeleteIfRes bool `json:"PrevDeleteIfRes"`
			} `json:"ResourceGroup"`
			TempDeploy struct {
				DeleteNestedItems bool `json:"DeleteNestedItems"`
			} `json:"TempDeploy"`
			VirtMachine struct {
				DeleteOsDisk        bool `json:"DeleteOsDisk"`
				GracefulShutdown    bool `json:"GracefulShutdown"`
				SkipShutForceDelete bool `json:"SkipShutForceDelete"`
			} `json:"VirtMachine"`
			Vmss struct {
				ForceDelete   bool `json:"ForceDelete"`
				RollInstances bool `json:"RollInstances"`
			} `json:"VMSS"`
		} `json:"Features"`
		ClientID          string   `json:"ClientId"`
		Environment       string   `json:"Environment"`
		SubID             string   `json:"SubId"`
		TenantID          string   `json:"TenantId"`
		AuxTenantID       []string `json:"AuxTenantId"`
		ClientCertPass    string   `json:"ClientCertPass"`
		ClientCertPath    string   `json:"ClientCertPath"`
		ClientSecret      string   `json:"ClientSecret"`
		MsiEndpoint       string   `json:"MsiEndpoint"`
		UseMsi            bool     `json:"UseMsi"`
		DisablePartnerID  bool     `json:"DisablePartnerId"`
		MetaHost          string   `json:"MetaHost"`
		PartnerID         string   `json:"PartnerId"`
		SkipProviderReg   bool     `json:"SkipProviderReg"`
		StorageUseAzureAd bool     `json:"StorageUseAzureAd"`
	} `json:"Azure"`
}

var (
	prov   AzProv
)

// BuildAz takes JSON data in from the ReadFile function and formats it into the needed provider block
// If fields are not intended to be set for provider block simply omit them, they will evaluate to false
func BuildAz(f string, prov string, fields *AzProv) int {

	feats := map[string]bool{
		"APIMngmtPurge":               fields.Azure.Features.APIMngmt.PurgeOnDestroy,
		"CogAccountPurge":             fields.Azure.Features.CogAccount.PurgeOnDestroy,
		"KvPurge":                     fields.Azure.Features.KeyVault.PurgeOnDestroy,
		"KvSoftDelete":                fields.Azure.Features.KeyVault.RecoverSoftDelete,
		"LogAnalyticsPermDelete":      fields.Azure.Features.LogAnalyticsWrkSpc.PermDeleteOnDestroy,
		"ResourceGroupPrevDelete":     fields.Azure.Features.ResourceGroup.PrevDeleteIfRes,
		"TempDeploy":                  fields.Azure.Features.TempDeploy.DeleteNestedItems,
		"VirtMachineDeleteOsDisk":     fields.Azure.Features.VirtMachine.DeleteOsDisk,
		"VirtMachineGracefulShutdown": fields.Azure.Features.VirtMachine.GracefulShutdown,
		"VirtMachineForceDelete":      fields.Azure.Features.VirtMachine.SkipShutForceDelete,
		"VmssForceDelete":             fields.Azure.Features.Vmss.ForceDelete,
		"VmssRollInstances":           fields.Azure.Features.Vmss.RollInstances,
		"UseMsi":                      fields.Azure.UseMsi,
		"DisablePartnerId":            fields.Azure.DisablePartnerID,
		"SkipProviderReg":             fields.Azure.SkipProviderReg,
		"StorageUseAzureAd":           fields.Azure.StorageUseAzureAd,
	}

	extras := map[string]string{
		"ClientId":       fields.Azure.ClientID,
		"Env":            fields.Azure.Environment,
		"SubId":          fields.Azure.SubID,
		"TenantId":       fields.Azure.TenantID,
		"ClientCertPass": fields.Azure.ClientCertPass,
		"ClientCertPath": fields.Azure.ClientCertPath,
		"ClientSecret":   fields.Azure.ClientSecret,
		"MsiEndpoint":    fields.Azure.MsiEndpoint,
		"MetaHost":       fields.Azure.MetaHost,
		"PartnerId":      fields.Azure.PartnerID,
	}

	auxTenId := fields.Azure.AuxTenantID

	writer := hclwrite.NewFile()

	hclFile, err := os.OpenFile(f, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	rootBod := writer.Body()

	provider := rootBod.AppendNewBlock("provider", []string{prov})

	providerBody := provider.Body()

	block := hclwrite.NewBlock("features", nil)

	features := block.Body()

	if feats["KvPurge"] && feats["KvSoftDelete"] == true {
		var kvopts []bool

		kvopts = append(kvopts, true, true)

		fmt.Println("Writing block for key vault options....")
		kv := hclwrite.NewBlock("key_vault", nil)

		set := kv.Body()

		if feats["KvPurge"] {
			val := cty.BoolVal(kvopts[0])
			set.SetAttributeValue("recover_soft_deleted_key_vaults", val)
		}

		if feats["KvSoftDelete"] {
			val := cty.BoolVal(kvopts[1])
			set.SetAttributeValue("purge_soft_delete_on_destroy", val)
		}

		features.AppendBlock(kv)
	} else if feats["KvPurge"] == true && feats["KvSoftDelete"] == false {
		var kvopts []bool

		kvopts = append(kvopts, true, false)

		fmt.Println("Writing block for key vault options....")
		kv := hclwrite.NewBlock("key_vault", nil)

		set := kv.Body()

		if feats["KvPurge"] {
			val := cty.BoolVal(kvopts[0])
			set.SetAttributeValue("recover_soft_deleted_key_vaults", val)
		}

		if feats["KvSoftDelete"] == false {
			val := cty.BoolVal(kvopts[1])
			set.SetAttributeValue("purge_soft_delete_on_destroy", val)
		}

		features.AppendBlock(kv)
	} else if feats["KvPurge"] == false && feats["KvSoftDelete"] == true {
		var kvopts []bool

		kvopts = append(kvopts, false, true)

		fmt.Println("Writing block for key vault options....")
		kv := hclwrite.NewBlock("key_vault", nil)

		set := kv.Body()

		if feats["KvPurge"] == false {
			val := cty.BoolVal(kvopts[0])
			set.SetAttributeValue("recover_soft_deleted_key_vaults", val)
		}

		if feats["KvSoftDelete"] {
			val := cty.BoolVal(kvopts[1])
			set.SetAttributeValue("purge_soft_delete_on_destroy", val)
		}

		features.AppendBlock(kv)
	}

	if feats["VirtMachineDeleteOsDisk"] && feats["VirtMachineGracefulShutdown"] && feats["VirtMachineForceDelete"] == true {
		var vmopts []bool

		vmopts = append(vmopts, true, true, true)

		fmt.Println("Writing block for vm options....")
		vm := hclwrite.NewBlock("virtual_machine", nil)

		set := vm.Body()

		if feats["VirtMachineDeleteOsDisk"] {
			val := cty.BoolVal(vmopts[0])
			set.SetAttributeValue("delete_os_disk_on_deletion", val)
		}

		if feats["VirtMachineGracefulShutdown"] {
			val := cty.BoolVal(vmopts[1])
			set.SetAttributeValue("graceful_shutdown", val)
		}

		if feats["VirtMachineForceDelete"] {
			val := cty.BoolVal(vmopts[2])
			set.SetAttributeValue("skip_shutdown_and_force_delete", val)
		}

		features.AppendBlock(vm)
	} else if feats["VirtMachineDeleteOsDisk"] && feats["VirtMachineGracefulShutdown"] == true && feats["VirtMachineForceDelete"] == false {
		var vmopts []bool

		vmopts = append(vmopts, true, true, false)

		fmt.Println("Writing block for vm options....")
		vm := hclwrite.NewBlock("virtual_machine", nil)

		set := vm.Body()

		if feats["VirtMachineDeleteOsDisk"] {
			val := cty.BoolVal(vmopts[0])
			set.SetAttributeValue("delete_os_disk_on_deletion", val)
		}

		if feats["VirtMachineGracefulShutdown"] {
			val := cty.BoolVal(vmopts[1])
			set.SetAttributeValue("graceful_shutdown", val)
		}

		if feats["VirtMachineForceDelete"] == false {
			val := cty.BoolVal(vmopts[2])
			set.SetAttributeValue("skip_shutdown_and_force_delete", val)
		}

		features.AppendBlock(vm)
	} else if feats["VirtMachineDeleteOsDisk"] == true && feats["VirtMachineGracefulShutdown"] && feats["VirtMachineForceDelete"] == false {
		var vmopts []bool

		vmopts = append(vmopts, true, false, false)

		fmt.Println("Writing block for vm options....")
		vm := hclwrite.NewBlock("virtual_machine", nil)

		set := vm.Body()

		if feats["VirtMachineDeleteOsDisk"] {
			val := cty.BoolVal(vmopts[0])
			set.SetAttributeValue("delete_os_disk_on_deletion", val)
		}

		if feats["VirtMachineGracefulShutdown"] == false {
			val := cty.BoolVal(vmopts[1])
			set.SetAttributeValue("graceful_shutdown", val)
		}

		if feats["VirtMachineForceDelete"] == false {
			val := cty.BoolVal(vmopts[2])
			set.SetAttributeValue("skip_shutdown_and_force_delete", val)
		}

		features.AppendBlock(vm)
	} else if feats["VirtMachineDeleteOsDisk"] == false && feats["VirtMachineGracefulShutdown"] && feats["VirtMachineForceDelete"] == true {
		var vmopts []bool

		vmopts = append(vmopts, false, true, true)

		fmt.Println("Writing block for vm options....")
		vm := hclwrite.NewBlock("virtual_machine", nil)

		set := vm.Body()

		if feats["VirtMachineDeleteOsDisk"] == false {
			val := cty.BoolVal(vmopts[0])
			set.SetAttributeValue("delete_os_disk_on_deletion", val)
		}

		if feats["VirtMachineGracefulShutdown"] {
			val := cty.BoolVal(vmopts[1])
			set.SetAttributeValue("graceful_shutdown", val)
		}

		if feats["VirtMachineForceDelete"] {
			val := cty.BoolVal(vmopts[2])
			set.SetAttributeValue("skip_shutdown_and_force_delete", val)
		}

		features.AppendBlock(vm)
	} else if feats["VirtMachineDeleteOsDisk"] && feats["VirtMachineGracefulShutdown"] == false && feats["VirtMachineForceDelete"] == true {
		var vmopts []bool

		vmopts = append(vmopts, false, false, true)

		fmt.Println("Writing block for vm options....")
		vm := hclwrite.NewBlock("virtual_machine", nil)

		set := vm.Body()

		if feats["VirtMachineDeleteOsDisk"] == false {
			val := cty.BoolVal(vmopts[0])
			set.SetAttributeValue("delete_os_disk_on_deletion", val)
		}

		if feats["VirtMachineGracefulShutdown"] == false {
			val := cty.BoolVal(vmopts[1])
			set.SetAttributeValue("graceful_shutdown", val)
		}

		if feats["VirtMachineForceDelete"] {
			val := cty.BoolVal(vmopts[2])
			set.SetAttributeValue("skip_shutdown_and_force_delete", val)
		}

		features.AppendBlock(vm)
	} else if feats["VirtMachineDeleteOsDisk"] && feats["VirtMachineForceDelete"] == false && feats["VirtMachineGracefulShutdown"] == true {
		var vmopts []bool

		vmopts = append(vmopts, false, true, false)

		fmt.Println("Writing block for vm options....")
		vm := hclwrite.NewBlock("virtual_machine", nil)

		set := vm.Body()

		if feats["VirtMachineDeleteOsDisk"] == false {
			val := cty.BoolVal(vmopts[0])
			set.SetAttributeValue("delete_os_disk_on_deletion", val)
		}

		if feats["VirtMachineGracefulShutdown"] {
			val := cty.BoolVal(vmopts[1])
			set.SetAttributeValue("graceful_shutdown", val)
		}

		if feats["VirtMachineForceDelete"] == false {
			val := cty.BoolVal(vmopts[2])
			set.SetAttributeValue("skip_shutdown_and_force_delete", val)
		}

		features.AppendBlock(vm)
	}

	if feats["VmssForceDelete"] && feats["VmssRollInstances"] == true {
		var vmssopts []bool

		vmssopts = append(vmssopts, true, true)

		fmt.Println("Writing block for vmss options....")
		vmss := hclwrite.NewBlock("virtual_machine_scale_set", nil)

		set := vmss.Body()

		if feats["VmssForceDelete"] {
			val := cty.BoolVal(vmssopts[0])
			set.SetAttributeValue("force_delete", val)
		}

		if feats["VmssRollInstances"] {
			val := cty.BoolVal(vmssopts[1])
			set.SetAttributeValue("roll_instances_when_required", val)
		}

		features.AppendBlock(vmss)
	} else if feats["VmssForceDelete"] == false && feats["VmssRollInstances"] == true {
		var vmssopts []bool

		vmssopts = append(vmssopts, false, true)

		fmt.Println("Writing block for vmss options....")
		vmss := hclwrite.NewBlock("virtual_machine_scale_set", nil)

		set := vmss.Body()

		if feats["VmssForceDelete"] == false {
			val := cty.BoolVal(vmssopts[0])
			set.SetAttributeValue("force_delete", val)
		}

		if feats["VmssRollInstances"] {
			val := cty.BoolVal(vmssopts[1])
			set.SetAttributeValue("roll_instances_when_required", val)
		}

		features.AppendBlock(vmss)
	} else if feats["VmssForceDelete"] == true && feats["VmssRollInstances"] == false {
		var vmssopts []bool

		vmssopts = append(vmssopts, true, false)

		fmt.Println("Writing block for vmss options....")
		vmss := hclwrite.NewBlock("virtual_machine_scale_set", nil)

		set := vmss.Body()

		if feats["VmssForceDelete"] {
			val := cty.BoolVal(vmssopts[0])
			set.SetAttributeValue("force_delete", val)
		}

		if feats["VmssRollInstances"] == false {
			val := cty.BoolVal(vmssopts[1])
			set.SetAttributeValue("roll_instances_when_required", val)
		}

		features.AppendBlock(vmss)
	}

	for k, v := range feats {
		switch v {
		case k == "ApiMngmtPurge":
			if v == true {
				val := cty.BoolVal(v)

				fmt.Println("Writing block for api_management....")
				api := hclwrite.NewBlock("api_management", nil)

				set := api.Body()

				set.SetAttributeValue("purge_soft_delete_on_destroy", val)

				features.AppendBlock(api)
			}
		case k == "CogAccountPurge":
			if v == true {
				val := cty.BoolVal(v)

				fmt.Println("Writing block for cognitive_account....")
				cog := hclwrite.NewBlock("cognitive_account", nil)

				set := cog.Body()

				set.SetAttributeValue("purge_soft_delete_on_destroy", val)

				features.AppendBlock(cog)
			}
		case k == "LogAnalyticsPermDelete":
			if v == true {
				val := cty.BoolVal(v)

				fmt.Println("Writing block for log analytics....")
				loganalytics := hclwrite.NewBlock("log_analytics_workspace", nil)

				set := loganalytics.Body()

				set.SetAttributeValue("permanently_delete_on_destroy", val)

				features.AppendBlock(loganalytics)
			}
		case k == "ResourceGroupPrevDelete":
			if v == true {
				val := cty.BoolVal(v)

				fmt.Println("Writing block for resource group....")
				rg := hclwrite.NewBlock("resource_group", nil)

				set := rg.Body()

				set.SetAttributeValue("prevent_deletion_if_contains_resources", val)
				features.AppendBlock(rg)
			}
		case k == "TempDeploy":
			if v == true {
				val := cty.BoolVal(v)

				fmt.Println("Writing block for template deployment....")
				tmp := hclwrite.NewBlock("template_deployment", nil)

				set := tmp.Body()

				set.SetAttributeValue("delete_nested_items_during_deletion", val)

				features.AppendBlock(tmp)
			}
		case k == "UseMsi":
			if v == true {
				val := cty.BoolVal(v)

				fmt.Println("Writing for msi field....")

				features.SetAttributeValue("use_msi", val)
			}
		case k == "DisablePartnerID":
			if v == true {
				val := cty.BoolVal(v)

				fmt.Println("Writing for disable_partner_id field....")

				features.SetAttributeValue("disable_partner_id", val)
			}
		case k == "SkipProviderReg":
			if v == true {
				val := cty.BoolVal(v)

				fmt.Println("Writing for skip_provider_registration field....")

				features.SetAttributeValue("skip_provider_registration", val)
			}
		case k == "StorageUseAzureAd":
			if v == true {
				val := cty.BoolVal(v)

				fmt.Println("Writing for storage_use_azuread field....")

				features.SetAttributeValue("storage_use_azuread", val)
			}
		default:
			for k, v := range feats {
				if v == false {
					fmt.Printf("%v set to %v, skipping setting of block/field...\n", k, v)
				}
			}
		}
	}

	for k, v := range extras {
		switch k {
		case "ClientId":
			if len(v) != 0 {
				val := cty.StringVal(v)

				fmt.Println("Writing field for client id...")

				providerBody.SetAttributeValue("client_id", val)
			}
		case "Env":
			if len(v) != 0 {
				val := cty.StringVal(v)

				fmt.Println("Writing field for environment...")

				providerBody.SetAttributeValue("environment", val)
			}
		case "SubId":
			if len(v) != 0 {
				val := cty.StringVal(v)

				fmt.Println("Writing field for subscription id...")

				providerBody.SetAttributeValue("subscription_id", val)
			}
		case "TenantId":
			if len(v) != 0 {
				val := cty.StringVal(v)

				fmt.Println("Writing field for tenant id...")

				providerBody.SetAttributeValue("tenant_id", val)
			}
		case "ClientCertPass":
			if len(v) != 0 {
				val := cty.StringVal(v)

				fmt.Println("Writing field for client cert password...")

				providerBody.SetAttributeValue("client_certificate_password", val)
			}
		case "ClientCertPath":
			if len(v) != 0 {
				val := cty.StringVal(v)

				fmt.Println("Writing field for client cert path...")

				providerBody.SetAttributeValue("client_certificate_path", val)
			}
		case "ClientSecret":
			if len(v) != 0 {
				val := cty.StringVal(v)

				fmt.Println("Writing field for client secret...")

				providerBody.SetAttributeValue("client_secret", val)
			}
		case "MsiEndpoint":
			if len(v) != 0 {
				val := cty.StringVal(v)

				fmt.Println("Writing field for msi endpoint...")

				providerBody.SetAttributeValue("msi_endpoint", val)
			}
		case "MetaHost":
			if len(v) != 0 {
				val := cty.StringVal(v)

				fmt.Println("Writing field for metadata host...")

				providerBody.SetAttributeValue("metadata_host", val)
			}
		case "PartnerId":
			if len(v) != 0 {
				val := cty.StringVal(v)

				fmt.Println("Writing field for partner id...")

				providerBody.SetAttributeValue("partner_id", val)
			}
		}
	}

	if len(auxTenId) != 0 {
		var val []cty.Value
		for _,v := range auxTenId {
			val = append(val, cty.StringVal(v))
		}
		l := cty.ListVal(val)
		providerBody.SetAttributeValue("auxiliary_tenant_ids", l)
	} else {
		fmt.Println("Skipping provisioning of auxiliary_tenant_ids field")
	}

	providerBody.AppendBlock(block)

	write, err := hclFile.Write(writer.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	return write
}

// ReadFile Function readFile reads the JSON file provided from struct files for provider creation
// Takes one argument which is a string in order to read the file and then Unmarshal the JSON
func ReadFile(path string) (*AzProv, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(file, &prov)
	if err != nil {
		log.Fatal(err)
	}

	return &prov, nil

}
