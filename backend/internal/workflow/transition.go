package workflow

// TransitionRule defines whether a transition from one node to another is valid.
type TransitionRule struct {
	From        string
	To          string
	IsForward   bool  // forward (approve) or backward (reject)
}

// IsValidTransition checks if a transition from `from` to `to` is allowed.
func IsValidTransition(from, to string, defs []NodeDefinition) bool {
	for _, rule := range GetAllowedTransitions(defs) {
		if rule.From == from && rule.To == to {
			return true
		}
	}
	return false
}

// GetAllowedTransitions computes all valid transitions from the node definitions.
// For each node:
//   - Forward: current node -> next node (if NextNode is non-empty)
//   - Backward: current node -> reject target (if CanReject and RejectTarget is non-empty)
func GetAllowedTransitions(defs []NodeDefinition) []TransitionRule {
	var rules []TransitionRule
	nodeMap := make(map[string]NodeDefinition)
	for _, n := range defs {
		nodeMap[n.Code] = n
	}
	for _, n := range defs {
		if n.NextNode != "" {
			rules = append(rules, TransitionRule{
				From:      n.Code,
				To:        n.NextNode,
				IsForward: true,
			})
		}
		if n.CanReject && n.RejectTarget != "" {
			rules = append(rules, TransitionRule{
				From:      n.Code,
				To:        n.RejectTarget,
				IsForward: false,
			})
		}
	}
	return rules
}

// GetNextApprovalNode returns the next node after approval.
func GetNextApprovalNode(currentCode string, defs []NodeDefinition) string {
	for _, n := range defs {
		if n.Code == currentCode {
			return n.NextNode
		}
	}
	return ""
}

// GetRejectTarget returns the node to return to on rejection.
func GetRejectTarget(currentCode string, defs []NodeDefinition) string {
	for _, n := range defs {
		if n.Code == currentCode {
			return n.RejectTarget
		}
	}
	return ""
}