package admin

import (
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/test"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestAdminActionProvider(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()

	log.DefaultLogger.Entry.Logger.Level = log.DebugLevel
	historyQ := &history.Q{tt.HorizonRepo()}

	Convey("Set commission Actions:", t, func() {
		actionProvider := NewAdminActionProvider(historyQ)
		Convey("Several data objects", func() {
			_, err := actionProvider.CreateNewParser(map[string]interface{} {
				string(SubjectCommission): map[string]interface{}{},
				string(SubjectTraits): map[string]interface{}{},
			})
			So(err, ShouldNotBeNil)
		})
		Convey("Unknown action type", func() {
			_, err := actionProvider.CreateNewParser(map[string]interface{} {
				"random_action": map[string]interface{}{},
			})
			So(err, ShouldNotBeNil)
		})
		Convey("Invalid type", func() {
			_, err := actionProvider.CreateNewParser(map[string]interface{} {
				string(SubjectAccountLimits): "random_data",
			})
			So(err, ShouldNotBeNil)
		})
		Convey("Create commission set action", func() {
			action, err := actionProvider.CreateNewParser(map[string]interface{} {
				string(SubjectCommission): map[string]interface{}{},
			})
			So(err, ShouldBeNil)
			switch action.(type) {
			case *SetCommissionAction:
				//ok
			default:
				//not ok
				assert.Fail(t, "Expected SetCommissionAction")
			}
		})
		Convey("Create limits set action", func() {
			action, err := actionProvider.CreateNewParser(map[string]interface{} {
				string(SubjectAccountLimits): map[string]interface{}{},
			})
			So(err, ShouldBeNil)
			switch action.(type) {
			case *SetLimitsAction:
			//ok
			default:
				//not ok
				assert.Fail(t, "Expected SetLimitsAction")
			}
		})
		Convey("Set traits action", func() {
			action, err := actionProvider.CreateNewParser(map[string]interface{} {
				string(SubjectTraits): map[string]interface{}{},
			})
			So(err, ShouldBeNil)
			switch action.(type) {
			case *SetTraitsAction:
			//ok
			default:
				//not ok
				assert.Fail(t, "Expected SetTraitsAction")
			}
		})

		Convey("Manage assets action", func() {
			action, err := actionProvider.CreateNewParser(map[string]interface{} {
				string(SubjectAsset): map[string]interface{}{},
			})
			So(err, ShouldBeNil)
			switch action.(type) {
			case *ManageAssetsAction:
			//ok
			default:
				//not ok
				assert.Fail(t, "Expected ManageAssetsAction")
			}
		})
	})
}
