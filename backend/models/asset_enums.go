package models

// AssetClass 资产类别
type AssetClass string

const (
	AssetClassEquity      AssetClass = "equity"      // 股票
	AssetClassBond        AssetClass = "bond"        // 债券
	AssetClassCommodity   AssetClass = "commodity"   // 商品
	AssetClassREIT        AssetClass = "reit"        // 房地产信托
	AssetClassCurrency    AssetClass = "currency"    // 货币
	AssetClassMultiAsset  AssetClass = "multi_asset" // 多资产
	AssetClassAlternative AssetClass = "alternative" // 另类投资
)

// Region 地区
type Region string

const (
	RegionGlobal       Region = "global"        // 全球
	RegionUS           Region = "us"            // 美国
	RegionChina        Region = "china"         // 中国
	RegionEurope       Region = "europe"        // 欧洲
	RegionJapan        Region = "japan"         // 日本
	RegionEmerging     Region = "emerging"      // 新兴市场
	RegionAsiaPacific  Region = "asia_pacific"  // 亚太
	RegionLatinAmerica Region = "latin_america" // 拉美
)

// ETFType ETF类型
type ETFType string

const (
	ETFTypeIndex     ETFType = "index"     // 指数基金
	ETFTypeSector    ETFType = "sector"    // 行业ETF
	ETFTypeFactor    ETFType = "factor"    // 因子ETF
	ETFTypeThematic  ETFType = "thematic"  // 主题ETF
	ETFTypeActive    ETFType = "active"    // 主动管理ETF
	ETFTypeLeveraged ETFType = "leveraged" // 杠杆ETF
	ETFTypeInverse   ETFType = "inverse"   // 反向ETF
)
