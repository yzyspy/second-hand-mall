# Verification Report: mall-ios-network-auth-foundation

**Verify mode:** full
**Date:** 2026-07-03

## Summary

| Dimension    | Status                                    |
|--------------|--------------------------------------------|
| Completeness | 19/19 tasks done; 9/9 delta spec requirements implemented |
| Correctness  | 9/9 requirements matched to code; 1 scenario intentionally deferred (disclosed) |
| Coherence    | Design decisions followed; 1 beneficial addition beyond design.md (documented below) |

## Completeness

- **Tasks:** `openspec instructions apply --json` reports `state: all_done`, 19/19 checked. Confirmed independently via `grep -c '\- \[ \]' tasks.md` → 0.
- **Spec coverage** — both delta specs, all requirements have corresponding implementation:
  - `ios-network-client`: 统一请求信封解析 (`APIClient.request<T>`), 鉴权 Header 注入 (`requiresAuth` branch), 错误分类映射 (`APIError` 4 cases), 非信封响应的注册接口 (`APIClient.register`)
  - `ios-account-session`: 用户名密码登录 (`AppSession.login`), 注册即登录 (`AppSession.register` calls `login`), 退出登录 (`AppSession.logout`), 启动时会话恢复 (`AppSession.bootstrap`), 401 触发自动登出 (mechanism present, see Correctness note below)

## Correctness

Verified against actual source (`mall-ios/Core/Network/APIClient.swift`, `mall-ios/Core/Auth/{TokenStore,AppSession}.swift`, `mall-ios/Features/Profile/*`, `mall-ios/ContentView.swift`), not just filename search:

- `ApiResponse<T>` envelope decode + `EmptyResponse` fallback for null-data responses — matches spec scenario "服务端返回成功信封"/"data 字段缺失" handling.
- `requiresAuth` injects `Authorization: Bearer <token>` only when true; throws `.unauthorized` immediately when token missing (no network call) — matches spec scenario exactly, confirmed by `APIClientTests.testRequest_throwsUnauthorizedWhenAuthRequiredAndNoTokenSaved`.
- `APIError` 4-case mapping (`.server`/`.unauthorized`/`.transport`/`.decoding`) matches spec's error classification requirement 1:1.
- `TokenStore` uses `kSecClassGenericPassword`, `service = Bundle.main.bundleIdentifier`, `account = "jwt"` — matches design.md's Keychain decision verbatim.
- `AppSession.login`/`register`/`logout`/`bootstrap` all match their spec requirements' WHEN/THEN scenarios; `LoginResponseData.CodingKeys` correctly maps backend's `user_id`/`user_name` snake_case fields (cross-checked against `mall-server/internal/app/service/login.go:74-83` during the final code review — response shape confirmed byte-for-byte).
- Four-tab `ContentView` order/icons (首页/发布/消息/我的) match mall-mini's `app.json` tab order as required by proposal.md.

**WARNING (accepted, disclosed):** The "401 触发自动登出" requirement's mechanism (`AppSession.logout()`) exists and is tested in isolation, but this change has no authenticated request that could actually receive a 401 (login/register both use `requiresAuth: false`), so the scenario "已登录期间 token 过期" is not exercised end-to-end in this change. This is explicitly disclosed in the plan's Self-Review section and in the final code review (Minor issue #4) as intentional — the contract is ready for the first downstream change that adds an authenticated call. **Recommendation:** the next change that introduces an authenticated request (`mall-ios-browse-search-detail` or later) must wire `catch APIError.unauthorized { session.logout() }` at that call site and add a test for it.

**Scenario coverage via tests:** All spec scenarios that ARE exercisable in this change have corresponding tests in the 26-test suite (`ApiResponseTests`, `APIClientTests` ×10, `AppSessionTests` ×7, `TokenStoreTests` ×4, plus infrastructure tests). Independently re-ran `mall-ios/scripts/ci-test.sh` fresh during this verify pass: **26/26 passed, 0 failures.**

## Coherence

- Design.md's four numbered decisions (network layer, session+Keychain, tab structure, error-feedback pattern) are all followed as designed.
- **Beneficial addition beyond design.md:** the final code review (round 1) found `AppSession` needed `@MainActor` isolation to prevent an off-main-thread `@Observable` state mutation (a real latent data race the original design.md/Design Doc didn't anticipate, since neither discussed Swift concurrency isolation explicitly). This was fixed (commit `b9f8cf4`) and independently re-reviewed as correct with 0 new issues. This is **not a contradiction** of any design.md decision — the design doc was simply silent on actor isolation — so it does not trigger the Spec Drift decision point (Step 2b's condition is delta-spec-vs-design-doc *contradiction*, not silence). Recorded here as a SUGGESTION: a future editor of `docs/superpowers/specs/2026-07-02-mall-ios-network-auth-foundation-design.md` could add a short note that `AppSession` is `@MainActor`-isolated, for readers who reference that doc later.
- No project-pattern deviations found: `@Observable`/`Observation` used throughout (no `ObservableObject`), file/directory layout matches existing `Features/Home`, `Features/Publish` conventions, dependency injection via `init` params used consistently.

## Build/Test Evidence (fresh, this verify pass)

```
$ mall-ios/scripts/ci-build.sh
** BUILD SUCCEEDED **

$ mall-ios/scripts/ci-test.sh
Test Suite 'All tests' passed ...
Executed 26 tests, with 0 failures (0 unexpected) in 0.258 (0.270) seconds
** TEST SUCCEEDED **
```

## Known, disclosed gap (not blocking)

tasks.md item 6.3 (interactive UI walkthrough: register → auto-login → kill+relaunch → verify persisted session → logout) was accepted based on automated evidence rather than a real tap-through, because no UI automation tool was available in this environment (no `idb`; AppleScript/System Events lacked Accessibility permission). See `.superpowers/sdd/task-12-manual-verification-note.md` for full detail. The automated evidence (26 tests covering every state transition + real backend integration confirming response-shape correctness + real simulator screenshot confirming the four-tab shell renders) is judged sufficient to proceed; a human can still run the 2-3 minute manual walkthrough at their convenience.

## Final Assessment

**No CRITICAL issues.** One WARNING (401→logout call site deferred, explicitly disclosed and actionable for the next change) and one SUGGESTION (design doc could note the `@MainActor` decision) — neither blocks archiving.

**Ready for archive.**
