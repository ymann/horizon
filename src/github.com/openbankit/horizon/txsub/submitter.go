package txsub

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/accounttypes"
	"github.com/openbankit/horizon/cache"
	"github.com/openbankit/horizon/commissions"
	conf "github.com/openbankit/horizon/config"
	"github.com/openbankit/horizon/db2/core"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/render/problem"
	"github.com/openbankit/horizon/txsub/results"
	"github.com/openbankit/horizon/txsub/transactions"
	"github.com/openbankit/horizon/txsub/transactions/statistics"
	"github.com/go-errors/errors"
	"golang.org/x/net/context"
)

const (
	StatusError     = "ERROR"
	StatusPending   = "PENDING"
	StatusDuplicate = "DUPLICATE"
)

// NewDefaultSubmitter returns a new, simple Submitter implementation
// that submits directly to the stellar-core at `url` using the http client
// `h`.
func NewDefaultSubmitter(
	h *http.Client,
	url string,
	coreDb *core.Q,
	historyDb *history.Q,
	config *conf.Config,
	sharedCache *cache.SharedCache,
) Submitter {
	return createSubmitter(h, url, coreDb, historyDb, config, sharedCache)
}

// coreSubmissionResponse is the json response from stellar-core's tx endpoint
type coreSubmissionResponse struct {
	Exception string `json:"exception"`
	Error     string `json:"error"`
	Status    string `json:"status"`
}

// submitter is the default implementation for the Submitter interface.  It
// submits directly to the configured stellar-core instance using the
// configured http client.
type submitter struct {
	http     *http.Client
	coreURL  string
	coreQ    *core.Q
	historyQ *history.Q
	config   *conf.Config
	Log      *log.Entry

	defaultTxValidator TransactionValidatorInterface
	commissionManager  *commissions.CommissionsManager
}

func createSubmitter(h *http.Client, url string, coreDb *core.Q, historyDb *history.Q, config *conf.Config, sharedCache *cache.SharedCache) *submitter {
	return &submitter{
		http:               h,
		coreURL:            url,
		coreQ:              coreDb,
		historyQ:           historyDb,
		config:             config,
		commissionManager:  commissions.New(sharedCache, historyDb),
		defaultTxValidator: NewTransactionValidator(transactions.NewManager(coreDb, historyDb, statistics.NewManager(historyDb, accounttype.GetAll(), config), config, sharedCache)),
		Log:                log.WithField("service", "submitter"),
	}
}

// Submit sends the provided envelope to stellar-core and parses the response into
// a SubmissionResult
func (sub *submitter) Submit(ctx context.Context, env *transactions.EnvelopeInfo) (result SubmissionResult) {
	start := time.Now()
	defer func() { result.Duration = time.Since(start) }()

	sub.Log.Debug("Setting commission")
	err := sub.commissionManager.SetCommissions(env.Tx)
	if err != nil {
		log.WithField("Error", err).Error("Failed to set commissions")
		result.Err = &problem.ServerError
		return
	}

	// check constraints for tx
	sub.Log.Debug("Checking tx")
	err = sub.defaultTxValidator.CheckTransaction(env)
	if err != nil {
		result.Err = err
		return
	}

	updatedEnv, err := writeTransaction(env.Tx)
	if err != nil {
		result.Err = err
		return
	}

	// construct the request
	u, err := url.Parse(sub.coreURL)
	if err != nil {
		result.Err = errors.Wrap(err, 1)
		return
	}

	u.Path = "/tx"
	q := u.Query()
	q.Add("blob", *updatedEnv)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		result.Err = errors.Wrap(err, 1)
		return
	}

	// perform the submission
	resp, err := sub.http.Do(req)
	if err != nil {
		result.Err = errors.Wrap(err, 1)
		return
	}
	defer resp.Body.Close()

	// parse response
	var cresp coreSubmissionResponse
	err = json.NewDecoder(resp.Body).Decode(&cresp)
	if err != nil {
		result.Err = errors.Wrap(err, 1)
		return
	}

	// interpet response
	if cresp.Exception != "" {
		result.Err = errors.Errorf("stellar-core exception: %s", cresp.Exception)
		return
	}

	switch cresp.Status {
	case StatusError:
		result.Err = &results.FailedTransactionError{cresp.Error}
	case StatusPending, StatusDuplicate:
		//noop.  A nil Err indicates success
	default:
		result.Err = errors.Errorf("Unrecognized stellar-core status response: %s", cresp.Status)
	}

	return
}

func writeTransaction(tx *xdr.TransactionEnvelope) (*string, error) {
	res, err := xdr.MarshalBase64(tx)
	if err != nil {
		log.WithField("Erorr", err).Error("Failed to marshal tx")
		err = &results.MalformedTransactionError{}
		return nil, err
	}
	return &res, nil
}
