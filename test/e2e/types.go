package e2e

import "k8s.io/apimachinery/pkg/runtime/schema"

type IdentifiableKind interface {
	GetName() string
	GetNamespace() string
	GetAnnotations() map[string]string
	GetObjectKind() schema.ObjectKind
}
