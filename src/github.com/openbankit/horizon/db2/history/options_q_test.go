package history

import (
	"testing"

	"github.com/openbankit/horizon/test"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
)

func TestOptionsQ(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()

	h := Q{tt.HorizonRepo()}
	Convey("Given empty database", t, func() {
		Convey("On Select. Not found - returns nil, nil", func() {
			options, err := h.OptionsByName("random_name")
			assert.Nil(t, err)
			assert.Nil(t, options)
		})
		Convey("False if not updated", func() {
			options := &Options{
				Name: strconv.FormatInt(rand.Int63(), 10),
			}
			isUpdated, err := h.OptionsUpdate(options)
			assert.Nil(t, err)
			assert.False(t, isUpdated)
		})
		Convey("False if not deleted", func() {
			isDeleted, err := h.OptionsDelete(strconv.FormatInt(rand.Int63(), 10))
			assert.Nil(t, err)
			assert.False(t, isDeleted)
		})
		Convey("Insert", func() {
			expectedOptions := &Options{
				Name: "random_name_random",
				Data: "Lots of random data",
			}

			err := h.OptionsInsert(expectedOptions)
			assert.Nil(t, err)
			// can select
			actualOptions, err := h.OptionsByName(expectedOptions.Name)
			assert.Nil(t, err)
			assert.Equal(t, expectedOptions, actualOptions)

			// can update
			expectedOptions.Data = "updated data"
			isUpdated, err := h.OptionsUpdate(expectedOptions)
			assert.Nil(t, err)
			assert.True(t, isUpdated)

			actualOptions, err = h.OptionsByName(expectedOptions.Name)
			assert.Nil(t, err)
			assert.Equal(t, expectedOptions, actualOptions)

			// can delete
			isDeleted, err := h.OptionsDelete(expectedOptions.Name)
			assert.Nil(t, err)
			assert.True(t, isDeleted)

			actualOptions, err = h.OptionsByName(expectedOptions.Name)
			assert.Nil(t, err)
			assert.Nil(t, actualOptions)
		})
	})
}
