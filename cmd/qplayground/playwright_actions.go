package main

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/playwright-community/playwright-go"
)

func init() {
	RegisterAction("playwright:goto", func() PluginAction { return &GotoAction{} })
	RegisterAction("playwright:click", func() PluginAction { return &ClickAction{} })
	RegisterAction("playwright:fill", func() PluginAction { return &FillAction{} })
	RegisterAction("playwright:type", func() PluginAction { return &TypeAction{} })
	RegisterAction("playwright:press", func() PluginAction { return &PressAction{} })
	RegisterAction("playwright:check", func() PluginAction { return &CheckAction{} })
	RegisterAction("playwright:uncheck", func() PluginAction { return &UncheckAction{} })
	RegisterAction("playwright:select_option", func() PluginAction { return &SelectOptionAction{} })
	RegisterAction("playwright:wait_for_selector", func() PluginAction { return &WaitForSelectorAction{} })
	RegisterAction("playwright:wait_for_timeout", func() PluginAction { return &WaitForTimeoutAction{} })
	RegisterAction("playwright:screenshot", func() PluginAction { return &ScreenshotAction{} })
	RegisterAction("playwright:evaluate", func() PluginAction { return &EvaluateAction{} })
	RegisterAction("playwright:hover", func() PluginAction { return &HoverAction{} })
	RegisterAction("playwright:scroll", func() PluginAction { return &ScrollAction{} })
	RegisterAction("playwright:get_text", func() PluginAction { return &GetTextAction{} })
	RegisterAction("playwright:get_attribute", func() PluginAction { return &GetAttributeAction{} })
	RegisterAction("playwright:wait_for_load_state", func() PluginAction { return &WaitForLoadStateAction{} })
	RegisterAction("playwright:set_viewport", func() PluginAction { return &SetViewportAction{} })
	RegisterAction("playwright:reload", func() PluginAction { return &ReloadAction{} })
	RegisterAction("playwright:go_back", func() PluginAction { return &GoBackAction{} })
	RegisterAction("playwright:go_forward", func() PluginAction { return &GoForwardAction{} })
}

// BaseAction provides common validation for selector-based actions.
type BaseAction struct{}

func (b *BaseAction) getSelector(actionConfig map[string]interface{}) (string, error) {
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return "", fmt.Errorf("action requires a 'selector' string in config")
	}
	return selector, nil
}

// GotoAction implements navigation to a URL
type GotoAction struct{}

func (a *GotoAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	url, ok := actionConfig["url"].(string)
	if !ok || url == "" {
		return fmt.Errorf("playwright:goto action requires a 'url' string in config")
	}

	runContext.Logger.Info("Executing playwright:goto", "url", url)

	options := playwright.PageGotoOptions{}
	if timeout, ok := actionConfig["timeout"].(float64); ok && timeout > 0 {
		options.Timeout = playwright.Float(timeout)
	}
	if waitUntil, ok := actionConfig["wait_until"].(string); ok {
		waitUntilState := playwright.WaitUntilState(waitUntil)
		options.WaitUntil = &waitUntilState
	}

	_, err := runContext.PlaywrightPage.Goto(url, options)
	return err
}

// ClickAction implements clicking on elements
type ClickAction struct {
	BaseAction
}

func (a *ClickAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, err := a.getSelector(actionConfig)
	if err != nil {
		return fmt.Errorf("playwright:click %w", err)
	}

	runContext.Logger.Info("Executing playwright:click", "selector", selector)

	options := playwright.LocatorClickOptions{}
	if button, ok := actionConfig["button"].(string); ok {
		btn := playwright.MouseButton(button)
		options.Button = &btn
	}
	if clickCount, ok := actionConfig["click_count"].(float64); ok {
		options.ClickCount = playwright.Int(int(clickCount))
	}
	if force, ok := actionConfig["force"].(bool); ok {
		options.Force = playwright.Bool(force)
	}

	return runContext.PlaywrightPage.Locator(selector).Click(options)
}

// FillAction implements filling input fields
type FillAction struct {
	BaseAction
}

