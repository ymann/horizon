package statistics

import (
	"github.com/openbankit/horizon/db2/core"
	"github.com/openbankit/horizon/db2/history"
)

type PaymentDirection string

const (
	PaymentDirectionOutgoing PaymentDirection = "outgoing"
	PaymentDirectionIncoming PaymentDirection = "incoming"
)

func (d *PaymentDirection) IsIncoming() bool {
	return *d == PaymentDirectionIncoming
}

type OperationData struct {
	Source *history.Account
	Index  int
	TxHash string
}

func NewOperationData(source *history.Account, index int, txHash string) OperationData {
	return OperationData{
		Source: source,
		Index:  index,
		TxHash: txHash,
	}
}

type PaymentData struct {
	OperationData
	Destination          *history.Account
	DestinationTrustLine *core.Trustline
	Amount               int64
	Asset                history.Asset
}

func NewPaymentData(destination *history.Account, destinationTrustLine *core.Trustline, opAsset history.Asset, opAmount int64, opData OperationData) PaymentData {
	return PaymentData{
		OperationData:        opData,
		Destination:          destination,
		DestinationTrustLine: destinationTrustLine,
		Amount:               opAmount,
		Asset:                opAsset,
	}
}

func (p *PaymentData) GetAccountTrustLine(direction PaymentDirection) *core.Trustline {
	if direction == PaymentDirectionIncoming {
		return p.DestinationTrustLine
	}
	return nil
}

func (p *PaymentData) GetAccount(direction PaymentDirection) *history.Account {
	if direction == PaymentDirectionOutgoing {
		return p.Source
	}
	return p.Destination
}

func (p *PaymentData) GetCounterparty(direction PaymentDirection) *history.Account {
	if direction == PaymentDirectionIncoming {
		return p.Source
	}
	return p.Destination
}
