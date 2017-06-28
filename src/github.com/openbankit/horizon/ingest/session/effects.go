package session

import (
	"github.com/openbankit/go-base/amount"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/ingest/session/helpers"
	"github.com/openbankit/horizon/ingest/session/ingestion"
)

// EffectIngestion is a helper struct to smooth the ingestion of effects.  this
// struct will track what the correct operation to use and order to use when
// adding effects into an ingestion.
type EffectIngestion struct {
	Dest        *ingestion.Ingestion
	OperationID int64
	err         error
	added       int
}

func NewEffectIngestion(dest *ingestion.Ingestion, operationId int64) *EffectIngestion {
	return &EffectIngestion{
		Dest:        dest,
		OperationID: operationId,
	}
}

// Add writes an effect to the database while automatically tracking the index
// to use.
func (ei *EffectIngestion) Add(aid xdr.AccountId, typ history.EffectType, details interface{}) bool {
	if ei.err != nil {
		return false
	}

	ei.added++
	var account *history.Account
	account, ei.err = ei.Dest.HistoryAccountCache.Get(aid.Address())
	if ei.err != nil {
		return false
	}

	ei.err = ei.Dest.Effect(account.ID, ei.OperationID, ei.added, typ, details)
	if ei.err != nil {
		return false
	}

	return true
}

// Finish marks this ingestion as complete, returning any error that was recorded.
func (ei *EffectIngestion) Finish() error {
	err := ei.err
	ei.err = nil
	return err
}

func (effects *EffectIngestion) Ingest(cursor *Cursor) {
	source := cursor.OperationSourceAccount()
	opbody := cursor.Operation().Body

	switch cursor.OperationType() {
	case xdr.OperationTypePathPayment:
		op := opbody.MustPathPaymentOp()
		dets := map[string]interface{}{"amount": amount.String(op.DestAmount)}
		helpers.AssetDetails(dets, op.DestAsset, "")
		effects.Add(op.Destination, history.EffectAccountCredited, dets)

		result := cursor.OperationResult().MustPathPaymentResult()
		dets = map[string]interface{}{"amount": amount.String(result.SendAmount())}
		helpers.AssetDetails(dets, op.SendAsset, "")
		effects.Add(source, history.EffectAccountDebited, dets)
		effects.ingestTrades(source, result.MustSuccess().Offers)
	case xdr.OperationTypeManageOffer:
		result := cursor.OperationResult().MustManageOfferResult().MustSuccess()
		effects.ingestTrades(source, result.OffersClaimed)
	case xdr.OperationTypeCreatePassiveOffer:
		claims := []xdr.ClaimOfferAtom{}
		result := cursor.OperationResult()

		// KNOWN ISSUE:  stellar-core creates results for CreatePassiveOffer operations
		// with the wrong result arm set.
		if result.Type == xdr.OperationTypeManageOffer {
			claims = result.MustManageOfferResult().MustSuccess().OffersClaimed
		} else {
			claims = result.MustCreatePassiveOfferResult().MustSuccess().OffersClaimed
		}

		effects.ingestTrades(source, claims)
	default:
		return
	}
}

func (effects *EffectIngestion) ingestTrades(buyer xdr.AccountId, claims []xdr.ClaimOfferAtom) {
	for _, claim := range claims {
		seller := claim.SellerId
		bd, sd := helpers.TradeDetails(buyer, seller, claim)
		effects.Add(buyer, history.EffectTrade, bd)
		effects.Add(seller, history.EffectTrade, sd)
	}
}

func (effects *EffectIngestion) ingestSignerEffects(cursor *Cursor, op xdr.SetOptionsOp) {
	source := cursor.OperationSourceAccount()

	be, ae, err := cursor.BeforeAndAfter(source.LedgerKey())
	if err != nil {
		effects.err = err
		return
	}

	beforeAccount := be.Data.MustAccount()
	afterAccount := ae.Data.MustAccount()

	before := beforeAccount.SignerSummary()
	after := afterAccount.SignerSummary()

	for addy := range before {
		weight, ok := after[addy]
		if !ok {
			effects.Add(source, history.EffectSignerRemoved, map[string]interface{}{
				"public_key": addy,
			})
			continue
		}
		effects.Add(source, history.EffectSignerUpdated, map[string]interface{}{
			"public_key": addy,
			"weight":     weight,
		})
	}
	// Add the "created" effects
	for addy, weight := range after {
		// if `addy` is in before, the previous for loop should have recorded
		// the update, so skip this key
		if _, ok := before[addy]; ok {
			continue
		}

		effects.Add(source, history.EffectSignerCreated, map[string]interface{}{
			"public_key": addy,
			"weight":     weight,
		})
	}

}
