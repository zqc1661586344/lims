package workflow

type NodeDefinition struct {
	Code         string
	Name         string
	DeptCode     string
	RoleHint     string
	CanReject    bool
	RejectTarget string
	NextNode     string
}

func GetDefinition() []NodeDefinition {
	return []NodeDefinition{
		{
			Code:      NodeTaskCreate,
			Name:      "任务创建",
			DeptCode:  DeptBusiness,
			RoleHint:  "business_manager",
			CanReject: false,
			NextNode:  NodeContractReview,
		},
		{
			Code:         NodeContractReview,
			Name:         "合同评审",
			DeptCode:     DeptTech,
			RoleHint:     "contract_reviewer",
			CanReject:    true,
			RejectTarget: NodeTaskCreate,
			NextNode:     NodeQCTask,
		},
		{
			Code:         NodeQCTask,
			Name:         "质控任务",
			DeptCode:     DeptQC,
			RoleHint:     "qc_staff",
			CanReject:    true,
			RejectTarget: NodeContractReview,
			NextNode:     NodeSamplingSchedule,
		},
		{
			Code:         NodeSamplingSchedule,
			Name:         "采样调度",
			DeptCode:     DeptField,
			RoleHint:     "sampler",
			CanReject:    true,
			RejectTarget: NodeQCTask,
			NextNode:     NodeFieldSampling,
		},
		{
			Code:         NodeFieldSampling,
			Name:         "现场采样",
			DeptCode:     DeptField,
			RoleHint:     "sampler",
			CanReject:    true,
			RejectTarget: NodeSamplingSchedule,
			NextNode:     NodeSampleReceiving,
		},
		{
			Code:         NodeSampleReceiving,
			Name:         "样品接收",
			DeptCode:     DeptSample,
			RoleHint:     "sample_manager",
			CanReject:    true,
			RejectTarget: NodeFieldSampling,
			NextNode:     NodeTaskAssign,
		},
		{
			Code:         NodeTaskAssign,
			Name:         "任务分配",
			DeptCode:     DeptLab,
			RoleHint:     "lab_technician",
			CanReject:    true,
			RejectTarget: NodeSampleReceiving,
			NextNode:     NodeDataEntry,
		},
		{
			Code:         NodeDataEntry,
			Name:         "数据录入",
			DeptCode:     DeptLab,
			RoleHint:     "lab_technician",
			CanReject:    true,
			RejectTarget: NodeTaskAssign,
			NextNode:     NodeDataReview,
		},
		{
			Code:         NodeDataReview,
			Name:         "数据复核",
			DeptCode:     DeptLab,
			RoleHint:     "data_reviewer",
			CanReject:    true,
			RejectTarget: NodeDataEntry,
			NextNode:     NodeDataAudit,
		},
		{
			Code:         NodeDataAudit,
			Name:         "数据审核",
			DeptCode:     DeptLab,
			RoleHint:     "data_auditor",
			CanReject:    true,
			RejectTarget: NodeDataReview,
			NextNode:     NodeReportPrepare,
		},
		{
			Code:         NodeReportPrepare,
			Name:         "报告编制",
			DeptCode:     DeptReport,
			RoleHint:     "report_preparer",
			CanReject:    true,
			RejectTarget: NodeDataAudit,
			NextNode:     NodeReportReview,
		},
		{
			Code:         NodeReportReview,
			Name:         "报告复核",
			DeptCode:     DeptLab,
			RoleHint:     "report_reviewer",
			CanReject:    true,
			RejectTarget: NodeReportPrepare,
			NextNode:     NodeReportAudit,
		},
		{
			Code:         NodeReportAudit,
			Name:         "报告审核",
			DeptCode:     DeptQC,
			RoleHint:     "report_auditor",
			CanReject:    true,
			RejectTarget: NodeReportReview,
			NextNode:     NodeReportSign,
		},
		{
			Code:         NodeReportSign,
			Name:         "报告签发",
			DeptCode:     DeptTech,
			RoleHint:     "authorized_signer",
			CanReject:    true,
			RejectTarget: NodeReportAudit,
			NextNode:     NodeReportPrint,
		},
		{
			Code:         NodeReportPrint,
			Name:         "报告发放",
			DeptCode:     DeptBusiness,
			RoleHint:     "business_manager",
			CanReject:    true,
			RejectTarget: NodeReportSign,
			NextNode:     NodeProjectArchive,
		},
		{
			Code:      NodeProjectArchive,
			Name:      "项目归档",
			DeptCode:  DeptReport,
			RoleHint:  "archive_manager",
			CanReject: false,
		},
	}
}

func BuildNodeMap() map[string]NodeDefinition {
	index := make(map[string]NodeDefinition, 16)
	for _, n := range GetDefinition() {
		index[n.Code] = n
	}
	return index
}

func FindPreviousNode(code string) string {
	defs := GetDefinition()
	for i, n := range defs {
		if n.Code == code {
			if i > 0 {
				return defs[i-1].Code
			}
			return ""
		}
	}
	return ""
}

func GetNodeIndex(code string) int {
	for i, n := range GetDefinition() {
		if n.Code == code {
			return i
		}
	}
	return -1
}
