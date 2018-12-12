package resolver

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/resolver"
)

const (
	Scheme = "service"
)

type ConsulResolver struct {
	consulBuilder *ConsulResolverBuilder
	lock          sync.RWMutex
	target        resolver.Target
	cc            resolver.ClientConn
	// consul        *consulapi.Client
	addr          chan []resolver.Address
	done          chan struct{}
	watchInterval time.Duration
}

type ConsulResolverBuilder struct {
	Address       string
	Client        *consulapi.Client
	ServiceName   string
	Tag           string
	passingOnly   bool
	opts          *consulapi.QueryOptions
	WatchInterval time.Duration
	// ConsulClientConfig *api.Config
}

func (r *ConsulResolver) ResolveNow(resolver.ResolveNowOption) {
	r.resolve()
}

func (r *ConsulResolver) Close() {
	close(r.done)
}

func (r *ConsulResolver) updater() {
	for {
		select {
		case addrs := <-r.addr:
			r.cc.NewAddress(addrs)
		case <-r.done:
			return
		}
	}
}

func (r *ConsulResolver) watcher() {
	ticker := time.NewTicker(r.watchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.resolve()
		case <-r.done:
			return
		}
	}
}

func (r *ConsulResolver) resolve() {
	fmt.Println(r.target.Scheme)
	fmt.Println(r.target.Authority)
	fmt.Println(r.target.Endpoint)
	r.lock.Lock()
	defer r.lock.Unlock()

	services, _, err := r.consulBuilder.Client.Health().Service(r.consulBuilder.ServiceName, r.consulBuilder.Tag, r.consulBuilder.passingOnly, r.consulBuilder.opts)
	if err != nil {
		return
	}

	addresses := make([]resolver.Address, 0, len(services))

	for _, s := range services {
		address := s.Service.Address
		port := s.Service.Port

		if address == "" {
			address = s.Service.Address
		}

		addresses = append(addresses, resolver.Address{
			Addr:       address + ":" + strconv.Itoa(port),
			ServerName: r.target.Endpoint,
		})
	}
	fmt.Println(addresses)

	r.addr <- addresses
}
func NewConsulBuilder(address string, tag string, watchinterval time.Duration) error {
	consulcfg := consulapi.DefaultConfig()
	consulcfg.Address = address

	consul, err := consulapi.NewClient(consulcfg)
	if err != nil {
		return err
	}
	builder := &ConsulResolverBuilder{
		Address: address,
		Client:  consul,
		// ServiceName:   servicename,
		Tag:           tag,
		passingOnly:   true,
		WatchInterval: watchinterval,
	}
	resolver.Register(builder)
	return nil
}

func (b *ConsulResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	b.ServiceName = target.Endpoint
	r := ConsulResolver{
		consulBuilder: b,
		target:        target,
		cc:            cc,
		addr:          make(chan []resolver.Address, 1),
		done:          make(chan struct{}, 1),
		watchInterval: b.WatchInterval,
	}
	go r.updater()
	go r.watcher()
	r.resolve()

	return &r, nil
}

func (b *ConsulResolverBuilder) Scheme() string {
	return Scheme
}
