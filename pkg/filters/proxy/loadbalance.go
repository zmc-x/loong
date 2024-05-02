package proxy

import (
	"errors"
	"hash/fnv"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// loadBalance policy
const (
	Direct               = ""
	Randomlbp            = "random"
	RoundRobinlbp        = "roundRobin"
	WeigthRoundRoubinlbp = "weightRoundRobin"
	IPHashlbp            = "ipHash"
)

var (
	ErrNoHost             = errors.New("no host")
	ErrAlgorithmSupported = errors.New("algorithm no supported")
)

// Balancer interface is load balancer for reverse proxy.
// Different strategies of load balancing can
// be realized through this interface
type Balancer interface {
	Add(string)
	Remove(string)

	// balance function selects the appropriate server
	Balance(string) (string, error)
}

// factory design pattern
type Factory func([]Host) Balancer

var factories = make(map[string]Factory)

// generate the loadBalancer
func BuildBalancer(algorithm string, hosts []Host) (Balancer, error) {
	factory, ok := factories[algorithm]
	if !ok {
		return nil, ErrAlgorithmSupported
	}
	return factory(hosts), nil
}

// basic Balancer
type BaseBalancer struct {
	sync.RWMutex
	Hosts []Host

	// storage delete host information
	// easy to call add/remove methods
	// key: hostName, value: weight
	delHosts map[string]int64
}

func NewBaseBalancer(hosts []Host) Balancer {
	return &BaseBalancer{Hosts: hosts}
}

// this method add host to Hosts
// dynamically adjust the number of hosts
func (b *BaseBalancer) Add(nHost string) {
	b.Lock()
	defer b.Unlock()
	for _, host := range b.Hosts {
		if host.Url == nHost {
			return
		}
	}
	b.Hosts = append(b.Hosts, Host{
		Url:    nHost,
		Weight: b.delHosts[nHost],
	})
}

func (b *BaseBalancer) Remove(oHost string) {
	b.Lock()
	defer b.Unlock()
	for i, host := range b.Hosts {
		if host.Url == oHost {
			// record the weight
			b.delHosts[oHost] = host.Weight
			b.Hosts = append(b.Hosts[:i], b.Hosts[i+1:]...)
			break
		}
	}
}

// this method contains direct forward request.
func (b *BaseBalancer) Balance(key string) (string, error) {
	n := len(b.Hosts)
	if n == 0 {
		return "", ErrNoHost
	}
	// return the first host
	return b.Hosts[0].Url, nil
}


type LoadBalancePolicyRandom struct {
	BaseBalancer
	rnd *rand.Rand
}

func NewLbpRandom(hosts []Host) Balancer {
	return &LoadBalancePolicyRandom{
		BaseBalancer: BaseBalancer{
			Hosts: hosts,
		},
		rnd: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (lbp *LoadBalancePolicyRandom) Balance(_ string) (string, error) {
	lbp.RLock()
	defer lbp.RUnlock()
	n := len(lbp.Hosts)
	if n == 0 {
		return "", ErrNoHost
	}
	return lbp.Hosts[lbp.rnd.Int()%n].Url, nil
}

type LoadBalancePolicyRoundRobin struct {
	BaseBalancer
	cnt atomic.Uint64
}

func NewLbpRoundRobin(hosts []Host) Balancer {
	return &LoadBalancePolicyRoundRobin{
		BaseBalancer: BaseBalancer{
			Hosts: hosts,
		},
		cnt: atomic.Uint64{},
	}
}

func (lbp *LoadBalancePolicyRoundRobin) Balance(_ string) (string, error) {
	lbp.RLock()
	defer lbp.RUnlock()
	n := len(lbp.Hosts)
	if n == 0 {
		return "", ErrNoHost
	}
	return lbp.Hosts[lbp.cnt.Add(1)%uint64(n)].Url, nil
}

type LoadBalancePolicyWeightRoundRobin struct {
	BaseBalancer
	totalWeight int64
	rnd         *rand.Rand
}

func NewLbpWeightRoundRobin(hosts []Host) Balancer {
	// calculate the sum of weight of hosts
	var totalWeight int64
	for _, host := range hosts {
		totalWeight += host.Weight
	}
	return &LoadBalancePolicyWeightRoundRobin{
		BaseBalancer: BaseBalancer{
			Hosts: hosts,
		},
		totalWeight: totalWeight,
		rnd:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (lbp *LoadBalancePolicyWeightRoundRobin) Balance(_ string) (string, error) {
	w := lbp.rnd.Int63n(lbp.totalWeight)
	for _, host := range lbp.Hosts {
		w -= host.Weight
		if w < 0 {
			return host.Url, nil
		}
	}
	return "", nil
}

type LoadBalancePolicyIPHash struct {
	BaseBalancer
}

func NewLbpIPHash(hosts []Host) Balancer {
	return &LoadBalancePolicyIPHash{
		BaseBalancer: BaseBalancer{
			Hosts: hosts,
		},
	}
}

func (lbp *LoadBalancePolicyIPHash) Balance(clientIP string) (string, error) {
	hash := fnv.New32()
	hash.Write([]byte(clientIP))
	n := len(lbp.Hosts)
	return lbp.Hosts[hash.Sum32()%uint32(n)].Url, nil
}

// initialize the factories
// register loadBalance policy
func init() {
	factories[Direct] = NewBaseBalancer
	factories[Randomlbp] = NewLbpRandom
	factories[RoundRobinlbp] = NewLbpRoundRobin
	factories[WeigthRoundRoubinlbp] = NewLbpWeightRoundRobin
	factories[IPHashlbp] = NewLbpIPHash
}
