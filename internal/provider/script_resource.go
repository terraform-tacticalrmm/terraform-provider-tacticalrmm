package provider

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"

    "github.com/hashicorp/terraform-plugin-framework/attr"
    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ScriptResource{}
var _ resource.ResourceWithImportState = &ScriptResource{}

func NewScriptResource() resource.Resource {
    return &ScriptResource{}
}

// ScriptResource defines the resource implementation.
type ScriptResource struct {
    client *ClientConfig
}

// ScriptResourceModel describes the resource data model based on the Django model
type ScriptResourceModel struct {
    Id                   types.Int64  `tfsdk:"id"`
    Name                 types.String `tfsdk:"name"`
    Description          types.String `tfsdk:"description"`
    Shell                types.String `tfsdk:"shell"`
    ScriptType           types.String `tfsdk:"script_type"`
    Category             types.String `tfsdk:"category"`
    ScriptBody           types.String `tfsdk:"script_body"`
    DefaultTimeout       types.Int64  `tfsdk:"default_timeout"`
    Favorite             types.Bool   `tfsdk:"favorite"`
    Hidden               types.Bool   `tfsdk:"hidden"`
    RunAsUser            types.Bool   `tfsdk:"run_as_user"`
    Args                 types.List   `tfsdk:"args"`
    EnvVars              types.List   `tfsdk:"env_vars"`
    SupportedPlatforms   types.List   `tfsdk:"supported_platforms"`
    Syntax               types.String `tfsdk:"syntax"`
}

func (r *ScriptResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_script"
}

func (r *ScriptResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Script resource for Tactical RMM",

        Attributes: map[string]schema.Attribute{
            "id": schema.Int64Attribute{
                MarkdownDescription: "Script identifier",
                Computed:            true,
            },
            "name": schema.StringAttribute{
                MarkdownDescription: "Script name",
                Required:            true,
            },
            "description": schema.StringAttribute{
                MarkdownDescription: "Script description",
                Optional:            true,
            },
            "shell": schema.StringAttribute{
                MarkdownDescription: "Shell type: powershell, cmd, python, shell, nushell, deno",
                Required:            true,
            },
            "script_type": schema.StringAttribute{
                MarkdownDescription: "Script type: userdefined, builtin",
                Optional:            true,
                Computed:            true,
            },
            "category": schema.StringAttribute{
                MarkdownDescription: "Script category",
                Optional:            true,
            },
            "script_body": schema.StringAttribute{
                MarkdownDescription: "The script content",
                Required:            true,
            },
            "default_timeout": schema.Int64Attribute{
                MarkdownDescription: "Default timeout in seconds",
                Optional:            true,
                Computed:            true,
            },
            "favorite": schema.BoolAttribute{
                MarkdownDescription: "Whether script is marked as favorite",
                Optional:            true,
                Computed:            true,
            },
            "hidden": schema.BoolAttribute{
                MarkdownDescription: "Whether script is hidden",
                Optional:            true,
                Computed:            true,
            },
            "run_as_user": schema.BoolAttribute{
                MarkdownDescription: "Run script as logged in user",
                Optional:            true,
                Computed:            true,
            },
            "args": schema.ListAttribute{
                MarkdownDescription: "Script arguments",
                Optional:            true,
                ElementType:         types.StringType,
            },
            "env_vars": schema.ListAttribute{
                MarkdownDescription: "Environment variables",
                Optional:            true,
                ElementType:         types.StringType,
            },
            "supported_platforms": schema.ListAttribute{
                MarkdownDescription: "Supported platforms",
                Optional:            true,
                ElementType:         types.StringType,
            },
            "syntax": schema.StringAttribute{
                MarkdownDescription: "Script syntax",
                Optional:            true,
            },
        },
    }
}

