package playwright

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/playwright-community/playwright-go"
)

// ClickAction implements clicking on elements
type ClickAction struct{}

func (a *ClickAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return fmt.Errorf("playwright:click action requires a 'selector' string in config")
	}

	runContext.Logger.Info("Executing playwright:click", "selector", selector)

	options := playwright.LocatorClickOptions{}

	// Handle optional parameters
	if button, ok := actionConfig["button"].(string); ok {
		mouseButton := playwright.MouseButtonLeft
		switch button {
		case "right":
			mouseButton = playwright.MouseButtonRight
		case "middle":
			mouseButton = playwright.MouseButtonMiddle
		}
		options.Button = &mouseButton
	}

	if clickCount, ok := actionConfig["click_count"].(float64); ok {
		count := int(clickCount)
		options.ClickCount = &count
	}

	if delay, ok := actionConfig["delay"].(float64); ok {
		delayMs := float64(delay)
		options.Delay = &delayMs
	}

	if force, ok := actionConfig["force"].(bool); ok {
		options.Force = &force
	}

	if modifiers, ok := actionConfig["modifiers"].([]interface{}); ok {
		var keyboardModifiers []playwright.KeyboardModifier
		for _, mod := range modifiers {
			if modStr, ok := mod.(string); ok {
				switch modStr {
				case "Alt":
					keyboardModifiers = append(keyboardModifiers, playwright.KeyboardModifierAlt)
				case "Control":
					keyboardModifiers = append(keyboardModifiers, playwright.KeyboardModifierControl)
				case "Meta":
					keyboardModifiers = append(keyboardModifiers, playwright.KeyboardModifierMeta)
				case "Shift":
					keyboardModifiers = append(keyboardModifiers, playwright.KeyboardModifierShift)
				}
			}
		}
		if len(keyboardModifiers) > 0 {
			options.Modifiers = keyboardModifiers
		}
	}

	if noWaitAfter, ok := actionConfig["no_wait_after"].(bool); ok {
		options.NoWaitAfter = &noWaitAfter
	}

	if position, ok := actionConfig["position"].(map[string]interface{}); ok {
		if x, xOk := position["x"].(float64); xOk {
			if y, yOk := position["y"].(float64); yOk {
				options.Position = &playwright.Position{
					X: float64(x),
					Y: float64(y),
				}
			}
		}
	}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	if trial, ok := actionConfig["trial"].(bool); ok {
		options.Trial = &trial
	}

	err := runContext.PlaywrightPage.Locator(selector).Click(options)
	if err != nil {
		return fmt.Errorf("failed to click element with selector '%s': %w", selector, err)
	}

	runContext.Logger.Info("Successfully clicked element", "selector", selector)
	return nil
}

// FillAction implements filling input fields
type FillAction struct{}

func (a *FillAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return fmt.Errorf("playwright:fill action requires a 'selector' string in config")
	}

	value, ok := actionConfig["value"].(string)
	if !ok {
		return fmt.Errorf("playwright:fill action requires a 'value' string in config")
	}

	runContext.Logger.Info("Executing playwright:fill", "selector", selector, "value", value)

	options := playwright.LocatorFillOptions{}

	if force, ok := actionConfig["force"].(bool); ok {
		options.Force = &force
	}

	if noWaitAfter, ok := actionConfig["no_wait_after"].(bool); ok {
		options.NoWaitAfter = &noWaitAfter
	}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	err := runContext.PlaywrightPage.Locator(selector).Fill(value, options)
	if err != nil {
		return fmt.Errorf("failed to fill element with selector '%s': %w", selector, err)
	}

	runContext.Logger.Info("Successfully filled element", "selector", selector)
	return nil
}

// TypeAction implements typing text into elements
type TypeAction struct{}

func (a *TypeAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return fmt.Errorf("playwright:type action requires a 'selector' string in config")
	}

	text, ok := actionConfig["text"].(string)
	if !ok {
		return fmt.Errorf("playwright:type action requires a 'text' string in config")
	}

	runContext.Logger.Info("Executing playwright:type", "selector", selector, "text", text)

	options := playwright.LocatorTypeOptions{}

	if delay, ok := actionConfig["delay"].(float64); ok {
		delayMs := float64(delay)
		options.Delay = &delayMs
	}

	if noWaitAfter, ok := actionConfig["no_wait_after"].(bool); ok {
		options.NoWaitAfter = &noWaitAfter
	}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	err := runContext.PlaywrightPage.Locator(selector).Type(text, options)
	if err != nil {
		return fmt.Errorf("failed to type into element with selector '%s': %w", selector, err)
	}

	runContext.Logger.Info("Successfully typed into element", "selector", selector)
	return nil
}

// WaitAction implements waiting for elements or conditions
type WaitAction struct{}

