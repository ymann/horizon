package session

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/openbankit/go-base/amount"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/admin"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/db2/history/details"
	"github.com/openbankit/horizon/ingest/participants"
	"github.com/openbankit/horizon/ingest/session/helpers"
	"github.com/openbankit/horizon/log"
	"encoding/json"
	"github.com/spf13/viper"
)

// Run starts an attempt to ingest the range of ledgers specified in this
// session.
func (is *Session) Run() error {
	err := is.Ingestion.Start()
	if err != nil {
		return err
	}

	defer is.Ingestion.Rollback()

	for is.Cursor.NextLedger() {
		err = is.clearLedger()
		if err != nil {
			return err
		}

		err = is.ingestLedger()
		if err != nil {
			return err
		}

		err = is.flush()
		if err != nil {
			return err
		}
	}

	return is.Ingestion.Close()

	// TODO: validate ledger chain

}

func (is *Session) clearLedger() error {
	if !is.ClearExisting {
		return nil
	}
	start := time.Now()
	err := is.Ingestion.Clear(is.Cursor.LedgerRange())
	if err != nil {
		is.Metrics.ClearLedgerTimer.Update(time.Since(start))
	}

	return err
}

func (is *Session) flush() error {
	return is.Ingestion.Flush()
}

// ingestLedger ingests the current ledger
func (is *Session) ingestLedger() error {
	start := time.Now()
	err := is.Ingestion.Ledger(
		is.Cursor.LedgerID(),
		is.Cursor.Ledger(),
		is.Cursor.SuccessfulTransactionCount(),
		is.Cursor.SuccessfulLedgerOperationCount(),
	)
	if err != nil {
		return err
	}

	// If this is ledger 1, create the root account
	if is.Cursor.LedgerSequence() == 1 {
		master := history.NewAccount(1, viper.GetString("bank-master-key"), xdr.AccountTypeAccountBank)
		err = is.Ingestion.Account(master, false, nil, nil)
		if err != nil {
			return err
		}

		commission := history.NewAccount(3, viper.GetString("bank-commission-key"), xdr.AccountTypeAccountCommission)
		if master.Address != commission.Address {
			err = is.Ingestion.Account(commission, false, nil, nil)
			if err != nil {
				return err
			}
		}
	}

	for is.Cursor.NextTx() {
		err = is.ingestTransaction()
		if err != nil {
			return err
		}
	}

	is.Ingested++
	if is.Metrics != nil {
		is.Metrics.IngestLedgerTimer.Update(time.Since(start))
	}

	return nil
}

