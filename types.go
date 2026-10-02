package ipmax

import "encoding/json"

type Product string

const (
	ProductGeoIP        Product = "geoip"
	ProductIntelligence Product = "intelligence"
)

type NetworkClass string

const (
	NetworkClassMobile       NetworkClass = "mobile"
	NetworkClassHosting      NetworkClass = "hosting"
	NetworkClassResidential  NetworkClass = "residential"
	NetworkClassBusiness     NetworkClass = "business"
	NetworkClassOrganization NetworkClass = "organization"
	NetworkClassGovernment   NetworkClass = "government"
	NetworkClassMilitary     NetworkClass = "military"
	NetworkClassEducation    NetworkClass = "education"
	NetworkClassLibrary      NetworkClass = "library"
	NetworkClassDedicated    NetworkClass = "dedicated"
	NetworkClassSatellite    NetworkClass = "satellite"
	NetworkClassIXP          NetworkClass = "ixp"
	NetworkClassSearchEngine NetworkClass = "search_engine"
	NetworkClassAICrawler    NetworkClass = "ai_crawler"
	NetworkClassUnknown      NetworkClass = "unknown"
)

type NetworkClassSource string

const (
	NetworkClassSourceDomesticISP        NetworkClassSource = "domestic_isp"
	NetworkClassSourceUsageType          NetworkClassSource = "usage_type"
	NetworkClassSourceMobileCarrier      NetworkClassSource = "mobile_carrier"
	NetworkClassSourceConnectionType     NetworkClassSource = "connection_type"
	NetworkClassSourceNetworkTag         NetworkClassSource = "network_tag"
	NetworkClassSourceThreatIntelligence NetworkClassSource = "threat_intelligence"
	NetworkClassSourceASNTag             NetworkClassSource = "asn_tag"
	NetworkClassSourceIXP                NetworkClassSource = "ixp"
	NetworkClassSourceNone               NetworkClassSource = "none"
)

type LiveIntelligenceTag string

const (
	LiveIntelligenceTagHosting           LiveIntelligenceTag = "hosting"
	LiveIntelligenceTagPublicDNSResolver LiveIntelligenceTag = "public_dns_resolver"
	LiveIntelligenceTagWebServer         LiveIntelligenceTag = "web_server"
	LiveIntelligenceTagVPN               LiveIntelligenceTag = "vpn"
	LiveIntelligenceTagResidentialProxy  LiveIntelligenceTag = "residential_proxy"
)

type RevokedNetworkTag string

const (
	RevokedNetworkTagBusinessBroadband RevokedNetworkTag = "business_broadband"
	RevokedNetworkTagHomeISP           RevokedNetworkTag = "home_isp"
	RevokedNetworkTagMobileISP         RevokedNetworkTag = "mobile_isp"
)

type AnonymizerExemptionReason string

const (
	AnonymizerExemptionCloudflareCDNOrigin AnonymizerExemptionReason = "cloudflare_cdn_origin"
	AnonymizerExemptionSearchEngineCrawler AnonymizerExemptionReason = "search_engine_crawler"
)

type CpePart string

const (
	CpePartApplication     CpePart = "application"
	CpePartOperatingSystem CpePart = "operating_system"
	CpePartHardware        CpePart = "hardware"
)

type RpkiStatus string

const (
	RpkiStatusValid         RpkiStatus = "valid"
	RpkiStatusInvalidASN    RpkiStatus = "invalid_asn"
	RpkiStatusInvalidLength RpkiStatus = "invalid_length"
	RpkiStatusUnknown       RpkiStatus = "unknown"
	RpkiStatusNotRouted     RpkiStatus = "not_routed"
)

type RpkiSource string

const (
	RpkiSourceRTRLive    RpkiSource = "RTR Live"
	RpkiSourceLocalCache RpkiSource = "Local Cache"
	RpkiSourceLocalDB    RpkiSource = "Local DB"
)

type PtrConfidenceLevel string

const (
	PtrConfidenceHigh   PtrConfidenceLevel = "high"
	PtrConfidenceMedium PtrConfidenceLevel = "medium"
	PtrConfidenceLow    PtrConfidenceLevel = "low"
)

type PtrDatasetName string

