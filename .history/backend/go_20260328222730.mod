module etf-insight

go 1.21

require (
	github.com/gin-contrib/cors v1.4.0
	github.com/gin-gonic/gin v1.8.1
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/robfig/cron/v3 v3.0.1
	github.com/shopspring/decimal v1.3.1
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/datatypes v1.2.0
	gorm.io/driver/sqlite v1.5.4
	gorm.io/gorm v1.25.5
)

replace (
	golang.org/x/arch => github.com/golang/arch v0.5.0
	golang.org/x/crypto => github.com/golang/crypto v0.14.0
	golang.org/x/net => github.com/golang/net v0.16.0
	golang.org/x/sys => github.com/golang/sys v0.13.0
	golang.org/x/text => github.com/golang/text v0.13.0
	google.golang.org/protobuf => github.com/gogo/protobuf v1.3.2
)
