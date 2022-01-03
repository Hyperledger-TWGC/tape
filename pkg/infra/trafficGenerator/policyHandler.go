package trafficGenerator

import (
	"context"

	"github.com/open-policy-agent/opa/rego"
)

func Check(input []string, rule string) (bool, error) {
	rego := rego.New(
		rego.Query("data.tape.allow"),
		rego.Module("", rule),
		rego.Input(input),
	)
	rs, err := rego.Eval(context.Background())
	if err != nil {
		return false, err
	}
	return rs.Allowed(), nil
}
