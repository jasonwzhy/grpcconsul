# grpcconsul
The package support new grpc balancer interface. 


```
func WithBalancer
func WithBalancer(b Balancer) DialOption
WithBalancer returns a DialOption which sets a load balancer with the v1 API. Name resolver will be ignored if this DialOption is specified.

Deprecated: use the new balancer APIs in balancer package and WithBalancerName.

func WithBalancerName
func WithBalancerName(balancerName string) DialOption
WithBalancerName sets the balancer that the ClientConn will be initialized with. Balancer registered with balancerName will be used. This function panics if no balancer was registered by balancerName.

The balancer cannot be overridden by balancer option specified by service config.

This is an EXPERIMENTAL API.
```
