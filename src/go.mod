module github.com/goharbor/harbor/src

go 1.14

require (
	github.com/Azure/azure-sdk-for-go v37.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.9.3 // indirect
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/FZambia/sentinel v1.1.0
	github.com/Masterminds/semver v1.4.2
	github.com/Unknwon/goconfig v0.0.0-20160216183935-5f601ca6ef4d // indirect
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/aliyun/alibaba-cloud-sdk-go v0.0.0-20190726115642-cd293c93fd97
	github.com/astaxie/beego v1.12.1
	github.com/aws/aws-sdk-go v1.32.5
	github.com/beego/i18n v0.0.0-20140604031826-e87155e8f0c0
	github.com/bmatcuk/doublestar v1.1.1
	github.com/bugsnag/bugsnag-go v1.5.2 // indirect
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	github.com/casbin/casbin v1.7.0
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cloudflare/cfssl v0.0.0-20190510060611-9c027c93ba9e // indirect
	github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/denverdino/aliyungo v0.0.0-20191227032621-df38c6fa730c // indirect
	github.com/dghubble/sling v1.1.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/distribution/distribution v2.7.1+incompatible
	github.com/docker/go v0.0.0-20160303222718-d30aec9fd63c // indirect
	github.com/docker/go-metrics v0.0.0-20181218153428-b84716841b82 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7
	github.com/ghodss/yaml v1.0.0
	github.com/go-asn1-ber/asn1-ber v1.5.1
	github.com/go-ldap/ldap/v3 v3.2.4
	github.com/go-openapi/errors v0.19.2
	github.com/go-openapi/loads v0.19.4
	github.com/go-openapi/runtime v0.19.5
	github.com/go-openapi/spec v0.19.3
	github.com/go-openapi/strfmt v0.19.3
	github.com/go-openapi/swag v0.19.5
	github.com/go-openapi/validate v0.19.5
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gocraft/work v0.5.1
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/golang-migrate/migrate/v4 v4.11.0
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/certificate-transparency-go v1.0.21 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/csrf v1.6.2
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/graph-gophers/dataloader v5.0.0+incompatible
	github.com/jinzhu/gorm v1.9.8 // indirect
	github.com/lib/pq v1.3.0
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/miekg/pkcs11 v0.0.0-20170220202408-7283ca79f35e // indirect
	github.com/ncw/swift v1.0.49 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/olekukonko/tablewriter v0.0.1
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/prometheus/client_golang v1.7.1
	github.com/robfig/cron v1.0.0
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/tencentcloud/tencentcloud-sdk-go v1.0.62
	github.com/theupdateframework/notary v0.6.1
	github.com/vmihailenco/msgpack/v5 v5.0.0-rc.2
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20200625001655-4c5254603344
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	gopkg.in/dancannon/gorethink.v3 v3.0.5 // indirect
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect
	gopkg.in/gorethink/gorethink.v3 v3.0.5 // indirect
	gopkg.in/square/go-jose.v2 v2.3.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
	helm.sh/helm/v3 v3.1.1
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.3+incompatible
	github.com/goharbor/harbor => ../
	google.golang.org/api => google.golang.org/api v0.0.0-20160322025152-9bf6e6e569ff
)
