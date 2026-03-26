# Phase 12: Beta Launch (5-10 Users)

**Priority:** P1 | **Effort:** 1h | **Status:** pending

**Blocked by:** All phases 1-11

Soft launch to 5-10 early adopters for feedback and validation before general release.

## Pre-Launch Checklist

- [ ] Phase 6 complete (full user journey works)
- [ ] Phase 7 complete (mobile responsive)
- [ ] Phase 8 complete (error messages clear)
- [ ] Phase 9 complete (monitoring active)
- [ ] Phase 10 complete (API docs available)
- [ ] Phase 11 complete (ops guide ready)
- [ ] All critical bugs fixed
- [ ] Stripe in test mode (or live with test cards disabled)
- [ ] SMTP verified (emails deliver)
- [ ] Server logs being monitored

## Launch Steps

### Day 1: Soft Launch

1. **Identify beta users** (5-10):
   - Friends, colleagues, or early community members
   - Mix of technical and non-technical
   - Timezone spread if possible

2. **Send invite email:**

```
Subject: You're Invited to Valt Beta!

Hi [Name],

We're excited to invite you to beta-test Valt, our new AI-friendly secret vault.

Here's what to expect:
- Sign up at https://valt.turbo.ai.vn
- Create a project and add secrets
- Try the valt CLI to retrieve secrets
- Provide feedback on the experience

We'll be actively monitoring for bugs and improving based on your feedback.

Questions? Reply to this email or join our Slack: [link]

Thanks for helping shape Valt!
```

3. **Monitor during first day:**
   - Watch server logs for errors
   - Check UptimeRobot for uptime
   - Set up Slack notification for errors

4. **Prepare feedback form:**
   - What worked well?
   - What was confusing?
   - Features you'd like?
   - Any bugs encountered?

### Days 2-7: Active Monitoring

- [ ] Daily check-in with beta users
- [ ] Log all reported issues with severity
- [ ] Fix critical bugs (data loss, auth failures) immediately
- [ ] Defer nice-to-have features to Phase 6
- [ ] Monitor server performance

### Day 8: Retrospective

Gather feedback:
- [ ] 1:1 or group call with beta users
- [ ] Collect written feedback (form or email)
- [ ] Log feature requests for roadmap
- [ ] Note pain points for Phase 6

## Success Criteria

- All 5-10 users successfully onboarded
- No critical bugs (data loss, auth failures)
- Uptime >99%
- Email notifications working
- Stripe checkout works (test mode)
- Users can use CLI and dashboard
- At least 3 pieces of actionable feedback

## Post-Beta Decision

**Go:** Proceed to general public launch
- Fix any critical bugs found
- Incorporate quick wins from feedback
- Plan Phase 6 (scale, compliance)

**No-Go:** Extend beta
- Fix critical issues
- Add missing features
- Re-test with new cohort

## Communication

- Public status: "Beta" badge in header
- Acknowledge known limitations (VSCode ext, etc.)
- Set expectations: "This is early, expect occasional bugs"
- Privacy policy: Clarify data handling during beta

## Notes

- Keep beta small (5-10) to manage support burden
- Encourage bug reports, not just feature requests
- Thank beta users publicly (blog post, credits page)
- Plan Phase 6: scale to 100+ users
