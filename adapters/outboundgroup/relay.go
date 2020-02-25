package outboundgroup

import (
	"context"
	"encoding/json"

	"github.com/Dreamacro/clash/adapters/outbound"
	"github.com/Dreamacro/clash/adapters/provider"
	"github.com/Dreamacro/clash/common/singledo"
	C "github.com/Dreamacro/clash/constant"
)

type Relay struct {
	*outbound.Base
	single    *singledo.Single
	providers []provider.ProxyProvider
}

func (r *Relay) DialContext(ctx context.Context, metadata *C.Metadata) (C.Conn, error) {
	proxies := r.proxies()
	proxyAdapterExtendeds := make([]C.ProxyAdapterExtended, len(proxies))
	for i, proxy := range proxies {
		proxyAdapterExtendeds[i] = proxy.(C.ProxyAdapterExtended)
	}

	c, err := proxyAdapterExtendeds[0].InitConn(ctx)
	if err != nil {
		return nil, err
	}

	proxyMeta := make([]C.Metadata, len(proxyAdapterExtendeds))
	for i := 0; i < len(proxyAdapterExtendeds) - 1; i++ {
		proxyMeta[i], err = proxyAdapterExtendeds[i+1].ToMetadata()
		if err != nil {
			return nil, err
		}
		c, err = proxyAdapterExtendeds[i].StreamConn(c, &proxyMeta[i])
		if err != nil {
			return nil, err
		}
	}
	c, err = proxyAdapterExtendeds[len(proxyAdapterExtendeds)-1].StreamConn(c, metadata)
	if err != nil {
		return nil, err
	}

	cc := outbound.NewConn(c, proxyAdapterExtendeds[0])
	cc.AppendToChains(r)
	return cc, nil
}

func (r *Relay) SupportUDP() bool {
	return false
}

func (r *Relay) MarshalJSON() ([]byte, error) {
	var all []string
	for _, proxy := range r.proxies() {
		all = append(all, proxy.Name())
	}
	return json.Marshal(map[string]interface{}{
		"type": r.Type().String(),
		"all":  all,
	})
}

func (r *Relay) proxies() []C.Proxy {
	elm, _, _ := r.single.Do(func() (interface{}, error) {
		return getProvidersProxies(r.providers), nil
	})

	return elm.([]C.Proxy)
}

func NewRelay(name string, providers []provider.ProxyProvider) *Relay {
	return &Relay{
		Base:      outbound.NewBase(name, C.Relay, false),
		single:    singledo.NewSingle(defaultGetProxiesDuration),
		providers: providers,
	}
}
