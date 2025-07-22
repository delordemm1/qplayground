package playwright

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/delordemm1/qplayground/internal/modules/automation"
	"github.com/playwright-community/playwright-go"
)

func init() {
	automation.RegisterAction("playwright:goto", func() automation.PluginAction { return &GotoAction{} })
	automation.RegisterAction("playwright:click", func() automation.PluginAction { return &ClickAction{} })
	automation.RegisterAction("playwright:fill", func() automation.PluginAction { return &FillAction{} })
	automation.RegisterAction("playwright:type", func() automation.PluginAction { return &TypeAction{} })
	automation.RegisterAction("playwright:press", func() automation.PluginAction { return &PressAction{} })
	automation.RegisterAction("playwright:check", func() automation.PluginAction { return &CheckAction{} })
	automation.RegisterAction("playwright:uncheck", func() automation.PluginAction { return &UncheckAction{} })
	automation.RegisterAction("playwright:select_option", func() automation.PluginAction { return &SelectOptionAction{} })
	automation.RegisterAction("playwright:wait_for_selector", func() automation.PluginAction { return &WaitForSelectorAction{} })
	automation.RegisterAction("playwright:wait_for_timeout", func() automation.PluginAction { return &WaitForTimeoutAction{} })
	automation.RegisterAction("playwright:screenshot", func() automation.PluginAction { return &ScreenshotAction{} })
	automation.RegisterAction("playwright:evaluate", func() automation.PluginAction { return &EvaluateAction{} })
	automation.RegisterAction("playwright:hover", func() automation.PluginAction { return &HoverAction{} })
	automation.RegisterAction("playwright:scroll", func() automation.PluginAction { return &ScrollAction{} })
	automation.RegisterAction("playwright:get_text", func() automation.PluginAction { return &GetTextAction{} })
	automation.RegisterAction("playwright:get_attribute", func() automation.PluginAction { return &GetAttributeAction{} })
	automation.RegisterAction("playwright:wait_for_load_state", func() automation.PluginAction { return &WaitForLoadStateAction{} })
	automation.RegisterAction("playwright:set_viewport", func() automation.PluginAction { return &SetViewportAction{} })
	automation.RegisterAction("playwright:reload", func() automation.PluginAction { return &ReloadAction{} })
	automation.RegisterAction("playwright:go_back", func() automation.PluginAction { return &GoBackAction{} })
	automation.RegisterAction("playwright:go_forward", func() automation.PluginAction { return &GoForwardAction{} })
	automation.RegisterAction("playwright:if_else", func() automation.PluginAction { return &IfElseAction{} })
	automation.RegisterAction("playwright:log", func() automation.PluginAction { return &LogAction{} })
	automation.RegisterAction("playwright:loop_until", func() automation.PluginAction { return &LoopUntilAction{} })
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

func (a *GotoAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *ClickAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *FillAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *TypeAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *PressAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *CheckAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *UncheckAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *SelectOptionAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *WaitForSelectorAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *WaitForTimeoutAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	timeout, ok := actionConfig["timeout"].(float64)
	if !ok || timeout <= 0 {
		return fmt.Errorf("playwright:wait_for_timeout action requires a positive 'timeout' number in config")
	}

	runContext.Logger.Info("Executing playwright:wait_for_timeout", "timeout", timeout)

	runContext.PlaywrightPage.WaitForTimeout(timeout)
	return nil
}

// ScreenshotAction implements taking screenshots
type ScreenshotAction struct{}

func (a *ScreenshotAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

	if format, ok := actionConfig["format"].(string); ok {
		imgFormat := playwright.ScreenshotType(format)
		options.Type = &imgFormat
	}

	// Take screenshot
	screenshotBytes, err := runContext.PlaywrightPage.Screenshot(options)
	if err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}

	// Check if we should upload to R2
	uploadToR2, _ := actionConfig["upload_to_r2"].(bool)
	if uploadToR2 {
		r2Key, ok := actionConfig["r2_key"].(string)
		if !ok || r2Key == "" {
			return fmt.Errorf("playwright:screenshot with upload_to_r2 requires an 'r2_key' string in config")
		}

		// Determine content type
		contentType := "image/png" // Default
		if format, ok := actionConfig["format"].(string); ok {
			switch format {
			case "jpeg":
				contentType = "image/jpeg"
			case "png":
				contentType = "image/png"
			}
		}

		// Upload to R2
		reader := bytes.NewReader(screenshotBytes)
		_, err := runContext.StorageService.UploadFile(ctx, r2Key, reader, contentType)
		if err != nil {
			return fmt.Errorf("failed to upload screenshot to R2: %w", err)
		}

		runContext.Logger.Info("Screenshot uploaded to R2", "key", r2Key, "size", len(screenshotBytes))
	}

	return nil
}

