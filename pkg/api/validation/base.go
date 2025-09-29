package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	Validate    *validator.Validate
	langToTrans map[string]ut.Translator
)

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())

	Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	if err := registerCustomValidations(); err != nil {
		panic(err)
	}

	if err := registerTranslations(); err != nil {
		panic(err)
	}
}

func GetTranslation(lang string) ut.Translator {
	if lang == "" {
		return langToTrans["en"]
	}

	return langToTrans[lang]
}

func registerCustomValidations() error {
	if err := Validate.RegisterValidation("dns1123subdomain", ValidateDNS1123Subdomain); err != nil {
		return err
	}

	if err := Validate.RegisterValidation("dns1123label", ValidateDNS1123Label); err != nil {
		return err
	}

	if err := Validate.RegisterValidation("dns1035label", ValidateDNS1035Label); err != nil {
		return err
	}

	if err := Validate.RegisterValidation("kubeconfig", ValidateKubeconfig); err != nil {
		return err
	}

	return nil
}

func registerTranslations() error {
	enLocale := en.New()

	uni := ut.New(enLocale, enLocale)

	enTrans, _ := uni.GetTranslator("en")

	langToTrans = map[string]ut.Translator{
		"en": enTrans,
	}

	if err := en_translations.RegisterDefaultTranslations(Validate, enTrans); err != nil {
		return fmt.Errorf("failed to register default English translations: %w", err)
	}

	// Register translation for dns1123subdomain
	if err := Validate.RegisterTranslation("dns1123subdomain", enTrans, func(ut ut.Translator) error {
		return ut.Add(
			"dns1123subdomain",
			"{0} must be a valid DNS subdomain (RFC 1123): lowercase alphanumeric characters, '-' or '.', max 253 characters",
			true,
		)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("dns1123subdomain", fe.Field())
		return t
	}); err != nil {
		return err
	}

	// Register translation for dns1123label
	if err := Validate.RegisterTranslation("dns1123label", enTrans, func(ut ut.Translator) error {
		return ut.Add(
			"dns1123label",
			"{0} must be a valid DNS label (RFC 1123): lowercase alphanumeric characters or '-', max 63 characters",
			true,
		)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("dns1123label", fe.Field())
		return t
	}); err != nil {
		return err
	}

	// Register translation for dns1035label
	if err := Validate.RegisterTranslation("dns1035label", enTrans, func(ut ut.Translator) error {
		return ut.Add(
			"dns1035label",
			"{0} must be a valid DNS label (RFC 1035): must start with a lowercase letter, followed by lowercase alphanumeric characters or '-', max 63 characters", //nolint:lll
			true,
		)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("dns1035label", fe.Field())
		return t
	}); err != nil {
		return err
	}

	// Register translation for kubeconfig
	if err := Validate.RegisterTranslation("kubeconfig", enTrans, func(ut ut.Translator) error {
		return ut.Add(
			"kubeconfig",
			"{0} must be a valid kubeconfig file in YAML format with required fields (clusters, contexts, users)",
			true,
		)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("kubeconfig", fe.Field())
		return t
	}); err != nil {
		return err
	}

	return nil
}

var (
	// RFC 1123 DNS Subdomain regex (max 253 chars)
	// Pattern: [a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
	dns1123SubdomainRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)

	// RFC 1123 DNS Label regex (max 63 chars)
	// Pattern: [a-z0-9]([-a-z0-9]*[a-z0-9])?
	dns1123LabelRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

	// RFC 1035 DNS Label regex (max 63 chars)
	// Pattern: [a-z]([-a-z0-9]*[a-z0-9])?
	dns1035LabelRegex = regexp.MustCompile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)
)

// ValidateDNS1123Subdomain validates RFC 1123 DNS subdomain names
// Used by most Kubernetes resources
func ValidateDNS1123Subdomain(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	// Check length
	if len(value) == 0 || len(value) > 253 {
		return false
	}

	// Check pattern
	return dns1123SubdomainRegex.MatchString(value)
}

// ValidateDNS1123Label validates RFC 1123 DNS label names
// Used by container names, port names, etc.
func ValidateDNS1123Label(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	// Check length
	if len(value) == 0 || len(value) > 63 {
		return false
	}

	// Check pattern
	return dns1123LabelRegex.MatchString(value)
}

// ValidateDNS1035Label validates RFC 1035 DNS label names
// Used by some specific Kubernetes resources
func ValidateDNS1035Label(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	// Check length
	if len(value) == 0 || len(value) > 63 {
		return false
	}

	// Check pattern
	return dns1035LabelRegex.MatchString(value)
}

// ValidateKubeconfig validates that a string contains a valid kubeconfig YAML
// with required clusters, contexts, and users sections.
func ValidateKubeconfig(fl validator.FieldLevel) bool {
	kubeconfigContent := fl.Field().String()

	if kubeconfigContent == "" {
		return false
	}

	// Try to load the kubeconfig from the string content
	config, err := clientcmd.Load([]byte(kubeconfigContent))
	if err != nil {
		return false
	}

	// Check if config has at least one cluster, context, and user
	if len(config.Clusters) == 0 || len(config.Contexts) == 0 || len(config.AuthInfos) == 0 {
		return false
	}

	// Check if current-context exists and is valid
	if config.CurrentContext != "" {
		if _, exists := config.Contexts[config.CurrentContext]; !exists {
			return false
		}
	}

	return true
}
