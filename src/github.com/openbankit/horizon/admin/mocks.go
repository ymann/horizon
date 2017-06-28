package admin

import "github.com/stretchr/testify/mock"

type AdminActionProviderMock struct {
	mock.Mock
}

func (m *AdminActionProviderMock) CreateNewParser(data map[string]interface{}) (AdminActionInterface, error) {
	a := m.Called(data)
	admin := a.Get(0)
	err := a.Error(1)
	if admin == nil {
		return nil, err
	}
	return admin.(AdminActionInterface), err
}

type AdminActionMock struct {
	mock.Mock
}

func (m *AdminActionMock) GetError() error {
	return m.Called().Error(0)
}

func (m *AdminActionMock) Validate() {
	return
}

func (m *AdminActionMock) Apply() {
	return
}