// EvaluateAction implements executing JavaScript
type EvaluateAction struct{}

func (a *EvaluateAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *HoverAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *ScrollAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *GetTextAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *GetAttributeAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *WaitForLoadStateAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *SetViewportAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *ReloadAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *GoBackAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
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

func (a *GoForwardAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	runContext.Logger.Info("Executing playwright:go_forward")

	options := playwright.PageGoForwardOptions{}
	if timeout, ok := actionConfig["timeout"].(float64); ok && timeout > 0 {
		options.Timeout = playwright.Float(timeout)
	}

	_, err := runContext.PlaywrightPage.GoForward(options)
	return err
}

// IfElseAction implements conditional logic with multiple else-if blocks
type IfElseAction struct{}

func (a *IfElseAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return fmt.Errorf("playwright:if_else action requires a 'selector' string in config")
	}

	conditionType, ok := actionConfig["condition_type"].(string)
	if !ok || conditionType == "" {
		return fmt.Errorf("playwright:if_else action requires a 'condition_type' string in config")
	}

	runContext.Logger.Info("Executing playwright:if_else", "selector", selector, "condition_type", conditionType)

	// Evaluate main condition
	conditionMet, err := a.evaluateCondition(runContext, selector, conditionType)
	if err != nil {
		return fmt.Errorf("failed to evaluate main condition: %w", err)
	}

	if conditionMet {
		// Execute if_actions
		if ifActions, ok := actionConfig["if_actions"].([]interface{}); ok {
			runContext.Logger.Info("Main condition is true, executing if_actions", "count", len(ifActions))
			return a.executeNestedActions(ctx, ifActions, runContext)
		}
		return nil
	}

	// Check else_if_conditions
	if elseIfConditions, ok := actionConfig["else_if_conditions"].([]interface{}); ok {
		for i, elseIfCondition := range elseIfConditions {
			elseIfMap, ok := elseIfCondition.(map[string]interface{})
			if !ok {
				continue
			}

			elseIfSelector, ok := elseIfMap["selector"].(string)
			if !ok || elseIfSelector == "" {
				continue
			}

			elseIfConditionType, ok := elseIfMap["condition_type"].(string)
			if !ok || elseIfConditionType == "" {
				continue
			}

			runContext.Logger.Info("Evaluating else-if condition", "index", i, "selector", elseIfSelector, "condition_type", elseIfConditionType)

			elseIfConditionMet, err := a.evaluateCondition(runContext, elseIfSelector, elseIfConditionType)
			if err != nil {
				runContext.Logger.Warn("Failed to evaluate else-if condition", "index", i, "error", err)
				continue
			}

			if elseIfConditionMet {
				// Execute this else-if's actions
				if elseIfActions, ok := elseIfMap["actions"].([]interface{}); ok {
					runContext.Logger.Info("Else-if condition is true, executing actions", "index", i, "count", len(elseIfActions))
					return a.executeNestedActions(ctx, elseIfActions, runContext)
				}
				return nil
			}
		}
	}

	// Execute else_actions if all conditions failed
	if elseActions, ok := actionConfig["else_actions"].([]interface{}); ok {
		runContext.Logger.Info("All conditions failed, executing else_actions", "count", len(elseActions))
		return a.executeNestedActions(ctx, elseActions, runContext)
	}

	runContext.Logger.Info("No conditions met and no else actions defined")
	return nil
}