func (a *FillAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, err := a.getSelector(actionConfig)
	if err != nil {
		return fmt.Errorf("playwright:fill %w", err)
	}
	value, ok := actionConfig["value"].(string)
	if !ok {
		return fmt.Errorf("playwright:fill action requires a 'value' string in config")
	}

	runContext.Logger.Info("Executing playwright:fill", "selector", selector, "value", value)

	options := playwright.LocatorFillOptions{}
	if force, ok := actionConfig["force"].(bool); ok {
		options.Force = playwright.Bool(force)
	}

	return runContext.PlaywrightPage.Locator(selector).Fill(value, options)
}

// TypeAction implements typing text
type TypeAction struct {
	BaseAction
}

func (a *TypeAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, err := a.getSelector(actionConfig)
	if err != nil {
		return fmt.Errorf("playwright:type %w", err)
	}
	text, ok := actionConfig["text"].(string)
	if !ok {
		return fmt.Errorf("playwright:type action requires a 'text' string in config")
	}

	runContext.Logger.Info("Executing playwright:type", "selector", selector, "text", text)

	options := playwright.LocatorTypeOptions{}
	if delay, ok := actionConfig["delay"].(float64); ok {
		options.Delay = playwright.Float(delay)
	}

	return runContext.PlaywrightPage.Locator(selector).Type(text, options)
}

// PressAction implements key presses
type PressAction struct {
	BaseAction
}

func (a *PressAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, err := a.getSelector(actionConfig)
	if err != nil {
		return fmt.Errorf("playwright:press %w", err)
	}
	key, ok := actionConfig["key"].(string)
	if !ok {
		return fmt.Errorf("playwright:press action requires a 'key' string in config")
	}

	runContext.Logger.Info("Executing playwright:press", "selector", selector, "key", key)

	options := playwright.LocatorPressOptions{}
	if delay, ok := actionConfig["delay"].(float64); ok {
		options.Delay = playwright.Float(delay)
	}

	return runContext.PlaywrightPage.Locator(selector).Press(key, options)
}

// CheckAction implements checking checkboxes
type CheckAction struct {
	BaseAction
}

func (a *CheckAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, err := a.getSelector(actionConfig)
	if err != nil {
		return fmt.Errorf("playwright:check %w", err)
	}

	runContext.Logger.Info("Executing playwright:check", "selector", selector)

	options := playwright.LocatorCheckOptions{}
	if force, ok := actionConfig["force"].(bool); ok {
		options.Force = playwright.Bool(force)
	}

	return runContext.PlaywrightPage.Locator(selector).Check(options)
}

// UncheckAction implements unchecking checkboxes
type UncheckAction struct {
	BaseAction
}

func (a *UncheckAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, err := a.getSelector(actionConfig)
	if err != nil {
		return fmt.Errorf("playwright:uncheck %w", err)
	}

	runContext.Logger.Info("Executing playwright:uncheck", "selector", selector)

	options := playwright.LocatorUncheckOptions{}
	if force, ok := actionConfig["force"].(bool); ok {
		options.Force = playwright.Bool(force)
	}

	return runContext.PlaywrightPage.Locator(selector).Uncheck(options)
}

// SelectOptionAction implements selecting options from dropdowns
type SelectOptionAction struct {
	BaseAction
}

func (a *SelectOptionAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, err := a.getSelector(actionConfig)
	if err != nil {
		return fmt.Errorf("playwright:select_option %w", err)
	}

	runContext.Logger.Info("Executing playwright:select_option", "selector", selector)

	var selectOptions playwright.SelectOptionValues

	// Support multiple ways to specify options
	if value, ok := actionConfig["value"].(string); ok {
		selectOptions.Values = &[]string{value}
	} else if values, ok := actionConfig["values"].([]interface{}); ok {
		stringValues := make([]string, len(values))
		for i, v := range values {
			if str, ok := v.(string); ok {
				stringValues[i] = str
			}
		}
		selectOptions.Values = &stringValues
	} else if label, ok := actionConfig["label"].(string); ok {
		selectOptions.Labels = &[]string{label}
	} else if index, ok := actionConfig["index"].(float64); ok {
		selectOptions.Indexes = &[]int{int(index)}
	} else {
		return fmt.Errorf("playwright:select_option action requires 'value', 'values', 'label', or 'index' in config")
	}

	_, err = runContext.PlaywrightPage.Locator(selector).SelectOption(selectOptions)
	return err
}

// WaitForSelectorAction implements waiting for elements
type WaitForSelectorAction struct {
	BaseAction
}

