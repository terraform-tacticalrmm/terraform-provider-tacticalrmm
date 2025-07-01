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
var _ datasource.DataSource = &ScriptsDataSource{}

func NewScriptsDataSource() datasource.DataSource {
    return &ScriptsDataSource{}
}

// ScriptsDataSource defines the data source implementation.
type ScriptsDataSource struct {
    client *ClientConfig
}

// ScriptsDataSourceModel describes the data source data model.
type ScriptsDataSourceModel struct {
    Id         types.Int64  `tfsdk:"id"`
    Name       types.String `tfsdk:"name"`
    ScriptType types.String `tfsdk:"script_type"`
    Shell      types.String `tfsdk:"shell"`
    Category   types.String `tfsdk:"category"`
    Hidden     types.Bool   `tfsdk:"hidden"`
    Scripts    types.List   `tfsdk:"scripts"`
}

// ScriptModel represents a single script in the list
// Note: List endpoint uses ScriptTableSerializer which excludes script_body
type ScriptModel struct {
    Id                   types.Int64  `tfsdk:"id"`
    Name                 types.String `tfsdk:"name"`
    Description          types.String `tfsdk:"description"`
    Shell                types.String `tfsdk:"shell"`
    ScriptType           types.String `tfsdk:"script_type"`
    Category             types.String `tfsdk:"category"`
    Filename             types.String `tfsdk:"filename"`
    DefaultTimeout       types.Int64  `tfsdk:"default_timeout"`
    Favorite             types.Bool   `tfsdk:"favorite"`
    Hidden               types.Bool   `tfsdk:"hidden"`
    RunAsUser            types.Bool   `tfsdk:"run_as_user"`
    Args                 types.List   `tfsdk:"args"`
    EnvVars              types.List   `tfsdk:"env_vars"`
    SupportedPlatforms   types.List   `tfsdk:"supported_platforms"`
    Syntax               types.String `tfsdk:"syntax"`
}

func (d *ScriptsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_scripts"
}

