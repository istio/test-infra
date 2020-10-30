package github

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	token = "hunter2"
)

type fakeToken struct{}

func (fakeToken) Token(ctx context.Context) (string, error) {
	return token, nil
}

func TestGitHubApp(t *testing.T) {
	ctx := context.Background()
	gh := &GitHubApp{tr: fakeToken{}}

	want := &corev1.Secret{
		Type: corev1.SecretTypeBasicAuth,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"username": []byte("x-access-token"),
			"password": []byte(token),
		},
	}

	s, err := gh.Secret(ctx, want.GetNamespace(), want.GetName())
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, s); diff != "" {
		t.Error(diff)
	}
}
