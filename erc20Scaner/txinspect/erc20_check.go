package txinspect

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// ERC20CheckResult mirrors scanner checkERC20BySelector diagnostics.
type ERC20CheckResult struct {
	IsERC20           bool     `json:"isERC20"`
	Error             string   `json:"error,omitempty"`
	CoreFound         string   `json:"coreSelectorsFound"`
	OptionalFound     string   `json:"optionalSelectorsFound"`
	ExtensionFound    string   `json:"extensionSelectorsFound"`
	RequiredSelectors []string `json:"-"`
}

// checkERC20BySelector follows scanner/process.go logic.
func (a *Analyzer) checkERC20BySelector(contractAddress *common.Address) ERC20CheckResult {
	out := ERC20CheckResult{}
	requiredCoreSelectors := []string{
		"18160ddd", "70a08231", "a9059cbb", "23b872dd", "095ea7b3", "dd62ed3e",
	}
	optionalSelectors := []string{"06fdde03", "95d89b41", "313ce567"}
	extensionSelectors := []string{"39509351", "a457c2d7"}
	out.RequiredSelectors = requiredCoreSelectors

	code, err := a.c.CodeAt(*contractAddress, nil)
	if err != nil {
		out.Error = fmt.Sprintf("get code: %v", err)
		return out
	}
	if len(code) == 0 {
		out.Error = "empty code at address"
		return out
	}
	codeHex := hex.EncodeToString(code)

	coreFound := 0
	for _, s := range requiredCoreSelectors {
		if strings.Contains(codeHex, s) {
			coreFound++
		}
	}
	optionalFound := 0
	for _, s := range optionalSelectors {
		if strings.Contains(codeHex, s) {
			optionalFound++
		}
	}
	extFound := 0
	for _, s := range extensionSelectors {
		if strings.Contains(codeHex, s) {
			extFound++
		}
	}
	out.CoreFound = fmt.Sprintf("%d/%d", coreFound, len(requiredCoreSelectors))
	out.OptionalFound = fmt.Sprintf("%d/%d", optionalFound, len(optionalSelectors))
	out.ExtensionFound = fmt.Sprintf("%d/%d", extFound, len(extensionSelectors))

	if coreFound < 4 {
		out.Error = fmt.Sprintf("insufficient ERC20 core functions: %d/%d", coreFound, len(requiredCoreSelectors))
		return out
	}

	_, err = a.unPackageAbi("totalSupply", contractAddress)
	if err != nil {
		out.Error = fmt.Sprintf("totalSupply call failed: %v", err)
		return out
	}
	zeroAddr := common.HexToAddress("0x0000000000000000000000000000000000000000")
	_, _ = a.unPackageAbiWithArgs("balanceOf", contractAddress, zeroAddr)

	out.IsERC20 = true
	return out
}
