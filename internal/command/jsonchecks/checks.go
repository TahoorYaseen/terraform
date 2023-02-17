package jsonchecks

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform/internal/states"
)

func MarshalForRenderer(results *states.CheckResults) []CheckResultStatic {
	jsonResults := make([]CheckResultStatic, 0, results.ConfigResults.Len())

	for _, elem := range results.ConfigResults.Elems {
		staticAddr := elem.Key
		aggrResult := elem.Value

		objects := make([]CheckResultDynamic, 0, aggrResult.ObjectResults.Len())
		for _, elem := range aggrResult.ObjectResults.Elems {
			dynamicAddr := elem.Key
			result := elem.Value

			problems := make([]CheckProblem, 0, len(result.FailureMessages))
			for _, msg := range result.FailureMessages {
				problems = append(problems, CheckProblem{
					Message: msg,
				})
			}
			sort.Slice(problems, func(i, j int) bool {
				return problems[i].Message < problems[j].Message
			})

			objects = append(objects, CheckResultDynamic{
				Address:  makeDynamicObjectAddr(dynamicAddr),
				Status:   checkStatusForJSON(result.Status),
				Problems: problems,
			})
		}

		sort.Slice(objects, func(i, j int) bool {
			return objects[i].Address["to_display"].(string) < objects[j].Address["to_display"].(string)
		})

		jsonResults = append(jsonResults, CheckResultStatic{
			Address:   makeStaticObjectAddr(staticAddr),
			Status:    checkStatusForJSON(aggrResult.Status),
			Instances: objects,
		})
	}

	sort.Slice(jsonResults, func(i, j int) bool {
		return jsonResults[i].Address["to_display"].(string) < jsonResults[j].Address["to_display"].(string)
	})

	return jsonResults
}

// Marshal is the main entry-point for this package, which takes
// the top-level model object for checks in state and plan, and returns a
// JSON representation of it suitable for use in public integration points.
func Marshal(results *states.CheckResults) []byte {
	ret, err := json.Marshal(MarshalForRenderer(results))
	if err != nil {
		// We totally control the input to json.Marshal, so any error here
		// is a bug in the code above.
		panic(fmt.Sprintf("invalid input to json.Marshal: %s", err))
	}
	return ret
}

// CheckResultStatic is the container for the static, configuration-driven
// idea of "checkable object" -- a resource block with conditions, for example --
// which ensures that we can always say _something_ about each checkable
// object in the configuration even if Terraform Core encountered an error
// before being able to determine the dynamic instances of the checkable object.
type CheckResultStatic struct {
	ExperimentalNote experimentalNote `json:"//"`

	// Address is the address of the checkable object this result relates to.
	Address StaticObjectAddr `json:"address"`

	// Status is the aggregate status for all of the dynamic objects belonging
	// to this static object.
	Status CheckStatus `json:"status"`

	// Instances contains the results for each individual dynamic object that
	// belongs to this static object.
	Instances []CheckResultDynamic `json:"instances,omitempty"`
}

// CheckResultDynamic describes the check result for a dynamic object, which
// results from Terraform Core evaluating the "expansion" (e.g. count or for_each)
// of the containing object or its own containing module(s).
type CheckResultDynamic struct {
	// Address augments the Address of the containing CheckResultStatic with
	// instance-specific extra properties or overridden properties.
	Address DynamicObjectAddr `json:"address"`

	// Status is the status for this specific dynamic object.
	Status CheckStatus `json:"status"`

	// Problems describes some optional details associated with a failure
	// status, describing what fails.
	//
	// This does not include the errors for status "error", because Terraform
	// Core emits those separately as normal diagnostics. However, if a
	// particular object has a mixture of conditions that failed and conditions
	// that were invalid then status can be "error" while simultaneously
	// returning problems in this property.
	Problems []CheckProblem `json:"problems,omitempty"`
}

// CheckProblem describes one of potentially several problems that led to
// a check being classified as status "fail".
type CheckProblem struct {
	// Message is the condition error message provided by the author.
	Message string `json:"message"`

	// We don't currently have any other problem-related data, but this is
	// intentionally an object to allow us to add other data over time, such
	// as the source location where the failing condition was defined.
}

type experimentalNote struct{}

func (n experimentalNote) MarshalJSON() ([]byte, error) {
	return []byte(`"EXPERIMENTAL: see docs for details"`), nil
}
