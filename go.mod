module github.com/ipfs-force-community/venus-messager

go 1.15

require (
	github.com/filecoin-project/filecoin-ffi v0.30.4-0.20200910194244-f640612a1a1f
	github.com/filecoin-project/go-address v0.0.5
	github.com/filecoin-project/go-jsonrpc v0.1.4-0.20210217175800-45ea43ac2bec
	github.com/filecoin-project/go-state-types v0.1.0
	github.com/filecoin-project/venus v0.9.1
	github.com/gin-gonic/gin v1.6.3
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2 // indirect
	github.com/hraban/lrucache v0.0.0-20201130153820-17052bf09781 // indirect
	github.com/hunjixin/automapper v0.0.0-20191127090318-9b979ce72ce2
	github.com/ipfs-force-community/venus-auth v0.0.0-20210401093821-d73c83e45b94
	github.com/ipfs-force-community/venus-wallet v0.0.0-20210310055425-c7247d399ba3
	github.com/ipfs/go-cid v0.0.7
	github.com/jonboulle/clockwork v0.2.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/onsi/ginkgo v1.15.0 // indirect
	github.com/onsi/gomega v1.10.5 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pelletier/go-toml v1.6.0
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/toorop/gin-logrus v0.0.0-20210225092905-2c785434f26f
	github.com/urfave/cli/v2 v2.3.0
	go.uber.org/fx v1.13.1
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5 // indirect
	golang.org/x/mod v0.4.1 // indirect
	golang.org/x/text v0.3.5 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	google.golang.org/api v0.29.0 // indirect
	google.golang.org/genproto v0.0.0-20200707001353-8e8330bf89df // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gorm.io/driver/mysql v1.0.4
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.20.12
	honnef.co/go/tools v0.1.3 // indirect
)

replace github.com/ipfs-force-community/venus-messager => ./

replace github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi
