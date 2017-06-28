package horizon

import "github.com/openbankit/horizon/render/problem"

// NotFoundAction renders a 404 response
type NotFoundAction struct {
	Action
}

// JSON is a method for actions.JSON
func (action *NotFoundAction) JSON() {
	problem.Render(action.Ctx, action.W, problem.NotFound)
}