func (a *WaitAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	waitType, ok := actionConfig["wait_type"].(string)
	if !ok || waitType == "" {
		return fmt.Errorf("playwright:wait action requires a 'wait_type' string in config")
	}

	runContext.Logger.Info("Executing playwright:wait", "wait_type", waitType)

	switch waitType {
	case "selector":
		return a.waitForSelector(actionConfig, runContext)
	case "timeout":
		return a.waitForTimeout(actionConfig, runContext)
	case "load_state":
		return a.waitForLoadState(actionConfig, runContext)
	case "url":
		return a.waitForURL(actionConfig, runContext)
	default:
		return fmt.Errorf("unsupported wait_type: %s", waitType)
	}
}

func (a *WaitAction) waitForSelector(actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return fmt.Errorf("playwright:wait with wait_type 'selector' requires a 'selector' string in config")
	}

	options := playwright.PageWaitForSelectorOptions{}

	if state, ok := actionConfig["state"].(string); ok {
		switch state {
		case "attached":
			waitForSelectorState := playwright.WaitForSelectorStateAttached
			options.State = &waitForSelectorState
		case "detached":
			waitForSelectorState := playwright.WaitForSelectorStateDetached
			options.State = &waitForSelectorState
		case "visible":
			waitForSelectorState := playwright.WaitForSelectorStateVisible
			options.State = &waitForSelectorState
		case "hidden":
			waitForSelectorState := playwright.WaitForSelectorStateHidden
			options.State = &waitForSelectorState
		}
	}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	_, err := runContext.PlaywrightPage.WaitForSelector(selector, options)
	if err != nil {
		return fmt.Errorf("failed to wait for selector '%s': %w", selector, err)
	}

	runContext.Logger.Info("Successfully waited for selector", "selector", selector)
	return nil
}

func (a *WaitAction) waitForTimeout(actionConfig map[string]interface{}, runContext *RunContext) error {
	timeout, ok := actionConfig["timeout"].(float64)
	if !ok {
		return fmt.Errorf("playwright:wait with wait_type 'timeout' requires a 'timeout' number in config")
	}

	runContext.PlaywrightPage.WaitForTimeout(float64(timeout))
	runContext.Logger.Info("Successfully waited for timeout", "timeout_ms", timeout)
	return nil
}

func (a *WaitAction) waitForLoadState(actionConfig map[string]interface{}, runContext *RunContext) error {
	options := playwright.PageWaitForLoadStateOptions{}

	if state, ok := actionConfig["state"].(string); ok {
		switch state {
		case "load":
			loadState := playwright.LoadStateLoad
			options.State = &loadState
		case "domcontentloaded":
			loadState := playwright.LoadStateDomcontentloaded
			options.State = &loadState
		case "networkidle":
			loadState := playwright.LoadStateNetworkidle
			options.State = &loadState
		}
	}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	err := runContext.PlaywrightPage.WaitForLoadState(options)
	if err != nil {
		return fmt.Errorf("failed to wait for load state: %w", err)
	}

	runContext.Logger.Info("Successfully waited for load state")
	return nil
}

func (a *WaitAction) waitForURL(actionConfig map[string]interface{}, runContext *RunContext) error {
	url, ok := actionConfig["url"].(string)
	if !ok || url == "" {
		return fmt.Errorf("playwright:wait with wait_type 'url' requires a 'url' string in config")
	}

	options := playwright.PageWaitForURLOptions{}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	if waitUntil, ok := actionConfig["wait_until"].(string); ok {
		switch waitUntil {
		case "load":
			waitUntilState := playwright.WaitUntilStateLoad
			options.WaitUntil = &waitUntilState
		case "domcontentloaded":
			waitUntilState := playwright.WaitUntilStateDomcontentloaded
			options.WaitUntil = &waitUntilState
		case "networkidle":
			waitUntilState := playwright.WaitUntilStateNetworkidle
			options.WaitUntil = &waitUntilState
		case "commit":
			waitUntilState := playwright.WaitUntilStateCommit
			options.WaitUntil = &waitUntilState
		}
	}

	err := runContext.PlaywrightPage.WaitForURL(url, options)
	if err != nil {
		return fmt.Errorf("failed to wait for URL '%s': %w", url, err)
	}

	runContext.Logger.Info("Successfully waited for URL", "url", url)
	return nil
}

// NavigateAction implements page navigation
type NavigateAction struct{}

func (a *NavigateAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	url, ok := actionConfig["url"].(string)
	if !ok || url == "" {
		return fmt.Errorf("playwright:navigate action requires a 'url' string in config")
	}

	runContext.Logger.Info("Executing playwright:navigate", "url", url)

	options := playwright.PageGotoOptions{}

	if referer, ok := actionConfig["referer"].(string); ok {
		options.Referer = &referer
	}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	if waitUntil, ok := actionConfig["wait_until"].(string); ok {
		switch waitUntil {
		case "load":
			waitUntilState := playwright.WaitUntilStateLoad
			options.WaitUntil = &waitUntilState
		case "domcontentloaded":
			waitUntilState := playwright.WaitUntilStateDomcontentloaded
			options.WaitUntil = &waitUntilState
		case "networkidle":
			waitUntilState := playwright.WaitUntilStateNetworkidle
			options.WaitUntil = &waitUntilState
		case "commit":
			waitUntilState := playwright.WaitUntilStateCommit
			options.WaitUntil = &waitUntilState
		}
	}

	_, err := runContext.PlaywrightPage.Goto(url, options)
	if err != nil {
		return fmt.Errorf("failed to navigate to URL '%s': %w", url, err)
	}

	runContext.Logger.Info("Successfully navigated to URL", "url", url)
	return nil
}

