package cmd

import (
	"os"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/spf13/cobra"
)

type protecodeExecuteScanOptions struct {
	ProtecodeExcludeCVEs                 string `json:"protecodeExcludeCVEs,omitempty"`
	ProtecodeFailOnSevereVulnerabilities bool   `json:"protecodeFailOnSevereVulnerabilities,omitempty"`
	DockerRegistryURL                    string `json:"dockerRegistryUrl,omitempty"`
	CleanupMode                          string `json:"cleanupMode,omitempty"`
	FilePath                             string `json:"filePath,omitempty"`
	IncludeLayers                        bool   `json:"includeLayers,omitempty"`
	AddSideBarLink                       bool   `json:"addSideBarLink,omitempty"`
	Verbose                              bool   `json:"verbose,omitempty"`
	ProtecodeTimeoutMinutes              string `json:"protecodeTimeoutMinutes,omitempty"`
	ProtecodeServerURL                   string `json:"protecodeServerUrl,omitempty"`
	ReportFileName                       string `json:"reportFileName,omitempty"`
	UseCallback                          bool   `json:"useCallback,omitempty"`
	FetchURL                             string `json:"fetchUrl,omitempty"`
	ProtecodeGroup                       string `json:"protecodeGroup,omitempty"`
	ReuseExisting                        bool   `json:"reuseExisting,omitempty"`
	User                                 string `json:"user,omitempty"`
	Password                             string `json:"password,omitempty"`
}

var myProtecodeExecuteScanOptions protecodeExecuteScanOptions
var protecodeExecuteScanStepConfigJSON string

