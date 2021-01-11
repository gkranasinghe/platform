module github.com/tidepool-org/platform

go 1.11

require (
	github.com/ant0ine/go-json-rest v3.3.2+incompatible
	github.com/aws/aws-sdk-go v1.29.23
	github.com/blang/semver v3.5.1+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/disintegration/imaging v1.5.0
	github.com/fatih/color v1.7.0 // indirect
	github.com/githubnemo/CompileDaemon v1.0.0
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/go-pg/pg/v10 v10.0.0-beta.2
	github.com/howeyc/fsnotify v0.9.0 // indirect
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/mitchellh/go-homedir v1.0.0
	github.com/mjibson/esc v0.1.0
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/urfave/cli v1.20.0
	golang.org/x/crypto v0.0.0-20201012173705-84dcc777aaee
	golang.org/x/image v0.0.0-20181116024801-cd38e8056d9b // indirect
	golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3
	golang.org/x/oauth2 v0.0.0-20190115181402-5dab4167f31c
	golang.org/x/tools v0.0.0-20190524140312-2c0ae7006135
	gopkg.in/tylerb/graceful.v1 v1.2.15
	gopkg.in/yaml.v2 v2.3.0
)

replace gopkg.in/fsnotify.v1 v1.4.7 => gopkg.in/fsnotify/fsnotify.v1 v1.4.7