// ScreenshotAction implements taking screenshots
type ScreenshotAction struct{}

func (a *ScreenshotAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	path, ok := actionConfig["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("playwright:screenshot action requires a 'path' string in config")
	}

	runContext.Logger.Info("Executing playwright:screenshot", "path", path)

	options := playwright.PageScreenshotOptions{
		Path: &path,
	}

	if fullPage, ok := actionConfig["full_page"].(bool); ok {
		options.FullPage = &fullPage
	}

	if quality, ok := actionConfig["quality"].(float64); ok {
		qualityInt := int(quality)
		options.Quality = &qualityInt
	}

	if typeStr, ok := actionConfig["type"].(string); ok {
		switch typeStr {
		case "png":
			screenshotType := playwright.ScreenshotTypePng
			options.Type = &screenshotType
		case "jpeg":
			screenshotType := playwright.ScreenshotTypeJpeg
			options.Type = &screenshotType
		}
	}

	if clip, ok := actionConfig["clip"].(map[string]interface{}); ok {
		if x, xOk := clip["x"].(float64); xOk {
			if y, yOk := clip["y"].(float64); yOk {
				if width, widthOk := clip["width"].(float64); widthOk {
					if height, heightOk := clip["height"].(float64); heightOk {
						options.Clip = &playwright.FloatRect{
							X:      float64(x),
							Y:      float64(y),
							Width:  float64(width),
							Height: float64(height),
						}
					}
				}
			}
		}
	}

	_, err := runContext.PlaywrightPage.Screenshot(options)
	if err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}

	runContext.Logger.Info("Successfully took screenshot", "path", path)
	return nil
}

// SelectAction implements selecting options from dropdowns
type SelectAction struct{}

func (a *SelectAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return fmt.Errorf("playwright:select action requires a 'selector' string in config")
	}

	runContext.Logger.Info("Executing playwright:select", "selector", selector)

	options := playwright.LocatorSelectOptionOptions{}

	if force, ok := actionConfig["force"].(bool); ok {
		options.Force = &force
	}

	if noWaitAfter, ok := actionConfig["no_wait_after"].(bool); ok {
		options.NoWaitAfter = &noWaitAfter
	}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	// Handle different ways to specify options
	var selectOptions []playwright.SelectOption

	if values, ok := actionConfig["values"].([]interface{}); ok {
		for _, value := range values {
			if valueStr, ok := value.(string); ok {
				selectOptions = append(selectOptions, playwright.SelectOption{Value: &valueStr})
			}
		}
	} else if labels, ok := actionConfig["labels"].([]interface{}); ok {
		for _, label := range labels {
			if labelStr, ok := label.(string); ok {
				selectOptions = append(selectOptions, playwright.SelectOption{Label: &labelStr})
			}
		}
	} else if indexes, ok := actionConfig["indexes"].([]interface{}); ok {
		for _, index := range indexes {
			if indexFloat, ok := index.(float64); ok {
				indexInt := int(indexFloat)
				selectOptions = append(selectOptions, playwright.SelectOption{Index: &indexInt})
			}
		}
	} else if value, ok := actionConfig["value"].(string); ok {
		selectOptions = append(selectOptions, playwright.SelectOption{Value: &value})
	} else if label, ok := actionConfig["label"].(string); ok {
		selectOptions = append(selectOptions, playwright.SelectOption{Label: &label})
	} else if index, ok := actionConfig["index"].(float64); ok {
		indexInt := int(index)
		selectOptions = append(selectOptions, playwright.SelectOption{Index: &indexInt})
	} else {
		return fmt.Errorf("playwright:select action requires one of: 'values', 'labels', 'indexes', 'value', 'label', or 'index' in config")
	}

	_, err := runContext.PlaywrightPage.Locator(selector).SelectOption(selectOptions, options)
	if err != nil {
		return fmt.Errorf("failed to select options for element with selector '%s': %w", selector, err)
	}

	runContext.Logger.Info("Successfully selected options", "selector", selector, "options_count", len(selectOptions))
	return nil
}

// CheckAction implements checking checkboxes
type CheckAction struct{}

