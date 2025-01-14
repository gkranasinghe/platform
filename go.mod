module github.com/tidepool-org/platform

go 1.15

require (
	github.com/Shopify/sarama v1.27.0
	github.com/ant0ine/go-json-rest v3.3.2+incompatible
	github.com/aws/aws-sdk-go v1.35.3
	github.com/blang/semver v3.5.1+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/disintegration/imaging v1.6.2
	github.com/githubnemo/CompileDaemon v1.4.0
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.2.0
	github.com/gowebpki/jcs v0.0.0-20210215032300-680d9436c864
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.0.1 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mjibson/esc v0.2.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/prometheus/client_golang v1.9.0
	github.com/rinchsan/device-check-go v1.2.3
	github.com/tidepool-org/clinic/client v0.0.0-20210604144345-728c47d30b4b
	github.com/tidepool-org/devices/api v0.0.0-20201125143608-d0e358bf4a30
	github.com/tidepool-org/go-common v0.9.0
	github.com/urfave/cli v1.22.4
	go.mongodb.org/mongo-driver v1.5.4
	go.uber.org/fx v1.13.1
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/tools v0.1.1
	google.golang.org/grpc v1.33.1
	gopkg.in/yaml.v2 v2.3.0
	syreclabs.com/go/faker v1.2.2
)

replace gopkg.in/fsnotify.v1 v1.4.7 => gopkg.in/fsnotify/fsnotify.v1 v1.4.7
