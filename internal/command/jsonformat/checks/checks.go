package checks

import (
	"fmt"
	"github.com/hashicorp/terraform/internal/command/jsonchecks"
)

func RenderHuman(check jsonchecks.CheckResultStatic) (string, bool) {
	if check.Address["kind"] != "check" {
		// We only render the check blocks in the human plan.
		return "", false
	}

	return fmt.Sprintf("%s: %s", check.Address["to_display"], check.Status), true
}
