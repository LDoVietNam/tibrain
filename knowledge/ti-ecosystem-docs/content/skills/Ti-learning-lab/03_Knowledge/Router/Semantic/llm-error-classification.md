# LLM Error Classification System

**Pattern ID**: `llm-error-classification`
**Source**: Skyvern (skyvern/forge/failure_classifier.py)
**Category**: Error Handling, AI Agent
**Complexity**: Intermediate

## Problem

AI agents encounter various types of failures (LLM errors, browser errors, navigation failures, etc.). Without systematic classification, it's difficult to:
- Determine appropriate recovery strategies
- Track failure patterns for analytics
- Provide meaningful error messages to users
- Distinguish between LLM failures and runtime crashes

## Solution

Implement a keyword-based error classification system with confidence scoring:

### Core Components

1. **Error Categories**
   ```python
   ERROR_CATEGORIES = {
       "ANTI_BOT_DETECTION": 0.7,
       "PROXY_ERROR": 0.9,
       "BROWSER_ERROR": 0.9,
       "NAVIGATION_FAILURE": 0.9,
       "PAGE_LOAD_TIMEOUT": 0.8,
       "AUTH_FAILURE": 0.7,
       "LLM_ERROR": 0.9,
       "CREDENTIAL_ERROR": 0.8,
       "DATA_EXTRACTION_FAILURE": 0.7,
       "ELEMENT_NOT_FOUND": 0.8,
       "WRONG_PAGE_STATE": 0.6,
       "MAX_STEPS_EXCEEDED": 0.9,
       "LLM_REASONING_ERROR": 0.6,
       "INFRASTRUCTURE_ERROR": 0.8,
       "PARAMETER_BINDING_ERROR": 0.95,
       "UNKNOWN": 0.5
   }
   ```

2. **Classification Logic**
   ```python
   def classify_from_failure_reason(
       failure_reason: str | None,
       exception: Exception | None = None,
       fallback_to_unknown: bool = False
   ) -> list[dict] | None:
       """Classify failure from failure_reason text and/or exception type.

       Returns list of categories sorted by confidence, or None if no classification.
       """
       if not failure_reason and not exception:
           return None

       reason = (failure_reason or "").lower()
       exc_name = type(exception).__name__ if exception else ""

       categories: list[dict] = []

       # Bot detection / CAPTCHA
       _auth_context_keywords = ["login", "auth", "password", "permission", "credential"]
       _has_auth_context = any(kw in reason for kw in _auth_context_keywords)
       _antibot_keywords = [
           "captcha", "cloudflare", "bot detect", "bot block",
           "ip block", "request block", "anti-bot", "human verification"
       ]

       if not _has_auth_context:
           _antibot_keywords.append("access denied")

       if any(kw in reason for kw in _antibot_keywords):
           categories.append({
               "category": "ANTI_BOT_DETECTION",
               "confidence_float": 0.7,
               "reasoning": "Keywords matched in failure reason"
           })

       # Proxy errors (check before browser errors)
       _proxy_exc_keywords = ["NoProxy", "ProxyError"]
       _proxy_reason_keywords = ["no proxy available", "proxy unavailable"]

       if any(kw in exc_name for kw in _proxy_exc_keywords) or \
          any(kw in reason for kw in _proxy_reason_keywords):
           categories.append({
               "category": "PROXY_ERROR",
               "confidence_float": 0.9,
               "reasoning": f"Exception: {exc_name}" if exc_name else "Keywords matched"
           })

       # Browser errors (only if not proxy error)
       elif any(kw in exc_name for kw in ["Browser", "CDP", "TargetClosed"]) or \
            any(kw in reason for kw in ["browser context closed", "page closed", "browser crash"]):
           categories.append({
               "category": "BROWSER_ERROR",
               "confidence_float": 0.9,
               "reasoning": f"Exception: {exc_name}" if exc_name else "Keywords matched"
           })

       # Navigation failure
       if "FailedToNavigateToUrl" in exc_name or \
          any(kw in reason for kw in ["failed to navigate", "404", "redirect loop"]):
           categories.append({
               "category": "NAVIGATION_FAILURE",
               "confidence_float": 0.9,
               "reasoning": f"Exception: {exc_name}" if "FailedToNavigate" in exc_name else "Keywords matched"
           })

       # Page load timeout
       if "Timeout" in exc_name or "timeout" in reason:
           categories.append({
               "category": "PAGE_LOAD_TIMEOUT",
               "confidence_float": 0.8,
               "reasoning": f"Exception: {exc_name}" if "Timeout" in exc_name else "Timeout in failure reason"
           })

       # Auth failure
       if any(kw in reason for kw in ["login fail", "authentication fail", "auth fail", "mfa", "password"]) or \
          ("access denied" in reason and _has_auth_context):
           categories.append({
               "category": "AUTH_FAILURE",
               "confidence_float": 0.7,
               "reasoning": "Keywords matched"
           })

       # LLM error
       if any(kw in exc_name for kw in ["LLM", "APIError", "RateLimit"]) or "rate limit" in reason:
           categories.append({
               "category": "LLM_ERROR",
               "confidence_float": 0.9,
               "reasoning": f"Exception: {exc_name}" if exc_name else "Keywords matched"
           })

       # ... more categories ...

       if not categories:
           if fallback_to_unknown:
               return [{"category": "UNKNOWN", "confidence_float": 0.5, "reasoning": "No keyword match found"}]
           return None

       # Sort by confidence descending
       categories.sort(key=lambda x: x["confidence_float"], reverse=True)
       return categories
   ```

