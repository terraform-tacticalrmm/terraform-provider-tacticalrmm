package provider

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/hashicorp/terraform-plugin-framework/attr"
    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ScriptDataSource{}

func NewScriptDataSource() datasource.DataSource {
    return &ScriptDataSource{}
}

// ScriptDataSource defines the data source implementation.
type ScriptDataSource struct {
    client *ClientConfig
}

// ScriptDataSourceModel describes the data source data model.
type ScriptDataSourceModel struct {
    Id                   types.Int64  `tfsdk:"id"`
    Name                 types.String `tfsdk:"name"`
    Description          types.String `tfsdk:"description"`
    Shell                types.String `tfsdk:"shell"`
    ScriptType           types.String `tfsdk:"script_type"`
    Category             types.String `tfsdk:"category"`
    Filename             types.String `tfsdk:"filename"`
    ScriptBody           types.String `tfsdk:"script_body"`
    ScriptHash           types.String `tfsdk:"script_hash"`
    DefaultTimeout       types.Int64  `tfsdk:"default_timeout"`
    Favorite             types.Bool   `tfsdk:"favorite"`
    Hidden               types.Bool   `tfsdk:"hidden"`
    RunAsUser            types.Bool   `tfsdk:"run_as_user"`
    Args                 types.List   `tfsdk:"args"`
    EnvVars              types.List   `tfsdk:"env_vars"`
    SupportedPlatforms   types.List   `tfsdk:"supported_platforms"`
    Syntax               types.String `tfsdk:"syntax"`
}

func (d *ScriptDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_script"
}

func (d *ScriptDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Script data source for Tactical RMM. Use this to look up existing scripts by ID or name.",

        Attributes: map[string]schema.Attribute{
            "id": schema.Int64Attribute{
                MarkdownDescription: "Script identifier. Either `id` or `name` must be specified.",
                Optional:            true,
                Computed:            true,
            },
            "name": schema.StringAttribute{
                MarkdownDescription: "Script name. Either `id` or `name` must be specified.",
                Optional:            true,
                Computed:            true,
            },
            "description": schema.StringAttribute{
                MarkdownDescription: "Script description",
                Computed:            true,
            },
            "shell": schema.StringAttribute{
                MarkdownDescription: "Shell type: powershell, cmd, python, shell, nushell, deno",
                Computed:            true,
            },
            "script_type": schema.StringAttribute{
                MarkdownDescription: "Script type: userdefined, builtin",
                Computed:            true,
            },
            "category": schema.StringAttribute{
                MarkdownDescription: "Script category",
                Computed:            true,
            },
            "filename": schema.StringAttribute{
                MarkdownDescription: "Script filename (for builtin scripts)",
                Computed:            true,
            },
            "script_body": schema.StringAttribute{
                MarkdownDescription: "The script content",
                Computed:            true,
            },
            "script_hash": schema.StringAttribute{
                MarkdownDescription: "Script hash for integrity verification",
                Computed:            true,
            },
            "default_timeout": schema.Int64Attribute{
                MarkdownDescription: "Default timeout in seconds",
                Computed:            true,
            },
            "favorite": schema.BoolAttribute{
                MarkdownDescription: "Whether script is marked as favorite",
                Computed:            true,
            },
            "hidden": schema.BoolAttribute{
                MarkdownDescription: "Whether script is hidden",
                Computed:            true,
            },
            "run_as_user": schema.BoolAttribute{
                MarkdownDescription: "Run script as logged in user",
                Computed:            true,
            },
            "args": schema.ListAttribute{
                MarkdownDescription: "Script arguments",
                Computed:            true,
                ElementType:         types.StringType,
            },
            "env_vars": schema.ListAttribute{
                MarkdownDescription: "Environment variables",
                Computed:            true,
                ElementType:         types.StringType,
            },
            "supported_platforms": schema.ListAttribute{
                MarkdownDescription: "Supported platforms",
                Computed:            true,
                ElementType:         types.StringType,
            },
            "syntax": schema.StringAttribute{
                MarkdownDescription: "Script syntax",
                Computed:            true,
            },
        },
    }
}

func (d *ScriptDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*ClientConfig)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Data Source Configure Type",
            fmt.Sprintf("Expected *ClientConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
        )
        return
    }

    d.client = client
}

