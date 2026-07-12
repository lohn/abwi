package ado

import (
	"context"
	"fmt"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
)

// HTMLFields returns the reference names of the project's HTML (large text)
// fields, which are the fields eligible for Markdown format.
func (c *Client) HTMLFields(ctx context.Context) (map[string]bool, error) {
	fields, err := c.WIT.GetWorkItemFields(ctx, workitemtracking.GetWorkItemFieldsArgs{Project: &c.Project})
	if err != nil {
		return nil, err
	}
	return htmlFieldSet(*fields), nil
}

func htmlFieldSet(fields []workitemtracking.WorkItemField2) map[string]bool {
	set := map[string]bool{}
	for _, f := range fields {
		if f.Type != nil && *f.Type == workitemtracking.FieldTypeValues.Html && f.ReferenceName != nil {
			set[*f.ReferenceName] = true
		}
	}
	return set
}

func (c *Client) Create(ctx context.Context, workItemType string, ops []webapi.JsonPatchOperation) (*workitemtracking.WorkItem, error) {
	return c.WIT.CreateWorkItem(ctx, workitemtracking.CreateWorkItemArgs{
		Project:  &c.Project,
		Type:     &workItemType,
		Document: &ops,
	})
}

func (c *Client) Update(ctx context.Context, id int, ops []webapi.JsonPatchOperation) (*workitemtracking.WorkItem, error) {
	return c.WIT.UpdateWorkItem(ctx, workitemtracking.UpdateWorkItemArgs{
		Project:  &c.Project,
		Id:       &id,
		Document: &ops,
	})
}

func (c *Client) Get(ctx context.Context, id int, withRelations bool) (*workitemtracking.WorkItem, error) {
	args := workitemtracking.GetWorkItemArgs{Id: &id, Project: &c.Project}
	if withRelations {
		args.Expand = &workitemtracking.WorkItemExpandValues.Relations
	}
	return c.WIT.GetWorkItem(ctx, args)
}

// Query runs a WIQL query and fetches the matched work items with the fields
// needed for list output.
func (c *Client) Query(ctx context.Context, wiql string, limit int) ([]workitemtracking.WorkItem, error) {
	res, err := c.WIT.QueryByWiql(ctx, workitemtracking.QueryByWiqlArgs{
		Wiql:    &workitemtracking.Wiql{Query: &wiql},
		Project: &c.Project,
		Top:     &limit,
	})
	if err != nil {
		return nil, err
	}
	if res.WorkItems == nil || len(*res.WorkItems) == 0 {
		return nil, nil
	}
	ids := make([]int, 0, len(*res.WorkItems))
	for _, r := range *res.WorkItems {
		ids = append(ids, *r.Id)
	}
	fields := []string{"System.Id", "System.WorkItemType", "System.State", "System.Title"}
	items, err := c.WIT.GetWorkItems(ctx, workitemtracking.GetWorkItemsArgs{Ids: &ids, Fields: &fields})
	if err != nil {
		return nil, err
	}
	return *items, nil
}

// WebURL returns the browser URL of a work item.
func (c *Client) WebURL(id int) string {
	return fmt.Sprintf("%s/%s/_workitems/edit/%d", c.Org, c.Project, id)
}
