# Test Plan: Milestone 6 - Motion Service with LinearConstraint

## Feature Summary
Replace switch-based pour-prep movement with motion service using LinearConstraint to maintain kettle level during transit.

## Architecture Decisions
- Keep resting position via switch (simple, reliable)
- Pour-prep movement via motion service with LinearConstraint
- Store only XYZ target point in config (orientation maintained by constraint)
- Motion service is required dependency when pour_prep_target is configured

## Test Plan (REVISED - minimal, no mocks)

| Test Name | Category | Custom Logic Tested |
|-----------|----------|---------------------|
| TestConfigValidate_MotionServiceRequired | Config validation | Validate() rejects config when motion_service missing but pour_prep_target set |
| TestConfigValidate_PourPrepTargetRequired | Config validation | Validate() rejects config when pour_prep_target missing but motion_service set |
| TestConfigValidate_PourPrepTargetCoordinates | Config validation | Validate() rejects pour_prep_target with zero coordinates (likely misconfigured) |
| TestConfigValidate_MotionServiceDependency | Config validation | Validate() returns motion_service in dependency list when configured |

## What We're NOT Testing (trust SDK, verify on hardware)
- That `motion.Move()` receives correct arguments (delegation testing)
- That LinearConstraint works as expected (SDK behavior)
- That the arm actually moves to the right place (physical validation)
- DoCommand routing to handlers (tests dispatch, not logic)

## Rationale for Minimal Tests
Creating mock motion services to verify argument passing is:
1. Delegation testing (tests that we call SDK correctly, not our logic)
2. Overengineering (elaborate mocks for simple pass-through)
3. Brittle (breaks if SDK changes signatures)

Physical validation on hardware is the right way to verify motion service integration works correctly.
