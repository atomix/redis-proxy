module github.com/atomix/redis-proxy

go 1.13

require (
	github.com/Azure/go-autorest v11.1.2+incompatible // indirect
	github.com/appscode/jsonpatch v3.0.1+incompatible // indirect
	github.com/atomix/api v0.0.0-20200211005812-591fe8b07ea8
	github.com/atomix/go-framework v0.0.0-20200211010411-ae512dcee9ad
	github.com/atomix/kubernetes-controller v0.0.0-20200401003423-03136b08c532
	github.com/gogo/protobuf v1.3.1
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/onosproject/onos-lib-go v0.0.0-20200312143358-18e0412086bb
	github.com/onsi/gomega v1.8.1 // indirect
	github.com/pelletier/go-toml v1.4.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	go.uber.org/multierr v1.4.0 // indirect
	golang.org/x/sys v0.0.0-20191113165036-4c7a9d0fe056 // indirect
	golang.org/x/tools v0.0.0-20191113183821-b2a5ed324b91 // indirect
	google.golang.org/grpc v1.28.0
	k8s.io/api v0.0.0-20181126151915-b503174bad59
	k8s.io/apimachinery v0.0.0-20181126123746-eddba98df674
	k8s.io/client-go v0.0.0-20181126152608-d082d5923d3c
	k8s.io/code-generator v0.0.0-20190612125529-c522cb6c26aa
	sigs.k8s.io/controller-runtime v0.1.8
)

replace github.com/atomix/kubernetes-controller => ../kubernetes-controller
