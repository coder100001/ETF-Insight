module etf-insight

go 1.24

require (
	github.com/gin-gonic/gin v1.5.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/robfig/cron/v3 v3.0.1
	github.com/shopspring/decimal v1.3.1
	github.com/stretchr/testify v1.8.4
	gopkg.in/yaml.v3 v3.0.1
)

replace golang.org/x/sys => github.com/golang/sys v0.25.0
replace golang.org/x/net => github.com/golang/net v0.31.0
replace golang.org/x/crypto => github.com/golang/crypto v0.28.0
replace golang.org/x/arch => github.com/golang/arch v0.10.0
replace golang.org/x/text => github.com/golang/text v0.21.0