func (a *WaitForSelectorAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, err := a.getSelector(actionConfig)
	if err != nil {
		return fmt.Errorf("playwright:wait_for_selector %w", err)
	}

	runContext.Logger.Info("Executing playwright:wait_for_selector", "selector", selector)

	options := playwright.PageWaitForSelectorOptions{}
	if timeout, ok := actionConfig["timeout"].(float64); ok && timeout > 0 {
		options.Timeout = playwright.Float(timeout)
	}
	if state, ok := actionConfig["state"].(string); ok {
		waitForState := playwright.WaitForSelectorState(state)
		options.State = &waitForState
	}

	_, err = runContext.PlaywrightPage.WaitForSelector(selector, options)
	return err
}

// WaitForTimeoutAction implements waiting for a specific duration
type WaitForTimeoutAction struct{}

func (a *WaitForTimeoutAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	timeout, ok := actionConfig["timeout"].(float64)
	if !ok || timeout <= 0 {
		return fmt.Errorf("playwright:wait_for_timeout action requires a positive 'timeout' number in config")
	}

	runContext.Logger.Info("Executing playwright:wait_for_timeout", "timeout", timeout)

	runContext.PlaywrightPage.WaitForTimeout(timeout)
	return nil
}

// ScreenshotAction implements taking screenshots - CLI version saves to local filesystem
type ScreenshotAction struct{}

func (a *ScreenshotAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	runContext.Logger.Info("Executing playwright:screenshot")

	options := playwright.PageScreenshotOptions{}

	// Configure screenshot options
	if fullPage, ok := actionConfig["full_page"].(bool); ok {
		options.FullPage = playwright.Bool(fullPage)
	} else {
		options.FullPage = playwright.Bool(true) // Default to full page
	}

	if quality, ok := actionConfig["quality"].(float64); ok {
		options.Quality = playwright.Int(int(quality))
	}

	format := "png" // Default format
	if f, ok := actionConfig["format"].(string); ok {
		format = f
		imgFormat := playwright.ScreenshotType(format)
		options.Type = &imgFormat
	}

	// Take screenshot
	screenshotBytes, err := runContext.PlaywrightPage.Screenshot(options)
	if err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}

	// Generate filename - ignore R2 fields, use local path
	filename := fmt.Sprintf("screenshot_%d.%s", time.Now().Unix(), format)
	
	// Check if custom path is provided (ignoring R2 fields)
	if path, ok := actionConfig["path"].(string); ok && path != "" {
		// Use custom path but ensure it has correct extension
		if !strings.HasSuffix(path, "."+format) {
			path = strings.TrimSuffix(path, filepath.Ext(path)) + "." + format
		}
		filename = path
	}

	// Determine content type
	contentType := "image/png" // Default
	switch format {
	case "jpeg":
		contentType = "image/jpeg"
	case "png":
		contentType = "image/png"
	}

	// Save to local filesystem
	reader := bytes.NewReader(screenshotBytes)
	localPath, err := runContext.StorageService.UploadFile(ctx, filename, reader, contentType)
	if err != nil {
		return fmt.Errorf("failed to save screenshot locally: %w", err)
	}

	// Store the local path in action config for logging
	actionConfig["local_path"] = localPath

	runContext.Logger.Info("Screenshot saved locally", "path", localPath, "size", len(screenshotBytes))
	return nil
}

// EvaluateAction implements executing JavaScript
type EvaluateAction struct{}

func (a *EvaluateAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	expression, ok := actionConfig["expression"].(string)
	if !ok || expression == "" {
		return fmt.Errorf("playwright:evaluate action requires an 'expression' string in config")
	}

	runContext.Logger.Info("Executing playwright:evaluate", "expression", expression)

	_, err := runContext.PlaywrightPage.Evaluate(expression)
	return err
}

// HoverAction implements hovering over elements
type HoverAction struct {
	BaseAction
}

func (a *HoverAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, err := a.getSelector(actionConfig)
	if err != nil {
		return fmt.Errorf("playwright:hover %w", err)
	}

	runContext.Logger.Info("Executing playwright:hover", "selector", selector)

	options := playwright.LocatorHoverOptions{}
	if force, ok := actionConfig["force"].(bool); ok {
		options.Force = playwright.Bool(force)
	}

	return runContext.PlaywrightPage.Locator(selector).Hover(options)
}

// ScrollAction implements scrolling
type ScrollAction struct{}

