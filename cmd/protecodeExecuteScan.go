package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	pkgutil "github.com/GoogleContainerTools/container-diff/pkg/util"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/protecode"
)

func protecodeExecuteScan(config protecodeExecuteScanOptions, cpEnvironment *protecodeExecuteScanCommonPipelineEnvironment, influx *protecodeExecuteScanInflux) error {
	c := command.Command{}
	// reroute command output to loging framework
	c.Stdout(log.Entry().Writer())
	c.Stderr(log.Entry().Writer())

	return runProtecodeScan(config, cpEnvironment, influx)
}

func runProtecodeScan(config protecodeExecuteScanOptions, cpEnvironment *protecodeExecuteScanCommonPipelineEnvironment, influx *protecodeExecuteScanInflux) error {

	//create client for sending api request
	client := createClient(config)

	getDockerImage(config, cpEnvironment)

	parsedResult, err := executeProtecodeScan(client, config, writeReportToFile)
	if err != nil {
		return err
	}

	setInfluxData(influx, parsedResult)

	setCommonPipelineEnvironmentData(cpEnvironment, parsedResult)

	return nil
}

func getDockerImage(config protecodeExecuteScanOptions, cpEnvironment *protecodeExecuteScanCommonPipelineEnvironment) error {

	cachePath := "./cache"
	completeUrl, err := getUrlAndFileNameFromDockerImage(config, cpEnvironment)
	if err != nil {
		return err
	}

	image, err := pkgutil.GetImage(completeUrl, config.IncludeLayers, cachePath)
	if err != nil {
		return err
	}

	if len(config.FilePath) <= 0 {
		fileName := fmt.Sprintf("%v.tar", strings.ReplaceAll(config.ScanImage, "/", "_"))
		config.FilePath = filepath.Join(image.FSPath, fileName)
	}

	return nil
}

func getUrlAndFileNameFromDockerImage(config protecodeExecuteScanOptions, cpEnvironment *protecodeExecuteScanCommonPipelineEnvironment) (string, error) {

	if len(config.ScanImage) <= 0 {
		config.ScanImage = cpEnvironment.container.imageNameTag
	}
	if len(config.DockerRegistryURL) <= 0 {
		config.DockerRegistryURL = fmt.Sprintf("%v://%v", config.DockerRegistryProtocol, cpEnvironment.container.registryURL)
	}

	completeUrl := config.ScanImage

	if len(config.DockerRegistryURL) > 0 {

		if strings.HasSuffix(config.DockerRegistryURL, "/") {
			completeUrl = fmt.Sprintf("%v%v", config.DockerRegistryURL, config.ScanImage)
		} else {
			completeUrl = fmt.Sprintf("%v/%v", config.DockerRegistryURL, config.ScanImage)
		}
	}

	if len(completeUrl) <= 0 {
		return completeUrl, errors.New("Protecode scan failed, there is no scan image configured")
	}

	return completeUrl, nil
}

func executeProtecodeScan(client protecode.Protecode, config protecodeExecuteScanOptions, writeReportToFile func(resp io.ReadCloser, reportFileName string) error) (map[string]int, error) {

	var parsedResult map[string]int = make(map[string]int)

	//load existing product by filename
	productId, err := client.LoadExistingProduct(config.ProtecodeGroup, config.FilePath, config.ReuseExisting)
	if err != nil {
		return parsedResult, err
	}

	// check if no existing is found or reuse existing is false
	productId, err = uploadScanOrDeclareFetch(config, productId, client)
	if err != nil {
		return parsedResult, err
	}
	if productId <= 0 {
		return parsedResult, errors.New("Protecode scan failed, the product id is not valid (product id <= zero)")
	}
	//pollForResult
	result, err := client.PollForResult(productId, config.Verbose)
	if err != nil {
		return parsedResult, err
	}
	//check if result is ok else notify
	if len(result.Status) > 0 && result.Status == "F" {
		log.Entry().Fatal("Protecode scan failed, please check the log and protecode backend for more details.")
		return parsedResult, errors.New("Protecode scan failed, please check the log and protecode backend for more details.")
	}

	//loadReport
	resp, err := client.LoadReport(config.ReportFileName, productId)
	if err != nil {
		return parsedResult, err
	}

	//save report to filesystem
	err = writeReportToFile(*resp, config.ReportFileName)
	if err != nil {
		return parsedResult, err
	}

	//clean scan from server
	err = client.DeleteScan(config.CleanupMode, productId)
	if err != nil {
		return parsedResult, err
	}

	//count vulnerabilities
	parsedResult = client.ParseResultForInflux(result, config.ProtecodeExcludeCVEs)

	return parsedResult, nil
}

