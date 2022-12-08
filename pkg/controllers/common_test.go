package controllers_test

import (
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

type lifecycleReader struct {
	lifecycle templates.Lifecycle
}

func (l *lifecycleReader) GetLifecycle() *templates.Lifecycle {
	return &l.lifecycle
}

func (l *lifecycleReader) GetDefaultParams() v1alpha1.TemplateParams {
	panic("not implemented")
}
func (l *lifecycleReader) GetResourceTemplate() v1alpha1.TemplateSpec {
	panic("not implemented")
}
func (l *lifecycleReader) GetHealthRule() *v1alpha1.HealthRule {
	panic("not implemented")
}
func (l *lifecycleReader) IsYTTTemplate() bool {
	panic("not implemented")
}
func (l *lifecycleReader) GetRetentionPolicy() v1alpha1.RetentionPolicy {
	panic("not implemented")
}