func (a *ScrollAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	runContext.Logger.Info("Executing playwright:scroll")

	// Default scroll to bottom
	deltaY := 1000.0
	if dy, ok := actionConfig["delta_y"].(float64); ok {
		deltaY = dy
	}

	deltaX := 0.0
	if dx, ok := actionConfig["delta_x"].(float64); ok {
		deltaX = dx
	}

	script := fmt.Sprintf("window.scrollBy(%f, %f)", deltaX, deltaY)
	_, err := runContext.PlaywrightPage.Evaluate(script)
	return err
}

// GetTextAction implements getting text content
type GetTextAction struct {
	BaseAction
}

func (a *GetTextAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, err := a.getSelector(actionConfig)
	if err != nil {
		return fmt.Errorf("playwright:get_text %w", err)
	}

	runContext.Logger.Info("Executing playwright:get_text", "selector", selector)

	text, err := runContext.PlaywrightPage.Locator(selector).TextContent()
	if err != nil {
		return err
	}

	runContext.Logger.Info("Retrieved text", "selector", selector, "text", text)
	return nil
}

// GetAttributeAction implements getting element attributes
type GetAttributeAction struct {
	BaseAction
}

func (a *GetAttributeAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, err := a.getSelector(actionConfig)
	if err != nil {
		return fmt.Errorf("playwright:get_attribute %w", err)
	}

	attribute, ok := actionConfig["attribute"].(string)
	if !ok || attribute == "" {
		return fmt.Errorf("playwright:get_attribute action requires an 'attribute' string in config")
	}

	runContext.Logger.Info("Executing playwright:get_attribute", "selector", selector, "attribute", attribute)

	value, err := runContext.PlaywrightPage.Locator(selector).GetAttribute(attribute)
	if err != nil {
		return err
	}

	runContext.Logger.Info("Retrieved attribute", "selector", selector, "attribute", attribute, "value", value)
	return nil
}

// WaitForLoadStateAction implements waiting for page load states
type WaitForLoadStateAction struct{}

func (a *WaitForLoadStateAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	state := "load" // Default state
	if s, ok := actionConfig["state"].(string); ok {
		state = s
	}

	runContext.Logger.Info("Executing playwright:wait_for_load_state", "state", state)

	options := playwright.PageWaitForLoadStateOptions{}
	if timeout, ok := actionConfig["timeout"].(float64); ok && timeout > 0 {
		options.Timeout = playwright.Float(timeout)
	}

	stateField := playwright.LoadState(state)
	err := runContext.PlaywrightPage.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: &stateField,
	}, options)
	return err
}

// SetViewportAction implements setting viewport size
type SetViewportAction struct{}

func (a *SetViewportAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	width, ok1 := actionConfig["width"].(float64)
	height, ok2 := actionConfig["height"].(float64)

	if !ok1 || !ok2 {
		return fmt.Errorf("playwright:set_viewport action requires 'width' and 'height' numbers in config")
	}

	runContext.Logger.Info("Executing playwright:set_viewport", "width", width, "height", height)

	err := runContext.PlaywrightPage.SetViewportSize(int(width), int(height))
	return err
}

// ReloadAction implements page reload
type ReloadAction struct{}

func (a *ReloadAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	runContext.Logger.Info("Executing playwright:reload")

	options := playwright.PageReloadOptions{}
	if timeout, ok := actionConfig["timeout"].(float64); ok && timeout > 0 {
		options.Timeout = playwright.Float(timeout)
	}

	_, err := runContext.PlaywrightPage.Reload(options)
	return err
}

// GoBackAction implements browser back navigation
type GoBackAction struct{}

func (a *GoBackAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	runContext.Logger.Info("Executing playwright:go_back")

	options := playwright.PageGoBackOptions{}
	if timeout, ok := actionConfig["timeout"].(float64); ok && timeout > 0 {
		options.Timeout = playwright.Float(timeout)
	}

	_, err := runContext.PlaywrightPage.GoBack(options)
	return err
}

// GoForwardAction implements browser forward navigation
type GoForwardAction struct{}

func (a *GoForwardAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	runContext.Logger.Info("Executing playwright:go_forward")

	options := playwright.PageGoForwardOptions{}
	if timeout, ok := actionConfig["timeout"].(float64); ok && timeout > 0 {
		options.Timeout = playwright.Float(timeout)
	}

	_, err := runContext.PlaywrightPage.GoForward(options)
	return err
}