func setInfluxData(influx *protecodeExecuteScanInflux, result map[string]int) {

	influx.protecode_data.fields.historical_vulnerabilities = fmt.Sprintf("%v", result["historical_vulnerabilities"])
	influx.protecode_data.fields.historical_vulnerabilities = fmt.Sprintf("%v", result["triaged_vulnerabilities"])
	influx.protecode_data.fields.historical_vulnerabilities = fmt.Sprintf("%v", result["excluded_vulnerabilities"])
	influx.protecode_data.fields.historical_vulnerabilities = fmt.Sprintf("%v", result["minor_vulnerabilities"])
	influx.protecode_data.fields.historical_vulnerabilities = fmt.Sprintf("%v", result["major_vulnerabilities"])
	influx.protecode_data.fields.historical_vulnerabilities = fmt.Sprintf("%v", result["vulnerabilities"])
}

func setCommonPipelineEnvironmentData(cpEnvironment *protecodeExecuteScanCommonPipelineEnvironment, result map[string]int) {

	cpEnvironment.appContainerProperties.protecodeCount = fmt.Sprintf("%v", result["count"])
	cpEnvironment.appContainerProperties.cvss2GreaterOrEqualSeven = fmt.Sprintf("%v", result["cvss2GreaterOrEqualSeven"])
	cpEnvironment.appContainerProperties.cvss3GreaterOrEqualSeven = fmt.Sprintf("%v", result["cvss3GreaterOrEqualSeven"])
	cpEnvironment.appContainerProperties.excluded_vulnerabilities = fmt.Sprintf("%v", result["excluded_vulnerabilities"])
	cpEnvironment.appContainerProperties.triaged_vulnerabilities = fmt.Sprintf("%v", result["triaged_vulnerabilities"])
	cpEnvironment.appContainerProperties.historical_vulnerabilities = fmt.Sprintf("%v", result["historical_vulnerabilities"])
}

func createClient(config protecodeExecuteScanOptions) protecode.Protecode {

	var duration time.Duration = time.Duration(10 * 60)

	if len(config.ProtecodeTimeoutMinutes) > 0 {
		s, _ := strconv.ParseInt(config.ProtecodeTimeoutMinutes, 10, 64)
		duration = time.Duration(s * 60)
	}

	pc := protecode.Protecode{}

	protecodeOptions := protecode.ProtecodeOptions{
		ServerURL: config.ProtecodeServerURL,
		Logger:    log.Entry().WithField("package", "SAP/jenkins-library/pkg/protecode"),
		Duration:  duration,
		Username:  config.User,
		Password:  config.Password,
	}

	pc.SetOptions(protecodeOptions)

	return pc
}

func uploadScanOrDeclareFetch(config protecodeExecuteScanOptions, productId int, client protecode.Protecode) (int, error) {

	// check if no existing is found or reuse existing is false
	if productId <= 0 || !config.ReuseExisting {
		if len(config.FetchURL) > 0 {
			fmt.Printf("triggering Protecode scan - url: %v, group: %v", config.FetchURL, config.ProtecodeGroup)
			result, err := client.DeclareFetchUrl(config.CleanupMode, config.ProtecodeGroup, config.FetchURL)
			if err != nil {
				return -1, err
			}
			productId = result.ProductId

		} else {
			fmt.Printf("triggering Protecode scan - file: %v, group: %v", config.FilePath, config.ProtecodeGroup)
			result, err := client.UploadScanFile(config.CleanupMode, config.ProtecodeGroup, config.FilePath)
			if err != nil {
				return -1, err
			}
			productId = result.ProductId
		}
	}

	return productId, nil
}

func writeReportToFile(resp io.ReadCloser, reportFileName string) error {
	f, err := os.Create(reportFileName)
	if err == nil {
		defer f.Close()
		_, err = io.Copy(f, resp)
	}

	return err
}