func (d *ScriptDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data ScriptDataSourceModel

    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Validate that either ID or name is provided
    if data.Id.IsNull() && data.Name.IsNull() {
        resp.Diagnostics.AddError(
            "Missing Script Identifier",
            "Either 'id' or 'name' must be specified to look up a script.",
        )
        return
    }

    var script map[string]interface{}

    if !data.Id.IsNull() {
        // Look up by ID
        httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/scripts/%d/", d.client.BaseURL, data.Id.ValueInt64()), nil)
        if err != nil {
            resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read script, got error: %s", err))
            return
        }

        httpResp, err := d.client.Do(httpReq)
        if err != nil {
            resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read script, got error: %s", err))
            return
        }
        defer httpResp.Body.Close()

        if httpResp.StatusCode == http.StatusNotFound {
            resp.Diagnostics.AddError("Script Not Found", fmt.Sprintf("Script with ID %d not found", data.Id.ValueInt64()))
            return
        }

        if httpResp.StatusCode != http.StatusOK {
            resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read script, status code: %d", httpResp.StatusCode))
            return
        }

        if err := json.NewDecoder(httpResp.Body).Decode(&script); err != nil {
            resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
            return
        }
    } else {
        // Look up by name - need to list all scripts and find the matching one
        httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/scripts/", d.client.BaseURL), nil)
        if err != nil {
            resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list scripts, got error: %s", err))
            return
        }

        httpResp, err := d.client.Do(httpReq)
        if err != nil {
            resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list scripts, got error: %s", err))
            return
        }
        defer httpResp.Body.Close()

        if httpResp.StatusCode != http.StatusOK {
            resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list scripts, status code: %d", httpResp.StatusCode))
            return
        }

        var scripts []map[string]interface{}
        if err := json.NewDecoder(httpResp.Body).Decode(&scripts); err != nil {
            resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse scripts list, got error: %s", err))
            return
        }

        // Find the script by name
        for _, s := range scripts {
            if name, ok := s["name"].(string); ok && name == data.Name.ValueString() {
                script = s
                break
            }
        }

        if script == nil {
            resp.Diagnostics.AddError("Script Not Found", fmt.Sprintf("Script with name '%s' not found", data.Name.ValueString()))
            return
        }
    }

    // Update model with response data
    if id, ok := script["id"].(float64); ok {
        data.Id = types.Int64Value(int64(id))
    }
    if name, ok := script["name"].(string); ok {
        data.Name = types.StringValue(name)
    }
    if description, ok := script["description"].(string); ok {
        data.Description = types.StringValue(description)
    } else {
        data.Description = types.StringNull()
    }
    if shell, ok := script["shell"].(string); ok {
        data.Shell = types.StringValue(shell)
    }
    if scriptType, ok := script["script_type"].(string); ok {
        data.ScriptType = types.StringValue(scriptType)
    }
    if category, ok := script["category"].(string); ok && category != "" {
        data.Category = types.StringValue(category)
    } else {
        data.Category = types.StringNull()
    }
    if filename, ok := script["filename"].(string); ok && filename != "" {
        data.Filename = types.StringValue(filename)
    } else {
        data.Filename = types.StringNull()
    }
    if scriptBody, ok := script["script_body"].(string); ok {
        data.ScriptBody = types.StringValue(scriptBody)
    }
    if scriptHash, ok := script["script_hash"].(string); ok && scriptHash != "" {
        data.ScriptHash = types.StringValue(scriptHash)
    } else {
        data.ScriptHash = types.StringNull()
    }
    if timeout, ok := script["default_timeout"].(float64); ok {
        data.DefaultTimeout = types.Int64Value(int64(timeout))
    }
    if favorite, ok := script["favorite"].(bool); ok {
        data.Favorite = types.BoolValue(favorite)
    }
    if hidden, ok := script["hidden"].(bool); ok {
        data.Hidden = types.BoolValue(hidden)
    }
    if runAsUser, ok := script["run_as_user"].(bool); ok {
        data.RunAsUser = types.BoolValue(runAsUser)
    }
    if syntax, ok := script["syntax"].(string); ok && syntax != "" {
        data.Syntax = types.StringValue(syntax)
    } else {
        data.Syntax = types.StringNull()
    }

    // Handle arrays
    if args, ok := script["args"].([]interface{}); ok && len(args) > 0 {
        argsList := make([]attr.Value, len(args))
        for i, arg := range args {
            if str, ok := arg.(string); ok {
                argsList[i] = types.StringValue(str)
            }
        }
        data.Args = types.ListValueMust(types.StringType, argsList)
    } else {
        data.Args = types.ListNull(types.StringType)
    }

    if envVars, ok := script["env_vars"].([]interface{}); ok && len(envVars) > 0 {
        envList := make([]attr.Value, len(envVars))
        for i, env := range envVars {
            if str, ok := env.(string); ok {
                envList[i] = types.StringValue(str)
            }
        }
        data.EnvVars = types.ListValueMust(types.StringType, envList)
    } else {
        data.EnvVars = types.ListNull(types.StringType)
    }

    if platforms, ok := script["supported_platforms"].([]interface{}); ok && len(platforms) > 0 {
        platList := make([]attr.Value, len(platforms))
        for i, plat := range platforms {
            if str, ok := plat.(string); ok {
                platList[i] = types.StringValue(str)
            }
        }
        data.SupportedPlatforms = types.ListValueMust(types.StringType, platList)
    } else {
        data.SupportedPlatforms = types.ListNull(types.StringType)
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
