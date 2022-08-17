package proofconfig

type None struct {
	*configDB
}

// IsHaveProofPermission check Permission
func (n *None) IsHaveProofPermission(_ string) bool {
	return true
}

// IsHaveDelProofPermission check  DelProofPermission
func (n *None) IsHaveDelProofPermission(_, _, _ string) bool {
	return true
}

func (n *None) GetOrganizationName(addr string) (string, error) {
	m, err := n.configDB.GetMember(addr)
	if err == nil {
		return m.Organization, nil
	}
	return "system", nil
}
