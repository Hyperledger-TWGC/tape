package trafficGenerator

import (
	"context"

	"github.com/open-policy-agent/opa/rego"
)

func Check(input []string) (bool, error) {
	// Create query that returns a single boolean value.
	rego := rego.New(
		rego.Query("data.tape.allow"),
		rego.Module("",
			`package tape

default allow = false
allow {
	input[_] == "org1"
}

allow {
	input[_] == "org2"
}

`,
		),
		rego.Input(input),
	)
	rs, err := rego.Eval(context.Background())
	if err != nil {
		return false, err
	}
	return rs.Allowed(), nil
}