const (
	PtrDatasetGeonames          PtrDatasetName = "geonames"
	PtrDatasetOurAirports       PtrDatasetName = "ourairports"
	PtrDatasetOperatorOverrides PtrDatasetName = "operator-overrides"
	PtrDatasetUNLOCODE          PtrDatasetName = "unlocode"
)

type PtrHintKind string

const (
	PtrHintIATA         PtrHintKind = "iata"
	PtrHintUNLOCODE     PtrHintKind = "unlocode"
	PtrHintCLLI         PtrHintKind = "clli"
	PtrHintCityName     PtrHintKind = "city_name"
	PtrHintOperatorCode PtrHintKind = "operator_code"
)

type LookupRequest struct {
	IP string `json:"ip"`
}

type ErrorResponse struct {
	Status bool     `json:"status"`
	Error  APIError `json:"error"`
}

type Price struct {
	Product    Product `json:"product"`
	Name       string  `json:"name"`
	UnitMicros float64 `json:"unitMicros"`
}

type Catalog struct {
	Currency      string  `json:"currency"`
	Prices        []Price `json:"prices"`
	PurchaseEmail string  `json:"purchaseEmail"`
	Available     bool    `json:"available"`
}

type CatalogResponse struct {
	Status bool    `json:"status"`
	Data   Catalog `json:"data"`
}

type Wallet struct {
	Product       Product `json:"product"`
	BalanceMicros float64 `json:"balanceMicros"`
}

type LedgerEntry struct {
	ID            string  `json:"id"`
	Product       Product `json:"product"`
	Kind          string  `json:"kind"`
	AmountMicros  float64 `json:"amountMicros"`
	BalanceMicros float64 `json:"balanceMicros"`
	Reference     string  `json:"reference"`
	CreatedAt     string  `json:"createdAt"`
}

type Account struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	KeyPrefix string        `json:"keyPrefix"`
	Wallets   []Wallet      `json:"wallets"`
	Entries   []LedgerEntry `json:"entries"`
}

type AccountResponse struct {
	Status bool    `json:"status"`
	Data   Account `json:"data"`
}

type GeoIPAs struct {
	ASN         string `json:"asn"`
	Name        string `json:"name"`
	NameEn      string `json:"name_en,omitempty"`
	CountryCode string `json:"countryCode,omitempty"`
	Domain      string `json:"domain"`
}

type Geolocation struct {
	City          string   `json:"city"`
	Region        string   `json:"region"`
	District      string   `json:"district"`
	RegionCode    string   `json:"region_code"`
	Adcode        string   `json:"adcode"`
	Country       string   `json:"country"`
	CountryCode   string   `json:"country_code"`
	Continent     string   `json:"continent"`
	ContinentCode string   `json:"continent_code"`
	Latitude      *float64 `json:"latitude"`
	Longitude     *float64 `json:"longitude"`
	Timezone      string   `json:"timezone"`
	PostalCode    string   `json:"postal_code"`
	Radius        *float64 `json:"radius"`
	AreaCode      string   `json:"area_code,omitempty"`
	PlusCode      string   `json:"plus_code,omitempty"`
	CountryEn     string   `json:"country_en,omitempty"`
	CountryZh     string   `json:"country_zh,omitempty"`
	RegionEn      string   `json:"region_en,omitempty"`
	RegionZh      string   `json:"region_zh,omitempty"`
	CityEn        string   `json:"city_en,omitempty"`
	CityZh        string   `json:"city_zh,omitempty"`
	DistrictEn    string   `json:"district_en,omitempty"`
	DistrictZh    string   `json:"district_zh,omitempty"`
}

type Currency struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Symbol string `json:"symbol"`
	Native string `json:"native"`
	Plural string `json:"plural"`
}

type TimeZone struct {
	Name        string `json:"name"`
	Abbr        string `json:"abbr"`
	Offset      string `json:"offset"`
	IsDST       bool   `json:"is_dst"`
	CurrentTime string `json:"current_time"`
}

type GeoIPData struct {
	As           GeoIPAs     `json:"as"`
	IP           string      `json:"ip"`
	Geo          Geolocation `json:"geo"`
	Netmask      string      `json:"netmask"`
	PrefixLength *int        `json:"prefix_length"`
	IsBogon      bool        `json:"is_bogon"`
	Currency     *Currency   `json:"currency,omitempty"`
	TimeZone     *TimeZone   `json:"time_zone,omitempty"`
	CallingCode  *string     `json:"calling_code,omitempty"`
}