// ProtecodeExecuteScanCommand Protecode is an Open Source Vulnerability Scanner that is capable of scanning binaries. It can be used to scan docker images but is supports many other programming languages especially those of the C family. You can find more details on its capabilities in the [OS3 - Open Source Software Security JAM](https://jam4.sapjam.com/groups/XgeUs0CXItfeWyuI4k7lM3/overview_page/aoAsA0k4TbezGFyOkhsXFs). For getting access to Protecode please visit the [guide](https://go.sap.corp/protecode).
func ProtecodeExecuteScanCommand() *cobra.Command {
	metadata := protecodeExecuteScanMetadata()
	var createProtecodeExecuteScanCmd = &cobra.Command{
		Use:   "protecodeExecuteScan",
		Short: "Protecode is an Open Source Vulnerability Scanner that is capable of scanning binaries. It can be used to scan docker images but is supports many other programming languages especially those of the C family. You can find more details on its capabilities in the [OS3 - Open Source Software Security JAM](https://jam4.sapjam.com/groups/XgeUs0CXItfeWyuI4k7lM3/overview_page/aoAsA0k4TbezGFyOkhsXFs). For getting access to Protecode please visit the [guide](https://go.sap.corp/protecode).",
		Long: `Protecode is an Open Source Vulnerability Scanner that is capable of scanning binaries. It can be used to scan docker images but is supports many other programming languages especially those of the C family. You can find more details on its capabilities in the [OS3 - Open Source Software Security JAM](https://jam4.sapjam.com/groups/XgeUs0CXItfeWyuI4k7lM3/overview_page/aoAsA0k4TbezGFyOkhsXFs). For getting access to Protecode please visit the [guide](https://go.sap.corp/protecode).

!!! info "New: Using protecodeExecuteScan for Docker images on JaaS"
    **This step now also works on "Jenkins as a Service (JaaS)"!**<br />
    For the JaaS use case where the execution happens in a Kubernetes cluster without access to a Docker daemon [skopeo](https://github.com/containers/skopeo) is now used silently in the background to save a Docker image retrieved from a registry.


!!! hint "Auditing findings (Triaging)"
    Triaging is now supported by the Protecode backend and also Piper does consider this information during the analysis of the scan results though product versions are not supported by Protecode. Therefore please make sure that the ` + "`" + `fileName` + "`" + ` you are providing does either contain a stable version or that it does not contain one at all. By ensuring that you are able to triage CVEs globally on the upload file's name without affecting any other artifacts scanned in the same Protecode group and as such triaged vulnerabilities will be considered during the next scan and will not fail the build anymore.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetStepName("protecodeExecuteScan")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "protecodeExecuteScan", &myProtecodeExecuteScanOptions, config.OpenPiperFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return protecodeExecuteScan(myProtecodeExecuteScanOptions)
		},
	}

	addProtecodeExecuteScanFlags(createProtecodeExecuteScanCmd)
	return createProtecodeExecuteScanCmd
}

func addProtecodeExecuteScanFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&myProtecodeExecuteScanOptions.ProtecodeExcludeCVEs, "protecodeExcludeCVEs", "[]", "DEPRECATED: Do use triaging within the Protecode UI instead")
	cmd.Flags().BoolVar(&myProtecodeExecuteScanOptions.ProtecodeFailOnSevereVulnerabilities, "protecodeFailOnSevereVulnerabilities", true, "Whether to fail the job on severe vulnerabilties or not")
	cmd.Flags().StringVar(&myProtecodeExecuteScanOptions.DockerRegistryURL, "dockerRegistryUrl", os.Getenv("PIPER_dockerRegistryUrl"), "The reference to the docker registry to scan with Protecode")
	cmd.Flags().StringVar(&myProtecodeExecuteScanOptions.CleanupMode, "cleanupMode", "binary", "Decides which parts are removed from the Protecode backend after the scan")
	cmd.Flags().StringVar(&myProtecodeExecuteScanOptions.FilePath, "filePath", os.Getenv("PIPER_filePath"), "The path to the file from local workspace to scan with Protecode")
	cmd.Flags().BoolVar(&myProtecodeExecuteScanOptions.IncludeLayers, "includeLayers", false, "Flag if the docker layers should be included")
	cmd.Flags().BoolVar(&myProtecodeExecuteScanOptions.AddSideBarLink, "addSideBarLink", true, "Whether to create a side bar link pointing to the report produced by Protecode or not")
	cmd.Flags().BoolVar(&myProtecodeExecuteScanOptions.Verbose, "verbose", false, "Whether to log verbose information or not")
	cmd.Flags().StringVar(&myProtecodeExecuteScanOptions.ProtecodeTimeoutMinutes, "protecodeTimeoutMinutes", "60", "The timeout to wait for the scan to finish")
	cmd.Flags().StringVar(&myProtecodeExecuteScanOptions.ProtecodeServerURL, "protecodeServerUrl", "https://protecode.c.eu-de-2.cloud.sap", "The URL to the Protecode backend")
	cmd.Flags().StringVar(&myProtecodeExecuteScanOptions.ReportFileName, "reportFileName", "protecode_report.pdf", "The file name of the report to be created")
	cmd.Flags().BoolVar(&myProtecodeExecuteScanOptions.UseCallback, "useCallback", false, "Whether to the Protecode backend's callback or poll for results")
	cmd.Flags().StringVar(&myProtecodeExecuteScanOptions.FetchURL, "fetchUrl", os.Getenv("PIPER_fetchUrl"), "The URL to fetch the file to scan with Protecode which must be accessible via public HTTP GET request")
	cmd.Flags().StringVar(&myProtecodeExecuteScanOptions.ProtecodeGroup, "protecodeGroup", os.Getenv("PIPER_protecodeGroup"), "The Protecode group ID of your team")
	cmd.Flags().BoolVar(&myProtecodeExecuteScanOptions.ReuseExisting, "reuseExisting", false, "Whether to reuse an existing product instead of creating a new one")
	cmd.Flags().StringVar(&myProtecodeExecuteScanOptions.User, "user", os.Getenv("PIPER_user"), "user which is used for the protecode scan")
	cmd.Flags().StringVar(&myProtecodeExecuteScanOptions.Password, "password", os.Getenv("PIPER_password"), "password which is used for the user")

	cmd.MarkFlagRequired("protecodeGroup")
	cmd.MarkFlagRequired("user")
	cmd.MarkFlagRequired("password")
}

// retrieve step metadata
func protecodeExecuteScanMetadata() config.StepData {
	var theMetaData = config.StepData{
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:      "protecodeExcludeCVEs",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "protecodeFailOnSevereVulnerabilities",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "dockerRegistryUrl",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "cleanupMode",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "filePath",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "includeLayers",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "addSideBarLink",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "verbose",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "protecodeTimeoutMinutes",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "protecodeServerUrl",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "reportFileName",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "useCallback",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "fetchUrl",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "protecodeGroup",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "reuseExisting",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "bool",
						Mandatory: false,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "user",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
					{
						Name:      "password",
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