func (a *CheckAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return fmt.Errorf("playwright:check action requires a 'selector' string in config")
	}

	runContext.Logger.Info("Executing playwright:check", "selector", selector)

	options := playwright.LocatorCheckOptions{}

	if force, ok := actionConfig["force"].(bool); ok {
		options.Force = &force
	}

	if noWaitAfter, ok := actionConfig["no_wait_after"].(bool); ok {
		options.NoWaitAfter = &noWaitAfter
	}

	if position, ok := actionConfig["position"].(map[string]interface{}); ok {
		if x, xOk := position["x"].(float64); xOk {
			if y, yOk := position["y"].(float64); yOk {
				options.Position = &playwright.Position{
					X: float64(x),
					Y: float64(y),
				}
			}
		}
	}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	if trial, ok := actionConfig["trial"].(bool); ok {
		options.Trial = &trial
	}

	err := runContext.PlaywrightPage.Locator(selector).Check(options)
	if err != nil {
		return fmt.Errorf("failed to check element with selector '%s': %w", selector, err)
	}

	runContext.Logger.Info("Successfully checked element", "selector", selector)
	return nil
}

// UncheckAction implements unchecking checkboxes
type UncheckAction struct{}

func (a *UncheckAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return fmt.Errorf("playwright:uncheck action requires a 'selector' string in config")
	}

	runContext.Logger.Info("Executing playwright:uncheck", "selector", selector)

	options := playwright.LocatorUncheckOptions{}

	if force, ok := actionConfig["force"].(bool); ok {
		options.Force = &force
	}

	if noWaitAfter, ok := actionConfig["no_wait_after"].(bool); ok {
		options.NoWaitAfter = &noWaitAfter
	}

	if position, ok := actionConfig["position"].(map[string]interface{}); ok {
		if x, xOk := position["x"].(float64); xOk {
			if y, yOk := position["y"].(float64); yOk {
				options.Position = &playwright.Position{
					X: float64(x),
					Y: float64(y),
				}
			}
		}
	}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	if trial, ok := actionConfig["trial"].(bool); ok {
		options.Trial = &trial
	}

	err := runContext.PlaywrightPage.Locator(selector).Uncheck(options)
	if err != nil {
		return fmt.Errorf("failed to uncheck element with selector '%s': %w", selector, err)
	}

	runContext.Logger.Info("Successfully unchecked element", "selector", selector)
	return nil
}

// HoverAction implements hovering over elements
type HoverAction struct{}

func (a *HoverAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return fmt.Errorf("playwright:hover action requires a 'selector' string in config")
	}

	runContext.Logger.Info("Executing playwright:hover", "selector", selector)

	options := playwright.LocatorHoverOptions{}

	if force, ok := actionConfig["force"].(bool); ok {
		options.Force = &force
	}

	if modifiers, ok := actionConfig["modifiers"].([]interface{}); ok {
		var keyboardModifiers []playwright.KeyboardModifier
		for _, mod := range modifiers {
			if modStr, ok := mod.(string); ok {
				switch modStr {
				case "Alt":
					keyboardModifiers = append(keyboardModifiers, playwright.KeyboardModifierAlt)
				case "Control":
					keyboardModifiers = append(keyboardModifiers, playwright.KeyboardModifierControl)
				case "Meta":
					keyboardModifiers = append(keyboardModifiers, playwright.KeyboardModifierMeta)
				case "Shift":
					keyboardModifiers = append(keyboardModifiers, playwright.KeyboardModifierShift)
				}
			}
		}
		if len(keyboardModifiers) > 0 {
			options.Modifiers = keyboardModifiers
		}
	}

	if noWaitAfter, ok := actionConfig["no_wait_after"].(bool); ok {
		options.NoWaitAfter = &noWaitAfter
	}

	if position, ok := actionConfig["position"].(map[string]interface{}); ok {
		if x, xOk := position["x"].(float64); xOk {
			if y, yOk := position["y"].(float64); yOk {
				options.Position = &playwright.Position{
					X: float64(x),
					Y: float64(y),
				}
			}
		}
	}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	if trial, ok := actionConfig["trial"].(bool); ok {
		options.Trial = &trial
	}

	err := runContext.PlaywrightPage.Locator(selector).Hover(options)
	if err != nil {
		return fmt.Errorf("failed to hover over element with selector '%s': %w", selector, err)
	}

	runContext.Logger.Info("Successfully hovered over element", "selector", selector)
	return nil
}

// FocusAction implements focusing on elements
type FocusAction struct{}

func (a *FocusAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return fmt.Errorf("playwright:focus action requires a 'selector' string in config")
	}

	runContext.Logger.Info("Executing playwright:focus", "selector", selector)

	options := playwright.LocatorFocusOptions{}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	err := runContext.PlaywrightPage.Locator(selector).Focus(options)
	if err != nil {
		return fmt.Errorf("failed to focus on element with selector '%s': %w", selector, err)
	}

	runContext.Logger.Info("Successfully focused on element", "selector", selector)
	return nil
}

// BlurAction implements removing focus from elements
type BlurAction struct{}

func (a *BlurAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return fmt.Errorf("playwright:blur action requires a 'selector' string in config")
	}

	runContext.Logger.Info("Executing playwright:blur", "selector", selector)

	options := playwright.LocatorBlurOptions{}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	err := runContext.PlaywrightPage.Locator(selector).Blur(options)
	if err != nil {
		return fmt.Errorf("failed to blur element with selector '%s': %w", selector, err)
	}

	runContext.Logger.Info("Successfully blurred element", "selector", selector)
	return nil
}

