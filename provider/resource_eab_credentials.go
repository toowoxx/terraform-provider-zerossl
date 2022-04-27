package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/pkg/errors"
)

type resourceEABCredentialsType struct {
	ID      types.String `tfsdk:"id"`
	APIKey  types.String `tfsdk:"api_key"`
	KID     types.String `tfsdk:"kid"`
	HMACKey types.String `tfsdk:"hmac_key"`
}

type eabCredentialResponse struct {
	Success bool   `json:"success"`
	KID     string `json:"eab_kid"`
	HMACKey string `json:"eab_hmac_key"`
	Error   struct {
		Code int    `json:"code"`
		Type string `json:"type"`
	} `json:"error"`
}

func (r resourceEABCredentialsType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"api_key": {
				Type:        types.StringType,
				Required:    true,
				Description: "ZeroSSL API key ([View it here](https://app.zerossl.com/developer)]",
			},
			"kid": {
				Type:        types.StringType,
				Computed:    true,
				Sensitive:   true,
				Description: "kid of EAB credentials",
			},
			"hmac_key": {
				Type:        types.StringType,
				Computed:    true,
				Sensitive:   true,
				Description: "hmac_key of EAB credentials",
			},
		},
	}, nil
}

func (r resourceEABCredentialsType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceEABCredentials{
		p: *(p.(*provider)),
	}, nil
}

type resourceEABCredentials struct {
	p provider
}

func (r resourceEABCredentials) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

func (r resourceEABCredentials) updateState(resourceState *resourceEABCredentialsType) error {
	if resourceState.ID.Unknown {
		resourceState.ID = types.String{Value: uuid.Must(uuid.NewRandom()).String()}
	}

	return nil
}

func (r resourceEABCredentials) GenerateEABCredentials(
	ctx context.Context,
	resourceState *resourceEABCredentialsType,
) error {
	query := url.Values{"access_key": []string{resourceState.APIKey.Value}}
	endpoint := fmt.Sprintf("%s/acme/eab-credentials?%s", ZeroSSLBaseURL, query.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return errors.Wrap(err, "could not create request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "could not request EAB credentials")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("getting EAB credentials failed with HTTP status code %d %s",
			resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var eabResponse eabCredentialResponse
	err = json.NewDecoder(resp.Body).Decode(&eabResponse)

	if err != nil {
		return errors.Wrap(err, "could not decode response")
	}
	if eabResponse.Error.Code != 0 {
		return fmt.Errorf("could not get EAB credentials; server responded with "+
			"error type %s, error code %d and HTTP status code %d %s",
			eabResponse.Error.Type, eabResponse.Error.Code, resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	resourceState.KID = types.String{Value: eabResponse.KID}
	resourceState.HMACKey = types.String{Value: eabResponse.HMACKey}

	return nil
}

func (r resourceEABCredentials) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	resourceState := resourceEABCredentialsType{}
	diags := req.Config.Get(ctx, &resourceState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.GenerateEABCredentials(ctx, &resourceState); err != nil {
		resp.Diagnostics.AddError("could not generate EAB credentials", err.Error())
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &resourceState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceEABCredentials) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var resourceState resourceEABCredentialsType
	diags := req.State.Get(ctx, &resourceState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &resourceState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceEABCredentials) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan resourceEABCredentialsType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.GenerateEABCredentials(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("could not generate EAB credentials", err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceEABCredentials) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state resourceEABCredentialsType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.State.RemoveResource(ctx)
}
