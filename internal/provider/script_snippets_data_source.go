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
var _ datasource.DataSource = &ScriptSnippetsDataSource{}

func NewScriptSnippetsDataSource() datasource.DataSource {
    return &ScriptSnippetsDataSource{}
}

// ScriptSnippetsDataSource defines the data source implementation.
type ScriptSnippetsDataSource struct {
    client *ClientConfig
}

// ScriptSnippetsDataSourceModel describes the data source data model.
type ScriptSnippetsDataSourceModel struct {
    Id       types.Int64  `tfsdk:"id"`
    Name     types.String `tfsdk:"name"`
    Snippets types.List   `tfsdk:"snippets"`
}

// ScriptSnippetModel represents a single snippet in the list
type ScriptSnippetModel struct {
    Id    types.Int64  `tfsdk:"id"`
    Name  types.String `tfsdk:"name"`
    Desc  types.String `tfsdk:"desc"`
    Code  types.String `tfsdk:"code"`
    Shell types.String `tfsdk:"shell"`
}

func (d *ScriptSnippetsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_script_snippets"
}

func (d *ScriptSnippetsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Script Snippets data source for Tactical RMM. Use this to fetch all script snippets or filter by ID or name.",

        Attributes: map[string]schema.Attribute{
            "id": schema.Int64Attribute{
                MarkdownDescription: "Optional: Filter snippets by a specific ID.",
                Optional:            true,
            },
            "name": schema.StringAttribute{
                MarkdownDescription: "Optional: Filter snippets by name (exact match).",
                Optional:            true,
            },
            "snippets": schema.ListNestedAttribute{
                MarkdownDescription: "List of script snippets matching the filter criteria, or all snippets if no filter is specified.",
                Computed:            true,
                NestedObject: schema.NestedAttributeObject{
                    Attributes: map[string]schema.Attribute{
                        "id": schema.Int64Attribute{
                            MarkdownDescription: "Script snippet identifier",
                            Computed:            true,
                        },
                        "name": schema.StringAttribute{
                            MarkdownDescription: "Snippet name",
                            Computed:            true,
                        },
                        "desc": schema.StringAttribute{
                            MarkdownDescription: "Snippet description",
                            Computed:            true,
                        },
                        "code": schema.StringAttribute{
                            MarkdownDescription: "Snippet code content",
                            Computed:            true,
                        },
                        "shell": schema.StringAttribute{
                            MarkdownDescription: "Shell type: powershell, cmd, python, shell",
                            Computed:            true,
                        },
                    },
                },
            },
        },
    }
}

func (d *ScriptSnippetsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ScriptSnippetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data ScriptSnippetsDataSourceModel

    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Fetch all script snippets
    httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/scripts/snippets/", d.client.BaseURL), nil)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list script snippets, got error: %s", err))
        return
    }

    httpResp, err := d.client.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list script snippets, got error: %s", err))
        return
    }
    defer httpResp.Body.Close()

    if httpResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list script snippets, status code: %d", httpResp.StatusCode))
        return
    }

    var snippets []map[string]interface{}
    if err := json.NewDecoder(httpResp.Body).Decode(&snippets); err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse script snippets list, got error: %s", err))
        return
    }

    // Filter snippets based on criteria
    var filteredSnippets []map[string]interface{}
    
    if !data.Id.IsNull() {
        // Filter by ID
        targetId := data.Id.ValueInt64()
        for _, snippet := range snippets {
            if id, ok := snippet["id"].(float64); ok && int64(id) == targetId {
                filteredSnippets = append(filteredSnippets, snippet)
                break
            }
        }
    } else if !data.Name.IsNull() {
        // Filter by name
        targetName := data.Name.ValueString()
        for _, snippet := range snippets {
            if name, ok := snippet["name"].(string); ok && name == targetName {
                filteredSnippets = append(filteredSnippets, snippet)
            }
        }
    } else {
        // No filter, return all snippets
        filteredSnippets = snippets
    }

    // Convert to ScriptSnippetModel list
    snippetsList := make([]ScriptSnippetModel, len(filteredSnippets))
    for i, snippet := range filteredSnippets {
        model := ScriptSnippetModel{}
        
        if id, ok := snippet["id"].(float64); ok {
            model.Id = types.Int64Value(int64(id))
        }
        if name, ok := snippet["name"].(string); ok {
            model.Name = types.StringValue(name)
        }
        if desc, ok := snippet["desc"].(string); ok {
            model.Desc = types.StringValue(desc)
        } else {
            model.Desc = types.StringNull()
        }
        if code, ok := snippet["code"].(string); ok {
            model.Code = types.StringValue(code)
        }
        if shell, ok := snippet["shell"].(string); ok {
            model.Shell = types.StringValue(shell)
        }
        
        snippetsList[i] = model
    }

    // Convert to list value
    snippetObjectType := types.ObjectType{
        AttrTypes: map[string]attr.Type{
            "id":    types.Int64Type,
            "name":  types.StringType,
            "desc":  types.StringType,
            "code":  types.StringType,
            "shell": types.StringType,
        },
    }

    snippetsListValue := make([]attr.Value, len(snippetsList))
    for i, snippet := range snippetsList {
        objValue, diags := types.ObjectValueFrom(ctx, snippetObjectType.AttrTypes, snippet)
        resp.Diagnostics.Append(diags...)
        snippetsListValue[i] = objValue
    }

    listValue, diags := types.ListValue(snippetObjectType, snippetsListValue)
    resp.Diagnostics.Append(diags...)
    data.Snippets = listValue

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
