package handler

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/umono-cms/umono/internal/models"
	"github.com/umono-cms/umono/internal/service"
	"github.com/umono-cms/umono/internal/view"
)

// bindingErrorText translates the binding error among errs for the Layout
// field and returns it with the root rule's call skeleton. Both are "" when
// errs has no binding error.
func bindingErrorText(c *fiber.Ctx, errs []error) (string, string) {
	for _, err := range errs {
		var be *service.BindingError
		if !errors.As(err, &be) {
			continue
		}
		text := translate(c, "admin.pages.layout.errors."+strings.ReplaceAll(string(be.Kind), "-", "_"))
		if len(be.Names) > 0 {
			text += ": " + strings.Join(be.Names, ", ")
		}
		if len(be.Offered) > 0 {
			text += ". " + translate(c, "admin.pages.layout.errors.offered_has") + ": " + strings.Join(be.Offered, ", ")
		}
		return text, be.Skeleton
	}
	return "", ""
}

// bindingString is the Layout field's value of a binding.
func bindingString(contractID *uint, layoutID *string) string {
	if contractID == nil || layoutID == nil {
		return ""
	}
	return fmt.Sprintf("%d:%s", *contractID, *layoutID)
}

// layoutField builds the Layout field of the saved page with value chosen.
func layoutField(translator any, ts *service.ThemeService, saved models.SitePage, value string) view.LayoutField {
	options, err := ts.BindingOptions(saved)
	if err != nil {
		log.Printf("error: layout options: %v", err)
	}
	return view.NewLayoutField(options, value, bindingString(saved.ContractID, saved.LayoutID), translator)
}
