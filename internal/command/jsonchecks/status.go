package jsonchecks

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform/internal/checks"
)

type CheckStatus string

func checkStatusForJSON(s checks.Status) CheckStatus {
	if ret, ok := checkStatuses[s]; ok {
		return ret
	}
	panic(fmt.Sprintf("unsupported check status %#v", s))
}

func (s CheckStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

var checkStatuses = map[checks.Status]CheckStatus{
	checks.StatusPass:    CheckStatus("pass"),
	checks.StatusFail:    CheckStatus("fail"),
	checks.StatusError:   CheckStatus("error"),
	checks.StatusUnknown: CheckStatus("unknown"),
}

var CheckStatusesReverse = map[CheckStatus]checks.Status{
	CheckStatus("pass"):    checks.StatusPass,
	CheckStatus("fail"):    checks.StatusFail,
	CheckStatus("error"):   checks.StatusError,
	CheckStatus("unknown"): checks.StatusUnknown,
}