func (is *Session) ingestOperation() error {
	err := is.Ingestion.Operation(
		is.Cursor.OperationID(),
		is.Cursor.TransactionID(),
		is.Cursor.OperationOrder(),
		is.Cursor.OperationSourceAccount(),
		is.Cursor.OperationType(),
		is.operationDetails(),
	)

	if err != nil {
		return err
	}

	switch is.Cursor.Operation().Body.Type {
	case xdr.OperationTypePayment:
		// Update statistics for both accounts
		op := is.Cursor.Operation().Body.MustPaymentOp()
		from := is.Cursor.OperationSourceAccount()
		to := op.Destination
		assetCode, err := getAssetCode(op.Asset)
		if err != nil {
			return err
		}
		err = is.ingestPayment(from.Address(), to.Address(), op.Amount, op.Amount, assetCode, assetCode)
		if err != nil {
			return err
		}
	case xdr.OperationTypePathPayment:
		op := is.Cursor.Operation().Body.MustPathPaymentOp()
		from := is.Cursor.OperationSourceAccount()
		to := op.Destination
		result := is.Cursor.OperationResult().MustPathPaymentResult()
		sourceAmount := result.SendAmount()
		destAmount := op.DestAmount
		sourceAsset, err := getAssetCode(op.SendAsset)
		if err != nil {
			return err
		}

		destAsset, err := getAssetCode(op.DestAsset)
		if err != nil {
			return err
		}

		err = is.ingestPayment(from.Address(), to.Address(), sourceAmount, destAmount, sourceAsset, destAsset)
		if err != nil {
			return err
		}
	case xdr.OperationTypeCreateAccount:
		// Import the new account if one was created
		op := is.Cursor.Operation().Body.MustCreateAccountOp()
		account := history.NewAccount(is.Cursor.OperationID(), op.Destination.Address(), xdr.AccountType(op.Body.AccountType))
		err = is.Ingestion.Account(account, true, nil, nil)
		if err != nil {
			return err
		}
	case xdr.OperationTypeAdministrative:
		logger := log.WithFields(log.F{
			"tx_hash":      is.Cursor.Transaction().TransactionHash,
			"operation_id": is.Cursor.OperationID(),
		})
		op := is.Cursor.Operation().Body.MustAdminOp()
		var opData  map[string]interface{}
		err = json.Unmarshal([]byte(op.OpData), &opData)
		if err != nil {
			return err
		}

		adminActionProvider := admin.NewAdminActionProvider(&history.Q{is.Ingestion.DB})
		adminAction, err := adminActionProvider.CreateNewParser(opData)
		if err != nil {
			return err
		}

		adminAction.Validate()
		if adminAction.GetError() != nil {
			logger.WithError(adminAction.GetError()).Error("Failed to validate admin action")
			break
		}
		adminAction.Apply()
		if adminAction.GetError() != nil {
			logger.WithError(adminAction.GetError()).Error("Failed to apply admin action")
			break
		}
	case xdr.OperationTypePaymentReversal:
		// Update statistics for both accounts
		op := is.Cursor.Operation().Body.MustPaymentReversalOp()
		reversalSource := is.Cursor.OperationSourceAccount()
		paymentSource := op.PaymentSource.Address()
		assetCode, err := getAssetCode(op.Asset)
		if err != nil {
			return err
		}

		err = is.ingestPaymentReversal(int64(op.PaymentId), reversalSource.Address(), paymentSource, assetCode, op.Amount)
		if err != nil {
			return err
		}
	case xdr.OperationTypeExternalPayment:
		// Update statistics for both accounts
		op := is.Cursor.Operation().Body.ExternalPaymentOp
		from := is.Cursor.OperationSourceAccount()
		to := op.ExchangeAgent

		assetCode, err := getAssetCode(op.Asset)
		if err != nil {
			return err
		}
		err = is.ingestPayment(from.Address(), to.Address(), op.Amount, op.Amount, assetCode, assetCode)
		if err != nil {
			return err
		}
	}

	err = is.ingestOperationParticipants()
	if err != nil {
		return err
	}

	return is.ingestEffects()
}

func (is *Session) ingestOperationParticipants() error {
	// Find the participants
	var p []xdr.AccountId
	p, err := participants.ForOperation(
		&is.Cursor.Transaction().Envelope.Tx,
		is.Cursor.Operation(),
	)
	if err != nil {
		return err
	}

	var aids []int64
	aids, err = is.lookupParticipantIDs(p)
	if err != nil {
		return err
	}

	return is.Ingestion.OperationParticipants(is.Cursor.OperationID(), aids)
}
func (is *Session) ingestTransaction() error {
	// skip ingesting failed transactions
	if !is.Cursor.Transaction().IsSuccessful() {
		return nil
	}

	err := is.Ingestion.Transaction(
		is.Cursor.TransactionID(),
		is.Cursor.Transaction(),
		is.Cursor.TransactionFee(),
	)
	if err != nil {
		return err
	}

	for is.Cursor.NextOp() {
		err = is.ingestOperation()
		if err != nil {
			return err
		}
	}

	return is.ingestTransactionParticipants()
}

func (is *Session) ingestTransactionParticipants() error {
	// Find the participants
	var p []xdr.AccountId
	p, err := participants.ForTransaction(
		&is.Cursor.Transaction().Envelope.Tx,
		&is.Cursor.Transaction().ResultMeta,
		&is.Cursor.TransactionFee().Changes,
	)
	if err != nil {
		return err
	}

	var aids []int64
	aids, err = is.lookupParticipantIDs(p)
	if err != nil {
		return err
	}

	return is.Ingestion.TransactionParticipants(is.Cursor.TransactionID(), aids)
}

func (is *Session) ingestEffects() error {
	effects := NewEffectIngestion(is.Ingestion, is.Cursor.OperationID())
	effects.Ingest(is.Cursor)
	return effects.Finish()
}

