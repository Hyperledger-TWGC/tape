package trafficGenerator

import (
	"context"

	"github.com/hyperledger-twgc/tape/pkg/infra/basic"
	"github.com/open-policy-agent/opa/rego"
)

func CheckPolicy(input *basic.Elements, rule string) (bool, error) {
	if input.Processed {
		return false, nil
	}
	rego := rego.New(
		rego.Query("data.tape.allow"),
		rego.Module("", rule),
		rego.Input(input.Orgs),
	)
	rs, err := rego.Eval(context.Background())
	if err != nil {
		return false, err
	}
	input.Processed = rs.Allowed()
	return rs.Allowed(), nil
}
