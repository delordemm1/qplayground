// Import all action configuration components
import PlaywrightGotoConfig from "$lib/components/ActionConfigs/PlaywrightGotoConfig.svelte";
import PlaywrightClickConfig from "$lib/components/ActionConfigs/PlaywrightClickConfig.svelte";
import PlaywrightFillConfig from "$lib/components/ActionConfigs/PlaywrightFillConfig.svelte";
import PlaywrightTypeConfig from "$lib/components/ActionConfigs/PlaywrightTypeConfig.svelte";
import PlaywrightPressConfig from "$lib/components/ActionConfigs/PlaywrightPressConfig.svelte";
import PlaywrightCheckConfig from "$lib/components/ActionConfigs/PlaywrightCheckConfig.svelte";
import PlaywrightUncheckConfig from "$lib/components/ActionConfigs/PlaywrightUncheckConfig.svelte";
import PlaywrightSelectOptionConfig from "$lib/components/ActionConfigs/PlaywrightSelectOptionConfig.svelte";
import PlaywrightHoverConfig from "$lib/components/ActionConfigs/PlaywrightHoverConfig.svelte";
import PlaywrightWaitConfig from "$lib/components/ActionConfigs/PlaywrightWaitConfig.svelte";
import PlaywrightGetTextConfig from "$lib/components/ActionConfigs/PlaywrightGetTextConfig.svelte";
import PlaywrightGetAttributeConfig from "$lib/components/ActionConfigs/PlaywrightGetAttributeConfig.svelte";
import PlaywrightScreenshotConfig from "$lib/components/ActionConfigs/PlaywrightScreenshotConfig.svelte";
import PlaywrightReloadConfig from "$lib/components/ActionConfigs/PlaywrightReloadConfig.svelte";
import PlaywrightGoBackConfig from "$lib/components/ActionConfigs/PlaywrightGoBackConfig.svelte";
import PlaywrightGoForwardConfig from "$lib/components/ActionConfigs/PlaywrightGoForwardConfig.svelte";
import PlaywrightScrollConfig from "$lib/components/ActionConfigs/PlaywrightScrollConfig.svelte";
import PlaywrightSetViewportConfig from "$lib/components/ActionConfigs/PlaywrightSetViewportConfig.svelte";
import PlaywrightEvaluateConfig from "$lib/components/ActionConfigs/PlaywrightEvaluateConfig.svelte";
import PlaywrightIfElseConfig from "$lib/components/ActionConfigs/PlaywrightIfElseConfig.svelte";
import PlaywrightLoopUntilConfig from "$lib/components/ActionConfigs/PlaywrightLoopUntilConfig.svelte";
import PlaywrightLogConfig from "$lib/components/ActionConfigs/PlaywrightLogConfig.svelte";
import R2UploadConfig from "$lib/components/ActionConfigs/R2UploadConfig.svelte";
import R2DeleteConfig from "$lib/components/ActionConfigs/R2DeleteConfig.svelte";

// Define all available action types
export const actionTypes = [
  // Navigation actions
  "playwright:goto",
  "playwright:reload",
  "playwright:go_back",
  "playwright:go_forward",
  
  // Interaction actions
  "playwright:click",
  "playwright:fill",
  "playwright:type",
  "playwright:press",
  "playwright:check",
  "playwright:uncheck",
  "playwright:select_option",
  "playwright:hover",
  
  // Waiting actions
  "playwright:wait_for_selector",
  "playwright:wait_for_timeout",
  "playwright:wait_for_load_state",
  
  // Information actions
  "playwright:get_text",
  "playwright:get_attribute",
  "playwright:screenshot",
  
  // Control flow actions
  "playwright:if_else",
  "playwright:loop_until",
  "playwright:log",
  
  // Utility actions
  "playwright:evaluate",
  "playwright:scroll",
  "playwright:set_viewport",
  
  // R2 Storage actions
  "r2:upload",
  "r2:delete",
];

// Define action types that can be used in nested contexts (if_else, loop_until)
export const nestedActionTypes = [
  "playwright:goto",
  "playwright:click",
  "playwright:fill",
  "playwright:type",
  "playwright:press",
  "playwright:check",
  "playwright:uncheck",
  "playwright:select_option",
  "playwright:hover",
  "playwright:wait_for_selector",
  "playwright:wait_for_timeout",
  "playwright:wait_for_load_state",
  "playwright:get_text",
  "playwright:get_attribute",
  "playwright:screenshot",
  "playwright:log",
  "playwright:evaluate",
  "playwright:scroll",
  "playwright:set_viewport",
  "r2:upload",
  "r2:delete",
];

