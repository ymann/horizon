package transactions

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/core"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/txsub/results"
	"github.com/openbankit/horizon/txsub/transactions/statistics"
	"github.com/openbankit/horizon/txsub/transactions/validators"
	"database/sql"
)

type PathPaymentOpFrame struct {
	*OperationFrame
	pathPayment               xdr.PathPaymentOp
	sendAsset                 history.Asset
	destAsset                 history.Asset
	destAccount               *history.Account
	destTrustline             core.Trustline
	isDestExists              bool

	accountTypeValidator      validators.AccountTypeValidatorInterface
	assetsValidator           validators.AssetsValidatorInterface
	traitsValidator           validators.TraitsValidatorInterface
	defaultOutLimitsValidator validators.OutgoingLimitsValidatorInterface
	defaultInLimitsValidator  validators.IncomingLimitsValidatorInterface
}

func NewPathPaymentOpFrame(opFrame *OperationFrame) *PathPaymentOpFrame {
	return &PathPaymentOpFrame{
		OperationFrame: opFrame,
		pathPayment:    opFrame.Op.Body.MustPathPaymentOp(),
	}
}

func (p *PathPaymentOpFrame) GetTraitsValidator() validators.TraitsValidatorInterface {
	if p.traitsValidator == nil {
		p.traitsValidator = validators.NewTraitsValidator()
	}
	return p.traitsValidator
}

func (p *PathPaymentOpFrame) GetAccountTypeValidator() validators.AccountTypeValidatorInterface {
	if p.accountTypeValidator == nil {
		p.accountTypeValidator = validators.NewAccountTypeValidator()
	}
	return p.accountTypeValidator
}

func (p *PathPaymentOpFrame) GetOutgoingLimitsValidator(paymentData *statistics.PaymentData, manager *Manager) validators.OutgoingLimitsValidatorInterface {
	if p.defaultOutLimitsValidator != nil {
		return p.defaultOutLimitsValidator
	}
	return validators.NewOutgoingLimitsValidator(paymentData, manager.StatsManager, manager.HistoryQ, manager.Config.AnonymousUserRestrictions, *p.now)
}

func (p *PathPaymentOpFrame) GetIncomingLimitsValidator(paymentData *statistics.PaymentData, manager *Manager) validators.IncomingLimitsValidatorInterface {
	if p.defaultInLimitsValidator != nil {
		return p.defaultInLimitsValidator
	}
	return validators.NewIncomingLimitsValidator(paymentData, manager.HistoryQ, manager.StatsManager, manager.Config.AnonymousUserRestrictions, *p.now)
}

func (p *PathPaymentOpFrame) GetAssetsValidator(historyQ history.QInterface) validators.AssetsValidatorInterface {
	if p.assetsValidator == nil {
		p.log.Debug("Creating new assets validator")
		p.assetsValidator = validators.NewAssetsValidator(historyQ)
	}
	return p.assetsValidator
}

func (p *PathPaymentOpFrame) DoCheckValid(manager *Manager) (bool, error) {
	// check if all assets are valid
	isAssetsValid, err := p.isAssetsValid(manager.HistoryQ)
	if err != nil {
		p.log.Error("Failed to validate assets")
		return false, err
	}

	if !isAssetsValid {
		p.getInnerResult().Code = xdr.PathPaymentResultCodePathPaymentMalformed
		p.Result.Info = results.AdditionalErrorInfoError(ASSET_NOT_ALLOWED)
		return false, nil
	}

	// check if destination exists or asset is anonymous
	p.isDestExists, err = p.tryLoadDestinationAccount(manager)
	if err != nil {
		return false, err
	}

	if !p.isDestExists && !p.destAsset.IsAnonymous {
		p.getInnerResult().Code = xdr.PathPaymentResultCodePathPaymentNoDestination
		return false, nil
	}

	// check if destination trust line exists or (dest account does not exist and asset is anonymous)
	// if p.isDestExists {
		// err = manager.CoreQ.TrustlineByAddressAndAsset(&p.destTrustline, p.pathPayment.Destination.Address(), p.destAsset.Code, p.destAsset.Issuer)
		// if err != nil {
		// 	if err != sql.ErrNoRows {
		// 		return false, err
		// 	}
		// 	p.getInnerResult().Code = xdr.PathPaymentResultCodePathPaymentNoTrust
		// 	return false, nil
		// }
	// }

	isLimitsValid, err := p.checkLimits(manager)
	if err != nil || !isLimitsValid {
		return isLimitsValid, err
	}

	p.getInnerResult().Code = xdr.PathPaymentResultCodePathPaymentSuccess
	return true, nil
}

func (p *PathPaymentOpFrame) tryLoadDestinationAccount(manager *Manager) (bool, error) {
	var err error
	p.destAccount, err = manager.AccountHistoryCache.Get(p.pathPayment.Destination.Address())
	if err == nil {
		return true, nil
	} else if err != sql.ErrNoRows {
		return false, err
	}
	p.destAccount = new(history.Account)
	p.destAccount.Address = p.pathPayment.Destination.Address()
	p.destAccount.AccountType = xdr.AccountTypeAccountAnonymousUser
	return false, nil
}

