module github.com/zwcway/castserver-go

go 1.20

require (
	github.com/fasthttp/websocket v1.5.1
	github.com/go-ini/ini v1.67.0
	github.com/hajimehoshi/oto/v2 v2.3.1
	github.com/jackpal/gateway v1.0.7
	github.com/jedib0t/go-pretty v4.3.0+incompatible
	github.com/muesli/go-app-paths v0.2.2
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.4
	github.com/valyala/fasthttp v1.44.0
	github.com/zwcway/fasthttp-upnp v0.0.0-20230327113116-cfeb3e90ec00
	go.uber.org/zap v1.24.0
	golang.org/x/exp v0.0.0-20230303215020-44a13b063f3e
	golang.org/x/net v0.8.0
	gorm.io/driver/postgres v1.5.0
	gorm.io/driver/sqlite v1.4.4
	gorm.io/driver/sqlserver v1.4.2
	gorm.io/gorm v1.25.7
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/asaskevich/govalidator v0.0.0-20200907205600-7a23bdc65eef // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-openapi/errors v0.20.2 // indirect
	github.com/go-openapi/strfmt v0.21.3 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.3.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/mattn/go-runewidth v0.0.7 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	github.com/microsoft/go-mssqldb v0.20.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/savsgio/gotils v0.0.0-20220530130905-52f3993e8d6d // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/zwcway/go-jsonpack v0.0.0-20230621104234-52358eb7753f
	go.mongodb.org/mongo-driver v1.10.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/zwcway/castserver-go => ./

// replace github.com/zwcway/fasthttp-upnp => ../upnp
