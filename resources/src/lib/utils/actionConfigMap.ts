// Shared action configuration mapping and types
import PlaywrightGotoConfig from "../components/ActionConfigs/PlaywrightGotoConfig.svelte";
import PlaywrightClickConfig from "../components/ActionConfigs/PlaywrightClickConfig.svelte";
import PlaywrightFillConfig from "../components/ActionConfigs/PlaywrightFillConfig.svelte";
import PlaywrightTypeConfig from "../components/ActionConfigs/PlaywrightTypeConfig.svelte";
import PlaywrightPressConfig from "../components/ActionConfigs/PlaywrightPressConfig.svelte";
import PlaywrightCheckConfig from "../components/ActionConfigs/PlaywrightCheckConfig.svelte";
import PlaywrightUncheckConfig from "../components/ActionConfigs/PlaywrightUncheckConfig.svelte";
import PlaywrightSelectOptionConfig from "../components/ActionConfigs/PlaywrightSelectOptionConfig.svelte";
import PlaywrightHoverConfig from "../components/ActionConfigs/PlaywrightHoverConfig.svelte";
import PlaywrightScrollConfig from "../components/ActionConfigs/PlaywrightScrollConfig.svelte";
import PlaywrightGetTextConfig from "../components/ActionConfigs/PlaywrightGetTextConfig.svelte";
import PlaywrightGetAttributeConfig from "../components/ActionConfigs/PlaywrightGetAttributeConfig.svelte";
import PlaywrightSetViewportConfig from "../components/ActionConfigs/PlaywrightSetViewportConfig.svelte";
import PlaywrightReloadConfig from "../components/ActionConfigs/PlaywrightReloadConfig.svelte";
import PlaywrightGoBackConfig from "../components/ActionConfigs/PlaywrightGoBackConfig.svelte";
import PlaywrightGoForwardConfig from "../components/ActionConfigs/PlaywrightGoForwardConfig.svelte";
import PlaywrightScreenshotConfig from "../components/ActionConfigs/PlaywrightScreenshotConfig.svelte";
import PlaywrightWaitConfig from "../components/ActionConfigs/PlaywrightWaitConfig.svelte";
import PlaywrightEvaluateConfig from "../components/ActionConfigs/PlaywrightEvaluateConfig.svelte";
import R2UploadConfig from "../components/ActionConfigs/R2UploadConfig.svelte";
import R2DeleteConfig from "../components/ActionConfigs/R2DeleteConfig.svelte";
import PlaywrightIfElseConfig from "../components/ActionConfigs/PlaywrightIfElseConfig.svelte";
import PlaywrightLogConfig from "../components/ActionConfigs/PlaywrightLogConfig.svelte";
import PlaywrightLoopUntilConfig from "../components/ActionConfigs/PlaywrightLoopUntilConfig.svelte";

// List of supported action types
export const actionTypes = [
  "playwright:goto",
  "playwright:click",
  "playwright:fill",
  "playwright:type",
  "playwright:press",
  "playwright:check",
  "playwright:uncheck",
  "playwright:select_option",
  "playwright:wait_for_selector",
  "playwright:if_else",
  "playwright:log",
  "playwright:wait_for_timeout",
  "playwright:wait_for_load_state",
  "playwright:screenshot",
  "playwright:evaluate",
  "playwright:hover",
  "playwright:scroll",
  "playwright:get_text",
  "playwright:get_attribute",
  "playwright:set_viewport",
  "playwright:reload",
  "playwright:go_back",
  "playwright:go_forward",
  "playwright:loop_until",
  "r2:upload",
  "r2:delete",
];

// List of action types that can be used in nested contexts (excluding if_else to prevent infinite nesting)
export const nestedActionTypes = actionTypes.filter(type => type !== "playwright:loop_until");

// Map action types to their respective config components
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
  "playwright:scroll": PlaywrightScrollConfig,
  "playwright:get_text": PlaywrightGetTextConfig,
  "playwright:get_attribute": PlaywrightGetAttributeConfig,
  "playwright:set_viewport": PlaywrightSetViewportConfig,
  "playwright:reload": PlaywrightReloadConfig,
  "playwright:go_back": PlaywrightGoBackConfig,
  "playwright:go_forward": PlaywrightGoForwardConfig,
  "playwright:screenshot": PlaywrightScreenshotConfig,
  "playwright:wait_for_selector": PlaywrightWaitConfig,
  "playwright:wait_for_timeout": PlaywrightWaitConfig,
  "playwright:wait_for_load_state": PlaywrightWaitConfig,
  "playwright:evaluate": PlaywrightEvaluateConfig,
  "r2:upload": R2UploadConfig,
  "r2:delete": R2DeleteConfig,
  "playwright:if_else": PlaywrightIfElseConfig,
  "playwright:log": PlaywrightLogConfig,
  "playwright:loop_until": PlaywrightLoopUntilConfig,
};

