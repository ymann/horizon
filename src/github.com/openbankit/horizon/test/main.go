// Package test contains simple test helpers that should not
// have any dependencies on horizon's packages.  think constants,
// custom matchers, generic helpers etc.
package test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openbankit/go-base/hash"
	"github.com/openbankit/go-base/keypair"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/config"
	hlog "github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/test/db"
	tdb "github.com/openbankit/horizon/test/db"
	"github.com/PuerkitoBio/throttled"
	"github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// StaticMockServer is a test helper that records it's last request
type StaticMockServer struct {
	*httptest.Server
	LastRequest *http.Request
}

// T provides a common set of functionality for each test in horizon
type T struct {
	T          *testing.T
	Assert     *assert.Assertions
	Require    *require.Assertions
	Ctx        context.Context
	HorizonDB  *sqlx.DB
	CoreDB     *sqlx.DB
	Logger     *hlog.Entry
	LogMetrics *hlog.Metrics
	LogBuffer  *bytes.Buffer
}

// Context provides a context suitable for testing in tests that do not create
// a full App instance (in which case your tests should be using the app's
// context).  This context has a logger bound to it suitable for testing.
func Context() context.Context {
	return hlog.Set(context.Background(), testLogger)
}

// ContextWithLogBuffer returns a context and a buffer into which the new, bound
// logger will write into.  This method allows you to inspect what data was
// logged more easily in your tests.
func ContextWithLogBuffer() (context.Context, *bytes.Buffer) {
	output := new(bytes.Buffer)
	l, _ := hlog.New()
	l.Logger.Out = output
	l.Logger.Formatter.(*logrus.TextFormatter).DisableColors = true
	l.Logger.Level = logrus.DebugLevel

	ctx := hlog.Set(context.Background(), l)
	return ctx, output

}

// Database returns a connection to the horizon test database
//
// DEPRECATED:  use `Horizon()` from test/db package
func Database() *sqlx.DB {
	return tdb.Horizon()
}

// Returns default bank master key seed.
func BankMasterSeedStr() string {
	bankMasterKeySeed := os.Getenv("BANK_MASTER_KEY_SEED")

	if bankMasterKeySeed == "" {
		hlog.Panic("BANK_MASTER_KEY_SEED must be set for tests")
	}

	return bankMasterKeySeed
}

func BankMasterSeed() *keypair.Full {
	raw := BankMasterSeedStr()
	bankKP, err := keypair.Parse(raw)
	if err != nil {
		hlog.WithStack(err).Panic("Failed to parse BANK_MASTER_KEY_SEED")
	}

	bankFull, ok := bankKP.(*keypair.Full)
	if !ok {
		hlog.Panic("BANK_MASTER_KEY_SEED must be valid seed")
	}
	return bankFull
}

func RedisURL() string {
	return os.Getenv("REDIS_URL")
}

// DatabaseURL returns the database connection the url any test
// use when connecting to the history/horizon database
//
// DEPRECATED:  use `HorizonURL()` from test/db package
func DatabaseURL() string {
	return tdb.HorizonURL()
}

// LoadScenario populates the test databases with pre-created scenarios.  Each
// scenario is in the scenarios subfolder of this package and are a pair of
// sql files, one per database.
func LoadScenario(scenarioName string) {
	loadScenario(scenarioName, true)
}

// LoadScenarioWithoutHorizon populates the test Stellar core database a with
// pre-created scenario.  Unlike `LoadScenario`, this
func LoadScenarioWithoutHorizon(scenarioName string) {
	loadScenario(scenarioName, false)
}

// Start initializes a new test helper object and conceptually "starts" a new
// test
func Start(t *testing.T) *T {
	result := &T{}

	result.T = t
	result.LogBuffer = new(bytes.Buffer)
	result.Logger, result.LogMetrics = hlog.New()
	result.Logger.Logger.Out = result.LogBuffer
	result.Logger.Logger.Formatter.(*logrus.TextFormatter).DisableColors = true
	result.Logger.Logger.Level = logrus.DebugLevel

	result.Ctx = hlog.Set(context.Background(), result.Logger)
	result.HorizonDB = Database()
	result.CoreDB = StellarCoreDatabase()
	result.Assert = assert.New(t)
	result.Require = require.New(t)

	return result
}

// StellarCoreDatabase returns a connection to the stellar core test database
//
// DEPRECATED:  use `StellarCore()` from test/db package
func StellarCoreDatabase() *sqlx.DB {
	return tdb.StellarCore()
}

// StellarCoreDatabaseURL returns the database connection the url any test
// use when connecting to the stellar-core database
//
// DEPRECATED:  use `StellarCoreURL()` from test/db package
func StellarCoreDatabaseURL() string {
	return tdb.StellarCoreURL()

}

// Used to create http.Request with signature
type RequestData struct {
	Path        string
	Signature   string
	PublicKey   string
	EncodedForm string
	Timestamp   int64
}

func GetAdminActionSignatureBase(bodyString string, timeCreated string) string {
	return "{method: 'post', body: '" + bodyString + "', timestamp: '" + timeCreated + "'}"
}

// Used to create valid RequestData
func NewRequestData(signer keypair.KP, form url.Values) RequestData {
	r := RequestData{
		PublicKey: signer.Address(),
		Timestamp: time.Now().Unix(),
	}
	r.EncodedForm = form.Encode()
	signatureBase := GetAdminActionSignatureBase(r.EncodedForm, strconv.FormatInt(r.Timestamp, 10))
	hashBase := hash.Hash([]byte(signatureBase))
	xdrSig, err := signer.SignDecorated(hashBase[:])
	if err != nil {
		hlog.Panic("Failed to sign")
	}
	r.Signature, err = xdr.MarshalBase64(xdrSig)
	if err != nil {
		hlog.Panic("Failed to marshal sign")
	}
	return r
}

// Creates http request from RequestData
func (r *RequestData) CreateRequest() *http.Request {
	body := strings.NewReader(r.EncodedForm)
	req, _ := http.NewRequest("POST", r.Path, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-AuthPublicKey", r.PublicKey)
	req.Header.Set("X-AuthSignature", r.Signature)
	req.Header.Set("X-AuthTimestamp", strconv.FormatInt(r.Timestamp, 10))
	return req
}

func NewTestConfig() config.Config {
	return config.Config{
		DatabaseURL:            db.HorizonURL(),
		StellarCoreDatabaseURL: db.StellarCoreURL(),
		RedisURL:               RedisURL(),
		RateLimit:              &throttled.RateQuota{
			MaxRate: throttled.PerHour(1000),
			MaxBurst: 1000,
		},
		LogLevel:               hlog.DebugLevel,
		AdminSignatureValid:    time.Duration(60) * time.Second,
		StatisticsTimeout:      time.Duration(60) * time.Second,
		ProcessedOpTimeout:     time.Duration(30) * time.Second,
		BankMasterKey:          "GAWIB7ETYGSWULO4VB7D6S42YLPGIC7TY7Y2SSJKVOTMQXV5TILYWBUA", //SAWVTL2JG2HTPPABJZKN3GJEDTHT7YD3TW5XWAWPKAE2NNZPWNNBOIXE
	}
}