type GeoIPResponse struct {
	Status bool      `json:"status"`
	Data   GeoIPData `json:"data"`
}

type AsnEstimatedCapacity struct {
	LowerGbps float64  `json:"lower_gbps"`
	UpperGbps *float64 `json:"upper_gbps,omitempty"`
	Label     string   `json:"label"`
}

type AsnExchangePort struct {
	CapacityGbps float64 `json:"capacity_gbps"`
	Operational  *bool   `json:"operational,omitempty"`
	RSPeer       *bool   `json:"rs_peer,omitempty"`
}

type AsnExchange struct {
	Name        string            `json:"name"`
	CountryCode string            `json:"countryCode,omitempty"`
	Type        string            `json:"type"`
	Ports       []AsnExchangePort `json:"ports"`
}

type AsnConnectivity struct {
	DeclaredTraffic   string               `json:"declared_traffic"`
	ExchangesTotal    *int                 `json:"exchanges_total,omitempty"`
	EstimatedCapacity AsnEstimatedCapacity `json:"estimated_capacity"`
	Exchanges         []AsnExchange        `json:"exchanges"`
}

type IntelligenceAs struct {
	ASN          string           `json:"asn"`
	Name         string           `json:"name"`
	NameEn       string           `json:"name_en,omitempty"`
	CountryCode  string           `json:"countryCode,omitempty"`
	Domain       string           `json:"domain"`
	Type         string           `json:"type"`
	Class        string           `json:"class,omitempty"`
	Tags         []string         `json:"tags,omitempty"`
	Connectivity *AsnConnectivity `json:"connectivity,omitempty"`
}

type PtrOperator struct {
	Name   string `json:"name"`
	ASN    string `json:"asn"`
	Suffix string `json:"suffix"`
}

type PtrHint struct {
	Value string      `json:"value"`
	Kind  PtrHintKind `json:"kind"`
}

