package workflow

// NodeDefinition defines a single node in the workflow.
type NodeDefinition struct {
	Code        string   // unique node code
	Name        string   // human-readable name
	DeptCode    string   // responsible department code
	CanReject   bool     // whether this node can reject (return to previous)
	RejectTarget string  // node to return to on rejection (empty = previous node)
	NextNode    string   // the next node after approval (empty = terminal)
}

// GetDefinition returns the full 16-node workflow definition.
// The order matters — nodes are indexed by position in the slice.
func GetDefinition() []NodeDefinition {
	return []NodeDefinition{
		{
			Code:      NodeTaskCreate,
			Name:      "任务创建",
			DeptCode:  DeptBusiness,
			CanReject: false,
			NextNode:  NodeContractReview,
		},
		{
			Code:         NodeContractReview,
			Name:         "合同评审",
			DeptCode:     DeptTech,
			CanReject:    true,
			RejectTarget: NodeTaskCreate,
			NextNode:     NodeQCTask,
		},
		{
			Code:         NodeQCTask,
			Name:         "质控任务",
			DeptCode:     DeptQC, // 质控室（与业务流程图一致）
			CanReject:    true,
			RejectTarget: NodeContractReview,
			NextNode:     NodeSamplingSchedule,
		},
		{
			Code:         NodeSamplingSchedule,
			Name:         "采样调度",
			DeptCode:     DeptField,
			CanReject:    true,
			RejectTarget: NodeQCTask,
			NextNode:     NodeFieldSampling,
		},
		{
			Code:         NodeFieldSampling,
			Name:         "现场采样",
			DeptCode:     DeptField,
			CanReject:    true,
			RejectTarget: NodeSamplingSchedule,
			NextNode:     NodeSampleReceiving,
		},
		{
			Code:         NodeSampleReceiving,
			Name:         "样品接收",
			DeptCode:     DeptSample,
			CanReject:    true,
			RejectTarget: NodeFieldSampling,
			NextNode:     NodeTaskAssign,
		},
		{
			Code:         NodeTaskAssign,
			Name:         "任务分配",
			DeptCode:     DeptLab,
			CanReject:    true,
			RejectTarget: NodeSampleReceiving,
			NextNode:     NodeDataEntry,
		},
		{
			Code:         NodeDataEntry,
			Name:         "数据录入",
			DeptCode:     DeptLab,
			CanReject:    true,
			RejectTarget: NodeTaskAssign,
			NextNode:     NodeDataReview,
		},
		{
			Code:         NodeDataReview,
			Name:         "数据复核",
			DeptCode:     DeptLab,
			CanReject:    true,
			RejectTarget: NodeDataEntry,
			NextNode:     NodeDataAudit,
		},
		{
			Code:         NodeDataAudit,
			Name:         "数据审核",
			DeptCode:     DeptLab,
			CanReject:    true,
			RejectTarget: NodeDataReview,
			NextNode:     NodeReportPrepare,
		},
		{
			Code:         NodeReportPrepare,
			Name:         "报告编制",
			DeptCode:     DeptReport,
			CanReject:    true,
			RejectTarget: NodeDataAudit,
			NextNode:     NodeReportReview,
		},
		{
			Code:         NodeReportReview,
			Name:         "报告复核",
			DeptCode:     DeptLab,
			CanReject:    true,
			RejectTarget: NodeReportPrepare,
			NextNode:     NodeReportAudit,
		},
		{
			Code:         NodeReportAudit,
			Name:         "报告审核",
			DeptCode:     DeptQC,
			CanReject:    true,
			RejectTarget: NodeReportReview,
			NextNode:     NodeReportSign,
		},
		{
			Code:         NodeReportSign,
			Name:         "报告签发",
			DeptCode:     DeptTech,
			CanReject:    true,
			RejectTarget: NodeReportAudit,
			NextNode:     NodeReportPrint,
		},
		{
			Code:         NodeReportPrint,
			Name:         "报告发放",
			DeptCode:     DeptBusiness,
			CanReject:    true,
			RejectTarget: NodeReportSign,
			NextNode:     NodeProjectArchive,
		},
		{
			Code:      NodeProjectArchive,
			Name:      "项目归档",
			DeptCode:  DeptReport,
			CanReject: false,
			// NextNode is empty — terminal node
		},
	}
}

// BuildNodeMap returns a map of node code -> NodeDefinition for fast lookup.
func BuildNodeMap() map[string]NodeDefinition {
	index := make(map[string]NodeDefinition, 16)
	for _, n := range GetDefinition() {
		index[n.Code] = n
	}
	return index
}

// FindPreviousNode returns the node code immediately before the given node.
// Returns empty string if the node is the first one.
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

// GetNodeIndex returns the position (0-based) of a node in the workflow.
func GetNodeIndex(code string) int {
	for i, n := range GetDefinition() {
		if n.Code == code {
			return i
		}
	}
	return -1
}