func (r *ScriptResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ScriptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data ScriptResourceModel

    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }
    
    // Store original state of arrays to preserve null vs empty
    argsWasNull := data.Args.IsNull()
    envVarsWasNull := data.EnvVars.IsNull()
    platformsWasNull := data.SupportedPlatforms.IsNull()

    // Create API request body
    body := map[string]interface{}{
        "name":            data.Name.ValueString(),
        "shell":           data.Shell.ValueString(),
        "script_body":     data.ScriptBody.ValueString(),
        "script_type":     "userdefined",
    }

    // Optional fields
    if !data.Description.IsNull() {
        body["description"] = data.Description.ValueString()
    }
    if !data.Category.IsNull() {
        body["category"] = data.Category.ValueString()
    }
    if !data.DefaultTimeout.IsNull() {
        body["default_timeout"] = data.DefaultTimeout.ValueInt64()
    } else {
        body["default_timeout"] = 90
    }
    if !data.Favorite.IsNull() {
        body["favorite"] = data.Favorite.ValueBool()
    }
    if !data.Hidden.IsNull() {
        body["hidden"] = data.Hidden.ValueBool()
    }
    if !data.RunAsUser.IsNull() {
        body["run_as_user"] = data.RunAsUser.ValueBool()
    }
    if !data.Syntax.IsNull() {
        body["syntax"] = data.Syntax.ValueString()
    }

    // Handle arrays
    if !data.Args.IsNull() {
        var args []string
        resp.Diagnostics.Append(data.Args.ElementsAs(ctx, &args, false)...)
        body["args"] = args
    }
    if !data.EnvVars.IsNull() {
        var envVars []string
        resp.Diagnostics.Append(data.EnvVars.ElementsAs(ctx, &envVars, false)...)
        body["env_vars"] = envVars
    }
    if !data.SupportedPlatforms.IsNull() {
        var platforms []string
        resp.Diagnostics.Append(data.SupportedPlatforms.ElementsAs(ctx, &platforms, false)...)
        body["supported_platforms"] = platforms
    }

    jsonBody, err := json.Marshal(body)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create script, got error: %s", err))
        return
    }

    // Create HTTP request
    httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/scripts/", r.client.BaseURL), bytes.NewBuffer(jsonBody))
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create script, got error: %s", err))
        return
    }

    // Make request
    httpResp, err := r.client.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create script, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create script, status code: %d", httpResp.StatusCode))
        return
    }

    // Response is just a message, so we need to get the created script
    // First, list all scripts to find our newly created one
    listReq, err := http.NewRequest("GET", fmt.Sprintf("%s/scripts/", r.client.BaseURL), nil)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list scripts, got error: %s", err))
        return
    }

    listResp, err := r.client.Do(listReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list scripts, got error: %s", err))
        return
    }
    defer listResp.Body.Close()

    var scripts []map[string]interface{}
    if err := json.NewDecoder(listResp.Body).Decode(&scripts); err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse scripts list, got error: %s", err))
        return
    }

    // Find the script we just created by name
    var createdScript map[string]interface{}
    for _, script := range scripts {
        if name, ok := script["name"].(string); ok && name == data.Name.ValueString() {
            createdScript = script
            break
        }
    }

    if createdScript == nil {
        resp.Diagnostics.AddError("Client Error", "Unable to find created script")
        return
    }

    // Update model with response data
    if id, ok := createdScript["id"].(float64); ok {
        data.Id = types.Int64Value(int64(id))
    }
    
    // Update fields from the response, with defaults for computed fields
    if scriptType, ok := createdScript["script_type"].(string); ok {
        data.ScriptType = types.StringValue(scriptType)
    } else {
        data.ScriptType = types.StringValue("userdefined")
    }
    
    if timeout, ok := createdScript["default_timeout"].(float64); ok {
        data.DefaultTimeout = types.Int64Value(int64(timeout))
    } else if data.DefaultTimeout.IsNull() || data.DefaultTimeout.IsUnknown() {
        data.DefaultTimeout = types.Int64Value(90)
    }
    
    if favorite, ok := createdScript["favorite"].(bool); ok {
        data.Favorite = types.BoolValue(favorite)
    } else {
        data.Favorite = types.BoolValue(false)
    }
    
    if hidden, ok := createdScript["hidden"].(bool); ok {
        data.Hidden = types.BoolValue(hidden)
    } else {
        data.Hidden = types.BoolValue(false)
    }
    
    if runAsUser, ok := createdScript["run_as_user"].(bool); ok {
        data.RunAsUser = types.BoolValue(runAsUser)
    } else {
        data.RunAsUser = types.BoolValue(false)
    }
    
    // Handle arrays from response - preserve null state from plan
    if !argsWasNull {
        if args, ok := createdScript["args"].([]interface{}); ok {
            argsList := make([]attr.Value, len(args))
            for i, arg := range args {
                if str, ok := arg.(string); ok {
                    argsList[i] = types.StringValue(str)
                }
            }
            data.Args = types.ListValueMust(types.StringType, argsList)
        } else {
            // Plan had empty list, preserve it
            data.Args = types.ListValueMust(types.StringType, []attr.Value{})
        }
    }
    // If args was null in plan, keep it null

    if !envVarsWasNull {
        if envVars, ok := createdScript["env_vars"].([]interface{}); ok {
            envList := make([]attr.Value, len(envVars))
            for i, env := range envVars {
                if str, ok := env.(string); ok {
                    envList[i] = types.StringValue(str)
                }
            }
            data.EnvVars = types.ListValueMust(types.StringType, envList)
        } else {
            // Plan had empty list, preserve it
            data.EnvVars = types.ListValueMust(types.StringType, []attr.Value{})
        }
    }
    // If env_vars was null in plan, keep it null

    if !platformsWasNull {
        if platforms, ok := createdScript["supported_platforms"].([]interface{}); ok {
            platList := make([]attr.Value, len(platforms))
            for i, plat := range platforms {
                if str, ok := plat.(string); ok {
                    platList[i] = types.StringValue(str)
                }
            }
            data.SupportedPlatforms = types.ListValueMust(types.StringType, platList)
        } else {
            // Plan had empty list, preserve it
            data.SupportedPlatforms = types.ListValueMust(types.StringType, []attr.Value{})
        }
    }
    // If supported_platforms was null in plan, keep it null

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScriptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var data ScriptResourceModel

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Create HTTP request
    httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/scripts/%d/", r.client.BaseURL, data.Id.ValueInt64()), nil)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read script, got error: %s", err))
        return
    }

    // Make request
    httpResp, err := r.client.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read script, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode == http.StatusNotFound {
        resp.State.RemoveResource(ctx)
        return
    }

    if httpResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read script, status code: %d", httpResp.StatusCode))
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
    if description, ok := result["description"].(string); ok {
        data.Description = types.StringValue(description)
    }
    if shell, ok := result["shell"].(string); ok {
        data.Shell = types.StringValue(shell)
    }
    if scriptType, ok := result["script_type"].(string); ok {
        data.ScriptType = types.StringValue(scriptType)
    }
    if category, ok := result["category"].(string); ok && category != "" {
        data.Category = types.StringValue(category)
    }
    if scriptBody, ok := result["script_body"].(string); ok {
        data.ScriptBody = types.StringValue(scriptBody)
    }
    if timeout, ok := result["default_timeout"].(float64); ok {
        data.DefaultTimeout = types.Int64Value(int64(timeout))
    }
    if favorite, ok := result["favorite"].(bool); ok {
        data.Favorite = types.BoolValue(favorite)
    }
    if hidden, ok := result["hidden"].(bool); ok {
        data.Hidden = types.BoolValue(hidden)
    }
    if runAsUser, ok := result["run_as_user"].(bool); ok {
        data.RunAsUser = types.BoolValue(runAsUser)
    }
    if syntax, ok := result["syntax"].(string); ok && syntax != "" {
        data.Syntax = types.StringValue(syntax)
    }

    // Handle arrays - preserve null if empty
    if args, ok := result["args"].([]interface{}); ok && len(args) > 0 {
        argsList := make([]attr.Value, len(args))
        for i, arg := range args {
            if str, ok := arg.(string); ok {
                argsList[i] = types.StringValue(str)
            }
        }
        data.Args = types.ListValueMust(types.StringType, argsList)
    }
    // Keep null if the API returns empty or no args

    if envVars, ok := result["env_vars"].([]interface{}); ok && len(envVars) > 0 {
        envList := make([]attr.Value, len(envVars))
        for i, env := range envVars {
            if str, ok := env.(string); ok {
                envList[i] = types.StringValue(str)
            }
        }
        data.EnvVars = types.ListValueMust(types.StringType, envList)
    }
    // Keep null if the API returns empty or no env_vars

    if platforms, ok := result["supported_platforms"].([]interface{}); ok && len(platforms) > 0 {
        platList := make([]attr.Value, len(platforms))
        for i, plat := range platforms {
            if str, ok := plat.(string); ok {
                platList[i] = types.StringValue(str)
            }
        }
        data.SupportedPlatforms = types.ListValueMust(types.StringType, platList)
    }
    // Keep null if the API returns empty or no supported_platforms

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScriptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var data ScriptResourceModel
    var state ScriptResourceModel

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
        "name":        data.Name.ValueString(),
        "shell":       data.Shell.ValueString(),
        "script_body": data.ScriptBody.ValueString(),
    }

    // Optional fields
    if !data.Description.IsNull() {
        body["description"] = data.Description.ValueString()
    }
    if !data.Category.IsNull() {
        body["category"] = data.Category.ValueString()
    }
    if !data.DefaultTimeout.IsNull() {
        body["default_timeout"] = data.DefaultTimeout.ValueInt64()
    }
    if !data.Favorite.IsNull() {
        body["favorite"] = data.Favorite.ValueBool()
    }
    if !data.Hidden.IsNull() {
        body["hidden"] = data.Hidden.ValueBool()
    }
    if !data.RunAsUser.IsNull() {
        body["run_as_user"] = data.RunAsUser.ValueBool()
    }
    if !data.Syntax.IsNull() {
        body["syntax"] = data.Syntax.ValueString()
    }

    // Handle arrays
    if !data.Args.IsNull() {
        var args []string
        resp.Diagnostics.Append(data.Args.ElementsAs(ctx, &args, false)...)
        body["args"] = args
    }
    if !data.EnvVars.IsNull() {
        var envVars []string
        resp.Diagnostics.Append(data.EnvVars.ElementsAs(ctx, &envVars, false)...)
        body["env_vars"] = envVars
    }
    if !data.SupportedPlatforms.IsNull() {
        var platforms []string
        resp.Diagnostics.Append(data.SupportedPlatforms.ElementsAs(ctx, &platforms, false)...)
        body["supported_platforms"] = platforms
    }

    jsonBody, err := json.Marshal(body)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update script, got error: %s", err))
        return
    }

    // Create HTTP request
    httpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/scripts/%d/", r.client.BaseURL, data.Id.ValueInt64()), bytes.NewBuffer(jsonBody))
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update script, got error: %s", err))
        return
    }

    // Make request
    httpResp, err := r.client.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update script, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update script, status code: %d", httpResp.StatusCode))
        return
    }

    // Get the updated script to ensure all computed fields are populated
    getReq, err := http.NewRequest("GET", fmt.Sprintf("%s/scripts/%d/", r.client.BaseURL, data.Id.ValueInt64()), nil)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated script, got error: %s", err))
        return
    }

    getResp, err := r.client.Do(getReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated script, got error: %s", err))
        return
    }
    defer getResp.Body.Close()

    if getResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated script, status code: %d", getResp.StatusCode))
        return
    }

    // Parse response
    var result map[string]interface{}
    if err := json.NewDecoder(getResp.Body).Decode(&result); err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
        return
    }

    // Update computed fields from the response
    if scriptType, ok := result["script_type"].(string); ok {
        data.ScriptType = types.StringValue(scriptType)
    } else {
        data.ScriptType = types.StringValue("userdefined")
    }
    
    if timeout, ok := result["default_timeout"].(float64); ok {
        data.DefaultTimeout = types.Int64Value(int64(timeout))
    } else if data.DefaultTimeout.IsNull() || data.DefaultTimeout.IsUnknown() {
        data.DefaultTimeout = types.Int64Value(90)
    }
    
    if favorite, ok := result["favorite"].(bool); ok {
        data.Favorite = types.BoolValue(favorite)
    } else if data.Favorite.IsNull() || data.Favorite.IsUnknown() {
        data.Favorite = types.BoolValue(false)
    }
    
    if hidden, ok := result["hidden"].(bool); ok {
        data.Hidden = types.BoolValue(hidden)
    } else if data.Hidden.IsNull() || data.Hidden.IsUnknown() {
        data.Hidden = types.BoolValue(false)
    }
    
    if runAsUser, ok := result["run_as_user"].(bool); ok {
        data.RunAsUser = types.BoolValue(runAsUser)
    } else if data.RunAsUser.IsNull() || data.RunAsUser.IsUnknown() {
        data.RunAsUser = types.BoolValue(false)
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScriptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var data ScriptResourceModel

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Create HTTP request
    httpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/scripts/%d/", r.client.BaseURL, data.Id.ValueInt64()), nil)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete script, got error: %s", err))
        return
    }

    // Make request
    httpResp, err := r.client.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete script, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete script, status code: %d", httpResp.StatusCode))
        return
    }
}

func (r *ScriptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    // Convert string ID to int64
    id, err := strconv.ParseInt(req.ID, 10, 64)
    if err != nil {
        resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
        return
    }
    
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