func (is *Session) lookupParticipantIDs(aids []xdr.AccountId) (ret []int64, err error) {
	found := map[int64]bool{}

	for _, in := range aids {
		var account *history.Account
		account, err = is.Ingestion.HistoryAccountCache.Get(in.Address())
		if err != nil {
			return
		}

		// De-duplicate
		if _, ok := found[account.ID]; ok {
			continue
		}

		found[account.ID] = true
		ret = append(ret, account.ID)
	}

	return
}

func getAssetCode(a xdr.Asset) (string, error) {
	var (
		t    string
		code string
		i    string
	)
	err := a.Extract(&t, &code, &i)

	return code, err
}

func (is *Session) feeDetails(xdrFee xdr.OperationFee) map[string]interface{} {
	fee := details.Fee{}
	fee.Populate(xdrFee)
	return fee.ToMap()
}

// operationDetails returns the details regarding the current operation, suitable
// for ingestion into a history_operation row
func (is *Session) operationDetails() map[string]interface{} {
	opDetails := map[string]interface{}{}
	c := is.Cursor
	source := c.OperationSourceAccount()

	fee := c.Transaction().Envelope.OperationFees[c.OperationOrder()-1]
	opDetails["fee"] = is.feeDetails(fee)

	switch c.OperationType() {
	case xdr.OperationTypeCreateAccount:
		op := c.Operation().Body.MustCreateAccountOp()
		opDetails["funder"] = source.Address()
		opDetails["account"] = op.Destination.Address()
		opDetails["account_type"] = uint32(op.Body.AccountType)
	case xdr.OperationTypePayment:
		op := c.Operation().Body.MustPaymentOp()
		opDetails["from"] = source.Address()
		opDetails["to"] = op.Destination.Address()
		opDetails["amount"] = amount.String(op.Amount)
		helpers.AssetDetails(opDetails, op.Asset, "")
	case xdr.OperationTypePathPayment:
		op := c.Operation().Body.MustPathPaymentOp()
		opDetails["from"] = source.Address()
		opDetails["to"] = op.Destination.Address()

		result := c.OperationResult().MustPathPaymentResult()

		opDetails["amount"] = amount.String(op.DestAmount)
		opDetails["source_amount"] = amount.String(result.SendAmount())
		opDetails["source_max"] = amount.String(op.SendMax)
		helpers.AssetDetails(opDetails, op.DestAsset, "")
		helpers.AssetDetails(opDetails, op.SendAsset, "source_")

		var path = make([]map[string]interface{}, len(op.Path))
		for i := range op.Path {
			path[i] = make(map[string]interface{})
			helpers.AssetDetails(path[i], op.Path[i], "")
		}
		opDetails["path"] = path
	case xdr.OperationTypeManageOffer:
		op := c.Operation().Body.MustManageOfferOp()
		opDetails["offer_id"] = op.OfferId
		opDetails["amount"] = amount.String(op.Amount)
		opDetails["price"] = op.Price.String()
		opDetails["price_r"] = map[string]interface{}{
			"n": op.Price.N,
			"d": op.Price.D,
		}
		helpers.AssetDetails(opDetails, op.Buying, "buying_")
		helpers.AssetDetails(opDetails, op.Selling, "selling_")

	case xdr.OperationTypeCreatePassiveOffer:
		op := c.Operation().Body.MustCreatePassiveOfferOp()
		opDetails["amount"] = amount.String(op.Amount)
		opDetails["price"] = op.Price.String()
		opDetails["price_r"] = map[string]interface{}{
			"n": op.Price.N,
			"d": op.Price.D,
		}
		helpers.AssetDetails(opDetails, op.Buying, "buying_")
		helpers.AssetDetails(opDetails, op.Selling, "selling_")
	case xdr.OperationTypeSetOptions:
		op := c.Operation().Body.MustSetOptionsOp()

		if op.InflationDest != nil {
			opDetails["inflation_dest"] = op.InflationDest.Address()
		}

		if op.SetFlags != nil && *op.SetFlags > 0 {
			is.operationFlagDetails(opDetails, int32(*op.SetFlags), "set")
		}

		if op.ClearFlags != nil && *op.ClearFlags > 0 {
			is.operationFlagDetails(opDetails, int32(*op.ClearFlags), "clear")
		}

		if op.MasterWeight != nil {
			opDetails["master_key_weight"] = *op.MasterWeight
		}

		if op.LowThreshold != nil {
			opDetails["low_threshold"] = *op.LowThreshold
		}

		if op.MedThreshold != nil {
			opDetails["med_threshold"] = *op.MedThreshold
		}

		if op.HighThreshold != nil {
			opDetails["high_threshold"] = *op.HighThreshold
		}

		if op.HomeDomain != nil {
			opDetails["home_domain"] = *op.HomeDomain
		}

		if op.Signer != nil {
			opDetails["signer_key"] = op.Signer.PubKey.Address()
			opDetails["signer_weight"] = op.Signer.Weight
		}
	case xdr.OperationTypeChangeTrust:
		op := c.Operation().Body.MustChangeTrustOp()
		helpers.AssetDetails(opDetails, op.Line, "")
		opDetails["trustor"] = source.Address()
		opDetails["trustee"] = opDetails["asset_issuer"]
		opDetails["limit"] = amount.String(op.Limit)
	case xdr.OperationTypeAllowTrust:
		op := c.Operation().Body.MustAllowTrustOp()
		helpers.AssetDetails(opDetails, op.Asset.ToAsset(source), "")
		opDetails["trustee"] = source.Address()
		opDetails["trustor"] = op.Trustor.Address()
		opDetails["authorize"] = op.Authorize
	case xdr.OperationTypeAccountMerge:
		aid := c.Operation().Body.MustDestination()
		opDetails["account"] = source.Address()
		opDetails["into"] = aid.Address()
	case xdr.OperationTypeInflation:
		// no inflation details, presently
	case xdr.OperationTypeManageData:
		op := c.Operation().Body.MustManageDataOp()
		opDetails["name"] = string(op.DataName)
		if op.DataValue != nil {
			opDetails["value"] = base64.StdEncoding.EncodeToString(*op.DataValue)
		} else {
			opDetails["value"] = nil
		}
	case xdr.OperationTypeAdministrative:
		op := c.Operation().Body.MustAdminOp()
		var adminOpDetails map[string]interface{}
		err := json.Unmarshal([]byte(op.OpData), &adminOpDetails)
		if err != nil {
			log.WithField("tx_hash", c.Transaction().TransactionHash).WithError(err).Error("Failed to unmarshal admin op details")
		}
		opDetails["details"] = adminOpDetails
	case xdr.OperationTypePaymentReversal:
		op := c.Operation().Body.MustPaymentReversalOp()
		opDetails["source_account"] = source.Address()
		opDetails["payment_source"] = op.PaymentSource.Address()
		opDetails["amount"] = amount.String(op.Amount)
		opDetails["commission"] = amount.String(op.CommissionAmount)
		opDetails["payment_id"] = int64(op.PaymentId)
		helpers.AssetDetails(opDetails, op.Asset, "")
	case xdr.OperationTypeExternalPayment:
		op := c.Operation().Body.MustExternalPaymentOp()
		opDetails["from"] = source.Address()
		opDetails["exchangeAgent"] = op.ExchangeAgent.Address()
		opDetails["destinationBank"] = op.DestinationBank.Address()
		opDetails["destinationAccount"] = op.DestinationAccount.Address()
		opDetails["amount"] = amount.String(op.Amount)
		helpers.AssetDetails(opDetails, op.Asset, "")
	default:
		panic(fmt.Errorf("Unknown operation type: %s", c.OperationType()))
	}

	return opDetails
}

// operationFlagDetails sets the account flag details for `f` on `result`.
func (is *Session) operationFlagDetails(result map[string]interface{}, f int32, prefix string) {
	var (
		n []int32
		s []string
	)

	if (f & int32(xdr.AccountFlagsAuthRequiredFlag)) > 0 {
		n = append(n, int32(xdr.AccountFlagsAuthRequiredFlag))
		s = append(s, "auth_required")
	}

	if (f & int32(xdr.AccountFlagsAuthRevocableFlag)) > 0 {
		n = append(n, int32(xdr.AccountFlagsAuthRevocableFlag))
		s = append(s, "auth_revocable")
	}

	if (f & int32(xdr.AccountFlagsAuthImmutableFlag)) > 0 {
		n = append(n, int32(xdr.AccountFlagsAuthImmutableFlag))
		s = append(s, "auth_immutable")
	}

	result[prefix+"_flags"] = n
	result[prefix+"_flags_s"] = s
}
