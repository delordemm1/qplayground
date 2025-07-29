package playwright

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/delordemm1/qplayground-cli/internal/automation"
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
	automation.RegisterAction("playwright:hover", func() automation.PluginAction { return &HoverAction{} })
	automation.RegisterAction("playwright:scroll", func() automation.PluginAction { return &ScrollAction{} })
	automation.RegisterAction("playwright:get_text", func() automation.PluginAction { return &GetTextAction{} })
	automation.RegisterAction("playwright:get_attribute", func() automation.PluginAction { return &GetAttributeAction{} })
	automation.RegisterAction("playwright:set_viewport", func() automation.PluginAction { return &SetViewportAction{} })
	automation.RegisterAction("playwright:reload", func() automation.PluginAction { return &ReloadAction{} })
	automation.RegisterAction("playwright:go_back", func() automation.PluginAction { return &GoBackAction{} })
	automation.RegisterAction("playwright:go_forward", func() automation.PluginAction { return &GoForwardAction{} })
	automation.RegisterAction("playwright:wait_for_selector", func() automation.PluginAction { return &WaitForSelectorAction{} })
	automation.RegisterAction("playwright:wait_for_timeout", func() automation.PluginAction { return &WaitForTimeoutAction{} })
	automation.RegisterAction("playwright:wait_for_load_state", func() automation.PluginAction { return &WaitForLoadStateAction{} })
	automation.RegisterAction("playwright:screenshot", func() automation.PluginAction { return &ScreenshotAction{} })
	automation.RegisterAction("playwright:evaluate", func() automation.PluginAction { return &EvaluateAction{} })
	automation.RegisterAction("playwright:log", func() automation.PluginAction { return &LogAction{} })
	automation.RegisterAction("playwright:if_else", func() automation.PluginAction { return &IfElseAction{} })
	automation.RegisterAction("playwright:loop_until", func() automation.PluginAction { return &LoopUntilAction{} })
}

// Helper function to send success event
func sendSuccessEvent(runContext *automation.RunContext, actionType, message string, duration time.Duration) {
	if runContext.EventCh != nil {
		select {
		case runContext.EventCh <- automation.RunEvent{
			Type:           automation.RunEventTypeLog,
			Timestamp:      time.Now(),
			StepName:       runContext.StepName,
			StepID:         runContext.StepID,
			ActionID:       runContext.ActionID,
			ActionName:     runContext.ActionName,
			ParentActionID: runContext.ParentActionID,
			ActionType:     actionType,
			Message:        message,
			Duration:       duration.Milliseconds(),
			LoopIndex:      runContext.LoopIndex,
			LocalLoopIndex: runContext.VariableContext.LocalLoopIndex,
		}:
		default:
		}
	}
}

// Helper function to send error event
func sendErrorEvent(runContext *automation.RunContext, actionType, errorMsg string, duration time.Duration) {
	if runContext.EventCh != nil {
		select {
		case runContext.EventCh <- automation.RunEvent{
			Type:           automation.RunEventTypeError,
			Timestamp:      time.Now(),
			StepName:       runContext.StepName,
			StepID:         runContext.StepID,
			ActionID:       runContext.ActionID,
			ActionName:     runContext.ActionName,
			ParentActionID: runContext.ParentActionID,
			ActionType:     actionType,
			Error:          errorMsg,
			Duration:       duration.Milliseconds(),
			LoopIndex:      runContext.LoopIndex,
			LocalLoopIndex: runContext.VariableContext.LocalLoopIndex,
		}:
		default:
		}
	}
}

// GotoAction implements page navigation
type GotoAction struct{}

func (a *GotoAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	startTime := time.Now()
	
	url, ok := actionConfig["url"].(string)
	if !ok || url == "" {
		return fmt.Errorf("playwright:goto action requires a 'url' string in config")
	}

	// Resolve variables in URL
	resolvedURL, err := runContext.Runner.ResolveVariablesInString(url, runContext.VariableContext, runContext.AutomationConfig)
	if err != nil {
		return fmt.Errorf("failed to resolve variables in URL: %w", err)
	}

	runContext.Logger.Info("Navigating to URL", "url", resolvedURL)

	// Parse optional parameters
	timeout := 30000 // default 30 seconds
	if t, ok := actionConfig["timeout"].(float64); ok {
		timeout = int(t)
	}

	waitUntil := "load"
	if wu, ok := actionConfig["wait_until"].(string); ok {
		waitUntil = wu
	}

	// Navigate to URL
	_, err = runContext.PlaywrightPage.Goto(resolvedURL, playwright.PageGotoOptions{
		Timeout:   playwright.Float(float64(timeout)),
		WaitUntil: playwright.WaitUntilState(waitUntil),
	})

	duration := time.Since(startTime)
	if err != nil {
		sendErrorEvent(runContext, "playwright:goto", fmt.Sprintf("failed to navigate to %s: %v", resolvedURL, err), duration)
		return fmt.Errorf("failed to navigate to %s: %w", resolvedURL, err)
	}

	sendSuccessEvent(runContext, "playwright:goto", fmt.Sprintf("Successfully navigated to %s", resolvedURL), duration)
	return nil
}

