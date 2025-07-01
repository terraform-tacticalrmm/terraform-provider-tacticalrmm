package provider

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ScriptSnippetDataSource{}

func NewScriptSnippetDataSource() datasource.DataSource {
    return &ScriptSnippetDataSource{}
}

// ScriptSnippetDataSource defines the data source implementation.
type ScriptSnippetDataSource struct {
    client *ClientConfig
}

// ScriptSnippetDataSourceModel describes the data source data model.
type ScriptSnippetDataSourceModel struct {
    Id    types.Int64  `tfsdk:"id"`
    Name  types.String `tfsdk:"name"`
    Desc  types.String `tfsdk:"desc"`
    Code  types.String `tfsdk:"code"`
    Shell types.String `tfsdk:"shell"`
}

func (d *ScriptSnippetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_script_snippet"
}

func (d *ScriptSnippetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Script Snippet data source for Tactical RMM. Use this to look up existing script snippets by ID or name.",

        Attributes: map[string]schema.Attribute{
            "id": schema.Int64Attribute{
                MarkdownDescription: "Script snippet identifier. Either `id` or `name` must be specified.",
                Optional:            true,
                Computed:            true,
            },
            "name": schema.StringAttribute{
                MarkdownDescription: "Snippet name. Either `id` or `name` must be specified.",
                Optional:            true,
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
    }
}

func (d *ScriptSnippetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ScriptSnippetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data ScriptSnippetDataSourceModel

    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Validate that either ID or name is provided
    if data.Id.IsNull() && data.Name.IsNull() {
        resp.Diagnostics.AddError(
            "Missing Script Snippet Identifier",
            "Either 'id' or 'name' must be specified to look up a script snippet.",
        )
        return
    }

    var snippet map[string]interface{}

    if !data.Id.IsNull() {
        // Look up by ID
        httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/scripts/snippets/%d/", d.client.BaseURL, data.Id.ValueInt64()), nil)
        if err != nil {
            resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read script snippet, got error: %s", err))
            return
        }

        httpResp, err := d.client.Do(httpReq)
        if err != nil {
            resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read script snippet, got error: %s", err))
            return
        }
        defer httpResp.Body.Close()

        if httpResp.StatusCode == http.StatusNotFound {
            resp.Diagnostics.AddError("Script Snippet Not Found", fmt.Sprintf("Script snippet with ID %d not found", data.Id.ValueInt64()))
            return
        }

        if httpResp.StatusCode != http.StatusOK {
            resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read script snippet, status code: %d", httpResp.StatusCode))
            return
        }

        if err := json.NewDecoder(httpResp.Body).Decode(&snippet); err != nil {
            resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
            return
        }
    } else {
        // Look up by name - need to list all snippets and find the matching one
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

        // Find the snippet by name
        for _, s := range snippets {
            if name, ok := s["name"].(string); ok && name == data.Name.ValueString() {
                snippet = s
                break
            }
        }

        if snippet == nil {
            resp.Diagnostics.AddError("Script Snippet Not Found", fmt.Sprintf("Script snippet with name '%s' not found", data.Name.ValueString()))
            return
        }
    }

    // Update model with response data
    if id, ok := snippet["id"].(float64); ok {
        data.Id = types.Int64Value(int64(id))
    }
    if name, ok := snippet["name"].(string); ok {
        data.Name = types.StringValue(name)
    }
    if desc, ok := snippet["desc"].(string); ok {
        data.Desc = types.StringValue(desc)
    } else {
        data.Desc = types.StringNull()
    }
    if code, ok := snippet["code"].(string); ok {
        data.Code = types.StringValue(code)
    }
    if shell, ok := snippet["shell"].(string); ok {
        data.Shell = types.StringValue(shell)
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