func (a *IfElseAction) evaluateCondition(runContext *automation.RunContext, selector, conditionType string) (bool, error) {
	locator := runContext.PlaywrightPage.Locator(selector)

	switch conditionType {
	case "is_enabled":
		return locator.IsEnabled()
	case "is_disabled":
		return locator.IsDisabled()
	case "is_visible":
		return locator.IsVisible()
	case "is_hidden":
		return locator.IsHidden()
	case "is_checked":
		return locator.IsChecked()
	case "is_editable":
		return locator.IsEditable()
	default:
		return false, fmt.Errorf("unsupported condition type: %s", conditionType)
	}
}

func (a *IfElseAction) executeNestedActions(ctx context.Context, actions []interface{}, runContext *automation.RunContext) error {
	for i, actionInterface := range actions {
		actionMap, ok := actionInterface.(map[string]interface{})
		if !ok {
			runContext.Logger.Warn("Invalid nested action format", "index", i)
			continue
		}

		actionType, ok := actionMap["action_type"].(string)
		if !ok || actionType == "" {
			runContext.Logger.Warn("Missing action_type in nested action", "index", i)
			continue
		}

		actionConfig, ok := actionMap["action_config"].(map[string]interface{})
		if !ok {
			actionConfig = make(map[string]interface{})
		}

		runContext.Logger.Info("Executing nested action", "index", i, "action_type", actionType)

		// Get the plugin action
		pluginAction, err := automation.GetAction(actionType)
		if err != nil {
			runContext.Logger.Error("Failed to get nested action", "action_type", actionType, "error", err)
			return fmt.Errorf("failed to get nested action '%s': %w", actionType, err)
		}

		// Execute the nested action
		err = pluginAction.Execute(ctx, actionConfig, runContext)
		if err != nil {
			runContext.Logger.Error("Nested action failed", "action_type", actionType, "error", err)
			return fmt.Errorf("nested action '%s' failed: %w", actionType, err)
		}

		runContext.Logger.Info("Nested action completed", "action_type", actionType)
	}

	return nil
}

// LogAction implements logging messages
type LogAction struct{}

func (a *LogAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	message, ok := actionConfig["message"].(string)
	if !ok || message == "" {
		return fmt.Errorf("playwright:log action requires a 'message' string in config")
	}

	level, _ := actionConfig["level"].(string)
	if level == "" {
		level = "info"
	}

	switch level {
	case "debug":
		runContext.Logger.Debug("User Log", "message", message)
	case "warn":
		runContext.Logger.Warn("User Log", "message", message)
	case "error":
		runContext.Logger.Error("User Log", "message", message)
	default:
		runContext.Logger.Info("User Log", "message", message)
	}

	return nil
}

// LoopUntilAction implements looping until a condition is met or force stop
type LoopUntilAction struct{}