func (a *GotoAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:goto cannot be used as a condition")
}

// ClickAction implements element clicking
type ClickAction struct{}

func (a *ClickAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	startTime := time.Now()
	
	selector, ok := actionConfig["selector"].(string)
	if !ok || selector == "" {
		return fmt.Errorf("playwright:click action requires a 'selector' string in config")
	}

	// Resolve variables in selector
	resolvedSelector, err := runContext.Runner.ResolveVariablesInString(selector, runContext.VariableContext, runContext.AutomationConfig)
	if err != nil {
		return fmt.Errorf("failed to resolve variables in selector: %w", err)
	}

	runContext.Logger.Info("Clicking element", "selector", resolvedSelector)

	// Parse optional parameters
	button := "left"
	if b, ok := actionConfig["button"].(string); ok {
		button = b
	}

	clickCount := 1
	if cc, ok := actionConfig["click_count"].(float64); ok {
		clickCount = int(cc)
	}

	force := false
	if f, ok := actionConfig["force"].(bool); ok {
		force = f
	}

	// Click element
	err = runContext.PlaywrightPage.Click(resolvedSelector, playwright.PageClickOptions{
		Button:     playwright.MouseButton(button),
		ClickCount: playwright.Int(clickCount),
		Force:      playwright.Bool(force),
	})

	duration := time.Since(startTime)
	if err != nil {
		sendErrorEvent(runContext, "playwright:click", fmt.Sprintf("failed to click %s: %v", resolvedSelector, err), duration)
		return fmt.Errorf("failed to click %s: %w", resolvedSelector, err)
	}

	sendSuccessEvent(runContext, "playwright:click", fmt.Sprintf("Successfully clicked %s", resolvedSelector), duration)
	return nil
}

func (a *ClickAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:click cannot be used as a condition")
}

// ScreenshotAction implements taking screenshots
type ScreenshotAction struct{}

func (a *ScreenshotAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	startTime := time.Now()
	
	runContext.Logger.Info("Taking screenshot")

	// Parse configuration
	fullPage := true
	if fp, ok := actionConfig["full_page"].(bool); ok {
		fullPage = fp
	}

	format := "png"
	if f, ok := actionConfig["format"].(string); ok {
		format = f
	}

	quality := 80
	if q, ok := actionConfig["quality"].(float64); ok {
		quality = int(q)
	}

	uploadToR2 := true
	if utr, ok := actionConfig["upload_to_r2"].(bool); ok {
		uploadToR2 = utr
	}

	r2Key, _ := actionConfig["r2_key"].(string)

	// Resolve variables in r2Key
	if r2Key != "" {
		resolvedKey, err := runContext.Runner.ResolveVariablesInString(r2Key, runContext.VariableContext, runContext.AutomationConfig)
		if err != nil {
			return fmt.Errorf("failed to resolve variables in r2_key: %w", err)
		}
		r2Key = resolvedKey
	}

	// Take screenshot
	screenshotOptions := playwright.PageScreenshotOptions{
		FullPage: playwright.Bool(fullPage),
		Type:     playwright.ScreenshotType(format),
	}

	if format == "jpeg" && quality > 0 {
		screenshotOptions.Quality = playwright.Int(quality)
	}

	screenshotBytes, err := runContext.PlaywrightPage.Screenshot(screenshotOptions)
	duration := time.Since(startTime)

	if err != nil {
		sendErrorEvent(runContext, "playwright:screenshot", fmt.Sprintf("failed to take screenshot: %v", err), duration)
		return fmt.Errorf("failed to take screenshot: %w", err)
	}

	var publicURL string
	if uploadToR2 && r2Key != "" && runContext.StorageService != nil {
		// Upload to storage
		contentType := "image/png"
		if format == "jpeg" {
			contentType = "image/jpeg"
		}

		reader := strings.NewReader(string(screenshotBytes))
		uploadedURL, err := runContext.StorageService.UploadFile(ctx, r2Key, reader, contentType)
		if err != nil {
			slog.Error("Failed to upload screenshot to storage", "key", r2Key, "error", err)
			// Continue without failing the action
		} else {
			publicURL = uploadedURL
			runContext.Logger.Info("Screenshot uploaded to storage", "key", r2Key, "url", publicURL)
		}
	}

	// Add to output files
	if publicURL != "" {
		runContext.LastOutputFiles = append(runContext.LastOutputFiles, publicURL)
	}

	// Send output file event
	if publicURL != "" && runContext.EventCh != nil {
		select {
		case runContext.EventCh <- automation.RunEvent{
			Type:           automation.RunEventTypeOutputFile,
			Timestamp:      time.Now(),
			StepName:       runContext.StepName,
			StepID:         runContext.StepID,
			ActionID:       runContext.ActionID,
			ActionName:     runContext.ActionName,
			ParentActionID: runContext.ParentActionID,
			ActionType:     "playwright:screenshot",
			OutputFile:     publicURL,
			Duration:       duration.Milliseconds(),
			LoopIndex:      runContext.LoopIndex,
			LocalLoopIndex: runContext.VariableContext.LocalLoopIndex,
		}:
		default:
		}
	}

	sendSuccessEvent(runContext, "playwright:screenshot", fmt.Sprintf("Successfully took screenshot (size: %d bytes)", len(screenshotBytes)), duration)
	return nil
}

