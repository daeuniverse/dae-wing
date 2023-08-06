package daemsg

import (
	"github.com/daeuniverse/dae-wing/graphql/scalar"
	"github.com/daeuniverse/dae/component/outbound/dialer"
	"github.com/daeuniverse/dae/control"
	"github.com/graph-gophers/graphql-go"
)

type Resolver struct {
	*control.Msg
}

func (r *Resolver) Type() string {
	return string(r.Msg.Type)
}

func (r *Resolver) Timestamp() graphql.Time {
	return graphql.Time{
		Time: r.Msg.Timestamp,
	}
}

func (r *Resolver) CheckResult() *CheckResultResolver {
	if cr, ok := r.Msg.Body.(*dialer.CheckResult); ok {
		return &CheckResultResolver{
			CheckResult: cr,
		}
	}
	return nil
}

type CheckResultResolver struct {
	*dialer.CheckResult
}

func (r *CheckResultResolver) DialerProperty() *dialer.Property {
	return r.CheckResult.DialerProperty
}

func (r *CheckResultResolver) CheckType() *NetworkTypeResolver {
	return &NetworkTypeResolver{
		NetworkType: r.CheckResult.CheckType,
	}
}

func (r *CheckResultResolver) Latency() scalar.Int64 {
	return scalar.Int64{
		Int64: r.CheckResult.Latency,
	}
}

func (r *CheckResultResolver) Alive() bool {
	return r.CheckResult.Alive
}

func (r *CheckResultResolver) Error() *string {
	if r.CheckResult.Err == nil {
		return nil
	}
	res := r.CheckResult.Err.Error()
	return &res
}

type NetworkTypeResolver struct {
	*dialer.NetworkType
}

func (r *NetworkTypeResolver) L4Proto() string {
	return string(r.NetworkType.L4Proto)
}

func (r *NetworkTypeResolver) IpVersion() string {
	return string(r.NetworkType.IpVersion)
}

func (r *NetworkTypeResolver) IsDns() bool {
	return r.NetworkType.IsDns
}