func (a *LoopUntilAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	runContext.Logger.Info("Executing playwright:loop_until")

	// Extract configuration
	selector, _ := actionConfig["selector"].(string)
	conditionType, _ := actionConfig["condition_type"].(string)
	maxLoops, _ := actionConfig["max_loops"].(float64)
	timeoutMs, _ := actionConfig["timeout_ms"].(float64)
	failOnForceStop, _ := actionConfig["fail_on_force_stop"].(bool)
	loopActionsInterface, _ := actionConfig["loop_actions"].([]interface{})

	// Validate that at least one force stop condition is provided
	if maxLoops <= 0 && timeoutMs <= 0 {
		return fmt.Errorf("playwright:loop_until requires either max_loops or timeout_ms to prevent infinite loops")
	}

	// Validate selector condition if provided
	if selector != "" && conditionType == "" {
		return fmt.Errorf("playwright:loop_until requires condition_type when selector is provided")
	}

	// Convert loop actions to proper format
	var loopActions []map[string]interface{}
	for _, actionInterface := range loopActionsInterface {
		if actionMap, ok := actionInterface.(map[string]interface{}); ok {
			loopActions = append(loopActions, actionMap)
		}
	}

	if len(loopActions) == 0 {
		return fmt.Errorf("playwright:loop_until requires at least one loop action")
	}

	runContext.Logger.Info("Starting loop",
		"selector", selector,
		"condition_type", conditionType,
		"max_loops", maxLoops,
		"timeout_ms", timeoutMs,
		"loop_actions_count", len(loopActions))

	// Initialize loop variables
	loopCount := 0
	startTime := time.Now()
	var timeoutDuration time.Duration
	if timeoutMs > 0 {
		timeoutDuration = time.Duration(timeoutMs) * time.Millisecond
	}

	for {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("loop cancelled")
		default:
		}

		loopCount++
		runContext.Logger.Info("Loop iteration", "count", loopCount)

		// Check selector condition if provided
		if selector != "" && conditionType != "" {
			conditionMet, err := a.evaluateCondition(runContext, selector, conditionType)
			if err != nil {
				runContext.Logger.Warn("Failed to evaluate loop condition", "error", err)
			} else if conditionMet {
				runContext.Logger.Info("Loop condition met, exiting loop", "selector", selector, "condition_type", conditionType, "loops_completed", loopCount)
				break
			}
		}

		// Check force stop conditions
		forceStop := false
		forceStopReason := ""

		if maxLoops > 0 && float64(loopCount) >= maxLoops {
			forceStop = true
			forceStopReason = fmt.Sprintf("reached maximum loops (%d)", int(maxLoops))
		}

		if timeoutMs > 0 && time.Since(startTime) >= timeoutDuration {
			forceStop = true
			if forceStopReason != "" {
				forceStopReason += " and "
			}
			forceStopReason += fmt.Sprintf("reached timeout (%dms)", int(timeoutMs))
		}

		if forceStop {
			message := fmt.Sprintf("Loop force stopped: %s", forceStopReason)
			if failOnForceStop {
				runContext.Logger.Error("Loop force stopped", "reason", forceStopReason, "loops_completed", loopCount)
				return fmt.Errorf(message)
			} else {
				runContext.Logger.Warn("Loop force stopped", "reason", forceStopReason, "loops_completed", loopCount)
				break
			}
		}

		// Execute loop actions
		for actionIndex, actionData := range loopActions {
			// Check for cancellation before each action
			select {
			case <-ctx.Done():
				return fmt.Errorf("loop cancelled during action execution")
			default:
			}

			actionType, ok := actionData["action_type"].(string)
			if !ok || actionType == "" {
				runContext.Logger.Warn("Skipping loop action with missing action_type", "action_index", actionIndex)
				continue
			}

			actionConfig, ok := actionData["action_config"].(map[string]interface{})
			if !ok {
				actionConfig = make(map[string]interface{})
			}

			runContext.Logger.Info("Executing loop action", "loop_count", loopCount, "action_index", actionIndex, "action_type", actionType)

			// Get the plugin action
			pluginAction, err := automation.GetAction(actionType)
			if err != nil {
				runContext.Logger.Error("Failed to get loop action", "action_type", actionType, "error", err)
				return fmt.Errorf("failed to get loop action '%s': %w", actionType, err)
			}

			// Execute the loop action
			err = pluginAction.Execute(ctx, actionConfig, runContext)
			if err != nil {
				runContext.Logger.Error("Loop action failed", "action_type", actionType, "loop_count", loopCount, "error", err)
				return fmt.Errorf("loop action '%s' failed in iteration %d: %w", actionType, loopCount, err)
			}

			runContext.Logger.Info("Loop action completed", "action_type", actionType, "loop_count", loopCount)
		}

		// Small delay to prevent busy-waiting and allow page updates
		time.Sleep(100 * time.Millisecond)
	}

	runContext.Logger.Info("Loop completed successfully", "total_loops", loopCount)
	return nil
}

func (a *LoopUntilAction) evaluateCondition(runContext *automation.RunContext, selector, conditionType string) (bool, error) {
	locator := runContext.PlaywrightPage.Locator(selector)

	switch conditionType {
	case "is_enabled":
		return locator.IsEnabled()
	case "is_disabled":
		return locator.IsDisabled()
	case "is_visible":
		return locator.IsVisible()
	case "is_hidden":
		return locator.IsHidden()
	case "is_checked":
		return locator.IsChecked()
	case "is_editable":
		return locator.IsEditable()
	default:
		return false, fmt.Errorf("unsupported condition type: %s", conditionType)
	}
}