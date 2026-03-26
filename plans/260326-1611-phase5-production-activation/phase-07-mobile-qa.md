# Phase 7: Mobile Responsive QA

**Priority:** P1 | **Effort:** 2h | **Status:** pending

Verify dashboard and auth flows work on mobile devices (iOS, Android, tablets).

## Test Devices

- iPhone 12/14/15 (Safari)
- Samsung Galaxy (Chrome)
- iPad (Safari)
- Desktop: Chrome, Firefox, Safari (for regression)

## Test Scenarios (per device)

### Login & Auth (20min)

- [ ] Open https://valt.turbo.ai.vn on mobile
- [ ] Click "Sign in"
- [ ] Enter email/password
- [ ] Page not cut off, buttons tappable
- [ ] Redirect to dashboard works
- [ ] Click "Sign in with Google"
- [ ] OAuth popup/redirect works on mobile
- [ ] Verify email notification arrives on same device

### Dashboard Navigation (20min)

- [ ] Dashboard loads without layout breaks
- [ ] Sidebar collapses on mobile (hamburger menu works)
- [ ] All main pages accessible:
  - Secrets list (cards responsive)
  - Create secret form (inputs visible)
  - Access requests (list readable)
  - Settings pages (forms not cramped)
- [ ] All buttons/links tappable (min 44x44px)

### Create Secret Form (20min)

- [ ] Form loads fully visible (no horizontal scroll)
- [ ] All fields accessible without pinch-zoom
- [ ] Policy selector works on mobile
- [ ] Submit button visible and tappable
- [ ] Success confirmation visible

### Approvals & Notifications (20min)

- [ ] Notification (email) arrives and opens link
- [ ] Approval page loads on mobile
- [ ] Approve/Reject buttons visible and work
- [ ] Confirmation message shows correctly

### Mobile-Specific Issues to Check

- [ ] No horizontal scroll anywhere
- [ ] Touch targets at least 44x44px
- [ ] No console errors in mobile dev tools
- [ ] Images scale correctly (not stretched/blurry)
- [ ] Forms don't have tiny text (min 16px)
- [ ] Modals/popovers not cut off at edges

## Device Orientation

- [ ] Test both portrait and landscape
- [ ] Layout reflows correctly when rotating

## Performance

- [ ] Page loads within 3s on 4G
- [ ] No janky animations or 60fps drops

## Accessibility (Bonus)

- [ ] All buttons have visible focus state
- [ ] Forms have labels (not just placeholders)
- [ ] Color contrast ratio >=4.5:1 for text

## Issues Found

Log any issues:
- Type (layout, button, form, etc.)
- Device + OS + browser
- Steps to reproduce
- Screenshot/video if possible
- Severity (blocker, high, medium, low)

## Sign-Off

- All critical issues fixed
- At least one test per device type completed
- No horizontal scrolling anywhere
