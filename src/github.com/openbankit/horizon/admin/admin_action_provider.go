package admin

import (
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"fmt"
	"github.com/go-errors/errors"
)

type AdminActionProviderInterface interface {
	CreateNewParser(data map[string]interface{}) (AdminActionInterface, error)
}

type AdminActionProvider struct {
	log      *log.Entry
	historyQ history.QInterface
}

func NewAdminActionProvider(historyQ history.QInterface) *AdminActionProvider {
	return &AdminActionProvider{
		log:      log.WithField("service", "admin_action_provider"),
		historyQ: historyQ,
	}
}

func (p *AdminActionProvider) CreateNewParser(data map[string]interface{}) (AdminActionInterface, error) {
	if len(data) > 1 {
		return nil, errors.New("Only one operation per time can be processed")
	}
	for key, rawValue := range data {
		value, err := getAdminActionData(rawValue, key)
		if err != nil {
			return nil, err
		}
		adminAction := NewAdminAction(value, p.historyQ)
		switch AdminActionSubject(key) {
		case SubjectCommission:
			return NewSetCommissionAction(adminAction), nil
		case SubjectAccountLimits:
			return NewSetLimitsAction(adminAction), nil
		case SubjectTraits:
			return NewSetTraitsAction(adminAction), nil
		case SubjectAsset:
			return NewManageAssetsAction(adminAction), nil
		case SubjectMaxPaymentReversalDuration:
			return NewManageMaxReversalDurationAction(adminAction), nil
		default:
			return nil, errors.New("unknown admin action")
		}
	}
	return nil, errors.New("data can't be empty")
}

func getAdminActionData(rawValue interface{}, key string) (result map[string]interface{}, err error) {
	switch rawValue.(type) {
	case map[string]interface{}:
		result = rawValue.(map[string]interface{})
	default:
		return nil, errors.New(fmt.Sprintf("Value of %s must be object", key))
	}
	return result, nil
}
