package errors

import "fmt"

func TemplateNotFound(name string, msg string) error {
	// TODO real implementation
	return fmt.Errorf("TEMPLATE NOT FOUND %s", name)
}

func TemplateSyntaxError(msg string, lineno int, name *string, filename *string) error {
	// TODO real implementation
	return fmt.Errorf("TEMPLATE SYNTAX ERROR %s %d", msg, lineno)
}

func TemplateAssertionError(msg string, lineno int, name *string, filename *string) error {
	return TemplateAssertionError(msg, lineno, name, filename)
}

func TemplateError(msg string) error {
	return fmt.Errorf("TEMPLATE ERROR: %s", msg)
}
