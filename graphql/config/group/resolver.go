package group

import (
	"github.com/v2rayA/dae-wing/graphql/config"
	daeConfig "github.com/v2rayA/dae/config"
)

type Resolver struct {
	Group *daeConfig.Group
}

func (r *Resolver) Name() string {
	return r.Group.Name
}

func (r *Resolver) Filter() *config.AndFunctionsResolver {
	return &config.AndFunctionsResolver{
		AndFunctions: r.Group.Filter,
	}
}

func (r *Resolver) Policy() *config.AndFunctionsOrPlaintextResolver {
	return &config.AndFunctionsOrPlaintextResolver{
		FunctionListOrString: r.Group.Policy,
	}
}
