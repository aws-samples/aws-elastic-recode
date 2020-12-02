module github.com/stevensu1977/elasticrecode

go 1.14

require (
	github.com/aws/aws-sdk-go v1.29.34
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2
	github.com/operator-framework/operator-sdk v0.17.0
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.6.0
	google.golang.org/appengine v1.6.5
	gopkg.in/square/go-jose.v2 v2.2.2
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v12.0.0+incompatible
)

replace github.com/satori/go.uuid v1.2.0 => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
