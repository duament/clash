package outboundgroup

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/Dreamacro/clash/adapters/outbound"
	"github.com/Dreamacro/clash/adapters/provider"
	"github.com/Dreamacro/clash/common/singledo"
	"github.com/Dreamacro/clash/component/dialer"
	C "github.com/Dreamacro/clash/constant"
)

type Relay struct {
	*outbound.Base
	single    *singledo.Single
	providers []provider.ProxyProvider
}

func (r *Relay) DialContext(ctx context.Context, metadata *C.Metadata) (C.Conn, error) {
	proxies := r.proxies()
	first := proxies[0]
	last := proxies[len(proxies)-1]

	c, err := dialer.DialContext(ctx, "tcp", first.Addr())
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %w", first.Addr(), err)
	}

	var currentMeta *C.Metadata
	for _, proxy := range proxies[1:len(proxies)] {
		currentMeta, err = addrToMetadata(proxy.Addr())
		if err != nil {
			return nil, err
		}

		c, err = first.StreamConn(c, currentMeta)
		if err != nil {
			return nil, fmt.Errorf("%s connect error: %w", first.Addr(), err)
		}

		first = proxy
	}

	c, err = last.StreamConn(c, metadata)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %w", last.Addr(), err)
	}

	cc := outbound.NewConn(c, proxies[0])
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

func addrToMetadata(rawAddress string) (addr *C.Metadata, err error) {
	host, port, err := net.SplitHostPort(rawAddress)
	if err != nil {
		err = fmt.Errorf("addrToMetadata failed: %w", err)
		return
	}

	ip := net.ParseIP(host)
	if ip != nil {
		if ip.To4() != nil {
			addr = &C.Metadata{
				AddrType: C.AtypIPv4,
				Host:     "",
				DstIP:    ip,
				DstPort:  port,
			}
			return
		} else {
			addr = &C.Metadata{
				AddrType: C.AtypIPv6,
				Host:     "",
				DstIP:    ip,
				DstPort:  port,
			}
			return
		}
	} else {
		addr = &C.Metadata{
			AddrType: C.AtypDomainName,
			Host:     host,
			DstIP:    nil,
			DstPort:  port,
		}
		return
	}
}