3. **Recovery Strategy Mapping**
   ```python
   RECOVERY_STRATEGIES = {
       "ANTI_BOT_DETECTION": "retry_with_different_proxy",
       "PROXY_ERROR": "retry_with_different_proxy",
       "BROWSER_ERROR": "restart_browser_session",
       "NAVIGATION_FAILURE": "verify_url_and_retry",
       "PAGE_LOAD_TIMEOUT": "increase_timeout_and_retry",
       "AUTH_FAILURE": "request_new_credentials",
       "LLM_ERROR": "switch_to_fallback_llm",
       "CREDENTIAL_ERROR": "request_credential_linking",
       "DATA_EXTRACTION_FAILURE": "retry_with_alternative_selector",
       "ELEMENT_NOT_FOUND": "scroll_and_retry",
       "WRONG_PAGE_STATE": "navigate_to_correct_page",
       "MAX_STEPS_EXCEEDED": "terminate_with_partial_results",
       "LLM_REASONING_ERROR": "provide_context_correction",
       "INFRASTRUCTURE_ERROR": "escalate_to_ops",
       "PARAMETER_BINDING_ERROR": "fix_workflow_parameters",
       "UNKNOWN": "generic_retry"
   }

   def get_recovery_strategy(category: str) -> str:
       """Get recovery strategy for error category."""
       return RECOVERY_STRATEGIES.get(category, "generic_retry")
   ```

### Implementation Details

**Context-Aware Classification**:
- Use auth context keywords to distinguish between bot detection and auth failure
- Check proxy errors before browser errors to avoid misclassification
- Use exception names for precise classification when available
- Use failure reason text for LLM-generated error descriptions

**Confidence Scoring**:
- High confidence (0.9+): Exception name match or specific keywords
- Medium confidence (0.7-0.8): Keyword matches with some ambiguity
- Low confidence (0.5-0.6): Generic keywords or uncertain patterns

**Multi-Category Support**:
- Return list of categories for ambiguous failures
- Sort by confidence to prioritize recovery strategies
- Allow fallback strategies if primary fails

### Usage Example

```python
try:
    result = await agent.run_task(task)
except Exception as e:
    failure_reason = str(e)
    classifications = classify_from_failure_reason(
        failure_reason=failure_reason,
        exception=e,
        fallback_to_unknown=True
    )

    if classifications:
        primary_category = classifications[0]["category"]
        confidence = classifications[0]["confidence_float"]
        strategy = get_recovery_strategy(primary_category)

        logger.info(
            f"Classified failure as {primary_category} "
            f"(confidence: {confidence}), "
            f"recovery strategy: {strategy}"
        )

        # Execute recovery strategy
        await execute_recovery(strategy, context)
    else:
        logger.error(f"Could not classify failure: {failure_reason}")
        raise
```

### Benefits

1. **Structured Error Handling**: Systematic approach to different failure types
2. **Recovery Automation**: Enables automatic recovery strategy selection
3. **Analytics**: Track failure patterns for monitoring and improvement
4. **User Communication**: Provide meaningful error messages
5. **Debugging**: Easier to identify root causes of failures

### Trade-offs

1. **False Positives**: Keyword matching may misclassify some errors
2. **Maintenance**: Need to update keywords as error patterns evolve
3. **Context Dependency**: Classification accuracy depends on error message quality
4. **Complexity**: Adds overhead to error handling paths

### Variations

**browser-use Approach**:
- Uses fallback LLM for rate limit errors
- Less systematic classification
- Focuses on LLM-specific errors (429, 500, 502, 503, 504)
- Simple retry logic without classification

**Skyvern Approach**:
- Comprehensive 16-category system
- Confidence scoring for ambiguous cases
- Context-aware classification (auth context, proxy vs browser)
- Recovery strategy mapping

### When to Use

- Building production AI agents with complex error scenarios
- Need for automated recovery strategies
- Want to track failure patterns for analytics
- Providing meaningful error messages to users
- Distinguishing between LLM and runtime failures

### When Not to Use

- Simple agents with limited error scenarios
- When all errors can be handled the same way
- When error messages are unreliable
- Quick prototypes without production requirements

### Related Patterns

- `fallback-llm-system`
- `speculative-execution`
- `multi-step-reasoning`

## Implementation Checklist

- [ ] Define error categories and confidence thresholds
- [ ] Implement keyword-based classification logic
- [ ] Add context-aware classification (auth context, etc.)
- [ ] Implement multi-category support with confidence sorting
- [ ] Create recovery strategy mapping
- [ ] Add logging for classification events
- [ ] Test with various error scenarios
- [ ] Monitor classification accuracy in production
- [ ] Update keywords based on observed patterns
- [ ] Document recovery strategies for each category

## References

- Skyvern: `skyvern/forge/failure_classifier.py`
- browser-use: `browser_use/agent/service.py` (fallback LLM logic)
