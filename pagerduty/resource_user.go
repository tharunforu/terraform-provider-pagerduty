package pagerduty

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-pagerduty/client"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Schema: map[string]*schema.Schema{
			"last_updated": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"email": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"role": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"contact_methods": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"summary": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
					},
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	fmt.Println("resource create")
	ApiClient := m.(*client.Client)
	var diags diag.Diagnostics
	name := d.Get("name").(string)
	email := d.Get("email").(string)
	Type := d.Get("type").(string)
	Role := d.Get("role").(string)

	payload_body := client.Whole_body{
		User: client.User{
			Name:  name,
			Email: email,
			Type:  Type,
			Role:  Role,
		},
	}
	var err error
	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		User_response, err := ApiClient.CreateUser(payload_body)
		if err != nil {
			if ApiClient.IsRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		d.Set("id", User_response.User.Id)
		d.SetId(User_response.User.Id)
		return nil
	})
	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return diag.FromErr(retryErr)
	}
	if err != nil {
		log.Println("[UPDATE ERROR]: ", err)
		return diag.Errorf("unable to update user")
	}

	resourceUserRead(ctx, d, m)
	return diags
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	fmt.Println("resource read")
	ApiClient := m.(*client.Client)
	Id := d.Id()

	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		User_response, err := ApiClient.GetUser(Id)
		if err != nil {
			if ApiClient.IsRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		d.Set("email", User_response.User.Email)
		d.Set("name", User_response.User.Name)
		d.Set("id", User_response.User.Id)
		d.Set("type", User_response.User.Type)
		d.Set("role", User_response.User.Role)
		contact_methods_list := make([]interface{}, len(User_response.User.Contact_methods))

		for i, com := range User_response.User.Contact_methods {
			contact := make(map[string]interface{})
			contact["type"] = com.Type
			contact["summary"] = com.Summary

			contact_methods_list[i] = contact
		}
		d.Set("contact_methods", contact_methods_list)

		return nil
	})
	if retryErr != nil {
		if strings.Contains(retryErr.Error(), "The requested resource was not found.") == true {
			d.SetId("")
			return diags
		}
		return diag.FromErr(retryErr)
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ApiClient := m.(*client.Client)
	fmt.Println("resource update")
	Id := d.Id()
	var diags diag.Diagnostics
	if d.HasChange("email") || d.HasChange("name") || d.HasChange("type") || d.HasChange("role") {
		name := d.Get("name").(string)
		email := d.Get("email").(string)
		Type := d.Get("type").(string)
		Role := d.Get("role").(string)
		payload_body := client.Whole_body{
			User: client.User{
				Name:  name,
				Email: email,
				Type:  Type,
				Role:  Role,
			},
		}

		var err error
		retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
			_, err := ApiClient.UpdateUser(payload_body, Id)
			if err != nil {
				if ApiClient.IsRetry(err) {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}

			return nil
		})
		if retryErr != nil {
			time.Sleep(2 * time.Second)
			return diag.FromErr(retryErr)
		}
		if err != nil {
			log.Println("[UPDATE ERROR]: ", err)
			return diag.Errorf("unable to update user")
		}

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}
	resourceUserRead(ctx, d, m)
	return diags
}
func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	fmt.Println("resource delete")
	ApiClient := m.(*client.Client)
	var diags diag.Diagnostics
	userID := d.Id()

	var err error
	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		if err = ApiClient.DeleteUser(userID); err != nil {
			if ApiClient.IsRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return diag.FromErr(retryErr)
	}

	if err != nil {
		log.Println("[DELETE ERROR]: ", err)
		return diag.Errorf("unable to delete user")
	}
	d.SetId("")
	return diags
}
func resourceUserImporter(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	id := d.Id()
	ApiClient := m.(*client.Client)
	User_response, err := ApiClient.GetUser(id)
	if err != nil {
		return nil, err
	}
	d.Set("email", User_response.User.Email)
	d.Set("name", User_response.User.Name)
	d.Set("id", User_response.User.Id)
	d.Set("type", User_response.User.Type)
	d.Set("role", User_response.User.Role)
	d.SetId(id)
	return []*schema.ResourceData{d}, nil
}