func (d *ScriptsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Scripts data source for Tactical RMM. Use this to fetch all scripts or filter by ID or name. Note: The list endpoint does not return script_body field.",

        Attributes: map[string]schema.Attribute{
            "id": schema.Int64Attribute{
                MarkdownDescription: "Optional: Filter scripts by a specific ID.",
                Optional:            true,
            },
            "name": schema.StringAttribute{
                MarkdownDescription: "Optional: Filter scripts by name (exact match).",
                Optional:            true,
            },
            "script_type": schema.StringAttribute{
                MarkdownDescription: "Optional: Filter scripts by type (userdefined or builtin).",
                Optional:            true,
            },
            "shell": schema.StringAttribute{
                MarkdownDescription: "Optional: Filter scripts by shell type (powershell, cmd, python, shell, nushell, deno).",
                Optional:            true,
            },
            "category": schema.StringAttribute{
                MarkdownDescription: "Optional: Filter scripts by category.",
                Optional:            true,
            },
            "hidden": schema.BoolAttribute{
                MarkdownDescription: "Optional: Filter scripts by hidden status.",
                Optional:            true,
            },
            "scripts": schema.ListNestedAttribute{
                MarkdownDescription: "List of scripts matching the filter criteria, or all scripts if no filter is specified.",
                Computed:            true,
                NestedObject: schema.NestedAttributeObject{
                    Attributes: map[string]schema.Attribute{
                        "id": schema.Int64Attribute{
                            MarkdownDescription: "Script identifier",
                            Computed:            true,
                        },
                        "name": schema.StringAttribute{
                            MarkdownDescription: "Script name",
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
                },
            },
        },
    }
}

func (d *ScriptsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ScriptsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data ScriptsDataSourceModel

    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Fetch all scripts
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

    // Filter scripts based on criteria
    var filteredScripts []map[string]interface{}
    
    // Start with all scripts if no ID filter
    if !data.Id.IsNull() {
        // Filter by ID (exclusive filter)
        targetId := data.Id.ValueInt64()
        for _, script := range scripts {
            if id, ok := script["id"].(float64); ok && int64(id) == targetId {
                filteredScripts = append(filteredScripts, script)
                break
            }
        }
    } else {
        // Apply other filters
        for _, script := range scripts {
            include := true
            
            // Filter by name
            if !data.Name.IsNull() {
                if name, ok := script["name"].(string); !ok || name != data.Name.ValueString() {
                    include = false
                }
            }
            
            // Filter by script type
            if include && !data.ScriptType.IsNull() {
                if scriptType, ok := script["script_type"].(string); !ok || scriptType != data.ScriptType.ValueString() {
                    include = false
                }
            }
            
            // Filter by shell
            if include && !data.Shell.IsNull() {
                if shell, ok := script["shell"].(string); !ok || shell != data.Shell.ValueString() {
                    include = false
                }
            }
            
            // Filter by category
            if include && !data.Category.IsNull() {
                if category, ok := script["category"].(string); !ok || category != data.Category.ValueString() {
                    include = false
                }
            }
            
            // Filter by hidden status
            if include && !data.Hidden.IsNull() {
                if hidden, ok := script["hidden"].(bool); !ok || hidden != data.Hidden.ValueBool() {
                    include = false
                }
            }
            
            if include {
                filteredScripts = append(filteredScripts, script)
            }
        }
    }

    // Convert to ScriptModel list
    scriptsList := make([]ScriptModel, len(filteredScripts))
    for i, script := range filteredScripts {
        model := ScriptModel{}
        
        if id, ok := script["id"].(float64); ok {
            model.Id = types.Int64Value(int64(id))
        }
        if name, ok := script["name"].(string); ok {
            model.Name = types.StringValue(name)
        }
        if description, ok := script["description"].(string); ok {
            model.Description = types.StringValue(description)
        } else {
            model.Description = types.StringNull()
        }
        if shell, ok := script["shell"].(string); ok {
            model.Shell = types.StringValue(shell)
        }
        if scriptType, ok := script["script_type"].(string); ok {
            model.ScriptType = types.StringValue(scriptType)
        }
        if category, ok := script["category"].(string); ok && category != "" {
            model.Category = types.StringValue(category)
        } else {
            model.Category = types.StringNull()
        }
        if filename, ok := script["filename"].(string); ok && filename != "" {
            model.Filename = types.StringValue(filename)
        } else {
            model.Filename = types.StringNull()
        }
        if timeout, ok := script["default_timeout"].(float64); ok {
            model.DefaultTimeout = types.Int64Value(int64(timeout))
        }
        if favorite, ok := script["favorite"].(bool); ok {
            model.Favorite = types.BoolValue(favorite)
        }
        if hidden, ok := script["hidden"].(bool); ok {
            model.Hidden = types.BoolValue(hidden)
        }
        if runAsUser, ok := script["run_as_user"].(bool); ok {
            model.RunAsUser = types.BoolValue(runAsUser)
        }
        if syntax, ok := script["syntax"].(string); ok && syntax != "" {
            model.Syntax = types.StringValue(syntax)
        } else {
            model.Syntax = types.StringNull()
        }

        // Handle arrays
        if args, ok := script["args"].([]interface{}); ok && len(args) > 0 {
            argsList := make([]attr.Value, len(args))
            for j, arg := range args {
                if str, ok := arg.(string); ok {
                    argsList[j] = types.StringValue(str)
                }
            }
            model.Args = types.ListValueMust(types.StringType, argsList)
        } else {
            model.Args = types.ListNull(types.StringType)
        }

        if envVars, ok := script["env_vars"].([]interface{}); ok && len(envVars) > 0 {
            envList := make([]attr.Value, len(envVars))
            for j, env := range envVars {
                if str, ok := env.(string); ok {
                    envList[j] = types.StringValue(str)
                }
            }
            model.EnvVars = types.ListValueMust(types.StringType, envList)
        } else {
            model.EnvVars = types.ListNull(types.StringType)
        }

        if platforms, ok := script["supported_platforms"].([]interface{}); ok && len(platforms) > 0 {
            platList := make([]attr.Value, len(platforms))
            for j, plat := range platforms {
                if str, ok := plat.(string); ok {
                    platList[j] = types.StringValue(str)
                }
            }
            model.SupportedPlatforms = types.ListValueMust(types.StringType, platList)
        } else {
            model.SupportedPlatforms = types.ListNull(types.StringType)
        }
        
        scriptsList[i] = model
    }

    // Convert to list value
    scriptObjectType := types.ObjectType{
        AttrTypes: map[string]attr.Type{
            "id":                   types.Int64Type,
            "name":                 types.StringType,
            "description":          types.StringType,
            "shell":                types.StringType,
            "script_type":          types.StringType,
            "category":             types.StringType,
            "filename":             types.StringType,
            "default_timeout":      types.Int64Type,
            "favorite":             types.BoolType,
            "hidden":               types.BoolType,
            "run_as_user":          types.BoolType,
            "args":                 types.ListType{ElemType: types.StringType},
            "env_vars":             types.ListType{ElemType: types.StringType},
            "supported_platforms":  types.ListType{ElemType: types.StringType},
            "syntax":               types.StringType,
        },
    }

    scriptsListValue := make([]attr.Value, len(scriptsList))
    for i, script := range scriptsList {
        objValue, diags := types.ObjectValueFrom(ctx, scriptObjectType.AttrTypes, script)
        resp.Diagnostics.Append(diags...)
        scriptsListValue[i] = objValue
    }

    listValue, diags := types.ListValue(scriptObjectType, scriptsListValue)
    resp.Diagnostics.Append(diags...)
    data.Scripts = listValue

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
