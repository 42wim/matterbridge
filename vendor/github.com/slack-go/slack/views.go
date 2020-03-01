package slack

import (
	"context"
	"encoding/json"
)

const (
	VTModal   ViewType = "modal"
	VTHomeTab ViewType = "home"
)

type ViewType string

type View struct {
	SlackResponse
	ID              string           `json:"id"`
	TeamID          string           `json:"team_id"`
	Type            ViewType         `json:"type"`
	Title           *TextBlockObject `json:"title"`
	Close           *TextBlockObject `json:"close"`
	Submit          *TextBlockObject `json:"submit"`
	Blocks          Blocks           `json:"blocks"`
	PrivateMetadata string           `json:"private_metadata"`
	CallbackID      string           `json:"callback_id"`
	State           interface{}      `json:"state"`
	Hash            string           `json:"hash"`
	ClearOnClose    bool             `json:"clear_on_close"`
	NotifyOnClose   bool             `json:"notify_on_close"`
	RootViewID      string           `json:"root_view_id"`
	PreviousViewID  string           `json:"previous_view_id"`
	AppID           string           `json:"app_id"`
	ExternalID      string           `json:"external_id"`
	BotID           string           `json:"bot_id"`
}

type ModalViewRequest struct {
	Type            ViewType         `json:"type"`
	Title           *TextBlockObject `json:"title"`
	Blocks          Blocks           `json:"blocks"`
	Close           *TextBlockObject `json:"close"`
	Submit          *TextBlockObject `json:"submit"`
	PrivateMetadata string           `json:"private_metadata"`
	CallbackID      string           `json:"callback_id"`
	ClearOnClose    bool             `json:"clear_on_close"`
	NotifyOnClose   bool             `json:"notify_on_close"`
	ExternalID      string           `json:"external_id"`
}

func (v *ModalViewRequest) ViewType() ViewType {
	return v.Type
}

type HomeTabViewRequest struct {
	Type            ViewType `json:"type"`
	Blocks          Blocks   `json:"blocks"`
	PrivateMetadata string   `json:"private_metadata"`
	CallbackID      string   `json:"callback_id"`
	ExternalID      string   `json:"external_id"`
}

func (v *HomeTabViewRequest) ViewType() ViewType {
	return v.Type
}

type openViewRequest struct {
	TriggerID string           `json:"trigger_id"`
	View      ModalViewRequest `json:"view"`
}

type publishViewRequest struct {
	UserID string             `json:"user_id"`
	View   HomeTabViewRequest `json:"view"`
	Hash   string             `json:"hash"`
}

type pushViewRequest struct {
	TriggerID string           `json:"trigger_id"`
	View      ModalViewRequest `json:"view"`
}

type updateViewRequest struct {
	View       ModalViewRequest `json:"view"`
	ExternalID string           `json:"external_id"`
	Hash       string           `json:"hash"`
	ViewID     string           `json:"view_id"`
}

type ViewResponse struct {
	SlackResponse
	View `json:"view"`
}

// OpenView opens a view for a user.
func (api *Client) OpenView(triggerID string, view ModalViewRequest) (*ViewResponse, error) {
	return api.OpenViewContext(context.Background(), triggerID, view)
}

// OpenViewContext opens a view for a user with a custom context.
func (api *Client) OpenViewContext(
	ctx context.Context,
	triggerID string,
	view ModalViewRequest,
) (*ViewResponse, error) {
	if triggerID == "" {
		return nil, ErrParametersMissing
	}
	req := openViewRequest{
		TriggerID: triggerID,
		View:      view,
	}
	encoded, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	endpoint := api.endpoint + "views.open"
	resp := &ViewResponse{}
	err = postJSON(ctx, api.httpclient, endpoint, api.token, encoded, resp, api)
	if err != nil {
		return nil, err
	}
	return resp, resp.Err()
}

// PublishView publishes a static view for a user.
func (api *Client) PublishView(userID string, view HomeTabViewRequest, hash string) (*ViewResponse, error) {
	return api.PublishViewContext(context.Background(), userID, view, hash)
}

// PublishViewContext publishes a static view for a user with a custom context.
func (api *Client) PublishViewContext(
	ctx context.Context,
	userID string,
	view HomeTabViewRequest,
	hash string,
) (*ViewResponse, error) {
	if userID == "" {
		return nil, ErrParametersMissing
	}
	req := publishViewRequest{
		UserID: userID,
		View:   view,
		Hash:   hash,
	}
	encoded, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	endpoint := api.endpoint + "views.publish"
	resp := &ViewResponse{}
	err = postJSON(ctx, api.httpclient, endpoint, api.token, encoded, resp, api)
	if err != nil {
		return nil, err
	}
	return resp, resp.Err()
}

// PushView pushes a view onto the stack of a root view.
func (api *Client) PushView(triggerID string, view ModalViewRequest) (*ViewResponse, error) {
	return api.PushViewContext(context.Background(), triggerID, view)
}

// PublishViewContext pushes a view onto the stack of a root view with a custom context.
func (api *Client) PushViewContext(
	ctx context.Context,
	triggerID string,
	view ModalViewRequest,
) (*ViewResponse, error) {
	if triggerID == "" {
		return nil, ErrParametersMissing
	}
	req := pushViewRequest{
		TriggerID: triggerID,
		View:      view,
	}
	encoded, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	endpoint := api.endpoint + "views.push"
	resp := &ViewResponse{}
	err = postJSON(ctx, api.httpclient, endpoint, api.token, encoded, resp, api)
	if err != nil {
		return nil, err
	}
	return resp, resp.Err()
}

// UpdateView updates an existing view.
func (api *Client) UpdateView(view ModalViewRequest, externalID, hash, viewID string) (*ViewResponse, error) {
	return api.UpdateViewContext(context.Background(), view, externalID, hash, viewID)
}

// UpdateViewContext updates an existing view with a custom context.
func (api *Client) UpdateViewContext(
	ctx context.Context,
	view ModalViewRequest,
	externalID, hash,
	viewID string,
) (*ViewResponse, error) {
	if externalID == "" && viewID == "" {
		return nil, ErrParametersMissing
	}
	req := updateViewRequest{
		View:       view,
		ExternalID: externalID,
		Hash:       hash,
		ViewID:     viewID,
	}
	encoded, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	endpoint := api.endpoint + "views.update"
	resp := &ViewResponse{}
	err = postJSON(ctx, api.httpclient, endpoint, api.token, encoded, resp, api)
	if err != nil {
		return nil, err
	}
	return resp, resp.Err()
}