// ClearAction implements clearing input fields
type ClearAction struct{}

func (a *ClearAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return fmt.Errorf("playwright:clear action requires a 'selector' string in config")
	}

	runContext.Logger.Info("Executing playwright:clear", "selector", selector)

	options := playwright.LocatorClearOptions{}

	if force, ok := actionConfig["force"].(bool); ok {
		options.Force = &force
	}

	if noWaitAfter, ok := actionConfig["no_wait_after"].(bool); ok {
		options.NoWaitAfter = &noWaitAfter
	}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	err := runContext.PlaywrightPage.Locator(selector).Clear(options)
	if err != nil {
		return fmt.Errorf("failed to clear element with selector '%s': %w", selector, err)
	}

	runContext.Logger.Info("Successfully cleared element", "selector", selector)
	return nil
}

// PressAction implements pressing keys
type PressAction struct{}

func (a *PressAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return fmt.Errorf("playwright:press action requires a 'selector' string in config")
	}

	key, ok := actionConfig["key"].(string)
	if !ok || key == "" {
		return fmt.Errorf("playwright:press action requires a 'key' string in config")
	}

	runContext.Logger.Info("Executing playwright:press", "selector", selector, "key", key)

	options := playwright.LocatorPressOptions{}

	if delay, ok := actionConfig["delay"].(float64); ok {
		delayMs := float64(delay)
		options.Delay = &delayMs
	}

	if noWaitAfter, ok := actionConfig["no_wait_after"].(bool); ok {
		options.NoWaitAfter = &noWaitAfter
	}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	err := runContext.PlaywrightPage.Locator(selector).Press(key, options)
	if err != nil {
		return fmt.Errorf("failed to press key '%s' on element with selector '%s': %w", key, selector, err)
	}

	runContext.Logger.Info("Successfully pressed key", "selector", selector, "key", key)
	return nil
}

// DragAndDropAction implements drag and drop operations
type DragAndDropAction struct{}

func (a *DragAndDropAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	source, ok := actionConfig["source"].(string)
	if !ok || source == "" {
		return fmt.Errorf("playwright:drag_and_drop action requires a 'source' string in config")
	}

	target, ok := actionConfig["target"].(string)
	if !ok || target == "" {
		return fmt.Errorf("playwright:drag_and_drop action requires a 'target' string in config")
	}

	runContext.Logger.Info("Executing playwright:drag_and_drop", "source", source, "target", target)

	options := playwright.LocatorDragToOptions{}

	if force, ok := actionConfig["force"].(bool); ok {
		options.Force = &force
	}

	if noWaitAfter, ok := actionConfig["no_wait_after"].(bool); ok {
		options.NoWaitAfter = &noWaitAfter
	}

	if sourcePosition, ok := actionConfig["source_position"].(map[string]interface{}); ok {
		if x, xOk := sourcePosition["x"].(float64); xOk {
			if y, yOk := sourcePosition["y"].(float64); yOk {
				options.SourcePosition = &playwright.Position{
					X: float64(x),
					Y: float64(y),
				}
			}
		}
	}

	if targetPosition, ok := actionConfig["target_position"].(map[string]interface{}); ok {
		if x, xOk := targetPosition["x"].(float64); xOk {
			if y, yOk := targetPosition["y"].(float64); yOk {
				options.TargetPosition = &playwright.Position{
					X: float64(x),
					Y: float64(y),
				}
			}
		}
	}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	if trial, ok := actionConfig["trial"].(bool); ok {
		options.Trial = &trial
	}

	err := runContext.PlaywrightPage.Locator(source).DragTo(runContext.PlaywrightPage.Locator(target), options)
	if err != nil {
		return fmt.Errorf("failed to drag from '%s' to '%s': %w", source, target, err)
	}

	runContext.Logger.Info("Successfully performed drag and drop", "source", source, "target", target)
	return nil
}

// ScrollAction implements scrolling operations
type ScrollAction struct{}

func (a *ScrollAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	runContext.Logger.Info("Executing playwright:scroll")

	scrollType, ok := actionConfig["scroll_type"].(string)
	if !ok || scrollType == "" {
		scrollType = "element" // default to element scrolling
	}

	switch scrollType {
	case "element":
		return a.scrollElement(actionConfig, runContext)
	case "page":
		return a.scrollPage(actionConfig, runContext)
	default:
		return fmt.Errorf("unsupported scroll_type: %s", scrollType)
	}
}

func (a *ScrollAction) scrollElement(actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return fmt.Errorf("playwright:scroll with scroll_type 'element' requires a 'selector' string in config")
	}

	options := playwright.LocatorScrollIntoViewIfNeededOptions{}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	err := runContext.PlaywrightPage.Locator(selector).ScrollIntoViewIfNeeded(options)
	if err != nil {
		return fmt.Errorf("failed to scroll element with selector '%s' into view: %w", selector, err)
	}

	runContext.Logger.Info("Successfully scrolled element into view", "selector", selector)
	return nil
}

