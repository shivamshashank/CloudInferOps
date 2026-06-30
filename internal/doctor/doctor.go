package doctor

import (
	"fmt"
	"github.com/shivamshashank/CloudInferOps/internal/utils"
)

type CheckStatus int

const (
	StatusOK CheckStatus = iota
	StatusWarn
	StatusError
	StatusInfo
)

type CheckResult struct {
	Name    string
	Status  CheckStatus
	Message string
}

// Report contains all check results and overall status
type Report struct {
	Results   []CheckResult
	HasErrors bool
	HasK8s    bool
}

// Print prints a stylized doctor report to the terminal
func (r *Report) Print() {
	fmt.Println(utils.ColorBold + "🩺  CloudInferOps Doctor" + utils.ColorReset)
	fmt.Println()

	for _, res := range r.Results {
		var prefix string
		switch res.Status {
		case StatusOK:
			prefix = utils.PrefixOK
		case StatusWarn:
			prefix = utils.PrefixWarn
		case StatusError:
			prefix = utils.PrefixError
		case StatusInfo:
			prefix = utils.PrefixInfo
		}
		fmt.Printf("%s%s\n", prefix, res.Message)
	}
	fmt.Println()

	if r.HasErrors {
		fmt.Printf("%sSome critical prerequisites failed. Please resolve them before proceeding.\n", utils.PrefixError)
	} else if r.HasK8s {
		fmt.Printf("%sRun: sudo cloudinferops deploy observability\n", utils.PrefixReady)
	} else {
		fmt.Printf("%sKubernetes cluster not detected.\n", utils.PrefixWarn)
		fmt.Printf("%sRun: sudo cloudinferops deploy observability (which can automatically set up a local cluster for you) or configure an existing cluster.\n", utils.PrefixInfo)
	}
}

// RunDoctor aggregates and executes all pre-flight diagnostic checks
func RunDoctor() *Report {
	report := &Report{}

	// System checks
	report.Results = append(report.Results, CheckOS())
	report.Results = append(report.Results, CheckInternet())

	// Tool checks
	report.Results = append(report.Results, CheckTool("kubectl", false))
	report.Results = append(report.Results, CheckTool("helm", false))

	// Hardware checks
	report.Results = append(report.Results, CheckMemory())
	report.Results = append(report.Results, CheckCPU())
	report.Results = append(report.Results, CheckDisk())

	// K8s specific checks
	k8sResults, hasK8s := CheckK8sCluster()
	report.Results = append(report.Results, k8sResults...)
	report.HasK8s = hasK8s

	if hasK8s {
		report.Results = append(report.Results, CheckK8sVersion())
		report.Results = append(report.Results, CheckIngressController())
	}

	// Determine if we have fatal errors
	for _, res := range report.Results {
		if res.Status == StatusError {
			report.HasErrors = true
			break
		}
	}

	return report
}