type PtrMetroLocation struct {
	City        string  `json:"city"`
	Region      *string `json:"region"`
	RegionCode  *string `json:"region_code"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Precision   string  `json:"precision"`
}

type PtrConfidence struct {
	Level PtrConfidenceLevel `json:"level"`
	Score float64            `json:"score"`
}

type PtrEvidenceSource struct {
	Dataset PtrDatasetName `json:"dataset"`
	Version string         `json:"version"`
}

type PtrEvidence struct {
	RuleID         string              `json:"rule_id"`
	RulesetVersion string              `json:"ruleset_version"`
	Sources        []PtrEvidenceSource `json:"sources"`
}

type PtrLocatedIntelligence struct {
	Hostname   string           `json:"hostname"`
	Status     string           `json:"status"`
	Operator   PtrOperator      `json:"operator"`
	Hint       PtrHint          `json:"hint"`
	Location   PtrMetroLocation `json:"location"`
	Confidence PtrConfidence    `json:"confidence"`
	Evidence   PtrEvidence      `json:"evidence"`
}

type PtrAsnMismatchIntelligence struct {
	Hostname    string      `json:"hostname"`
	Status      string      `json:"status"`
	Operator    PtrOperator `json:"operator"`
	ObservedASN string      `json:"observed_asn"`
	Evidence    PtrEvidence `json:"evidence"`
}

type PtrUnmatchedIntelligence struct {
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
}

type PtrIntelligence struct {
	Hostname    string
	Status      string
	Located     *PtrLocatedIntelligence
	AsnMismatch *PtrAsnMismatchIntelligence
	Unmatched   *PtrUnmatchedIntelligence
}

func (p *PtrIntelligence) UnmarshalJSON(data []byte) error {
	var head PtrUnmatchedIntelligence
	if err := json.Unmarshal(data, &head); err != nil {
		return err
	}
	*p = PtrIntelligence{Hostname: head.Hostname, Status: head.Status}
	switch head.Status {
	case "located":
		p.Located = new(PtrLocatedIntelligence)
		return json.Unmarshal(data, p.Located)
	case "asn_mismatch":
		p.AsnMismatch = new(PtrAsnMismatchIntelligence)
		return json.Unmarshal(data, p.AsnMismatch)
	case "unmatched":
		p.Unmatched = &head
	}
	return nil
}

func (p PtrIntelligence) MarshalJSON() ([]byte, error) {
	switch {
	case p.Located != nil:
		return json.Marshal(p.Located)
	case p.AsnMismatch != nil:
		return json.Marshal(p.AsnMismatch)
	default:
		return json.Marshal(PtrUnmatchedIntelligence{Hostname: p.Hostname, Status: p.Status})
	}
}

type CpeObservation struct {
	Cpe     string  `json:"cpe"`
	Part    CpePart `json:"part"`
	Vendor  string  `json:"vendor"`
	Product string  `json:"product"`
	Version *string `json:"version"`
}

type RpkiRtr struct {
	Source      string `json:"source"`
	Server      string `json:"server,omitempty"`
	LastUpdated int64  `json:"lastUpdated"`
}

type RpkiValidator struct {
	Source      string `json:"source"`
	Engine      string `json:"engine"`
	LastUpdated int64  `json:"lastUpdated"`
}

type RpkiInfo struct {
	Status      RpkiStatus     `json:"status"`
	Source      RpkiSource     `json:"source,omitempty"`
	RouteSource string         `json:"route_source,omitempty"`
	OriginASN   string         `json:"origin_asn,omitempty"`
	RoutePrefix string         `json:"route_prefix,omitempty"`
	RoaPrefix   string         `json:"roa_prefix,omitempty"`
	RoaASN      string         `json:"roa_asn,omitempty"`
	MaxLength   *int           `json:"max_length,omitempty"`
	Ta          string         `json:"ta,omitempty"`
	Rtr         *RpkiRtr       `json:"rtr,omitempty"`
	Validator   *RpkiValidator `json:"validator,omitempty"`
}

type AbuseCategoryCount struct {
	ID    int `json:"id"`
	Count int `json:"count"`
}

type AbuseReports struct {
	ConfidenceScore   int                  `json:"confidence_score"`
	TotalReports      int                  `json:"total_reports"`
	DistinctReporters int                  `json:"distinct_reporters"`
	LastReportedAt    *string              `json:"last_reported_at"`
	IsWhitelisted     *bool                `json:"is_whitelisted"`
	Categories        []AbuseCategoryCount `json:"categories"`
}

type ResourceTransfer struct {
	TransferTime      string `json:"transfer_time"`
	TransferType      string `json:"transfer_type"`
	SourceResource    string `json:"source_resource"`
	RecipientResource string `json:"recipient_resource"`
	SourceRegistry    string `json:"source_registry"`
	RecipientRegistry string `json:"recipient_registry"`
	SourceHolder      string `json:"source_holder"`
	RecipientHolder   string `json:"recipient_holder"`
}

type SegmentProbe struct {
	IP    string                `json:"ip"`
	Ports []int                 `json:"ports"`
	Tags  []LiveIntelligenceTag `json:"tags"`
}

type LiveIntelligence struct {
	Ptr          []PtrIntelligence     `json:"ptr"`
	Nameservers  []string              `json:"nameservers"`
	Cpes         []CpeObservation      `json:"cpes"`
	Tags         []LiveIntelligenceTag `json:"tags"`
	RevokedTags  []RevokedNetworkTag   `json:"revoked_tags"`
	UsageType    *string               `json:"usage_type"`
	Rpki         *RpkiInfo             `json:"rpki,omitempty"`
	Abuse        *AbuseReports         `json:"abuse,omitempty"`
	Transfers    []ResourceTransfer    `json:"transfers,omitempty"`
	SegmentProbe *SegmentProbe         `json:"segment_probe,omitempty"`
}

type NetworkInfo struct {
	ConnectionType string   `json:"connection_type"`
	UsageType      string   `json:"usage_type"`
	AddressType    string   `json:"address_type,omitempty"`
	NetSpeed       string   `json:"net_speed"`
	Tags           []string `json:"tags"`
}

type NetworkClassEvidence struct {
	Class  NetworkClass       `json:"class"`
	Source NetworkClassSource `json:"source"`
	Scope  string             `json:"scope"`
	Value  string             `json:"value"`
}

type NetworkClassification struct {
	Primary    NetworkClass           `json:"primary"`
	Confidence string                 `json:"confidence"`
	Source     NetworkClassSource     `json:"source"`
	Conflicts  []NetworkClass         `json:"conflicts"`
	Evidence   []NetworkClassEvidence `json:"evidence"`
}

type MobileCarrier struct {
	MCC   string `json:"mcc"`
	MNC   string `json:"mnc"`
	Brand string `json:"brand"`
}

type Blocklist struct {
	Name string `json:"name"`
	Site string `json:"site"`
	Type string `json:"type"`
}

type ThreatScores struct {
	VPNScore    *float64 `json:"vpn_score"`
	ProxyScore  *float64 `json:"proxy_score"`
	ThreatScore *float64 `json:"threat_score"`
	TrustScore  *float64 `json:"trust_score"`
}

type ThreatInfo struct {
	IsTor                *bool                       `json:"is_tor"`
	IsVPN                *bool                       `json:"is_vpn"`
	IsICloudRelay        *bool                       `json:"is_icloud_relay"`
	IsCloudflareWarp     *bool                       `json:"is_cloudflare_warp"`
	IsProxy              *bool                       `json:"is_proxy"`
	IsDatacenter         *bool                       `json:"is_datacenter"`
	IsAnonymous          *bool                       `json:"is_anonymous"`
	IsKnownAttacker      *bool                       `json:"is_known_attacker"`
	IsKnownAbuser        *bool                       `json:"is_known_abuser"`
	IsThreat             *bool                       `json:"is_threat"`
	IsBogon              bool                        `json:"is_bogon"`
	Blocklists           []Blocklist                 `json:"blocklists"`
	BlocklistCount       int                         `json:"blocklist_count"`
	Scores               ThreatScores                `json:"scores"`
	AnonymizerExemptions []AnonymizerExemptionReason `json:"anonymizer_exemptions,omitempty"`
}

type Company struct {
	Name    *string `json:"name"`
	Address *string `json:"address"`
	Domain  *string `json:"domain"`
	Phone   *string `json:"phone"`
	Type    *string `json:"type"`
	Network *string `json:"network"`
	Country *string `json:"country"`
	Source  string  `json:"source,omitempty"`
}

type AbuseContact struct {
	Name    *string `json:"name"`
	Address *string `json:"address"`
	Country *string `json:"country"`
	Email   *string `json:"email"`
	Network *string `json:"network"`
	Phone   *string `json:"phone"`
	Source  string  `json:"source,omitempty"`
}

type DNSPtrEvidence struct {
	Hostname         string   `json:"hostname"`
	Sources          []string `json:"sources"`
	ForwardConfirmed bool     `json:"forward_confirmed"`
	FirstSeen        *string  `json:"first_seen"`
	LastSeen         *string  `json:"last_seen"`
}

type DNSHostnameEvidence struct {
	Hostname         string   `json:"hostname"`
	Kinds            []string `json:"kinds"`
	Sources          []string `json:"sources"`
	ForwardConfirmed bool     `json:"forward_confirmed"`
	FirstSeen        *string  `json:"first_seen"`
	LastSeen         *string  `json:"last_seen"`
}

type DNSRecords struct {
	Ptr         []string              `json:"ptr"`
	PtrEvidence []DNSPtrEvidence      `json:"ptr_evidence"`
	Hostnames   []DNSHostnameEvidence `json:"hostnames"`
	Sources     []string              `json:"sources"`
	SnapshotID  *string               `json:"snapshot_id"`
	UpdatedAt   *string               `json:"updated_at"`
}

type IntelligenceData struct {
	As           IntelligenceAs        `json:"as"`
	Intelligence LiveIntelligence      `json:"intelligence"`
	IP           string                `json:"ip"`
	Network      *NetworkInfo          `json:"network,omitempty"`
	NetworkClass NetworkClassification `json:"network_class"`
	Mobile       *MobileCarrier        `json:"mobile,omitempty"`
	Threat       *ThreatInfo           `json:"threat,omitempty"`
	Rpki         *RpkiInfo             `json:"rpki,omitempty"`
	Company      *Company              `json:"company,omitempty"`
	Abuse        *AbuseContact         `json:"abuse,omitempty"`
	DNS          *DNSRecords           `json:"dns,omitempty"`
	IsBogon      bool                  `json:"is_bogon"`
	IsAnycast    *bool                 `json:"is_anycast"`
	IsMobile     *bool                 `json:"is_mobile"`
	IsHosting    *bool                 `json:"is_hosting"`
	IsSatellite  *bool                 `json:"is_satellite"`
}

type IntelligenceResponse struct {
	Status bool             `json:"status"`
	Data   IntelligenceData `json:"data"`
}