func (a *ScrollAction) scrollPage(actionConfig map[string]interface{}, runContext *RunContext) error {
	x, xOk := actionConfig["x"].(float64)
	y, yOk := actionConfig["y"].(float64)

	if !xOk && !yOk {
		return fmt.Errorf("playwright:scroll with scroll_type 'page' requires 'x' and/or 'y' coordinates")
	}

	if !xOk {
		x = 0
	}
	if !yOk {
		y = 0
	}

	// Use evaluate to scroll the page
	script := fmt.Sprintf("window.scrollTo(%f, %f)", x, y)
	_, err := runContext.PlaywrightPage.Evaluate(script)
	if err != nil {
		return fmt.Errorf("failed to scroll page to coordinates (%f, %f): %w", x, y, err)
	}

	runContext.Logger.Info("Successfully scrolled page", "x", x, "y", y)
	return nil
}

// GetTextAction implements getting text content from elements
type GetTextAction struct{}

func (a *GetTextAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return fmt.Errorf("playwright:get_text action requires a 'selector' string in config")
	}

	variableName, ok := actionConfig["variable_name"].(string)
	if !ok || variableName == "" {
		return fmt.Errorf("playwright:get_text action requires a 'variable_name' string in config")
	}

	runContext.Logger.Info("Executing playwright:get_text", "selector", selector, "variable_name", variableName)

	options := playwright.LocatorTextContentOptions{}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	text, err := runContext.PlaywrightPage.Locator(selector).TextContent(options)
	if err != nil {
		return fmt.Errorf("failed to get text content from element with selector '%s': %w", selector, err)
	}

	// Store the text in the variable
	if text != nil {
		runContext.Variables[variableName] = *text
		runContext.Logger.Info("Successfully got text content", "selector", selector, "variable_name", variableName, "text", *text)
	} else {
		runContext.Variables[variableName] = ""
		runContext.Logger.Info("Successfully got text content (empty)", "selector", selector, "variable_name", variableName)
	}

	return nil
}

// GetAttributeAction implements getting attribute values from elements
type GetAttributeAction struct{}

func (a *GetAttributeAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return fmt.Errorf("playwright:get_attribute action requires a 'selector' string in config")
	}

	attributeName, ok := actionConfig["attribute_name"].(string)
	if !ok || attributeName == "" {
		return fmt.Errorf("playwright:get_attribute action requires an 'attribute_name' string in config")
	}

	variableName, ok := actionConfig["variable_name"].(string)
	if !ok || variableName == "" {
		return fmt.Errorf("playwright:get_attribute action requires a 'variable_name' string in config")
	}

	runContext.Logger.Info("Executing playwright:get_attribute", "selector", selector, "attribute_name", attributeName, "variable_name", variableName)

	options := playwright.LocatorGetAttributeOptions{}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	attribute, err := runContext.PlaywrightPage.Locator(selector).GetAttribute(attributeName, options)
	if err != nil {
		return fmt.Errorf("failed to get attribute '%s' from element with selector '%s': %w", attributeName, selector, err)
	}

	// Store the attribute value in the variable
	if attribute != nil {
		runContext.Variables[variableName] = *attribute
		runContext.Logger.Info("Successfully got attribute", "selector", selector, "attribute_name", attributeName, "variable_name", variableName, "value", *attribute)
	} else {
		runContext.Variables[variableName] = ""
		runContext.Logger.Info("Successfully got attribute (null)", "selector", selector, "attribute_name", attributeName, "variable_name", variableName)
	}

	return nil
}

// EvaluateAction implements executing JavaScript in the page context
type EvaluateAction struct{}

func (a *EvaluateAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	script, ok := actionConfig["script"].(string)
	if !ok || script == "" {
		return fmt.Errorf("playwright:evaluate action requires a 'script' string in config")
	}

	runContext.Logger.Info("Executing playwright:evaluate", "script", script)

	// Handle optional variable name for storing result
	variableName, _ := actionConfig["variable_name"].(string)

	// Handle optional arguments
	var args []interface{}
	if argsInterface, ok := actionConfig["args"].([]interface{}); ok {
		args = argsInterface
	}

	result, err := runContext.PlaywrightPage.Evaluate(script, args...)
	if err != nil {
		return fmt.Errorf("failed to evaluate script: %w", err)
	}

	// Store result in variable if specified
	if variableName != "" {
		runContext.Variables[variableName] = result
		runContext.Logger.Info("Successfully evaluated script and stored result", "variable_name", variableName, "result", result)
	} else {
		runContext.Logger.Info("Successfully evaluated script", "result", result)
	}

	return nil
}

// SetViewportAction implements setting the viewport size
type SetViewportAction struct{}

func (a *SetViewportAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	width, widthOk := actionConfig["width"].(float64)
	height, heightOk := actionConfig["height"].(float64)

	if !widthOk || !heightOk {
		return fmt.Errorf("playwright:set_viewport action requires 'width' and 'height' numbers in config")
	}

	runContext.Logger.Info("Executing playwright:set_viewport", "width", width, "height", height)

	err := runContext.PlaywrightPage.SetViewportSize(int(width), int(height))
	if err != nil {
		return fmt.Errorf("failed to set viewport size to %dx%d: %w", int(width), int(height), err)
	}

	runContext.Logger.Info("Successfully set viewport size", "width", int(width), "height", int(height))
	return nil
}

