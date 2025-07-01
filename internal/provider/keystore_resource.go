package provider

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strconv"

    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &KeyStoreResource{}
var _ resource.ResourceWithImportState = &KeyStoreResource{}

func NewKeyStoreResource() resource.Resource {
    return &KeyStoreResource{}
}

// KeyStoreResource defines the resource implementation.
type KeyStoreResource struct {
    client *ClientConfig
}

// KeyStoreResourceModel describes the resource data model based on GlobalKVStore Django model
type KeyStoreResourceModel struct {
    Id    types.Int64  `tfsdk:"id"`
    Name  types.String `tfsdk:"name"`
    Value types.String `tfsdk:"value"`
}

func (r *KeyStoreResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_keystore"
}

func (r *KeyStoreResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Global Key-Value Store resource for Tactical RMM",

        Attributes: map[string]schema.Attribute{
            "id": schema.Int64Attribute{
                MarkdownDescription: "KeyStore identifier",
                Computed:            true,
            },
            "name": schema.StringAttribute{
                MarkdownDescription: "Key name (max 25 characters)",
                Required:            true,
            },
            "value": schema.StringAttribute{
                MarkdownDescription: "Key value",
                Required:            true,
                Sensitive:           true,
            },
        },
    }
}

func (r *KeyStoreResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*ClientConfig)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Resource Configure Type",
            fmt.Sprintf("Expected *ClientConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
        )
        return
    }

    r.client = client
}

func (r *KeyStoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data KeyStoreResourceModel

    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Create API request body
    body := map[string]interface{}{
        "name":  data.Name.ValueString(),
        "value": data.Value.ValueString(),
    }

    jsonBody, err := json.Marshal(body)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create keystore entry, got error: %s", err))
        return
    }

    // Create HTTP request
    httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/core/keystore/", r.client.BaseURL), bytes.NewBuffer(jsonBody))
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create keystore entry, got error: %s", err))
        return
    }

    // Make request
    httpResp, err := r.client.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create keystore entry, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create keystore entry, status code: %d", httpResp.StatusCode))
        return
    }

    // Response is just "ok", so we need to get the created entry
    // List all keystore entries to find our newly created one
    listReq, err := http.NewRequest("GET", fmt.Sprintf("%s/core/keystore/", r.client.BaseURL), nil)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list keystore entries, got error: %s", err))
        return
    }

    listResp, err := r.client.Do(listReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list keystore entries, got error: %s", err))
        return
    }
    defer listResp.Body.Close()

    var entries []map[string]interface{}
    if err := json.NewDecoder(listResp.Body).Decode(&entries); err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse keystore entries list, got error: %s", err))
        return
    }

    // Find the entry we just created by name
    var createdEntry map[string]interface{}
    for _, entry := range entries {
        if name, ok := entry["name"].(string); ok && name == data.Name.ValueString() {
            createdEntry = entry
            break
        }
    }

    if createdEntry == nil {
        resp.Diagnostics.AddError("Client Error", "Unable to find created keystore entry")
        return
    }

    // Update model with response data
    if id, ok := createdEntry["id"].(float64); ok {
        data.Id = types.Int64Value(int64(id))
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeyStoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var data KeyStoreResourceModel

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Get all keystore entries since there's no individual GET endpoint
    httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/core/keystore/", r.client.BaseURL), nil)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read keystore entries, got error: %s", err))
        return
    }

    // Make request
    httpResp, err := r.client.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read keystore entries, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read keystore entries, status code: %d", httpResp.StatusCode))
        return
    }

    // Parse response
    var entries []map[string]interface{}
    if err := json.NewDecoder(httpResp.Body).Decode(&entries); err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
        return
    }

    // Find our entry by ID
    var found bool
    for _, entry := range entries {
        if id, ok := entry["id"].(float64); ok && int64(id) == data.Id.ValueInt64() {
            found = true
            if name, ok := entry["name"].(string); ok {
                data.Name = types.StringValue(name)
            }
            if value, ok := entry["value"].(string); ok {
                data.Value = types.StringValue(value)
            }
            break
        }
    }

    if !found {
        resp.State.RemoveResource(ctx)
        return
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeyStoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var data KeyStoreResourceModel
    var state KeyStoreResourceModel

    // Get the planned values
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Get the current state to retrieve the ID
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Use the ID from the current state
    data.Id = state.Id

    // Create API request body
    body := map[string]interface{}{
        "name":  data.Name.ValueString(),
        "value": data.Value.ValueString(),
    }

    jsonBody, err := json.Marshal(body)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update keystore entry, got error: %s", err))
        return
    }

    // Create HTTP request
    updateURL := fmt.Sprintf("%s/core/keystore/%d/", r.client.BaseURL, data.Id.ValueInt64())
    httpReq, err := http.NewRequest("PUT", updateURL, bytes.NewBuffer(jsonBody))
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update keystore entry, got error: %s", err))
        return
    }

    // Make request
    httpResp, err := r.client.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update keystore entry, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode != http.StatusOK {
        // Read the response body for more details
        bodyBytes, _ := io.ReadAll(httpResp.Body)
        resp.Diagnostics.AddError("Client Error", 
            fmt.Sprintf("Unable to update keystore entry ID %d, status code: %d, URL: %s, response: %s", 
                data.Id.ValueInt64(), httpResp.StatusCode, updateURL, string(bodyBytes)))
        return
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeyStoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var data KeyStoreResourceModel

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Create HTTP request
    httpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/core/keystore/%d/", r.client.BaseURL, data.Id.ValueInt64()), nil)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete keystore entry, got error: %s", err))
        return
    }

    // Make request
    httpResp, err := r.client.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete keystore entry, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete keystore entry, status code: %d", httpResp.StatusCode))
        return
    }
}

func (r *KeyStoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    // Convert string ID to int64
    id, err := strconv.ParseInt(req.ID, 10, 64)
    if err != nil {
        resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
        return
    }
    
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