func (a *ScreenshotAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:screenshot cannot be used as a condition")
}

// Placeholder implementations for other actions
type FillAction struct{}
func (a *FillAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	// Implementation would go here
	return fmt.Errorf("playwright:fill not yet implemented in CLI")
}
func (a *FillAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:fill cannot be used as a condition")
}

type TypeAction struct{}
func (a *TypeAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:type not yet implemented in CLI")
}
func (a *TypeAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:type cannot be used as a condition")
}

type PressAction struct{}
func (a *PressAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:press not yet implemented in CLI")
}
func (a *PressAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:press cannot be used as a condition")
}

type CheckAction struct{}
func (a *CheckAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:check not yet implemented in CLI")
}
func (a *CheckAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:check cannot be used as a condition")
}

type UncheckAction struct{}
func (a *UncheckAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:uncheck not yet implemented in CLI")
}
func (a *UncheckAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:uncheck cannot be used as a condition")
}

type SelectOptionAction struct{}
func (a *SelectOptionAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:select_option not yet implemented in CLI")
}
func (a *SelectOptionAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:select_option cannot be used as a condition")
}

type HoverAction struct{}
func (a *HoverAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:hover not yet implemented in CLI")
}
func (a *HoverAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:hover cannot be used as a condition")
}

type ScrollAction struct{}
func (a *ScrollAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:scroll not yet implemented in CLI")
}
func (a *ScrollAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:scroll cannot be used as a condition")
}

type GetTextAction struct{}
func (a *GetTextAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:get_text not yet implemented in CLI")
}
func (a *GetTextAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:get_text cannot be used as a condition")
}

type GetAttributeAction struct{}
func (a *GetAttributeAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:get_attribute not yet implemented in CLI")
}
func (a *GetAttributeAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:get_attribute cannot be used as a condition")
}

type SetViewportAction struct{}
func (a *SetViewportAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:set_viewport not yet implemented in CLI")
}
func (a *SetViewportAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:set_viewport cannot be used as a condition")
}

type ReloadAction struct{}
func (a *ReloadAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:reload not yet implemented in CLI")
}
func (a *ReloadAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:reload cannot be used as a condition")
}

type GoBackAction struct{}
func (a *GoBackAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:go_back not yet implemented in CLI")
}
func (a *GoBackAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:go_back cannot be used as a condition")
}

type GoForwardAction struct{}
func (a *GoForwardAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:go_forward not yet implemented in CLI")
}
func (a *GoForwardAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:go_forward cannot be used as a condition")
}

type WaitForSelectorAction struct{}
func (a *WaitForSelectorAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:wait_for_selector not yet implemented in CLI")
}
func (a *WaitForSelectorAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:wait_for_selector cannot be used as a condition")
}

type WaitForTimeoutAction struct{}
func (a *WaitForTimeoutAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:wait_for_timeout not yet implemented in CLI")
}
func (a *WaitForTimeoutAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:wait_for_timeout cannot be used as a condition")
}

type WaitForLoadStateAction struct{}
func (a *WaitForLoadStateAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:wait_for_load_state not yet implemented in CLI")
}
func (a *WaitForLoadStateAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:wait_for_load_state cannot be used as a condition")
}

type EvaluateAction struct{}
func (a *EvaluateAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:evaluate not yet implemented in CLI")
}
func (a *EvaluateAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:evaluate cannot be used as a condition")
}

type LogAction struct{}
func (a *LogAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:log not yet implemented in CLI")
}
func (a *LogAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:log cannot be used as a condition")
}

type IfElseAction struct{}
func (a *IfElseAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:if_else not yet implemented in CLI")
}
func (a *IfElseAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:if_else cannot be used as a condition")
}

type LoopUntilAction struct{}
func (a *LoopUntilAction) Execute(ctx context.Context, actionConfig map[string]interface{}, runContext *automation.RunContext) error {
	return fmt.Errorf("playwright:loop_until not yet implemented in CLI")
}
func (a *LoopUntilAction) EvaluateCondition(ctx context.Context, conditionConfig map[string]interface{}, runContext *automation.RunContext) (bool, error) {
	return false, fmt.Errorf("playwright:loop_until cannot be used as a condition")
}