// ReloadAction implements reloading the page
type ReloadAction struct{}

func (a *ReloadAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	runContext.Logger.Info("Executing playwright:reload")

	options := playwright.PageReloadOptions{}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	if waitUntil, ok := actionConfig["wait_until"].(string); ok {
		switch waitUntil {
		case "load":
			waitUntilState := playwright.WaitUntilStateLoad
			options.WaitUntil = &waitUntilState
		case "domcontentloaded":
			waitUntilState := playwright.WaitUntilStateDomcontentloaded
			options.WaitUntil = &waitUntilState
		case "networkidle":
			waitUntilState := playwright.WaitUntilStateNetworkidle
			options.WaitUntil = &waitUntilState
		case "commit":
			waitUntilState := playwright.WaitUntilStateCommit
			options.WaitUntil = &waitUntilState
		}
	}

	_, err := runContext.PlaywrightPage.Reload(options)
	if err != nil {
		return fmt.Errorf("failed to reload page: %w", err)
	}

	runContext.Logger.Info("Successfully reloaded page")
	return nil
}

// GoBackAction implements navigating back in browser history
type GoBackAction struct{}

func (a *GoBackAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	runContext.Logger.Info("Executing playwright:go_back")

	options := playwright.PageGoBackOptions{}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	if waitUntil, ok := actionConfig["wait_until"].(string); ok {
		switch waitUntil {
		case "load":
			waitUntilState := playwright.WaitUntilStateLoad
			options.WaitUntil = &waitUntilState
		case "domcontentloaded":
			waitUntilState := playwright.WaitUntilStateDomcontentloaded
			options.WaitUntil = &waitUntilState
		case "networkidle":
			waitUntilState := playwright.WaitUntilStateNetworkidle
			options.WaitUntil = &waitUntilState
		case "commit":
			waitUntilState := playwright.WaitUntilStateCommit
			options.WaitUntil = &waitUntilState
		}
	}

	_, err := runContext.PlaywrightPage.GoBack(options)
	return err
}

// GoForwardAction implements navigating forward in browser history
type GoForwardAction struct{}

func (a *GoForwardAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
	runContext.Logger.Info("Executing playwright:go_forward")

	options := playwright.PageGoForwardOptions{}

	if timeout, ok := actionConfig["timeout"].(float64); ok {
		timeoutMs := float64(timeout)
		options.Timeout = &timeoutMs
	}

	if waitUntil, ok := actionConfig["wait_until"].(string); ok {
		switch waitUntil {
		case "load":
			waitUntilState := playwright.WaitUntilStateLoad
			options.WaitUntil = &waitUntilState
		case "domcontentloaded":
			waitUntilState := playwright.WaitUntilStateDomcontentloaded
			options.WaitUntil = &waitUntilState
		case "networkidle":
			waitUntilState := playwright.WaitUntilStateNetworkidle
			options.WaitUntil = &waitUntilState
		case "commit":
			waitUntilState := playwright.WaitUntilStateCommit
			options.WaitUntil = &waitUntilState
		}
	}

	_, err := runContext.PlaywrightPage.GoForward(options)
	return err
}

// IfElseAction implements conditional logic with multiple else-if blocks
type IfElseAction struct{}

func (a *IfElseAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
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

	// Execute final_actions regardless of condition outcomes
	if finalActions, ok := actionConfig["final_actions"].([]interface{}); ok {
		runContext.Logger.Info("Executing final_actions", "count", len(finalActions))
		return a.executeNestedActions(ctx, finalActions, runContext)
	}

	runContext.Logger.Info("No conditions met and no else actions defined")
	return nil
}

