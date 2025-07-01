package provider

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"

    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ScriptSnippetResource{}
var _ resource.ResourceWithImportState = &ScriptSnippetResource{}

func NewScriptSnippetResource() resource.Resource {
    return &ScriptSnippetResource{}
}

// ScriptSnippetResource defines the resource implementation.
type ScriptSnippetResource struct {
    client *ClientConfig
}

// ScriptSnippetResourceModel describes the resource data model based on ScriptSnippet Django model
type ScriptSnippetResourceModel struct {
    Id    types.Int64  `tfsdk:"id"`
    Name  types.String `tfsdk:"name"`
    Desc  types.String `tfsdk:"desc"`
    Code  types.String `tfsdk:"code"`
    Shell types.String `tfsdk:"shell"`
}

func (r *ScriptSnippetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_script_snippet"
}

func (r *ScriptSnippetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Script Snippet resource for Tactical RMM",

        Attributes: map[string]schema.Attribute{
            "id": schema.Int64Attribute{
                MarkdownDescription: "Script snippet identifier",
                Computed:            true,
            },
            "name": schema.StringAttribute{
                MarkdownDescription: "Snippet name (max 40 characters, unique)",
                Required:            true,
            },
            "desc": schema.StringAttribute{
                MarkdownDescription: "Snippet description (max 50 characters)",
                Optional:            true,
            },
            "code": schema.StringAttribute{
                MarkdownDescription: "Snippet code content",
                Required:            true,
            },
            "shell": schema.StringAttribute{
                MarkdownDescription: "Shell type: powershell, cmd, python, shell",
                Optional:            true,
                Computed:            true,
            },
        },
    }
}

func (r *ScriptSnippetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ScriptSnippetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data ScriptSnippetResourceModel

    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Create API request body
    body := map[string]interface{}{
        "name": data.Name.ValueString(),
        "code": data.Code.ValueString(),
    }

    // Optional fields
    if !data.Desc.IsNull() {
        body["desc"] = data.Desc.ValueString()
    }
    if !data.Shell.IsNull() {
        body["shell"] = data.Shell.ValueString()
    } else {
        body["shell"] = "powershell" // Default value
    }

    jsonBody, err := json.Marshal(body)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create script snippet, got error: %s", err))
        return
    }

    // Create HTTP request
    httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/scripts/snippets/", r.client.BaseURL), bytes.NewBuffer(jsonBody))
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create script snippet, got error: %s", err))
        return
    }

    // Make request
    httpResp, err := r.client.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create script snippet, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create script snippet, status code: %d", httpResp.StatusCode))
        return
    }

    // Response is just a message, so we need to get the created snippet
    // List all snippets to find our newly created one
    listReq, err := http.NewRequest("GET", fmt.Sprintf("%s/scripts/snippets/", r.client.BaseURL), nil)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list script snippets, got error: %s", err))
        return
    }

    listResp, err := r.client.Do(listReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list script snippets, got error: %s", err))
        return
    }
    defer listResp.Body.Close()

    var snippets []map[string]interface{}
    if err := json.NewDecoder(listResp.Body).Decode(&snippets); err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse script snippets list, got error: %s", err))
        return
    }

    // Find the snippet we just created by name
    var createdSnippet map[string]interface{}
    for _, snippet := range snippets {
        if name, ok := snippet["name"].(string); ok && name == data.Name.ValueString() {
            createdSnippet = snippet
            break
        }
    }

    if createdSnippet == nil {
        resp.Diagnostics.AddError("Client Error", "Unable to find created script snippet")
        return
    }

    // Update model with response data
    if id, ok := createdSnippet["id"].(float64); ok {
        data.Id = types.Int64Value(int64(id))
    }

    // Set defaults if not provided
    if data.Shell.IsNull() {
        data.Shell = types.StringValue("powershell")
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScriptSnippetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var data ScriptSnippetResourceModel

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Create HTTP request
    httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/scripts/snippets/%d/", r.client.BaseURL, data.Id.ValueInt64()), nil)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read script snippet, got error: %s", err))
        return
    }

    // Make request
    httpResp, err := r.client.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read script snippet, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode == http.StatusNotFound {
        resp.State.RemoveResource(ctx)
        return
    }

    if httpResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read script snippet, status code: %d", httpResp.StatusCode))
        return
    }

    // Parse response
    var result map[string]interface{}
    if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
        return
    }

    // Update model with response data
    if name, ok := result["name"].(string); ok {
        data.Name = types.StringValue(name)
    }
    if desc, ok := result["desc"].(string); ok {
        data.Desc = types.StringValue(desc)
    }
    if code, ok := result["code"].(string); ok {
        data.Code = types.StringValue(code)
    }
    if shell, ok := result["shell"].(string); ok {
        data.Shell = types.StringValue(shell)
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScriptSnippetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var data ScriptSnippetResourceModel
    var state ScriptSnippetResourceModel

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
        "name": data.Name.ValueString(),
        "code": data.Code.ValueString(),
    }

    // Optional fields
    if !data.Desc.IsNull() {
        body["desc"] = data.Desc.ValueString()
    }
    if !data.Shell.IsNull() {
        body["shell"] = data.Shell.ValueString()
    }

    jsonBody, err := json.Marshal(body)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update script snippet, got error: %s", err))
        return
    }

    // Create HTTP request
    httpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/scripts/snippets/%d/", r.client.BaseURL, data.Id.ValueInt64()), bytes.NewBuffer(jsonBody))
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update script snippet, got error: %s", err))
        return
    }

    // Make request
    httpResp, err := r.client.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update script snippet, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update script snippet, status code: %d", httpResp.StatusCode))
        return
    }

    // Get the updated script snippet to ensure all computed fields are populated
    getReq, err := http.NewRequest("GET", fmt.Sprintf("%s/scripts/snippets/%d/", r.client.BaseURL, data.Id.ValueInt64()), nil)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated script snippet, got error: %s", err))
        return
    }

    getResp, err := r.client.Do(getReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated script snippet, got error: %s", err))
        return
    }
    defer getResp.Body.Close()

    if getResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated script snippet, status code: %d", getResp.StatusCode))
        return
    }

    // Parse response
    var result map[string]interface{}
    if err := json.NewDecoder(getResp.Body).Decode(&result); err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
        return
    }

    // Update computed fields from the response
    if shell, ok := result["shell"].(string); ok {
        data.Shell = types.StringValue(shell)
    } else if data.Shell.IsNull() || data.Shell.IsUnknown() {
        data.Shell = types.StringValue("powershell")
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScriptSnippetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var data ScriptSnippetResourceModel

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Create HTTP request
    httpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/scripts/snippets/%d/", r.client.BaseURL, data.Id.ValueInt64()), nil)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete script snippet, got error: %s", err))
        return
    }

    // Make request
    httpResp, err := r.client.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete script snippet, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete script snippet, status code: %d", httpResp.StatusCode))
        return
    }
}

func (r *ScriptSnippetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    // Convert string ID to int64
    id, err := strconv.ParseInt(req.ID, 10, 64)
    if err != nil {
        resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
        return
    }
    
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