func (p *PathPaymentOpFrame) getInnerResult() *xdr.PathPaymentResult {
	if p.Result.Result.Tr.PathPaymentResult == nil {
		p.Result.Result.Tr.PathPaymentResult = &xdr.PathPaymentResult{}
	}
	return p.Result.Result.Tr.PathPaymentResult
}

func (p *PathPaymentOpFrame) isAssetsValid(historyQ history.QInterface) (bool, error) {
	// check if assets are valid
	assetsValidator := p.GetAssetsValidator(historyQ)
	var err error
	sendAsset, err := assetsValidator.GetValidAsset(p.pathPayment.SendAsset)
	if err != nil || sendAsset == nil {
		return false, err
	}

	p.sendAsset = *sendAsset

	destAsset, err := assetsValidator.GetValidAsset(p.pathPayment.DestAsset)
	if err != nil || destAsset == nil {
		return false, err
	}

	p.destAsset = *destAsset

	if p.pathPayment.Path != nil {
		return assetsValidator.IsAssetsValid(p.pathPayment.Path...)
	}

	return true, nil
}

func (p *PathPaymentOpFrame) checkLimits(manager *Manager) (bool, error) {

	// 1. Check account types
	p.log.Debug("Validating account types")
	accountTypesRestricted := p.GetAccountTypeValidator().VerifyAccountTypesForPayment(p.SourceAccount.AccountType, p.destAccount.AccountType)
	if accountTypesRestricted != nil {
		p.getInnerResult().Code = xdr.PathPaymentResultCodePathPaymentMalformed
		p.Result.Info = results.AdditionalErrorInfoError(accountTypesRestricted)
		return false, nil
	}

	// 2. Check traits for accounts
	p.log.WithField("sourceAccount", p.SourceAccount.Address).WithField("destAccount", p.destAccount.Address).Debug("Checking traits")
	accountRestricted, err := p.GetTraitsValidator().CheckTraits(p.SourceAccount, p.destAccount)
	if err != nil {
		return false, err
	}

	if accountRestricted != nil {
		p.getInnerResult().Code = xdr.PathPaymentResultCodePathPaymentMalformed
		p.Result.Info = results.AdditionalErrorInfoError(accountRestricted)
		return false, nil
	}

	// 3. Check restrictions for sender
	operationData := statistics.NewOperationData(p.SourceAccount, p.Index, p.ParentTxFrame.TxHash)
	outPaymentData := statistics.NewPaymentData(p.destAccount, &p.destTrustline, p.sendAsset, int64(p.pathPayment.SendMax), operationData)
	outgoingValidator := p.GetOutgoingLimitsValidator(&outPaymentData, manager)
	outLimitsResult, err := outgoingValidator.VerifyLimits()
	if err != nil {
		return false, err
	}

	if outLimitsResult != nil {
		p.getInnerResult().Code = xdr.PathPaymentResultCodePathPaymentMalformed
		p.Result.Info = results.AdditionalErrorInfoError(outLimitsResult)
		return false, nil
	}

	inPaymentData := statistics.NewPaymentData(p.destAccount, &p.destTrustline, p.destAsset, int64(p.pathPayment.DestAmount), operationData)
	incomingValidator := p.GetIncomingLimitsValidator(&inPaymentData, manager)
	inLimitsResult, err := incomingValidator.VerifyLimits()
	if err != nil {
		return false, err
	}

	if inLimitsResult != nil {
		p.getInnerResult().Code = xdr.PathPaymentResultCodePathPaymentMalformed
		p.Result.Info = results.AdditionalErrorInfoError(inLimitsResult)
		return false, nil
	}

	return true, nil
}

func (p *PathPaymentOpFrame) DoRollbackCachedData(manager *Manager) error {
	p.log.Debug("Rollingback path payment")
	// 3. Check restrictions for sender
	operationData := statistics.NewOperationData(p.SourceAccount, p.Index, p.ParentTxFrame.TxHash)
	outPaymentData := statistics.NewPaymentData(p.destAccount, &p.destTrustline, p.sendAsset, int64(p.pathPayment.SendMax), operationData)
	err := manager.StatsManager.CancelOp(&outPaymentData, statistics.PaymentDirectionOutgoing, *p.now)
	if err != nil {
		p.log.WithError(err).Error("Failed to rollback outgoing payment part")
		return err
	}

	inPaymentData := statistics.NewPaymentData(p.destAccount, &p.destTrustline, p.destAsset, int64(p.pathPayment.DestAmount), operationData)
	err = manager.StatsManager.CancelOp(&inPaymentData, statistics.PaymentDirectionIncoming, *p.now)
	if err != nil {
		p.log.WithError(err).Error("Failed to rollback incoming payment part")
		return err
	}
	return nil
}
