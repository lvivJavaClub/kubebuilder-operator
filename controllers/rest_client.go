package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/lvivJavaClub/kubebuilder-operator/api/v1alpha1"
	"io"
	"net/http"
	"strings"
)

type Post struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func url(bp *v1alpha1.BlogPost) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local:%d/api/post/%s", bp.Spec.BlogAppName, bp.Namespace, 3001, bp.Name)
	//return "http://localhost:3001/api/post/" + bp.Name
}

func createNewPost(ctx context.Context, bp *v1alpha1.BlogPost) (*http.Response, error) {
	post := Post{
		Name:    bp.Spec.Name,
		Content: bp.Spec.Content,
	}
	postJson, _ := json.Marshal(post)
	req, err := http.NewRequestWithContext(ctx, "POST", url(bp), bytes.NewReader(postJson))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return res, err
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, res.Body)
		return res, fmt.Errorf("unexpected HTTP error: %s, %s", res.Status, buf.String())
	}
	return res, nil
}

func deletePost(ctx context.Context, bp *v1alpha1.BlogPost) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", url(bp), nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, res.Body)
		return fmt.Errorf("unexpected HTTP error: %s, %s", res.Status, buf.String())
	}
	return nil
}