func (a *IfElseAction) evaluateCondition(runContext *RunContext, selector, conditionType string) (bool, error) {
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

func (a *IfElseAction) executeNestedActions(ctx context.Context, actions []interface{}, runContext *RunContext) error {
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

		// Resolve variables in nested action config
		resolvedActionConfig, err := resolveVariablesInConfig(actionConfig, runContext.VariableContext, runContext.AutomationConfig)
		if err != nil {
			runContext.Logger.Error("Failed to resolve variables in nested action config", "action_type", actionType, "error", err)
			return fmt.Errorf("failed to resolve variables in nested action '%s': %w", actionType, err)
		}

		runContext.Logger.Info("Executing nested action", "index", i, "action_type", actionType)

		// Get the plugin action
		pluginAction, err := GetAction(actionType)
		if err != nil {
			runContext.Logger.Error("Failed to get nested action", "action_type", actionType, "error", err)
			return fmt.Errorf("failed to get nested action '%s': %w", actionType, err)
		}

		// Execute the nested action
		err = pluginAction.Execute(ctx, resolvedActionConfig, runContext)
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

func (a *LogAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
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

func (a *LoopUntilAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *RunContext) error {
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

		// Update local loop index in variable context
		runContext.VariableContext.LocalLoopIndex = loopCount

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

			// Resolve variables in loop action config
			resolvedActionConfig, err := resolveVariablesInConfig(actionConfig, runContext.VariableContext, runContext.AutomationConfig)
			if err != nil {
				runContext.Logger.Error("Failed to resolve variables in loop action config", "action_type", actionType, "error", err)
				return fmt.Errorf("failed to resolve variables in loop action '%s': %w", actionType, err)
			}

			runContext.Logger.Info("Executing loop action", "loop_count", loopCount, "action_index", actionIndex, "action_type", actionType)

			// Get the plugin action
			pluginAction, err := GetAction(actionType)
			if err != nil {
				runContext.Logger.Error("Failed to get loop action", "action_type", actionType, "error", err)
				return fmt.Errorf("failed to get loop action '%s': %w", actionType, err)
			}

			// Execute the loop action
			err = pluginAction.Execute(ctx, resolvedActionConfig, runContext)
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

func (a *LoopUntilAction) evaluateCondition(runContext *RunContext, selector, conditionType string) (bool, error) {
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

// Helper function to resolve variables in config (moved from runner.go for reuse)
func resolveVariablesInConfig(config map[string]interface{}, varContext *VariableContext, automationConfig *ExportedAutomationMeta) (map[string]interface{}, error) {
	resolved := make(map[string]interface{})

	for key, value := range config {
		switch v := value.(type) {
		case string:
			resolvedValue, err := resolveVariablesInString(v, varContext, automationConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve variables in field '%s': %w", key, err)
			}
			resolved[key] = resolvedValue
		case map[string]interface{}:
			// Recursively resolve nested objects
			nestedResolved, err := resolveVariablesInConfig(v, varContext, automationConfig)
			if err != nil {
				return nil, err
			}
			resolved[key] = nestedResolved
		default:
			// For non-string values, keep as-is
			resolved[key] = value
		}
	}

	return resolved, nil
}

// Helper function to resolve variables in string (moved from runner.go for reuse)
func resolveVariablesInString(input string, varContext *VariableContext, automationConfig *ExportedAutomationMeta) (string, error) {
	// Pattern to match {{variableName}} or {{faker.method}}
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	result := re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract variable name (remove {{ and }})
		varName := strings.Trim(match, "{}")

		// Handle environment variables
		switch varName {
		case "loopIndex":
			return strconv.Itoa(varContext.LoopIndex)
		case "localLoopIndex":
			return strconv.Itoa(varContext.LocalLoopIndex)
		case "timestamp":
			return varContext.Timestamp
		case "runId":
			return varContext.RunID
		case "userId":
			return varContext.UserID
		case "projectId":
			return varContext.ProjectID
		case "automationId":
			return varContext.AutomationID
		}

		// Handle faker variables
		if strings.HasPrefix(varName, "faker.") {
			fakerMethod := strings.TrimPrefix(varName, "faker.")
			return generateFakerValue(fakerMethod)
		}

		// Handle static variables
		if value, exists := varContext.StaticVars[varName]; exists {
			return value
		}

		// Handle dynamic variables from config
		for _, variable := range automationConfig.Variables {
			if variable.Key == varName {
				switch variable.Type {
				case "static":
					return variable.Value
				case "dynamic":
					// Variable.Value contains the faker method (e.g., "{{faker.email}}")
					if strings.HasPrefix(variable.Value, "{{faker.") && strings.HasSuffix(variable.Value, "}}") {
						fakerMethod := strings.TrimPrefix(strings.TrimSuffix(variable.Value, "}}"), "{{faker.")
						return generateFakerValue(fakerMethod)
					}
					return variable.Value
				case "environment":
					// Variable.Value contains the environment variable (e.g., "{{timestamp}}")
					v, err := resolveVariablesInString(variable.Value, varContext, automationConfig)
					if err != nil {
						return ""
					}
					return v
				}
			}
		}

		// If no match found, return the original placeholder
		return match
	})

	return result, nil
}

// generateFakerValue generates a fake value based on the faker method
func generateFakerValue(method string) string {
	gofakeit.Seed(time.Now().UnixNano()) // Ensure randomness

	switch method {
	case "name":
		return gofakeit.Name()
	case "email":
		return gofakeit.Email()
	case "phone":
		return gofakeit.Phone()
	case "address":
		return gofakeit.Address().Address
	case "company":
		return gofakeit.Company()
	case "username":
		return gofakeit.Username()
	case "password":
		return gofakeit.Password(true, true, true, true, false, 12)
	case "uuid":
		return gofakeit.UUID()
	case "number":
		return strconv.Itoa(gofakeit.Number(1, 1000))
	case "date":
		return gofakeit.Date().Format("2006-01-02")
	case "lastName":
		return gofakeit.LastName()
	case "firstName":
		return gofakeit.FirstName()
	default:
		return fmt.Sprintf("{{faker.%s}}", method)
	}
}