// Map action types to their configuration components
export const actionConfigComponents: Record<string, any> = {
  "playwright:goto": PlaywrightGotoConfig,
  "playwright:click": PlaywrightClickConfig,
  "playwright:fill": PlaywrightFillConfig,
  "playwright:type": PlaywrightTypeConfig,
  "playwright:press": PlaywrightPressConfig,
  "playwright:check": PlaywrightCheckConfig,
  "playwright:uncheck": PlaywrightUncheckConfig,
  "playwright:select_option": PlaywrightSelectOptionConfig,
  "playwright:hover": PlaywrightHoverConfig,
  "playwright:wait_for_selector": PlaywrightWaitConfig,
  "playwright:wait_for_timeout": PlaywrightWaitConfig,
  "playwright:wait_for_load_state": PlaywrightWaitConfig,
  "playwright:get_text": PlaywrightGetTextConfig,
  "playwright:get_attribute": PlaywrightGetAttributeConfig,
  "playwright:screenshot": PlaywrightScreenshotConfig,
  "playwright:reload": PlaywrightReloadConfig,
  "playwright:go_back": PlaywrightGoBackConfig,
  "playwright:go_forward": PlaywrightGoForwardConfig,
  "playwright:scroll": PlaywrightScrollConfig,
  "playwright:set_viewport": PlaywrightSetViewportConfig,
  "playwright:evaluate": PlaywrightEvaluateConfig,
  "playwright:if_else": PlaywrightIfElseConfig,
  "playwright:loop_until": PlaywrightLoopUntilConfig,
  "playwright:log": PlaywrightLogConfig,
  "r2:upload": R2UploadConfig,
  "r2:delete": R2DeleteConfig,
};

// Validation function for action configurations
export function validateActionConfig(actionType: string, config: Record<string, any>): string[] {
  const errors: string[] = [];

  switch (actionType) {
    case "playwright:goto":
      if (!config.url) {
        errors.push("URL is required for goto action.");
      }
      break;

    case "playwright:click":
      if (!config.selector) {
        errors.push("Selector is required for click action.");
      }
      break;

    case "playwright:fill":
      if (!config.selector) {
        errors.push("Selector is required for fill action.");
      }
      if (!config.value) {
        errors.push("Value is required for fill action.");
      }
      break;

    case "playwright:type":
      if (!config.selector) {
        errors.push("Selector is required for type action.");
      }
      if (!config.text) {
        errors.push("Text is required for type action.");
      }
      break;

    case "playwright:press":
      if (!config.selector) {
        errors.push("Selector is required for press action.");
      }
      if (!config.key) {
        errors.push("Key is required for press action.");
      }
      break;

    case "playwright:check":
    case "playwright:uncheck":
      if (!config.selector) {
        errors.push("Selector is required for check/uncheck action.");
      }
      break;

    case "playwright:select_option":
      if (!config.selector) {
        errors.push("Selector is required for select option action.");
      }
      if (config.selection_type === "value" && !config.value) {
        errors.push("Value is required when selection type is 'value'.");
      }
      if (config.selection_type === "label" && !config.label) {
        errors.push("Label is required when selection type is 'label'.");
      }
      if (config.selection_type === "index" && config.index === undefined) {
        errors.push("Index is required when selection type is 'index'.");
      }
      break;

    case "playwright:hover":
      if (!config.selector) {
        errors.push("Selector is required for hover action.");
      }
      break;

    case "playwright:wait_for_selector":
      if (!config.selector) {
        errors.push("Selector is required for wait for selector action.");
      }
      break;

    case "playwright:wait_for_timeout":
      if (!config.timeout_ms) {
        errors.push("Timeout is required for wait for timeout action.");
      }
      break;

    case "playwright:get_text":
    case "playwright:get_attribute":
      if (!config.selector) {
        errors.push("Selector is required for get text/attribute action.");
      }
      if (actionType === "playwright:get_attribute" && !config.attribute) {
        errors.push("Attribute name is required for get attribute action.");
      }
      break;

    case "playwright:screenshot":
      if (config.upload_to_r2 && !config.r2_key) {
        errors.push("R2 key is required when upload to R2 is enabled.");
      }
      break;

    case "playwright:set_viewport":
      if (!config.width || !config.height) {
        errors.push("Width and height are required for set viewport action.");
      }
      break;

    case "playwright:evaluate":
      if (!config.expression) {
        errors.push("JavaScript expression is required for evaluate action.");
      }
      break;

    case "playwright:if_else":
      // Selector is required unless it's a loop index or random condition
      if (!["loop_index_is_even", "loop_index_is_odd", "loop_index_is_prime", "random"].includes(config.condition_type)) {
        if (!config.selector) {
          errors.push("Selector is required for this condition type.");
        }
      }
      if (!config.condition_type) {
        errors.push("Condition type is required for if_else action.");
      }
      break;

    case "playwright:loop_until":
      // At least one force stop condition is required
      if (!config.max_loops && !config.timeout_ms) {
        errors.push("At least one force stop condition (max loops or timeout) is required.");
      }
      break;

    case "playwright:log":
      if (!config.message) {
        errors.push("Log message is required for log action.");
      }
      break;

    case "r2:upload":
      if (!config.key) {
        errors.push("Object key is required for R2 upload action.");
      }
      if (!config.content) {
        errors.push("Content is required for R2 upload action.");
      }
      break;

    case "r2:delete":
      if (!config.key) {
        errors.push("Object key is required for R2 delete action.");
      }
      break;

    default:
      // No specific validation for unknown action types
      break;
  }

  return errors;
}