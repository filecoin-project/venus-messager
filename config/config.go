package config

import (
	"time"

	"github.com/ipfs-force-community/metrics"
)

type Config struct {
	DB             DbConfig               `toml:"db"`
	JWT            JWTConfig              `toml:"jwt"`
	API            APIConfig              `toml:"api"`
	Node           NodeConfig             `toml:"node"`
	MessageService MessageServiceConfig   `toml:"messageService"`
	Gateway        GatewayConfig          `toml:"gateway"`
	RateLimit      RateLimitConfig        `toml:"rateLimit"`
	Trace          *metrics.TraceConfig   `toml:"tracing"`
	Metrics        *metrics.MetricsConfig `toml:"metrics"`
	Libp2pNet      *Libp2pNetConfig       `toml:"libp2p"`
	Publisher      *PublisherConfig       `toml:"publisher"`
}

type NodeConfig struct {
	Url   string `toml:"url"`
	Token string `toml:"token"`
}

type APIConfig struct {
	Address string
}

type DbConfig struct {
	Type   string       `toml:"type"`
	MySql  MySqlConfig  `toml:"mysql"`
	Sqlite SqliteConfig `toml:"sqlite"`
}

type SqliteConfig struct {
	Debug bool `toml:"debug"`
}

type MySqlConfig struct {
	ConnectionString string        `toml:"connectionString"`
	MaxOpenConn      int           `toml:"maxOpenConn"`
	MaxIdleConn      int           `toml:"maxIdleConn"`
	ConnMaxLifeTime  time.Duration `toml:"connMaxLifeTime"`
	Debug            bool          `toml:"debug"`
}

type JWTConfig struct {
	AuthURL string `toml:"authURL"`
	Token   string `toml:"token"`
}

const (
	MinWaitingChainHeadStableDuration = time.Second * 2
	MaxWaitingChainHeadStableDuration = time.Second * 25
	DefWaitingChainHeadStableDuration = time.Second * 8
)

const (
	DefaultTimeout         = time.Second
	SignMessageTimeout     = time.Second * 3
	EstimateMessageTimeout = time.Second * 30
)

type MessageServiceConfig struct {
	WaitingChainHeadStableDuration time.Duration `toml:"WaitingChainHeadStableDuration"`

	DefaultTimeout         time.Duration `toml:"DefaultTimeout"`
	SignMessageTimeout     time.Duration `toml:"SignMessageTimeout"`
	EstimateMessageTimeout time.Duration `toml:"EstimateMessageTimeout"`

	SkipProcessHead bool `toml:"skipProcessHead"`
	SkipPushMessage bool `toml:"skipPushMessage"`
}

type Libp2pNetConfig struct {
	ListenAddress      string   `toml:"listenAddresses"`
	BootstrapAddresses []string `toml:"bootstrapAddresses"`

	// MinPeerThreshold determine when to expand peers.
	// default set to 0 which means use network default config.
	MinPeerThreshold int `toml:"minPeerThreshold"`

	// ExpandPeriod determine how often to expand peers.
	// default set to "0s" which means use network default config.
	// otherwise, it should be a duration string like "5s", "30s".
	ExpandPeriod time.Duration `toml:"expandPeriod"`
	// TODO: EnableRelay
}

type PublisherConfig struct {
	// CacheReleasePeriod is the period to release massage cache with unit of second.
	// default is 5.
	// set a negative int means disable cache.
	Concurrency int `toml:"concurrency"`

	// CacheReleasePeriod is the period to release massage cache with unit of second.
	// default is 0 which means auto decide by network parameters that is 1/3 of the block time.
	// set a negative int means disable cache
	CacheReleasePeriod int64 `toml:"cacheReleasePeriod"`

	EnableP2P       bool `toml:"enablePubsub"`
	EnableMultiNode bool `toml:"enableMultiNode"`
}

type MessageStateConfig struct {
	BackTime int `toml:"backTime"` // 向前找多久的数据写到内存,单位秒

	DefaultExpiration, CleanupInterval int // message 缓存的有效时间和清理间隔
}

type GatewayConfig struct {
	Token string   `toml:"token"`
	Url   []string `toml:"url"`
}

type RateLimitConfig struct {
	Redis string `toml:"redis"`
}

func DefaultConfig() *Config {
	return &Config{
		DB: DbConfig{
			Type: "sqlite",
			MySql: MySqlConfig{
				ConnectionString: "",
				MaxOpenConn:      10,
				MaxIdleConn:      10,
				ConnMaxLifeTime:  time.Second * 60,
				Debug:            false,
			},
		},
		JWT: JWTConfig{
			AuthURL: "http://127.0.0.1:8989",
		},
		API: APIConfig{
			Address: "/ip4/127.0.0.1/tcp/39812",
		},
		Node: NodeConfig{
			Url:   "/ip4/127.0.0.1/tcp/3453",
			Token: "",
		},
		MessageService: MessageServiceConfig{
			WaitingChainHeadStableDuration: DefWaitingChainHeadStableDuration,

			DefaultTimeout:         DefaultTimeout,
			SignMessageTimeout:     SignMessageTimeout,
			EstimateMessageTimeout: EstimateMessageTimeout,

			SkipProcessHead: false,
			SkipPushMessage: false,
		},
		Gateway: GatewayConfig{
			Token: "",
			Url:   []string{"/ip4/127.0.0.1/tcp/45132"},
		},
		RateLimit: RateLimitConfig{Redis: ""},
		Trace:     metrics.DefaultTraceConfig(),
		Metrics:   metrics.DefaultMetricsConfig(),
		Libp2pNet: &Libp2pNetConfig{
			ListenAddress:      "/ip4/0.0.0.0/tcp/0",
			BootstrapAddresses: []string{},
			MinPeerThreshold:   0,
			ExpandPeriod:       0 * time.Second,
		},
		Publisher: &PublisherConfig{
			Concurrency:        5,
			CacheReleasePeriod: 0,
			EnableP2P:          false,
			EnableMultiNode:    true,
		},
	}
}
