package admin

import (
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/render/problem"
	"github.com/go-errors/errors"
	"time"
)

type ManageMaxReversalDurationAction struct {
	AdminAction
	maxReversalDurationSec int64
}

func NewManageMaxReversalDurationAction(adminAction AdminAction) *ManageMaxReversalDurationAction {
	return &ManageMaxReversalDurationAction{
		AdminAction: adminAction,
	}
}

func (action *ManageMaxReversalDurationAction) Validate() {
	action.loadParams()
	if action.Err != nil {
		return
	}
}

func (action *ManageMaxReversalDurationAction) Apply() {
	if action.Err != nil {
		return
	}

	exists, err := action.exists()
	if err != nil {
		action.Log.WithError(err).Error("Failed to check if max reversal duration option exists")
		action.Err = &problem.ServerError
		return
	}

	maxReversalDuration := history.NewMaxReversalDuration()
	maxReversalDuration.SetMaxDuration(time.Duration(action.maxReversalDurationSec) * time.Second)
	option := history.Options(*maxReversalDuration)
	if exists {
		_, err = action.HistoryQ().OptionsUpdate(&option)
	} else {
		err = action.HistoryQ().OptionsInsert(&option)
	}

	if err != nil {
		action.Log.WithError(err).Error("Failed to insert/update max reversal duration")
		action.Err = &problem.ServerError
		return
	}
}

func (action *ManageMaxReversalDurationAction) exists() (bool, error) {
	stored, err := action.HistoryQ().OptionsByName(history.OPTIONS_MAX_REVERSAL_DURATION)
	return stored != nil, err
}

func (action *ManageMaxReversalDurationAction) loadParams() {
	action.maxReversalDurationSec = action.GetInt64("max_reversal_duration")
	if action.Err != nil {
		return
	}

	if action.maxReversalDurationSec < 0 {
		action.SetInvalidField("max_reversal_duration", errors.New("Must not be negative"))
		return
	}
}