// Validation function for action configurations
export function validateActionConfig(
  actionType: string,
  config: Record<string, any>
): string[] {
  const errors: string[] = [];

  switch (actionType) {
    case "playwright:goto":
      if (!config.url) errors.push("URL is required");
      break;
    case "playwright:click":
    case "playwright:fill":
    case "playwright:type":
    case "playwright:press":
    case "playwright:check":
    case "playwright:uncheck":
    case "playwright:select_option":
    case "playwright:hover":
    case "playwright:get_text":
    case "playwright:get_attribute":
    case "playwright:wait_for_selector":
      if (!config.selector) errors.push("Selector is required");
      if (actionType === "playwright:fill" && !config.value)
        errors.push("Value is required");
      if (actionType === "playwright:type" && !config.text)
        errors.push("Text is required");
      if (actionType === "playwright:press" && !config.key)
        errors.push("Key is required");
      if (actionType === "playwright:get_attribute" && !config.attribute)
        errors.push("Attribute name is required");
      if (actionType === "playwright:select_option") {
        if (!config.value && !config.values && !config.label && config.index === undefined) {
          errors.push("Value, label, or index is required");
        }
      }
      break;
    case "playwright:wait_for_timeout":
      if (!config.timeout_ms || config.timeout_ms <= 0)
        errors.push("Timeout (ms) is required and must be positive");
      break;
    case "playwright:set_viewport":
      if (!config.width || config.width <= 0)
        errors.push("Width is required and must be positive");
      if (!config.height || config.height <= 0)
        errors.push("Height is required and must be positive");
      break;
    case "playwright:screenshot":
      if (config.upload_to_r2 && !config.r2_key)
        errors.push("R2 key is required when uploading to R2");
      break;
    case "playwright:evaluate":
      if (!config.expression)
        errors.push("JavaScript expression is required");
      break;
    case "r2:upload":
      if (!config.key) errors.push("Object key is required");
      if (!config.content) errors.push("Content is required");
      break;
    case "r2:delete":
      if (!config.key) errors.push("Object key is required");
      break;
    case "playwright:if_else":
      if (!config.condition_type) errors.push("Condition type is required");
      
      // Selector is only required for non-loop-index and non-random conditions
      const requiresSelector = !config.condition_type?.startsWith("loop_index_is_") && 
                              config.condition_type !== "random";
      if (requiresSelector && !config.selector) {
        errors.push("Selector is required for this condition type");
      }
      
      // Validate nested actions have action_type
      if (config.if_actions) {
        for (const action of config.if_actions) {
          if (!action.action_type) {
            errors.push("All IF actions must have an action type");
            break;
          }
        }
      }
      if (config.else_if_conditions) {
        for (const condition of config.else_if_conditions) {
          const elseIfRequiresSelector = !condition.condition_type?.startsWith("loop_index_is_") && 
                                        condition.condition_type !== "random";
          if (elseIfRequiresSelector && !condition.selector) {
            errors.push("Selector is required for ELSE IF conditions of this type");
            break;
          }
          if (!condition.condition_type) {
            errors.push("All ELSE IF conditions must have a condition type");
            break;
          }
          if (condition.actions) {
            for (const action of condition.actions) {
              if (!action.action_type) {
                errors.push("All ELSE IF actions must have an action type");
                break;
              }
            }
          }
        }
      }
      if (config.else_actions) {
        for (const action of config.else_actions) {
          if (!action.action_type) {
            errors.push("All ELSE actions must have an action type");
            break;
          }
        }
      }
      if (config.final_actions) {
        for (const action of config.final_actions) {
          if (!action.action_type) {
            errors.push("All FINAL actions must have an action type");
            break;
          }
        }
      }
      break;
    case "playwright:log":
      if (!config.message) errors.push("Log message is required");
      break;
    case "playwright:loop_until":
      // At least one force stop mechanism is required
      if (!config.max_loops && !config.timeout_ms) {
        errors.push("Either max loops or timeout must be specified to prevent infinite loops");
      }
      if (config.max_loops && config.max_loops <= 0) {
        errors.push("Max loops must be a positive number");
      }
      if (config.timeout_ms && config.timeout_ms <= 0) {
        errors.push("Timeout must be a positive number");
      }
      // If selector is provided, condition_type is required
      if (config.selector && !config.condition_type) {
        errors.push("Condition type is required when selector is provided");
      }
      // Validate nested actions have action_type
      if (config.loop_actions) {
        for (const action of config.loop_actions) {
          if (!action.action_type) {
            errors.push("All loop actions must have an action type");
            break;
          }
        }
      }
      break;
  }

  return errors;
}